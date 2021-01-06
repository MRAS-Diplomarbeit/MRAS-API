package mysql

import (
	"fmt"
	. "github.com/mras-diplomarbeit/mras-api/logger"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

type SQLService interface {
	Connect(conf map[string]interface{}) (*gorm.DB, error)
	InitializeModel() error
}

type SqlServices struct {
	Con *gorm.DB
}

func GormService() *SqlServices {
	return &SqlServices{}
}

func (service *SqlServices) InitializeModel() (*SqlServices, error) {
	Log.WithFields(logrus.Fields{"module": "gorm"}).Debug("Initializing Database")
	startTime := time.Now()

	if service.Con == nil {
		return service, fmt.Errorf("Connection not initialized")
	}

	service.Con.AutoMigrate(&User{}, &Permissions{}, &UserGroup{}, &Speaker{}, &SpeakerGroup{}, &Room{})

	service.Con.Exec("drop procedure if exists checkifalive")
	service.Con.Exec("drop event if exists alivecheck")
	service.Con.Exec(procedure)
	service.Con.Exec("create event alivecheck on schedule every 30 SECOND on completion preserve  enable  do CALL checkifalive();")

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	Log.WithFields(logrus.Fields{"module": "gorm"}).Infof("Database successfully initialized! [%s]", duration.String())
	return service, nil
}

func (service *SqlServices) Connect(conf map[string]interface{}) *SqlServices {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf["user"],
		conf["password"],
		conf["host"],
		conf["port"],
		conf["dbname"])
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: NewLogger(),
	})

	if err != nil {
		Log.WithFields(logrus.Fields{"module": "gorm"}).Panic(err)
	}
	service.Con = db
	return service
}
