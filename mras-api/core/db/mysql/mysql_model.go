package mysql

import (
	"gopkg.in/guregu/null.v4"
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
	Permissions   Permissions  `gorm:"foreignKey:PermID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	RefreshToken  string       `json:"-"`
	ResetCode     string       `json:"-"`
	UserGroups    []*UserGroup `gorm:"many2many:user_usergroups" json:"-"`
	UserGroupIDs  []int32      `gorm:"-" json:"usergroup_ids"`
	PasswordReset bool         `gorm:"default:false" json:"-"`
}

type UserGroup struct {
	ID          int32       `gorm:"primaryKey;autoIncrement:true;unique" json:"id"`
	Name        string      `gorm:"size:100" json:"name"`
	PermID      int32       `json:"perm_id"`
	Permissions Permissions `gorm:"foreignKey:PermID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Users       []*User     `gorm:"many2many:user_usergroups" json:"-"`
	UserIDs     []int32     `gorm:"-" json:"user_id"`
}

type Speaker struct {
	ID           int32           `gorm:"primaryKey;autoIncrement:true;unique" json:"id"`
	Name         string          `gorm:"size:100" json:"name"`
	Description  string          `json:"description"`
	Position     Position        `gorm:"embedded" json:"position"`
	RoomID       int32           `json:"room_id"`
	IPAddress    string          `json:"-"`
	CreatedAt    time.Time       `json:"-"`
	LastLifesign time.Time       `json:"-"`
	Alive        bool            `gorm:"default:true;not null" json:"-"`
	SpeakerGroup []*SpeakerGroup `gorm:"many2many:speakergroup_speakers" json:"-"`
}

type Position struct {
	PosX null.Float `json:"x"`
	PosY null.Float `json:"y"`
}

type SpeakerGroup struct {
	ID         int32      `gorm:"primaryKey;autoIncrement:true;unique" json:"id"`
	Name       string     `gorm:"not null;size:100" json:"name"`
	Speaker    []*Speaker `gorm:"many2many:speakergroup_speakers" json:"-"`
	SpeakerIds []int32    `gorm:"-" json:"speaker_ids"`
}

type Room struct {
	ID          int32      `gorm:"primaryKey;autoIncrement:true;unique;not null" json:"id"`
	Name        string     `gorm:"not null;size:100" json:"name"`
	Description string     `json:"description"`
	Dimensions  Dimensions `gorm:"embedded" json:"position"`
	CreatedAt   time.Time
}

type Dimensions struct {
	Height null.Float `json:"height"`
	Width  null.Float `json:"width"`
}

const procedure = "create definer = root@`%` procedure checkifalive() begin declare speaker_id int; declare diff int; declare finished integer default 0; declare curId cursor for SELECT id from speakers; declare continue handler for not found set finished = 1; open curId; updAlive:loop FETCH curId into speaker_id; select TIMESTAMPDIFF(SECOND,(SELECT last_lifesign from speakers where id = speaker_id),CURRENT_TIMESTAMP) into diff; if diff >= 30 then update speakers set alive = 0 where id = speaker_id; else update speakers set alive = 1 where id = speaker_id; end if; if finished = 1 then LEAVE updAlive; end if; end loop updAlive; close curId; end;"
