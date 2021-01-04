package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/config"
	"github.com/mras-diplomarbeit/mras-api/db/redis"
	. "github.com/mras-diplomarbeit/mras-api/logger"
	"github.com/mras-diplomarbeit/mras-api/service"
	"github.com/sirupsen/logrus"
	"net/http"
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
		tokenString := authHeader[len("Bearer "):]

		token, _ := service.JWTAuthService(config.JWTAccessSecret).ValidateToken(tokenString)
		if token.Valid {

			_, err := rdis.Get(tokenString)
			if err != nil {
				Log.WithFields(logrus.Fields{"module": "middleware"}).Warn("JWT not Found in Reids (Epxired)")
				c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH002", Message: "JWT not found in Redis (Expired)"})
			}

			claims := token.Claims.(jwt.MapClaims)
			c.Set("userid", claims["userid"])
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
