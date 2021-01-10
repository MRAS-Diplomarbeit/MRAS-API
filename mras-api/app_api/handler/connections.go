package handler

import (
	"github.com/mras-diplomarbeit/mras-api/core/config"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
	"github.com/mras-diplomarbeit/mras-api/core/db/redis"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
)

var rdis *redis.RedisServices
var db *mysql.SqlServices

//Connect to redis database
func connectRedis() {
	var err error
	rdis, err = redis.RedisDBService().Initialize(config.Redis)
	if err != nil {
		Log.WithField("module", "redis").WithError(err)
	}

}

//create connections to mysql database
func connectMySql() {
	db = mysql.GormService().Connect(config.MySQL)
}
