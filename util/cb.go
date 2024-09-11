package util

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type CBImpl struct {
	Redis redis.Client
}

type CBconfig struct {
	Counter   int
	Threshold int
	TimeOut   time.Duration
}

type CB interface {
	Register(ctx context.Context, name string, threshold int, timeout time.Duration)
	IsOpen(ctx context.Context, name string) bool
	Count(ctx context.Context, name string)
	GetCounter(ctx context.Context, name string) int
}

func NewCB(rdb redis.Client) CB {
	return &CBImpl{
		Redis: rdb,
	}
}

func (cb *CBImpl) Register(ctx context.Context, name string, threshold int, timeout time.Duration) {

	_, err := cb.Redis.Get(ctx, "cb_config_"+name).Result()
	if err != redis.Nil {
		return
	} else if err != nil && err != redis.Nil {
		panic(err.Error())
	}

	config := &CBconfig{Counter: 0, Threshold: threshold, TimeOut: timeout}
	serialized, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}

	err = cb.Redis.Set(ctx, "cb_config_"+name, string(serialized), 0).Err()
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
	val, err := cb.Redis.Get(ctx, "cb_config_"+name).Result()
	if err != nil {
		panic(err)
	}

	var cbConfig CBconfig
	json.Unmarshal([]byte(val), &cbConfig)

	if cbConfig.Counter < cbConfig.Threshold {
		return false
	}

	err = cb.Redis.Set(ctx, "cb_tripped_"+name, true, cbConfig.TimeOut).Err()
	if err != nil {
		panic(err)
	}

	cbConfig.Counter = 0

	serialized, err := json.Marshal(cbConfig)
	if err != nil {
		panic(err)
	}

	err = cb.Redis.Set(ctx, "cb_config_"+name, string(serialized), 0).Err()
	if err != nil {
		panic(err)
	}

	return true
}

func (cb *CBImpl) Count(ctx context.Context, name string) {
	val, err := cb.Redis.Get(ctx, "cb_config_"+name).Result()
	if err != nil {
		panic(err)
	}

	var cbConfig CBconfig

	json.Unmarshal([]byte(val), &cbConfig)

	cbConfig.Counter += 1
	serialized, err := json.Marshal(cbConfig)
	if err != nil {
		panic(err)
	}

	err = cb.Redis.Set(ctx, "cb_config_"+name, string(serialized), 0).Err()
	if err != nil {
		panic(err)
	}
}

func (cb *CBImpl) GetCounter(ctx context.Context, name string) int {
	val, err := cb.Redis.Get(ctx, "cb_config_"+name).Result()
	if err != nil {
		panic(err)
	}

	var cbConfig CBconfig

	json.Unmarshal([]byte(val), &cbConfig)
	return cbConfig.Counter
}
