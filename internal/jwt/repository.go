package jwt

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var (
	ErrTokenExpired         = errors.New("access token expired")
	ErrTokenRevoked         = errors.New("token has been revoked")
	ErrTokenInvalid         = errors.New("invalid token")
	ErrTokenBlocked         = errors.New("failed to check token status")
	ErrRefreshTokenInvalid  = errors.New("refresh token expired or already used, please login again")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
)

const (
	refreshTokenPrefix = "refresh:"
	sessionPrefix      = "session:"
	refreshTokenLength = 32
)

type JWTService interface {
	GenerateTokenPair(ctx context.Context, userID, role string, rememberMe bool) (*TokenPair, error)
	ValidateAccessToken(ctx context.Context, tokenStr string) (*Claims, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error)
	RevokeTokens(ctx context.Context, accessToken, refreshToken string) error
	RevokeUserSession(ctx context.Context, userID string) error
}

type jwtClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	gojwt.RegisteredClaims
}

type refreshTokenPayload struct {
	UserID    string        `json:"user_id"`
	Role      string        `json:"role"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at"`
	TTL       time.Duration `json:"ttl"`
}

type jwtService struct {
	secretKey             string
	accessTokenExpiry       time.Duration
	defaultRefreshExpiry    time.Duration
	shortRefreshTokenExpiry time.Duration
	redisClient             *redis.Client
}

func NewJWTService(
	secretKey string,
	accessTokenExpiry       time.Duration,
	defaultRefreshExpiry    time.Duration,
	shortRefreshTokenExpiry time.Duration,
	redisClient             *redis.Client,
) JWTService {
	return &jwtService{
		secretKey:               secretKey,
		accessTokenExpiry:       accessTokenExpiry,
		defaultRefreshExpiry:    defaultRefreshExpiry,
		shortRefreshTokenExpiry: shortRefreshTokenExpiry,
		redisClient:             redisClient,
	}
}

// hashToken hashes a raw token string using SHA-256 before storing or querying Redis.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *jwtService) generateAccessToken(userID, role string) (string, error) {
	now := time.Now()
	claims := &jwtClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(now.Add(s.accessTokenExpiry)),
			IssuedAt:  gojwt.NewNumericDate(now),
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

func (s *jwtService) parseAccessToken(tokenStr string) (*jwtClaims, error) {
	claims := &jwtClaims{}
	_, err := gojwt.ParseWithClaims(
		tokenStr,
		claims,
		func(t *gojwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
				return nil, ErrTokenInvalid
			}
			return []byte(s.secretKey), nil
		},
		gojwt.WithoutClaimsValidation(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenInvalid, err)
	}
	return claims, nil
}

func (s *jwtService) createSession(ctx context.Context, userID string, duration time.Duration) error {
	return s.redisClient.Set(ctx, sessionPrefix+userID, "active", duration).Err()
}

func (s *jwtService) deleteSession(ctx context.Context, userID string) error {
	return s.redisClient.Del(ctx, sessionPrefix+userID).Err()
}

func (s *jwtService) isSessionActive(ctx context.Context, userID string) (bool, error) {
	val, err := s.redisClient.Get(ctx, sessionPrefix+userID).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "active", nil
}

func (s *jwtService) generateRefreshToken(ctx context.Context, userID, role string, expiryDuration time.Duration) (string, error) {
	b := make([]byte, refreshTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	rawToken := hex.EncodeToString(b)

	now := time.Now()
	payload := refreshTokenPayload{
		UserID:    userID,
		Role:      role,
		CreatedAt: now,
		ExpiresAt: now.Add(expiryDuration),
		TTL:       expiryDuration,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal refresh token payload: %w", err)
	}

	hashedKey := refreshTokenPrefix + hashToken(rawToken)
	if err := s.redisClient.Set(ctx, hashedKey, data, expiryDuration).Err(); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return rawToken, nil
}

func (s *jwtService) getRefreshTokenPayload(ctx context.Context, rawRefreshToken string) (*refreshTokenPayload, error) {
	hashedKey := refreshTokenPrefix + hashToken(rawRefreshToken)
	data, err := s.redisClient.Get(ctx, hashedKey).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrRefreshTokenInvalid
	}
	if err != nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}

	var payload refreshTokenPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("corrupted refresh token data: %w", err)
	}
	return &payload, nil
}

// GenerateTokenPair creates new access and refresh tokens, updating the active session.
// Dynamic TTL is set based on rememberMe (1h if false, default duration if true).
func (s *jwtService) GenerateTokenPair(ctx context.Context, userID, role string, rememberMe bool) (*TokenPair, error) {
	refreshExpiry := s.defaultRefreshExpiry
	if !rememberMe {
		refreshExpiry = s.shortRefreshTokenExpiry
	}

	accessToken, err := s.generateAccessToken(userID, role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(ctx, userID, role, refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.createSession(ctx, userID, refreshExpiry); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &TokenPair{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateAccessToken checks signature, active session status, and expiration sequentially.
func (s *jwtService) ValidateAccessToken(ctx context.Context, tokenStr string) (*Claims, error) {
	claims, err := s.parseAccessToken(tokenStr)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	active, err := s.isSessionActive(ctx, claims.UserID)
	if err != nil {
		return nil, ErrTokenBlocked
	}
	if !active {
		return nil, ErrTokenRevoked
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, ErrTokenExpired
	}

	return &Claims{
		UserID:    claims.UserID,
		Role:      claims.Role,
		ExpiresAt: claims.ExpiresAt.Time,
		IssuedAt:  claims.IssuedAt.Time,
	}, nil
}

// RefreshTokens replaces the current refresh token with a new pair and preserves TTL settings.
func (s *jwtService) RefreshTokens(ctx context.Context, rawRefreshToken string) (*TokenPair, error) {
	payload, err := s.getRefreshTokenPayload(ctx, rawRefreshToken)
	if err != nil {
		return nil, err
	}

	hashedKey := refreshTokenPrefix + hashToken(rawRefreshToken)
	if err := s.redisClient.Del(ctx, hashedKey).Err(); err != nil {
		return nil, fmt.Errorf("failed to invalidate old refresh token: %w", err)
	}

	isRememberMe := payload.TTL == s.defaultRefreshExpiry
	return s.GenerateTokenPair(ctx, payload.UserID, payload.Role, isRememberMe)
}

// RevokeTokens removes both the user session and refresh token on logout.
func (s *jwtService) RevokeTokens(ctx context.Context, accessToken, rawRefreshToken string) error {
	claims, err := s.parseAccessToken(accessToken)
	if err != nil {
		return ErrTokenInvalid
	}

	if err := s.deleteSession(ctx, claims.UserID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	hashedKey := refreshTokenPrefix + hashToken(rawRefreshToken)
	if err := s.redisClient.Del(ctx, hashedKey).Err(); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (s *jwtService) RevokeUserSession(ctx context.Context, userID string) error {
    if err := s.deleteSession(ctx, userID); err != nil {
        return fmt.Errorf("failed to revoke user session: %w", err)
    }
    return nil
}