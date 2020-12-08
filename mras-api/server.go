package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mras-diplomarbeit/mras-api/config"
	"github.com/mras-diplomarbeit/mras-api/db/mysql"
	"github.com/mras-diplomarbeit/mras-api/db/redis"
	"github.com/mras-diplomarbeit/mras-api/middleware"
	"log"
	"net/http"
)

func init() {

}

func main() {
	defer mysql.Con.Close()
	defer redis.Rdb.Close()

	router := mux.NewRouter()
	router.Use(middleware.LoggerMiddleware)

	router.HandleFunc("/test", testhandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":"+fmt.Sprint(config.Port), router))
}

func testhandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Working"))
}
