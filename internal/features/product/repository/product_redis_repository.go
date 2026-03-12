package productrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/codepnw/go-starter-kit/internal/helper"
	"github.com/redis/go-redis/v9"
)

//go:generate mockgen -source=product_redis_repository.go -destination=product_redis_repository_mock.go -package=productrepository
type ProductRedisRepository interface {
	SetStock(ctx context.Context, productID int64, initQty int) error
	IncreaseStock(ctx context.Context, productID int64, qty int) error
	DecreaseStock(ctx context.Context, productID int64, qty int) (int64, error)

	GetJSONCache(ctx context.Context, key string, data any) error
	SetJSONCache(ctx context.Context, key string, data any, exp time.Duration) error
	
	DeleteCache(ctx context.Context, key string) error
}

type productRedisRepository struct {
	client *redis.Client
}

func NewProductRedisRepository(client *redis.Client) ProductRedisRepository {
	return &productRedisRepository{client: client}
}

// SetStock implements OrderRedisRepository.
func (r *productRedisRepository) SetStock(ctx context.Context, productID int64, initQty int) error {
	key := helper.RedisProductStockKey(productID)
	exp := time.Hour * 24

	err := r.client.Set(ctx, key, initQty, exp).Err()
	if err != nil {
		return fmt.Errorf("redis set stock failed: %w", err)
	}
	return nil
}

// IncreaseStock implements OrderRedisRepository.
func (r *productRedisRepository) IncreaseStock(ctx context.Context, productID int64, qty int) error {
	key := helper.RedisProductStockKey(productID)

	err := r.client.IncrBy(ctx, key, int64(qty)).Err()
	if err != nil {
		return fmt.Errorf("redis increase stock failed: %w", err)
	}
	return nil
}

// DecreaseStock implements OrderRedisRepository.
func (r *productRedisRepository) DecreaseStock(ctx context.Context, productID int64, qty int) (int64, error) {
	key := helper.RedisProductStockKey(productID)

	remainStock, err := r.client.DecrBy(ctx, key, int64(qty)).Result()
	if err != nil {
		return 0, fmt.Errorf("redis decrease stock failed: %w", err)
	}
	return remainStock, nil
}

// GetProductsCache implements ProductRedisRepository.
func (r *productRedisRepository) GetJSONCache(ctx context.Context, key string, data any) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return err
	}
	return nil
}

// SetProductsCache implements ProductRedisRepository.
func (r *productRedisRepository) SetJSONCache(ctx context.Context, key string, data any, exp time.Duration) error {
	val, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, val, exp).Err()
}

func (r *productRedisRepository) DeleteCache(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}