package users

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"E-COMMERCE-API/pkg"
)

// Prefix key Redis for isolated namespace
const (
	registerTokenPrefix      = "register:token:"
	resetPasswordTokenPrefix = "reset_password:token:"
)

// ==== Interface ====

type UserTokenRepository interface {
	// Register token — TTL 24 hours, value = userID
	SetRegisterToken(ctx context.Context, token string, userID string, ttl time.Duration) error
	GetRegisterToken(ctx context.Context, token string) (userID string, err error)
	DeleteRegisterToken(ctx context.Context, token string) error

	// Reset password token — TTL 15 minutes, value = email
	SetResetPasswordToken(ctx context.Context, token string, email string, ttl time.Duration) error
	GetResetPasswordToken(ctx context.Context, token string) (email string, err error)
	DeleteResetPasswordToken(ctx context.Context, token string) error
}

// ==== Implementation ====

type userTokenRepository struct {
	rdb *redis.Client
}

func NewUserTokenRepository(rdb *redis.Client) UserTokenRepository {
	return &userTokenRepository{rdb: rdb}
}

// ==== Register Token ====

func (r *userTokenRepository) SetRegisterToken(ctx context.Context, token string, userID string, ttl time.Duration) error {
	key := registerTokenKey(token)
	if err := r.rdb.Set(ctx, key, userID, ttl).Err(); err != nil {
		return pkg.ErrInternal
	}
	return nil
}

func (r *userTokenRepository) GetRegisterToken(ctx context.Context, token string) (string, error) {
	key := registerTokenKey(token)
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", pkg.ErrInvalidOrExpiredToken
		}
		return "", pkg.ErrInternal
	}
	return val, nil
}

func (r *userTokenRepository) DeleteRegisterToken(ctx context.Context, token string) error {
	key := registerTokenKey(token)
	return r.rdb.Del(ctx, key).Err()
}

// ==== Reset Password Token ====

func (r *userTokenRepository) SetResetPasswordToken(ctx context.Context, token string, email string, ttl time.Duration) error {
	key := resetPasswordTokenKey(token)
	if err := r.rdb.Set(ctx, key, email, ttl).Err(); err != nil {
		return pkg.ErrInternal
	}
	return nil
}

func (r *userTokenRepository) GetResetPasswordToken(ctx context.Context, token string) (string, error) {
	key := resetPasswordTokenKey(token)
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", pkg.ErrInvalidOrExpiredToken
		}
		return "", pkg.ErrInternal
	}
	return val, nil
}

func (r *userTokenRepository) DeleteResetPasswordToken(ctx context.Context, token string) error {
	key := resetPasswordTokenKey(token)
	return r.rdb.Del(ctx, key).Err()
}

// ==== Key Helpers ====

// hashToken converts raw token into sha256 hex string to prevent leaks in Redis.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func registerTokenKey(token string) string {
	return fmt.Sprintf("%s%s", registerTokenPrefix, hashToken(token))
}

func resetPasswordTokenKey(token string) string {
	return fmt.Sprintf("%s%s", resetPasswordTokenPrefix, hashToken(token))
}