package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/core/config"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
)

type Env struct {
	db *mysql.SqlServices
}

func (env *Env) Initialize() {
	env.db = mysql.GormService().Connect(config.MySQL)
}

func (env *Env) Lifesign(c *gin.Context) {

}

func (env *Env) DiscoverNew(c *gin.Context) {

}
