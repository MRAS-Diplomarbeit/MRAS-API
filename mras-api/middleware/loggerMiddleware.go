package middleware

import (
	"github.com/gin-gonic/gin"
	. "github.com/mras-diplomarbeit/mras-api/logger"
	"time"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		startTime := time.Now()

		c.Next()

		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)

		reqMethod := c.Request.Method
		reqUri := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		//"| %3d | %13v | %15s | %s | %s |"

		if statusCode == 200 {
			Log.Infof("| %d | %v | %s | %s | %s |",
				statusCode,
				latencyTime,
				clientIP,
				reqMethod,
				reqUri,
			)
		} else if statusCode >= 500 {
			Log.Errorf("| %d | %v | %s | %s | %s |",
				statusCode,
				latencyTime,
				clientIP,
				reqMethod,
				reqUri,
			)
		} else {
			Log.Warnf("| %d | %v | %s | %s | %s |",
				statusCode,
				latencyTime,
				clientIP,
				reqMethod,
				reqUri,
			)
		}

	}
}
