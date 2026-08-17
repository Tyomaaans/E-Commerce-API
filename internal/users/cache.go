package users

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"E-COMMERCE-API/internal/domain"
)

const (
	UserCachePrefix = "user:cache:"
	UserCacheTTL    = 15 * time.Minute
)
		
type userCacheRepository struct {
	rdb *redis.Client
}

func NewUserCacheRepository(rdb *redis.Client) domain.UserCacheRepository {
	return &userCacheRepository{rdb: rdb}
}

func (r *userCacheRepository) SetUserCache(ctx context.Context, cache *domain.UserCache) error {
	key := fmt.Sprintf("%s%s", UserCachePrefix, cache.ID)
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, data, UserCacheTTL).Err()
}

func (r *userCacheRepository) GetUserCache(ctx context.Context, userID string) (*domain.UserCache, error) {
	key := fmt.Sprintf("%s%s", UserCachePrefix, userID)
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var cache domain.UserCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

func (r *userCacheRepository) DeleteUserCache(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", UserCachePrefix, userID)
	return r.rdb.Del(ctx, key).Err()
}