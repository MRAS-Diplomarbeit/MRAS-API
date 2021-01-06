package middleware

import (
	"github.com/gin-gonic/gin"
	. "github.com/mras-diplomarbeit/mras-api/logger"
	"github.com/sirupsen/logrus"
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

		if latencyTime < time.Millisecond*500 && statusCode == 200 {
			Log.WithFields(logrus.Fields{"module": "router"}).Infof("| %d | %v | %s | %s | %s |",
				statusCode,
				latencyTime,
				clientIP,
				reqMethod,
				reqUri,
			)
		} else if statusCode >= 500 {
			Log.WithFields(logrus.Fields{"module": "router"}).Errorf("| %d | %v | %s | %s | %s |",
				statusCode,
				latencyTime,
				clientIP,
				reqMethod,
				reqUri,
			)
		} else {
			if latencyTime < time.Millisecond*500 {
				Log.WithFields(logrus.Fields{"module": "router"}).Warnf("| %d | %v | %s | %s | %s |",
					statusCode,
					latencyTime,
					clientIP,
					reqMethod,
					reqUri,
				)
			} else {
				Log.WithFields(logrus.Fields{"module": "router"}).Warnf("| %d | SLOW | %v | %s | %s | %s |",
					statusCode,
					latencyTime,
					clientIP,
					reqMethod,
					reqUri,
				)
			}

		}

	}
}
