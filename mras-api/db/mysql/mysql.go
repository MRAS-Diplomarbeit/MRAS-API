package mysql

import (
	"fmt"
	. "github.com/mras-diplomarbeit/mras-api/logger"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io/ioutil"
)

type SQLService interface {
	Connect(conf map[string]interface{}) (*gorm.DB, error)
	InitializeModel() error
}

type sqlServices struct {
	Connection *gorm.DB
}

func GormService() *sqlServices {
	return &sqlServices{}
}

func (service *sqlServices) InitializeModel() (*sqlServices, error) {
	Log.WithFields(logrus.Fields{"module": "gorm"}).Debug("Initializing Database")

	if service.Connection == nil {
		return service, fmt.Errorf("Connection not initialized")
	}

	service.Connection.AutoMigrate(&User{}, &Permissions{}, &UserGroup{}, &Speaker{}, &SpeakerGroup{}, &Room{})

	procedure, err := ioutil.ReadFile("procedure.sql")
	if err != nil {
		return service, err
	} else {
		service.Connection.Exec("drop procedure if exists checkifalive")
		service.Connection.Exec("drop event if exists alivecheck")
		service.Connection.Exec(string(procedure))
		service.Connection.Exec("create event alivecheck on schedule every 30 SECOND on completion preserve  enable  do CALL checkifalive();")
	}
	Log.WithFields(logrus.Fields{"module": "gorm"}).Info("Database successfully initialized!")
	return service, nil
}

func (service *sqlServices) Connect(conf map[string]interface{}) *sqlServices {
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
	service.Connection = db
	return service
}
