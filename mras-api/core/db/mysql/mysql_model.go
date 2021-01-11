package mysql

import (
	"database/sql"
	"time"
)

type Permissions struct {
	ID              int32           `gorm:"primaryKey;autoIncrement:true;unique" json:"-"`
	Admin           bool            `gorm:"default:false;not null" json:"admin"`
	CanEdit         bool            `gorm:"default:false;not null" json:"canedit"`
	Speakers        []*Speaker      `gorm:"many2many:perm_speakers" json:"-"`
	SpeakerIDs      []int32         `gorm:"-" json:"speaker_ids"`
	SpeakerGroups   []*SpeakerGroup `gorm:"many2many:perm_speakergroups" json:"-"`
	SpeakerGroupIDs []int32         `gorm:"-" json:"speakergroup_ids"`
	Rooms           []*Room         `gorm:"many2many:perm_rooms" json:"-"`
	RoomIDs         []int32         `gorm:"-" json:"room_ids"`
}

type User struct {
	ID            int32        `gorm:"primaryKey;autoIncrement:true;unique" json:"id"`
	Username      string       `gorm:"size:15;unique" json:"username"`
	Password      string       `gorm:"size:64" json:"-"`
	CreatedAt     time.Time    `json:"-"`
	AvatarID      string       `gorm:"size:10;default:default" json:"avatar_id"`
	PermID        int32        `json:"-"`
	Permissions   Permissions  `gorm:"foreignKey:PermID" json:"-"`
	RefreshToken  string       `json:"-"`
	ResetCode     string       `json:"-"`
	UserGroups    []*UserGroup `gorm:"many2many:user_usergroups" json:"-"`
	UserGroupIDs  []int32      `gorm:"-" json:"usergroup_ids"`
	PasswordReset bool         `gorm:"default:false" json:"-"`
}

type UserGroup struct {
	ID          int32  `gorm:"primaryKey;autoIncrement:true;unique"`
	Name        string `gorm:"size:100"`
	PermID      int32
	Permissions Permissions `gorm:"foreignKey:PermID"`
	Users       []*User     `gorm:"many2many:user_usergroups"`
}

type Speaker struct {
	ID           int32           `gorm:"primaryKey;autoIncrement:true;unique" json:"id"`
	Name         string          `gorm:"size:100" json:"name"`
	Description  string          `json:"description"`
	PosX         sql.NullFloat64 `json:"position_x"`
	PosY         sql.NullFloat64 `json:"position_y"`
	RoomID       int32           `json:"room_id"`
	IPAddress    string          `json:"-"`
	CreatedAt    time.Time       `json:"-"`
	LastLifesign time.Time       `json:"-"`
	Alive        bool            `gorm:"default:true;not null" json:"-"`
	SpeakerGroup []*SpeakerGroup `gorm:"many2many:speakergroup_speakers" json:"-"`
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

const procedure = "create definer = root@`%` procedure checkifalive() begin declare speaker_id int; declare diff int; declare finished integer default 0; declare curId cursor for SELECT id from speakers; declare continue handler for not found set finished = 1; open curId; updAlive: loop FETCH curId into speaker_id; select CURRENT_TIME - TIME_TO_SEC((SELECT last_lifesign from speakers)) into diff; insert into difference(test) value (diff); if diff >= 30 then update speakers set alive = false where id = speaker_id; else update speakers set alive = true where id = speaker_id; end if; if finished = 1 then LEAVE updAlive; end if; end loop updAlive; close curId; end;"
