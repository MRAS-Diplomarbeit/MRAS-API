package handler

import (
	"github.com/mras-diplomarbeit/mras-api/core/config"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
	"gorm.io/gorm"
)

type Env struct {
	db *gorm.DB
}

func (env *Env) Initialize() {
	env.db = mysql.GormService().Connect(config.MySQL).Con
}
