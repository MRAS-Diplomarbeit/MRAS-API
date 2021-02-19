package client_api

import (
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/client_api/handler"
	coremiddleware "github.com/mras-diplomarbeit/mras-api/core/middleware"
)

func SetupClientRouter() *gin.Engine {

	router := gin.New()
	router.Use(coremiddleware.LoggerMiddleware("client"))

	env := &handler.Env{}
	env.Initialize()

	router.GET("/discover", env.DiscoverNew)
	router.GET("/discover/:id", env.Lifesign)

	return router
}
