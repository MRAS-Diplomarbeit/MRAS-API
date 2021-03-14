package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
	errs "github.com/mras-diplomarbeit/mras-api/core/error"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (env *Env) GetAllSpeakerGroups(c *gin.Context) {

	type resAllSpeakerGroups struct {
		Count  int                  `json:"count"`
		Groups []mysql.SpeakerGroup `json:"groups"`
	}

	var groups []mysql.SpeakerGroup
	userid, _ := c.Get("userid")

	result := env.db.Where("speaker_groups.id in (select speakergoup_id from speakergroup_user_perms where user_id = @userid)",
		sql.Named("userid", userid)).Preload(clause.Associations).Find(&groups)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	for i := 0; i < len(groups); i++ {
		for _, speaker := range groups[i].Speaker {
			groups[i].SpeakerIds = append(groups[i].SpeakerIds, speaker.ID)
		}
	}

	c.JSON(http.StatusOK, resAllSpeakerGroups{Count: len(groups), Groups: groups})
}

func (env *Env) CreateSpeakerGroup(c *gin.Context) {

	type reqCreateSpeakerGroup struct {
		Name       string  `json:"name"`
		SpeakerIds []int32 `json:"speaker_ids"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request reqCreateSpeakerGroup
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var speakergroup mysql.SpeakerGroup

	speakergroup.Name = request.Name
	speakergroup.SpeakerIds = request.SpeakerIds

	result := env.db.Find(&speakergroup.Speaker, speakergroup.SpeakerIds)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	result = env.db.Save(&speakergroup)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, &speakergroup)
}

func (env *Env) UpdateSpeakerGroup(c *gin.Context) {

	type reqUpdtSpeakerGroup struct {
		ID         int32       `json:"id"`
		Name       null.String `json:"name"`
		SpeakerIds []int32     `json:"speaker_ids"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request reqUpdtSpeakerGroup
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var orgGroup mysql.SpeakerGroup
	orgGroup.ID = request.ID

	result := env.db.Find(&orgGroup)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.DBSQ001)
		return
	}

	if request.Name.Valid {
		orgGroup.Name = request.Name.String
	}

	if request.SpeakerIds != nil {
		if len(request.SpeakerIds) == 0 {
			orgGroup.Speaker = nil
		} else {
			orgGroup.Speaker = nil
			orgGroup.SpeakerIds = request.SpeakerIds
			result = env.db.Find(&orgGroup.Speaker, orgGroup.SpeakerIds)
			if result.Error != nil {
				Log.WithField("module", "sql").WithError(result.Error)
				c.AbortWithStatusJSON(http.StatusBadRequest, errs.DBSQ001)
				return
			}
		}
	}

	result = env.db.Save(&orgGroup)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, &orgGroup)

}

func (env *Env) GetSpeakerGroup(c *gin.Context) {

	var group mysql.SpeakerGroup

	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	group.ID = int32(tmp)

	result := env.db.Preload(clause.Associations).Find(&group)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.DBSQ001)
		return
	}

	for _, speaker := range group.Speaker {
		group.SpeakerIds = append(group.SpeakerIds, speaker.ID)
	}

	c.JSON(http.StatusOK, &group)
}

func (env *Env) DeleteSpeakerGroup(c *gin.Context) {

	var group mysql.SpeakerGroup

	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	group.ID = int32(tmp)

	result := env.db.Delete(&group)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.DBSQ001)
		return
	}

}

func (env *Env) EnablePlaybackSpeakerGroup(c *gin.Context) {

}
