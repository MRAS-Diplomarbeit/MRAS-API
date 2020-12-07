package db

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"lukaskoenig.at/mras-api/config"
)

var ctx = context.Background()

func RedisInit() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis["host"], config.Redis["port"]),
		Password: config.Redis["password"].(string), // no password set
		DB:       config.Redis["db"].(int),          // use default DB
	})

	err := rdb.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}
