package handler

import (
	"github.com/mras-diplomarbeit/mras-api/core/config"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
	"github.com/mras-diplomarbeit/mras-api/core/db/redis"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"gorm.io/gorm"
)

type Env struct {
	db   *gorm.DB
	rdis *redis.RedisServices
}

func (env *Env) Initialize() {
	env.db = mysql.GormService().Connect(config.MySQL).Con

	var err error
	env.rdis, err = redis.RedisDBService().Initialize(config.Redis)
	if err != nil {
		Log.WithField("module", "redis").WithError(err)
	}
}
