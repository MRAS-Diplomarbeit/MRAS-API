package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
	errs "github.com/mras-diplomarbeit/mras-api/core/error"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
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

	result := env.db.Con.Find(&groups)
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

	result := env.db.Con.Save(&reqUserGroup.Perms)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	newUserGroup := mysql.UserGroup{Name: reqUserGroup.Name, PermID: reqUserGroup.Perms.ID, UserIDs: reqUserGroup.UserIds}

	result = env.db.Con.Save(&newUserGroup)
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

	result := env.db.Con.Find(&userGroup)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, userGroup)
}

func (env *Env) UpdateUserGroup(c *gin.Context) {

	type updPerm struct {
		Admin           null.Bool `json:"admin"`
		CanEdit         null.Bool `json:"canedit"`
		SpeakerIDs      []int32   `json:"speaker_ids"`
		SpeakerGroupIDs []int32   `json:"speakergroup_ids"`
		RoomIDs         []int32   `json:"room_ids"`
	}

	type reqUpdtUserGroup struct {
		ID      int         `json:"id"`
		Name    null.String `json:"name"`
		Perms   updPerm     `json:"perms"`
		UserIds []int32     `json:"user_ids"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var uptdUserGroup reqUpdtUserGroup
	err = json.Unmarshal(jsonData, &uptdUserGroup)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var ogUserGroup mysql.UserGroup
	ogUserGroup.ID = int32(uptdUserGroup.ID)

	result := env.db.Con.Find(&ogUserGroup)
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

	err = env.db.Con.Model(&ogUserGroup).Association("Permissions").Find(&ogUserGroup.Permissions)
	if err != nil {
		Log.WithField("module", "sql").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if uptdUserGroup.Name.Valid {
		ogUserGroup.Name = uptdUserGroup.Name.String
	}
	if uptdUserGroup.Perms.Admin.Valid {
		ogUserGroup.Permissions.Admin = uptdUserGroup.Perms.Admin.Bool
	}
	if uptdUserGroup.Perms.CanEdit.Valid {
		ogUserGroup.Permissions.CanEdit = uptdUserGroup.Perms.CanEdit.Bool
	}
	//TODO: Fix Update Permission
	//if len(uptdUserGroup.Perms.SpeakerIDs) != len(ogUserGroup.Permissions.Speakers) {
	//	ogUserGroup.Permissions.SpeakerIDs = uptdUserGroup.Perms.SpeakerIDs
	//}
	//if uptdUserGroup.Perms.SpeakerGroupIDs != nil {
	//	ogUserGroup.Permissions.SpeakerGroupIDs = uptdUserGroup.Perms.SpeakerGroupIDs
	//}
	//if uptdUserGroup.Perms.RoomIDs != nil {
	//	ogUserGroup.Permissions.RoomIDs = uptdUserGroup.Perms.RoomIDs
	//}

	result = env.db.Con.Save(&ogUserGroup.Permissions)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	result = env.db.Con.Save(&ogUserGroup)
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

	result := env.db.Con.Delete(&userGroup)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

}
