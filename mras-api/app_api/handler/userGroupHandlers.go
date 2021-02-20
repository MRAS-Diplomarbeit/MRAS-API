package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
	errs "github.com/mras-diplomarbeit/mras-api/core/error"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (env *Env) GetAllUserGroups(c *gin.Context) {

	type resAllUserGroups struct {
		Count      int               `json:"count"`
		UserGroups []mysql.UserGroup `json:"groups"`
	}

	var groups []mysql.UserGroup

	result := env.db.Find(&groups)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, resAllUserGroups{Count: len(groups), UserGroups: groups})
}

func (env *Env) CreateUserGroup(c *gin.Context) {

	type reqCreateUserGroup struct {
		Name    string            `json:"name"`
		Perms   mysql.Permissions `json:"perms"`
		UserIds []int32           `json:"user_ids"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var reqUserGroup reqCreateUserGroup
	err = json.Unmarshal(jsonData, &reqUserGroup)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	result := env.db.Save(&reqUserGroup.Perms)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	newUserGroup := mysql.UserGroup{Name: reqUserGroup.Name, PermID: reqUserGroup.Perms.ID, UserIDs: reqUserGroup.UserIds}

	result = env.db.Save(&newUserGroup)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, newUserGroup)
}

func (env *Env) GetUserGroup(c *gin.Context) {

	var userGroup mysql.UserGroup

	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	userGroup.ID = int32(tmp)

	result := env.db.Find(&userGroup)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, userGroup)
}

func (env *Env) UpdateUserGroup(c *gin.Context) {

	//type updPerm struct {
	//	Admin           null.Bool `json:"admin"`
	//	CanEdit         null.Bool `json:"canedit"`
	//	SpeakerIDs      []int32   `json:"speaker_ids"`
	//	SpeakerGroupIDs []int32   `json:"speakergroup_ids"`
	//	RoomIDs         []int32   `json:"room_ids"`
	//}

	type reqUpdtUserGroup struct {
		ID      int         `json:"id"`
		Name    null.String `json:"name"`
		Perms   updtPerm    `json:"perms"`
		UserIds []int32     `json:"user_ids"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var updtUserGroup reqUpdtUserGroup
	err = json.Unmarshal(jsonData, &updtUserGroup)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var ogUserGroup mysql.UserGroup
	ogUserGroup.ID = int32(updtUserGroup.ID)

	result := env.db.Preload(clause.Associations).Find(&ogUserGroup)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ008)
			return
		}
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if error := env.updatePermissions(&ogUserGroup.Permissions, updtUserGroup.Perms); error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, error)
		return
	}

	if updtUserGroup.Name.Valid {
		ogUserGroup.Name = updtUserGroup.Name.String
	}

	if updtUserGroup.UserIds != nil {
		if len(updtUserGroup.UserIds) == 0 {
			err := env.db.Model(&updtUserGroup).Association("Users").Clear()
			if err != nil {
				Log.WithField("module", "sql").WithError(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			}
		} else {
			result = env.db.Find(&ogUserGroup, updtUserGroup.UserIds)
			if result.Error != nil {
				Log.WithField("module", "sql").WithError(result.Error)
				c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			}
		}
	}

	for _, users := range ogUserGroup.Users {
		ogUserGroup.UserIDs = append(ogUserGroup.UserIDs, users.ID)
	}

	result = env.db.Save(&ogUserGroup)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, ogUserGroup)

}

func (env *Env) DeleteUserGroup(c *gin.Context) {

	var userGroup mysql.UserGroup

	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	userGroup.ID = int32(tmp)

	result := env.db.Delete(&userGroup)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

}
