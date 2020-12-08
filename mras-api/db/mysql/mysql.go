package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mras-diplomarbeit/mras-api/config"
	log "github.com/mras-diplomarbeit/mras-api/logger"
)

var Con *sql.DB

func init() {

	sqlInfo := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.MySQL["user"],
		config.MySQL["password"],
		config.MySQL["host"],
		config.MySQL["port"],
		config.MySQL["dbname"])

	var err error
	Con, err = sql.Open("mysql", sqlInfo)
	if err != nil {
		log.ErrorLogger.Println(err)
		panic(err)
	}
	err = Con.Ping()
	if err != nil {
		log.ErrorLogger.Println(err)
		panic(err)
	}
	log.InfoLogger.Println("MySQL connection established!")
}