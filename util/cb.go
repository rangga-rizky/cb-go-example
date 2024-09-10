package util

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

//https://www.npmjs.com/package/@fastify/circuit-breaker
// 1. threshold
// 2. counter
// 3. status

type CBImpl struct {
	Redis redis.Client
}

type CB interface {
	Register(ctx context.Context, name string, threshold int)
	IsOpen(ctx context.Context, name string) bool
	Count(ctx context.Context, name string)
	GetCounter(ctx context.Context, name string) int
}

func NewCB(rdb redis.Client) CB {
	return &CBImpl{
		Redis: rdb,
	}
}

func (cb *CBImpl) Register(ctx context.Context, name string, threshold int) {
	err := cb.Redis.Set(ctx, "cb_threshold_"+name, threshold, 0).Err()
	if err != nil {
		panic(err)
	}

	err = cb.Redis.Set(ctx, "cb_counter_"+name, 0, 0).Err()
	if err != nil {
		panic(err)
	}
}

func (cb *CBImpl) IsOpen(ctx context.Context, name string) bool {

	_, err := cb.Redis.Get(ctx, "cb_tripped_"+name).Result()
	if err != redis.Nil {
		return true
	} else if err != nil && err != redis.Nil {
		panic(err.Error())
	}

	// check counter dan thresshold
	val, err := cb.Redis.Get(ctx, "cb_counter_"+name).Result()
	if err != nil {
		panic(err)
	}
	counter, _ := strconv.Atoi(val)

	val, err = cb.Redis.Get(ctx, "cb_threshold_"+name).Result()
	if err != nil {
		panic(err)
	}
	threshold, _ := strconv.Atoi(val)
	if counter < threshold {
		return false
	}

	err = cb.Redis.Set(ctx, "cb_tripped_"+name, true, 10*time.Minute).Err()
	if err != nil {
		panic(err)
	}

	err = cb.Redis.Set(ctx, "cb_counter_"+name, 0, 0).Err()
	if err != nil {
		panic(err)
	}

	return false
}

func (cb *CBImpl) Count(ctx context.Context, name string) {
	val, err := cb.Redis.Get(ctx, "cb_counter_"+name).Result()
	if err != nil {
		panic(err)
	}
	counter, _ := strconv.Atoi(val)
	counter += 1
	err = cb.Redis.Set(ctx, "cb_counter_"+name, counter, 0).Err()
	if err != nil {
		panic(err)
	}
}

func (cb *CBImpl) GetCounter(ctx context.Context, name string) int {
	val, err := cb.Redis.Get(ctx, "cb_counter_"+name).Result()
	if err != nil {
		panic(err)
	}
	counter, _ := strconv.Atoi(val)
	return counter
}
