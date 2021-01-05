package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/config"
	"github.com/mras-diplomarbeit/mras-api/db/redis"
	. "github.com/mras-diplomarbeit/mras-api/logger"
	"github.com/mras-diplomarbeit/mras-api/utils"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

var rdis *redis.RedisServices

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdis == nil {
			connectRedis()
		}

		Log.WithFields(logrus.Fields{"module": "middleware"}).Debug("Checking Authorization Header")

		authHeader := c.GetHeader("Authorization")
		if len(authHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH001", Message: "Missing or Invalid Authorization Header"})
			return
		}
		tokenString := strings.Split(c.Request.Header["Authorization"][0], " ")[1]
		Log.WithField("module","middleware").Debug("JWT Token: ",tokenString)

		token, _ := utils.JWTAuthService(config.JWTAccessSecret).ValidateToken(tokenString)
		Log.WithField("module","middleware").Debug(token)
		if token != nil && token.Valid {

			_, err := rdis.Get(tokenString)
			if err != nil {
				Log.WithFields(logrus.Fields{"module": "middleware"}).Warn("JWT not Found in Reids (Epxired)")
				c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH002", Message: "JWT not found in Redis (Expired)"})
				return
			}

			claims := token.Claims.(jwt.MapClaims)
			c.Set("userid", int32(claims["userid"].(float64)))
			c.Set("deviceid", claims["deviceid"])

		} else {
			Log.WithFields(logrus.Fields{"module": "middleware"}).Warn("Invalid JWT")
			c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH001", Message: "Missing or Invalid Authorization Header"})
			return
		}
		c.Next()
	}
}

func connectRedis() {
	var err error
	rdis, err = redis.RedisDBService().Initialize(config.Redis)
	if err != nil {
		Log.WithField("module", "redis").Panic(err)
	}

}
