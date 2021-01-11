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

	//Get all Speakers from Database
	result := db.Con.Where("").Find(&speakers)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}
}

func UpdateSpeaker(c *gin.Context) {

}

func GetSpeaker(c *gin.Context) {

}

func EnablePlaybackSpeaker(c *gin.Context) {

}

func RemoveSpeaker(c *gin.Context) {

}
