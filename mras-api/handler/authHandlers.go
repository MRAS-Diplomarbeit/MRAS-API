package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/config"
	"github.com/mras-diplomarbeit/mras-api/service"
	"net/http"
)

type Test struct {
	Stringtest string `json:"teststring"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Working"))
}

func GenerateAccessToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Working"))
}

func TestHandler(c *gin.Context) {
	ret := Test{"test123"}
	c.JSON(200, ret)
}

func LoginHandler(c *gin.Context) {
	accessToken, _ := service.JWTAuthService(config.JWTAccessSecret).GenerateToken("123", "test123")

	c.JSON(200, accessToken)
}
