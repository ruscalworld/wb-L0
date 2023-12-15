package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"wb-l0/internal/config"
	"wb-l0/internal/order"

	"github.com/redis/go-redis/v9"
)

// RedisRepository - обёртка над клиентом go-redis, отвечающая за сериализацию/десериализацию данных,
// позволяющая сохранять и получать данные из Redis.
//
// Был реализован изначально из-за невнимательности при чтении задания. :)
// Сейчас не используется, оставил просто на память.
type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{
		client: client,
	}
}

func NewRedisRepositoryFromConfig(cfg config.RedisConnection) *RedisRepository {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.Address,
	})

	return NewRedisRepository(client)
}

func (r *RedisRepository) GetOrder(ctx context.Context, uid string) (*order.Order, error) {
	serializedOrder, err := r.client.Get(ctx, uid).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, order.ErrNotFound
		}

		return nil, fmt.Errorf("error fetching cached order from redis: %s", err)
	}

	var o order.Order
	err = json.Unmarshal([]byte(serializedOrder), &o)
	if err != nil {
		return nil, fmt.Errorf("error parsing cached order: %s", err)
	}

	return &o, nil
}

func (r *RedisRepository) CreateOrder(ctx context.Context, o *order.Order) error {
	serializedOrder, err := json.Marshal(o)
	if err != nil {
		return fmt.Errorf("error marshalling order for caching: %s", err)
	}

	err = r.client.Set(ctx, o.OrderUID, serializedOrder, 0).Err()
	if err != nil {
		return fmt.Errorf("error saving order to redis: %s", err)
	}

	return nil
}
