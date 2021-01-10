package app_api

import (
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/app_api/handler"
	"github.com/mras-diplomarbeit/mras-api/app_api/middleware"
	coremiddleware "github.com/mras-diplomarbeit/mras-api/core/middleware"
)

func SetupApiRouter() *gin.Engine {

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(coremiddleware.LoggerMiddleware("app"))

	noAuthRouter := router.Group("/api/v1")

	noAuthRouter.POST("/user/login", handler.LoginUser)
	noAuthRouter.POST("/user/register", handler.RegisterUser)
	noAuthRouter.POST("/user/refresh", handler.GenerateAccessToken)
	noAuthRouter.POST("/user/password/reset/:username", handler.ResetUserPassword)
	noAuthRouter.POST("/user/password/new/:username", handler.NewUserPassword)

	authRouter := router.Group("/api/v1")
	authRouter.Use(middleware.AuthMiddleware())

	authRouter.GET("/user", handler.GetAllUsers)
	authRouter.GET("/user/:id", handler.GetUser)
	authRouter.DELETE("/user/:id", handler.DeleteUser)
	authRouter.GET("/user/:id/logout", handler.LogoutUser)

	authRouter.GET("/user/:id/permissions", handler.GetPermissions)
	authRouter.PATCH("/user/:id/permissions", handler.UpdatePermissions)

	authRouter.GET("/group/user", handler.GetAllUserGroups)
	authRouter.POST("/group/user", handler.CreateUserGroup)
	authRouter.POST("/group/user/:id", handler.GetUserGroup)
	authRouter.PATCH("/group/user/:id", handler.UpdateUserGroup)
	authRouter.DELETE("/group/user/:id", handler.DeleteUserGroup)

	authRouter.GET("/room", handler.GetAllRooms)
	authRouter.POST("/room", handler.CreateRoom)
	authRouter.PATCH("/room/:id", handler.UpdateRoom)
	authRouter.GET("/room/:id", handler.GetRoom)
	authRouter.POST("/room/:id", handler.EnablePlaybackRoom)
	authRouter.DELETE("/room/:id", handler.DeleteRoom)

	authRouter.GET("/speaker", handler.GetAllSpeakers)
	authRouter.PATCH("/speaker", handler.UpdateSpeaker)
	authRouter.GET("/speaker/:id", handler.GetSpeaker)
	authRouter.POST("/speaker/:id", handler.EnablePlaybackSpeaker)
	authRouter.DELETE("/speaker/:id", handler.RemoveSpeaker)

	authRouter.GET("/group/speaker", handler.GetAllSpeakerGroups)
	authRouter.POST("/group/speaker", handler.CreateSpeakerGroup)
	authRouter.GET("/group/speaker/:id", handler.GetSpeakerGroup)
	authRouter.POST("/group/speaker/:id", handler.EnablePlaybackSpeakerGroup)
	authRouter.PATCH("/group/speaker/:id", handler.UpdateSpeakerGroup)
	authRouter.DELETE("/group/speaker/:id", handler.DeleteSpeakerGroup)

	return router
}
