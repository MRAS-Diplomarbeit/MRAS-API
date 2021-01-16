package main

import (
	"flag"
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

	var threads int
	var cfgPath string
	var debug bool
	flag.IntVar(&threads, "thread", 2, "number of threads used by application")
	flag.IntVar(&threads, "t", 2, "number of threads used by application")
	flag.StringVar(&cfgPath, "config", ".", "path to config file")
	flag.StringVar(&cfgPath, "c", ".", "path to config file")
	flag.BoolVar(&debug, "debug", false, "activate debug mode")
	flag.BoolVar(&debug, "d", false, "activate debug mode")

	flag.Parse()

	runtime.GOMAXPROCS(threads)

	config.LoadConfig(cfgPath)
	if debug {
		config.Loglevel = "DEBUG"
	}
	InitLogger()

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
		err = apiRouter.Run(":" + fmt.Sprint(config.AppPort))
		if err != nil {
			Log.WithField("module", "router").Error(err)
		}
		wg.Done()
	}()

	clientRouter := client_api.SetupClientRouter()
	go func() {
		err = clientRouter.Run(":" + fmt.Sprint(config.ClientPort))
		if err != nil {
			Log.WithField("module", "router").Error(err)
		}
		wg.Done()
	}()

	wg.Wait()
}
