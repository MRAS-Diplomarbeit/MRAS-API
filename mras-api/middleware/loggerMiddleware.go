package middleware

import (
	log "github.com/mras-diplomarbeit/mras-api/logger"
	"net/http"
	"time"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Now().Sub(startTime)
		go log.InfoLogger.Println(r.Method + " " + r.RequestURI + " " + r.RemoteAddr + " "+ duration.String())
	})
}
