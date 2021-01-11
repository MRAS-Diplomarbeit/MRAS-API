package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
	errs "github.com/mras-diplomarbeit/mras-api/core/error"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"net/http"
)

func GetAllSpeakers(c *gin.Context) {

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	type resAllSpeakers struct {
		Count    int             `json:"count"`
		Speakers []mysql.Speaker `json:"speakers"`
	}

	var speakers []mysql.Speaker
	userid, _ := c.Get("userid")

	//Get all Speakers from Database
	result := db.Con.Where("(speakers.id in (select speaker_id from perm_speakers where permissions_id = (select perm_id from users where users.id = ?)) or speakers.id in (select speaker_id from perm_speakers where permissions_id = (select perm_id from user_groups where user_groups.id in (select user_group_id from user_usergroups where user_id = ?)))) and speakers.alive = true", userid, userid).Find(&speakers)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	for _, speaker := range speakers {
		if speaker.PosX.Valid && speaker.PosY.Valid {
			speaker.Position.X = speaker.PosX.Float64
			speaker.Position.Y = speaker.PosY.Float64
		}
	}

	c.JSON(http.StatusOK, resAllSpeakers{Count: len(speakers), Speakers: speakers})
}

func UpdateSpeaker(c *gin.Context) {

}

func GetSpeaker(c *gin.Context) {

}

func EnablePlaybackSpeaker(c *gin.Context) {

}

func RemoveSpeaker(c *gin.Context) {

}
