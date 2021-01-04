package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/mras-diplomarbeit/mras-api/config"
	. "github.com/mras-diplomarbeit/mras-api/logger"
	"time"
)

type RedisService interface {
	Initialize(conf map[string]interface{}) (*RedisServices, error)
}
type RedisServices struct {
	ctx context.Context
	rdb *redis.Client
}

func RedisDBService() *RedisServices {
	return &RedisServices{}
}

func (service *RedisServices) Initialize(conf map[string]interface{}) (*RedisServices, error) {
	service.ctx = context.Background()
	service.rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis["host"], config.Redis["port"]),
		Password: conf["password"].(string),
		DB:       conf["db"].(int),
	})

	err := service.rdb.Ping(service.ctx).Err()
	if err != nil {
		return service, err
	}

	Log.WithField("module", "redis").Info("Redis connection established!")
	return service, nil
}

func (service *RedisServices) AddPair(key string, value string, expiration time.Duration) error {
	return service.rdb.Set(service.ctx, key, value, expiration).Err()
}

func (service *RedisServices) Remove(key string) error {
	Log.WithField("module", "redis").Debug("Removing key from Redis")
	return service.rdb.Del(service.ctx, key).Err()
}

func (service *RedisServices) Get(key string) (string, error) {
	Log.WithField("module", "redis").Debug("Fetching value from Redis")
	val := service.rdb.Get(service.ctx, key)
	if val.Err() == redis.Nil {
		return "", fmt.Errorf("Key not found")
	}
	return val.Val(), nil
}

func (service *RedisServices) Close() {
	Log.WithField("module", "redis").Debug("Closing Redis connection")
	service.rdb.Close()
}
