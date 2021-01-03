package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/mras-diplomarbeit/mras-api/config"
	. "github.com/mras-diplomarbeit/mras-api/logger"
	"github.com/sirupsen/logrus"
)

type RedisService interface {
	Initialize(conf map[string]interface{}) (*redisServices, error)
}
type redisServices struct {
	Ctx context.Context
	Rdb *redis.Client
}

func RedisDBService() *redisServices {
	return &redisServices{}
}

func (service *redisServices) Initialize(conf map[string]interface{}) (*redisServices, error) {
	service.Ctx = context.Background()
	service.Rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis["host"], config.Redis["port"]),
		Password: conf["password"].(string),
		DB:       conf["db"].(int),
	})

	err := service.Rdb.Ping(service.Ctx).Err()

	if err != nil {
		return service, err
	}

	Log.WithFields(logrus.Fields{"module": "redis"}).Info("Redis connection established!")
	return service, nil
}
