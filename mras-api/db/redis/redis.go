package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/mras-diplomarbeit/mras-api/config"
	log "github.com/mras-diplomarbeit/mras-api/logger"
)

var Ctx = context.Background()
var Rdb *redis.Client

func init() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis["host"], config.Redis["port"]),
		Password: config.Redis["password"].(string),
		DB:       config.Redis["db"].(int),
	})

	err := Rdb.Ping(Ctx).Err()
	if err != nil {
		log.ErrorLogger.Println(err)
		panic(err)
	}

	log.InfoLogger.Println("Redis connection established!")

}
