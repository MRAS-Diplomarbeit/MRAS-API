package client_api

import (
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/client_api/handler"
	coremiddleware "github.com/mras-diplomarbeit/mras-api/core/middleware"
)

func SetupClientRouter() *gin.Engine {

	router := gin.New()
	router.Use(coremiddleware.LoggerMiddleware("client"))
	router.GET("/test", handler.TestHandler)

	return router
}
