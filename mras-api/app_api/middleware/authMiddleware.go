package middleware

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/core/config"
	"github.com/mras-diplomarbeit/mras-api/core/db/redis"
	errs "github.com/mras-diplomarbeit/mras-api/core/error"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"github.com/mras-diplomarbeit/mras-api/core/utils"
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH001)
			return
		}
		tokenString := strings.Split(c.Request.Header["Authorization"][0], " ")[1]

		token, _ := utils.JWTAuthService(config.JWTAccessSecret).ValidateToken(tokenString)
		if token != nil && token.Valid {

			claims := token.Claims.(jwt.MapClaims)
			userid := int32(claims["userid"].(float64))
			deviceid := claims["deviceid"]

			redistoken, err := rdis.Get(fmt.Sprint(userid))
			if err != nil {
				Log.WithFields(logrus.Fields{"module": "middleware"}).Warn("JWT not Found in Reids (Epxired)")
				c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH002)
				return
			}

			if redistoken != tokenString {
				c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH002)
				return
			}

			c.Set("userid", userid)
			c.Set("deviceid", deviceid)

		} else {
			Log.WithFields(logrus.Fields{"module": "middleware"}).Warn("Invalid JWT")
			c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH001)
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
