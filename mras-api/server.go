package main

import (
	"github.com/mras-diplomarbeit/mras-api/db/mysql"
	"github.com/mras-diplomarbeit/mras-api/db/redis"
	log "github.com/mras-diplomarbeit/mras-api/logger"
)

func init() {

}

func main() {
	mysql.Con.Ping()
	redis.Rdb.Ping(redis.Ctx)
	log.ErrorLogger.Println("You Muppet!")

}
