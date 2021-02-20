package handler

import (
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

type updtPerm struct {
	Admin           null.Bool `json:"admin"`
	CanEdit         null.Bool `json:"canedit"`
	SpeakerIDs      []int32   `json:"speaker_ids"`
	SpeakerGroupIDs []int32   `json:"speakergroup_ids"`
	RoomIDs         []int32   `json:"room_ids"`
}

//This function handles GET requests sent to the /api/v1/user/:id/permissions endpoint
func (env *Env) GetUserPermissions(c *gin.Context) {

	type getPermsResponse struct {
		Perms mysql.Permissions `json:"perms"`
	}

	//Convert ID Parameter into int32
	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	userid := int32(tmp)

	//Check if UserID exists
	var exists int64
	result := env.db.Model(mysql.User{}).Where("id = ?", userid).Count(&exists)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if exists == 0 {
		Log.WithField("module", "handler").Error("User not found in Database")
		c.AbortWithStatusJSON(http.StatusNotFound, errs.DBSQ006)
		return
	}

	//var perm mysql.Permissions
	var user mysql.User
	user.ID = userid
	//result = env.db.Model(&user).Preload(clause.Associations).Find(&user)

	//err = env.db.Model(&mysql.User{}).Where("id = ?", userid).Association("Permissions").Find(&perm)
	result = env.db.Model(&user).Preload(clause.Associations).Find(&user)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	result = env.db.Model(&user.Permissions).Preload(clause.Associations).Find(&user.Permissions)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	for _, speaker := range user.Permissions.Speakers {
		user.Permissions.SpeakerIDs = append(user.Permissions.SpeakerIDs, speaker.ID)
	}

	for _, room := range user.Permissions.Rooms {
		user.Permissions.RoomIDs = append(user.Permissions.RoomIDs, room.ID)
	}

	for _, speakergroup := range user.Permissions.SpeakerGroups {
		user.Permissions.SpeakerGroupIDs = append(user.Permissions.SpeakerGroupIDs, speakergroup.ID)
	}

	c.JSON(http.StatusOK, getPermsResponse{Perms: user.Permissions})
}

//This function handles PATCH requests sent to the /api/v1/user/:id/permissions endpoint
func (env *Env) UpdateUserPermissions(c *gin.Context) {

	type reqUpdtPrems struct {
		UserID int32    `json:"user_id"`
		Perms  updtPerm `json:"perms"`
	}

	//Convert ID Parameter into int32
	var user mysql.User
	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}

	user.ID = int32(tmp)

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request reqUpdtPrems
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	if user.ID != request.UserID {
		var rights int64
		result := env.db.Model(&mysql.Permissions{}).Where("(permissions.id = (select perm_id from users where users.id = ?) " +
			"or permissions.id in (select perm_id from user_groups where user_groups.id " +
			"in (select user_group_id from user_usergroups where user_id = ?))) " +
			"and admin;").Count(&rights)
		if result.Error != nil {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}

		if rights == 0 {
			Log.WithField("module", "handler").Error("User not Authorized for this Action")
			c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH009)
			return
		}
	}

	result := env.db.Preload(clause.Associations).Find(&user)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if error := env.updatePermissions(&user.Permissions, request.Perms); error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, error)
		return
	}

	c.JSON(http.StatusOK, user.Permissions)

}

func (env *Env) updatePermissions(orgPerms *mysql.Permissions, updatePerms updtPerm) interface{} {

	result := env.db.Preload(clause.Associations).Find(&orgPerms)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		return errs.DBSQ001
	}

	if updatePerms.Admin.Valid {
		orgPerms.Admin = updatePerms.Admin.Bool
	}
	if updatePerms.CanEdit.Valid {
		orgPerms.CanEdit = updatePerms.CanEdit.Bool
	}
	if updatePerms.RoomIDs != nil {
		if len(updatePerms.RoomIDs) == 0 {
			err := env.db.Model(&orgPerms).Association("Rooms").Clear()
			if err != nil {
				Log.WithField("module", "sql").WithError(err)
				return errs.DBSQ001

			}
		} else {
			result = env.db.Find(&orgPerms.Rooms, updatePerms.RoomIDs)
			if result.Error != nil {
				Log.WithField("module", "sql").WithError(result.Error)
				return errs.DBSQ001
			}
		}
	}
	if updatePerms.SpeakerIDs != nil {
		if len(updatePerms.SpeakerIDs) == 0 {
			err := env.db.Model(&orgPerms).Association("Speakers").Clear()
			if err != nil {
				Log.WithField("module", "sql").WithError(err)
				return errs.DBSQ001
			}
		} else {
			result = env.db.Find(&orgPerms.Speakers, updatePerms.SpeakerIDs)
			if result.Error != nil {
				Log.WithField("module", "sql").WithError(result.Error)
				return errs.DBSQ001
			}
		}
	}
	if updatePerms.SpeakerGroupIDs != nil {
		if len(updatePerms.SpeakerGroupIDs) == 0 {
			err := env.db.Model(&orgPerms).Association("SpeakerGroups").Clear()
			if err != nil {
				Log.WithField("module", "sql").WithError(err)
				return errs.DBSQ001
			}
		} else {
			result = env.db.Find(&orgPerms.SpeakerGroups, updatePerms.SpeakerGroupIDs)
			if result.Error != nil {
				Log.WithField("module", "sql").WithError(result.Error)
				return errs.DBSQ001
			}
		}
	}

	for _, speaker := range orgPerms.Speakers {
		orgPerms.SpeakerIDs = append(orgPerms.SpeakerIDs, speaker.ID)
	}

	for _, room := range orgPerms.Rooms {
		orgPerms.RoomIDs = append(orgPerms.RoomIDs, room.ID)
	}

	for _, speakergroup := range orgPerms.SpeakerGroups {
		orgPerms.SpeakerGroupIDs = append(orgPerms.SpeakerGroupIDs, speakergroup.ID)
	}

	result = env.db.Save(&orgPerms)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		return errs.DBSQ001
	}

	return nil
}
