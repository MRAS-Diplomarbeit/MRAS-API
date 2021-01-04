package mysql

import (
	"database/sql"
	"time"
)

type Permissions struct {
	ID           int32           `gorm:"primaryKey;autoIncrement:true;unique"`
	Admin        bool            `gorm:"default:false;not null"`
	CanEdit      bool            `gorm:"default:false;not null"`
	Speaker      []*Speaker      `gorm:"many2many:perm_speakers"`
	SpeakerGroup []*SpeakerGroup `gorm:"many2many:perm_speakergroups"`
	Room         []*Room         `gorm:"many2many:perm_rooms"`
}

type User struct {
	ID           int32        `gorm:"primaryKey;autoIncrement:true;unique" json:"id"`
	Username     string       `gorm:"size:15;unique" json:"username"`
	Password     string       `gorm:"size:64" json:"-"`
	CreatedAt    time.Time    `json:"-"`
	AvatarID     string       `gorm:"size:10;default:default" json:"avatar_id"`
	PermID       int32        `json:"-"`
	Permissions  Permissions  `gorm:"foreignKey:PermID" json:"-"`
	RefreshToken string       `json:"-"`
	ResetCode    string       `json:"-"`
	UserGroups   []*UserGroup `gorm:"many2many:user_usergroups" json:"-"`
	UserGroupIDs []int32      `gorm:"-" json:"usergroup_ids"`
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
