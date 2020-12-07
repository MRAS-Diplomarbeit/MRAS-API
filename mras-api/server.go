package main

import (
	"database/sql"
	"github.com/go-redis/redis/v8"
	"lukaskoenig.at/mras-api/config"
	"lukaskoenig.at/mras-api/db"
)

var Con *sql.Conn
var Rdb *redis.Client

func main() {
	err := config.Init()
	if err != nil {
		panic(err)
	}

	Con, err := db.MySQLInit()
	if err != nil {
		panic(err)
	}
	Con.Ping()

	db.RedisInit()

}
