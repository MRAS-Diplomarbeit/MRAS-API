package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/config"
	"github.com/mras-diplomarbeit/mras-api/db/mysql"
	"github.com/mras-diplomarbeit/mras-api/db/redis"
	"github.com/mras-diplomarbeit/mras-api/handler"
	"github.com/mras-diplomarbeit/mras-api/middleware"
)

func init() {

}

func main() {
	defer mysql.Con.Close()
	defer redis.Rdb.Close()

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware())

	noAuthRouter := router.Group("/api/v1")

	noAuthRouter.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	noAuthRouter.GET("/test", handler.TestHandler)
	noAuthRouter.GET("/user/login", handler.LoginHandler)

	authRouter := router.Group("/api/v1")

	authRouter.Use(middleware.AuthMiddleware())
	authRouter.POST("/authtest", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"userid":   c.GetString("userid"),
			"deviceid": c.GetString("deviceid"),
		})
	})

	router.Run(":" + fmt.Sprint(config.Port))

	//noAuthRouter := mux.NewRouter()
	//noAuthRouter.Use(middleware.LoggerMiddleware)
	//
	//noAuthRouter.HandleFunc("/test", testhandler).Methods("GET")
	//noAuthRouter.HandleFunc("/api/v1/login", handler.LoginHandler).Methods("POST")
	//noAuthRouter.HandleFunc("/api/v1/register", handler.RegisterHandler).Methods("POST")
	//noAuthRouter.HandleFunc("/api/v1/refresh",handler.GenerateAccessToken).Methods("POST")
	//noAuthRouter.HandleFunc("/api/v1/user/{username}/password/reset",handler.GenerateAccessToken).Methods("POST")
	//noAuthRouter.HandleFunc("/api/v1/user/{username}/password/new",handler.GenerateAccessToken).Methods("POST")
	//
	//authRouter := noAuthRouter.PathPrefix("/api/v1").Subrouter()
	//authRouter.Use(middleware.AuthMiddleware)

	//log.Fatal(http.ListenAndServe(":"+fmt.Sprint(config.Port), noAuthRouter))
}

//func testhandler(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Content-Type", "text/plain")
//	logger.ErrorLogger.Println("errortest")
//	w.WriteHeader(http.StatusOK)
//	w.Write([]byte("Working"))
//}
