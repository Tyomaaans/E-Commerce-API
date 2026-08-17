package domain

import (
	"context"
	"time"
)

type UserCache struct {
    ID                 string     `json:"id"`
    Email              string     `json:"email"`
    Role               Role       `json:"role"`
    IsEmailVerified    bool       `json:"is_email_verified"`
    EmailVerifiedAt    *time.Time `json:"email_verified_at"`
    IsProfileCompleted bool       `json:"is_profile_completed"`
    ProfileCompletedAt *time.Time `json:"profile_completed_at"`
    Providers          []string   `json:"providers"`
}

type UserCacheRepository interface {
    SetUserCache(ctx context.Context, cache *UserCache) error
    GetUserCache(ctx context.Context, userID string) (*UserCache, error)
    DeleteUserCache(ctx context.Context, userID string) error
}