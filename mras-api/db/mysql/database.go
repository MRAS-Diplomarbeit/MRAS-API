package mysql

import (
	"database/sql"
	"fmt"
	"github.com/mras-diplomarbeit/mras-api/config"
	. "github.com/mras-diplomarbeit/mras-api/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"io/ioutil"
	"log"
	"time"
)

var MySQL *gorm.DB

func InitDB() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.MySQL["user"],
		config.MySQL["password"],
		config.MySQL["host"],
		config.MySQL["port"],
		"mras_test")
	MySQL, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: NewLogger(),
	})

	MySQL.Logger.LogMode(gormlogger.Info)

	if err != nil {
		log.Panic("Failed to Connect do MySQL ", err)
	}

	MySQL.AutoMigrate(&User{}, &Permissions{}, &UserGroup{}, &Speaker{}, &SpeakerGroup{}, &Room{})

	procedure, err := ioutil.ReadFile("procedure.sql")
	if err != nil {
		Log.Warn(err)
	} else {
		MySQL.Exec("drop procedure if exists checkifalive")
		MySQL.Exec("drop event if exists alivecheck")
		MySQL.Exec(string(procedure))
		MySQL.Exec("create event alivecheck on schedule every 30 SECOND on completion preserve  enable  do CALL checkifalive();")
	}

}

type Permissions struct {
	ID           int32           `gorm:"primaryKey;autoIncrement:true;unique"`
	Admin        bool            `gorm:"default:false;not null"`
	CanEdit      bool            `gorm:"default:false;not null"`
	Speaker      []*Speaker      `gorm:"many2many:perm_speakers"`
	SpeakerGroup []*SpeakerGroup `gorm:"many2many:perm_speakergroups"`
	Room         []*Room         `gorm:"many2many:perm_rooms"`
}

type User struct {
	ID           int32  `gorm:"primaryKey;autoIncrement:true;unique"`
	Username     string `gorm:"size:50"`
	Password     string `gorm:"size:64"`
	CreatedAt    time.Time
	AvatarID     string `gorm:"size:10"`
	PermID       int32
	Permissions  Permissions `gorm:"foreignKey:PermID"`
	RefreshToken string
	ResetCode    string
	UserGroups   []*UserGroup `gorm:"many2many:user_usergroups"`
}

type UserGroup struct {
	ID          int32  `gorm:"primaryKey;autoIncrement:true;unique"`
	Name        string `gorm:"size:100"`
	PermID      int32
	Permissions Permissions `gorm:"foreignKey:PermID"`
	Users       []*User     `gorm:"many2many:user_usergroups"`
}

type Speaker struct {
	ID           int32  `gorm:"primaryKey;autoIncrement:true;unique"`
	Name         string `gorm:"size:100"`
	Description  string
	PosX         sql.NullFloat64
	PosY         sql.NullFloat64
	RoomID       int32
	IPAddress    string
	CreatedAt    time.Time
	LastLifesign time.Time
	Alive        bool            `gorm:"default:true;not null"`
	SpeakerGroup []*SpeakerGroup `gorm:"many2many:speakergroup_speakers"`
}

type SpeakerGroup struct {
	ID      int32      `gorm:"primaryKey;autoIncrement:true;unique"`
	Name    string     `gorm:"not null;size:100"`
	Speaker []*Speaker `gorm:"many2many:speakergroup_speakers"`
}

type Room struct {
	ID          int32  `gorm:"primaryKey;autoIncrement:true;unique;not null"`
	Name        string `gorm:"not null;size:100"`
	Description string
	DimHeight   sql.NullFloat64
	DimWidth    sql.NullFloat64
	CreatedAt   time.Time
}
