package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/internal/jwt"
	"E-COMMERCE-API/internal/users"
)

// UserFetcher is a narrow interface for retrieving a user record from persistent storage.
// Keeping it separate from the full repository prevents middleware from having access
// to write operations it has no business touching.
type UserFetcher interface {
	GetUserByID(ctx context.Context, id string) (*domain.UserEntity, error)
}

// CacheFetcher abstracts the Redis cache layer for user lookups.
// Both read and write are needed here so the middleware can repopulate
// a cold cache entry on a cache miss.
type CacheFetcher interface {
	GetUserCache(ctx context.Context, userID string) (*domain.UserCache, error)
	SetUserCache(ctx context.Context, cache *domain.UserCache) error
}

// AuthMiddleware holds the dependencies required to authenticate requests
// and enforce access control rules on protected routes.
type AuthMiddleware struct {
	jwtService jwt.JWTService
	userRepo   UserFetcher
	userCache  CacheFetcher
}

// NewAuthMiddleware wires up the auth middleware with its required dependencies.
func NewAuthMiddleware(jwtService jwt.JWTService, userRepo UserFetcher, userCache CacheFetcher) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		userRepo:   userRepo,
		userCache:  userCache,
	}
}

// ==== Middleware Chains ====

// Authenticate validates the Bearer token on every incoming request.
// On success it injects userID, role, and a populated userCache into the Gin context
// so downstream handlers and middleware can read them without hitting the DB again.
//
// Cache strategy: Redis is checked first. On a miss the user is fetched from the DB
// and the cache entry is refreshed transparently — no handler involvement needed.
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, ok := extractBearerToken(c)
		if !ok {
			abortWithError(c, http.StatusUnauthorized, "missing or invalid authorization header", "AUTH_TOKEN_MISSING")
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(c.Request.Context(), accessToken)
		if err != nil {
			switch {
			case errors.Is(err, jwt.ErrTokenExpired):
				abortWithError(c, http.StatusUnauthorized, "access token expired", "AUTH_TOKEN_EXPIRED")
			case errors.Is(err, jwt.ErrTokenRevoked):
				abortWithError(c, http.StatusUnauthorized, "token has been revoked", "AUTH_TOKEN_REVOKED")
			default:
				abortWithError(c, http.StatusUnauthorized, "invalid token", "AUTH_TOKEN_INVALID")
			}
			return
		}

		// Attempt to resolve the user from Redis to avoid a DB round-trip.
		uCache, err := m.userCache.GetUserCache(c.Request.Context(), claims.UserID)
		if err != nil {
			// Cache miss or expiry — fall back to the database and repopulate the cache.
			userEntity, dbErr := m.userRepo.GetUserByID(c.Request.Context(), claims.UserID)
			if dbErr != nil {
				// The token was valid but the user no longer exists — likely deleted after issuance.
				abortWithError(c, http.StatusUnauthorized, "user no longer exists", "AUTH_USER_NOT_FOUND")
				return
			}

			// Rebuild the cache entry so the next request is served from Redis.
			uCache = users.ToUserCache(userEntity)
			_ = m.userCache.SetUserCache(c.Request.Context(), uCache)
		}

		// Expose identity and the cache object to all downstream handlers.
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("userCache", uCache)

		c.Next()
	}
}

// RequireVerifiedEmail blocks requests from users who have not confirmed their email address.
// Must be chained after Authenticate so the userCache is guaranteed to be present in context.
func (m *AuthMiddleware) RequireVerifiedEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("userCache")
		if !exists {
			abortWithError(c, http.StatusUnauthorized, "unauthorized", "AUTH_MISSING_CONTEXT")
			return
		}

		uCache, ok := val.(*domain.UserCache)
		if !ok || !uCache.IsEmailVerified {
			abortWithError(c, http.StatusForbidden, "email not verified", "AUTH_EMAIL_NOT_VERIFIED")
			return
		}

		c.Next()
	}
}

// RequireCompletedProfile blocks requests from domain who skipped the post-registration
// profile setup. Must be chained after Authenticate.
func (m *AuthMiddleware) RequireCompletedProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("userCache")
		if !exists {
			abortWithError(c, http.StatusUnauthorized, "unauthorized", "AUTH_MISSING_CONTEXT")
			return
		}

		uCache, ok := val.(*domain.UserCache)
		if !ok || !uCache.IsProfileCompleted {
			abortWithError(c, http.StatusForbidden, "profile incomplete", "AUTH_PROFILE_INCOMPLETE")
			return
		}

		c.Next()
	}
}

// RequireRole enforces role-based access control on a route.
// Accepts a variadic list of allowed roles so a single route can permit multiple roles
// without registering separate middleware per role. Must be chained after Authenticate.
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			abortWithError(c, http.StatusForbidden, "role not found in token", "AUTHZ_ROLE_MISSING")
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			abortWithError(c, http.StatusForbidden, "invalid role format", "AUTHZ_ROLE_INVALID")
			return
		}

		for _, r := range roles {
			if roleStr == r {
				c.Next()
				return
			}
		}

		// None of the allowed roles matched — deny access.
		abortWithError(c, http.StatusForbidden, "access denied", "AUTHZ_ROLE_FORBIDDEN")
	}
}

// ==== Internal Helpers ====

// extractBearerToken pulls the token string out of the Authorization header.
// Returns false if the header is absent, malformed, or contains an empty token value.
func extractBearerToken(c *gin.Context) (string, bool) {
	header := c.GetHeader("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" {
		return "", false
	}
	return token, true
}

// abortWithError halts the middleware chain and writes a structured JSON error response.
// The machine-readable code field allows clients to handle specific error cases
// without relying on brittle message string matching.
func abortWithError(c *gin.Context, status int, message, code string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": message,
		"code":  code,
	})
}