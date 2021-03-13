package mysql

import (
	"bytes"
	"fmt"
	"github.com/markbates/pkger"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
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

	err := service.Con.AutoMigrate(&User{}, &Permissions{}, &UserGroup{}, &Speaker{}, &SpeakerGroup{}, &Room{}, &Sessions{})
	if err != nil {
		Log.WithField("module", "gorm").WithError(err)
		return nil, err
	}

	service.Con.Exec("drop procedure if exists checkifalive")
	service.Con.Exec("drop event if exists alivecheck")

	procedureCheckIfAlive, err := readScript("procedure_checkifalive.sql")
	if err != nil {
		Log.Fatal(err)
	}
	service.Con.Exec(procedureCheckIfAlive)

	eventAliveCheck, err := readScript("event_alivecheck.sql")
	if err != nil {
		Log.Fatal(err)
	}
	service.Con.Exec(eventAliveCheck)

	viewUserUserGroupsPerms, err := readScript("view_user_usergroups_perms.sql")
	if err != nil {
		Log.Fatal(err)
	}
	service.Con.Exec(viewUserUserGroupsPerms)

	viewRoomUserPerms, err := readScript("view_room_user_perms.sql")
	if err != nil {
		Log.Fatal(err)
	}
	service.Con.Exec(viewRoomUserPerms)

	viewSpeakerUserPerms, err := readScript("view_speaker_user_perms.sql")
	if err != nil {
		Log.Fatal(err)
	}
	service.Con.Exec(viewSpeakerUserPerms)

	viewSpeakerGroupUserPerms, err := readScript("view_speakergroup_user_perms.sql")
	if err != nil {
		Log.Fatal(err)
	}
	service.Con.Exec(viewSpeakerGroupUserPerms)

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

func readScript(filename string) (string, error) {
	f, err := pkger.Open("/static/" + filename)
	if err != nil {
		return "", err
	}

	defer f.Close()

	buf := bytes.NewBuffer(nil)

	_, err = io.Copy(buf, f)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
