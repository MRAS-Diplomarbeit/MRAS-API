package main

import (
	"fmt"
	"github.com/mras-diplomarbeit/mras-api/app_api"
	"github.com/mras-diplomarbeit/mras-api/client_api"
	"github.com/mras-diplomarbeit/mras-api/core/config"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
	"github.com/mras-diplomarbeit/mras-api/core/db/redis"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(2)
	var wg sync.WaitGroup
	wg.Add(2)

	_, err := mysql.GormService().Connect(config.MySQL).InitializeModel()
	if err != nil {
		Log.WithField("module", "gorm").Panic(err)
	}

	rdis, err := redis.RedisDBService().Initialize(config.Redis)
	if err != nil {
		Log.WithField("module", "redis").Panic(err)
	}
	rdis.Close()

	apiRouter := app_api.SetupApiRouter()
	go func() {
		err = apiRouter.Run(":" + fmt.Sprint(config.Port))
		if err != nil {
			Log.WithField("module", "router").Error(err)
		}
		wg.Done()
	}()

	clientRouter := client_api.SetupClientRouter()

	go func() {
		err = clientRouter.Run(":3001")
		if err != nil {
			Log.WithField("module", "router").Error(err)
		}
		wg.Done()
	}()

	wg.Wait()
}
