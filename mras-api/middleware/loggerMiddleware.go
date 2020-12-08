package middleware

import (
	log "github.com/mras-diplomarbeit/mras-api/logger"
	"net/http"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		go log.InfoLogger.Println(r.Method + " " + r.RequestURI + " " + r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
