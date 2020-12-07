package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"lukaskoenig.at/mras-api/config"
)

func MySQLInit() (*sql.DB, error) {

	sqlInfo := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.MySQL["user"], config.MySQL["password"], config.MySQL["host"], config.MySQL["port"], config.MySQL["dbname"])

	con, err := sql.Open("mysql", sqlInfo)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	err = con.Ping()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return con, nil
}
