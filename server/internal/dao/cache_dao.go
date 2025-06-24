package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spurge/p4rsec/server/internal/database"
	"github.com/spurge/p4rsec/server/internal/models"
)

type CacheDAO struct {
	redis *database.RedisDB
}

func NewCacheDAO(redis *database.RedisDB) *CacheDAO {
	return &CacheDAO{redis: redis}
}

const (
	UserCachePrefix    = "user:"
	UsersCacheKey      = "users:list"
	DefaultCacheExpiry = 1 * time.Hour
)

// User caching methods
func (d *CacheDAO) SetUser(ctx context.Context, user *models.User) error {
	key := fmt.Sprintf("%s%s", UserCachePrefix, user.ID.String())

	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	return d.redis.Set(ctx, key, data, DefaultCacheExpiry)
}

func (d *CacheDAO) GetUser(ctx context.Context, userID string) (*models.User, error) {
	key := fmt.Sprintf("%s%s", UserCachePrefix, userID)

	data, err := d.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("user not found in cache: %w", err)
	}

	var user models.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

func (d *CacheDAO) DeleteUser(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", UserCachePrefix, userID)
	return d.redis.Delete(ctx, key)
}

func (d *CacheDAO) SetUsers(ctx context.Context, users []*models.User, page, limit int) error {
	key := fmt.Sprintf("%s:%d:%d", UsersCacheKey, page, limit)

	data, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("failed to marshal users: %w", err)
	}

	return d.redis.Set(ctx, key, data, 30*time.Minute) // Shorter expiry for lists
}

func (d *CacheDAO) GetUsers(ctx context.Context, page, limit int) ([]*models.User, error) {
	key := fmt.Sprintf("%s:%d:%d", UsersCacheKey, page, limit)

	data, err := d.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("users not found in cache: %w", err)
	}

	var users []*models.User
	if err := json.Unmarshal([]byte(data), &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal users: %w", err)
	}

	return users, nil
}

func (d *CacheDAO) InvalidateUsersList(ctx context.Context) error {
	// Use pattern matching to delete all users list cache entries
	pattern := fmt.Sprintf("%s:*", UsersCacheKey)

	// Note: In production, you might want to use SCAN instead of KEYS for better performance
	keys, err := d.redis.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	if len(keys) > 0 {
		return d.redis.Delete(ctx, keys...)
	}

	return nil
}

// Generic cache methods
func (d *CacheDAO) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return d.redis.Set(ctx, key, value, expiration)
}

func (d *CacheDAO) Get(ctx context.Context, key string) (string, error) {
	return d.redis.Get(ctx, key)
}

func (d *CacheDAO) Delete(ctx context.Context, keys ...string) error {
	return d.redis.Delete(ctx, keys...)
}

func (d *CacheDAO) Exists(ctx context.Context, keys ...string) (int64, error) {
	return d.redis.Exists(ctx, keys...)
}

// Rate limiting helper
func (d *CacheDAO) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	count, err := d.redis.Incr(ctx, key)
	if err != nil {
		return 0, err
	}

	// Set expiration only on first increment
	if count == 1 {
		if err := d.redis.Expire(ctx, key, window); err != nil {
			return count, err
		}
	}

	return count, nil
}

// Session management
func (d *CacheDAO) SetSession(ctx context.Context, sessionID string, userID string, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return d.redis.Set(ctx, key, userID, expiration)
}

func (d *CacheDAO) GetSession(ctx context.Context, sessionID string) (string, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	return d.redis.Get(ctx, key)
}

func (d *CacheDAO) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return d.redis.Delete(ctx, key)
}
