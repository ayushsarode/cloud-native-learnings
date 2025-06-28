package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"microservice-order-api/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client	*redis.Client
}

func orderIDkey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

func(r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)

	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}
	

	key := orderIDkey(order.OrderID)

	res := r.Client.SetNX(ctx, key, string(data), 0)

	if err := res.Err(); err != nil {
		return fmt.Errorf("failed to set:  %w", err)
	}
	return nil
}

var ErrNotExist = errors.New("order does not exist")

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIDkey(id)

	value ,err := r.Client.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("get order: %w", err)
	}

	var order model.Order

	err = json.Unmarshal([]byte(value), &order)

	if err != nil {
		return model.Order{}, fmt.Errorf("failed to decode order json: %w", err)
	}

	return order, nil
}


func (r *RedisRepo) DeleteByID (ctx context.Context, id uint64) error {
	key := orderIDkey(id)

	err := r.Client.Del(ctx, key).Err()

	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("get order: %w", err)
	}

	return nil
}

func(r *RedisRepo) Update (ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)


	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}

	key := orderIDkey(order.OrderID)

	err = r.Client.SetXX(ctx ,key, string(data), 0).Err()

	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("set order: %w", err)
	}

	return nil
}
