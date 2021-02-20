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

	env := &handler.Env{}
	env.Initialize()

	noAuthRouter.POST("/user/login", env.LoginUser)
	noAuthRouter.POST("/user/register", env.RegisterUser)
	noAuthRouter.POST("/user/refresh", env.GenerateAccessToken)
	noAuthRouter.POST("/user/password/reset/:username", env.ResetUserPassword)
	noAuthRouter.POST("/user/password/new/:username", env.NewUserPassword)

	authRouter := router.Group("/api/v1")
	authRouter.Use(middleware.AuthMiddleware())

	authRouter.GET("/user", env.GetAllUsers)
	authRouter.GET("/user/:id", env.GetUser)
	authRouter.DELETE("/user/:id", env.DeleteUser)
	authRouter.GET("/user/:id/logout", env.LogoutUser)

	authRouter.GET("/user/:id/permissions", env.GetUserPermissions)
	authRouter.PATCH("/user/:id/permissions", env.UpdateUserPermissions)

	authRouter.GET("/group/user", env.GetAllUserGroups)
	authRouter.POST("/group/user", env.CreateUserGroup)
	authRouter.PATCH("/group/user", env.UpdateUserGroup)
	authRouter.GET("/group/user/:id", env.GetUserGroup)
	authRouter.DELETE("/group/user/:id", env.DeleteUserGroup)

	authRouter.GET("/room", env.GetAllRooms)
	authRouter.POST("/room", env.CreateRoom)
	authRouter.PATCH("/room/:id", env.UpdateRoom)
	authRouter.GET("/room/:id", env.GetRoom)
	authRouter.POST("/room/:id", env.EnablePlaybackRoom)
	authRouter.DELETE("/room/:id", env.DeleteRoom)

	authRouter.GET("/speaker", env.GetAllSpeakers)
	authRouter.PATCH("/speaker", env.UpdateSpeaker)
	authRouter.GET("/speaker/:id", env.GetSpeaker)
	authRouter.DELETE("/speaker/:id", env.RemoveSpeaker)

	authRouter.POST("/speaker/:id/playback", env.EnablePlaybackSpeaker)
	authRouter.DELETE("/speaker/:id/playback", env.StopPlaybackSpeaker)

	authRouter.GET("/group/speaker", env.GetAllSpeakerGroups)
	authRouter.POST("/group/speaker", env.CreateSpeakerGroup)
	authRouter.GET("/group/speaker/:id", env.GetSpeakerGroup)
	authRouter.POST("/group/speaker/:id", env.EnablePlaybackSpeakerGroup)
	authRouter.PATCH("/group/speaker/:id", env.UpdateSpeakerGroup)
	authRouter.DELETE("/group/speaker/:id", env.DeleteSpeakerGroup)

	return router
}
