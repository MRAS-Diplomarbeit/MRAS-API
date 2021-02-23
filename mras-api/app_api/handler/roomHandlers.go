package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
	errs "github.com/mras-diplomarbeit/mras-api/core/error"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"io/ioutil"
	"net/http"
)

func (env *Env) GetAllRooms(c *gin.Context) {

	type resAllRooms struct {
		Count int             `json:"count"`
		Rooms []mysql.Room `json:"rooms"`
	}

	var rooms []mysql.Room
	userid, _ := c.Get("userid")

	//Get all Rooms from Database
	result := env.db.Where("(rooms.id in (select room_id from perm_rooms where permissions_id ="+
		"(select perm_id from users where users.id = @userid)) or "+
		"rooms.id = any (select room_id from perm_rooms where permissions_id = any"+
		"(select perm_id from user_groups where user_groups.id = any"+
		"(select user_group_id from user_usergroups where user_id = @userid))) or"+
		"(select admin from permissions where id = "+
		"(select perm_id from users where users.id = @userid)) = true or"+
		"(select admin from permissions where permissions.id = any"+
		"(select perm_id from user_groups where user_groups.id = any"+
		"(select user_group_id from user_usergroups where user_id = @userid))) = true)",
		sql.Named("userid",userid)).Find(&rooms)

	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, resAllRooms{Count: len(rooms), Rooms: rooms})
}

func (env *Env) CreateRoom(c *gin.Context) {

	type reqCreateRoom struct{
		Name string `json:"name"`
		Description string `json:"description"`
		Dimensions mysql.Dimensions `json:"dimensions"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request reqCreateRoom
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	result := env.db.Save(&mysql.Room{Name: request.Name,Description: request.Description,Dimensions: request.Dimensions})
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.DBSQ001)
		return
	}
}

func (env *Env) UpdateRoom(c *gin.Context) {

}

func (env *Env) GetRoom(c *gin.Context) {

}

func (env *Env) DeleteRoom(c *gin.Context) {

}

func (env *Env) EnablePlaybackRoom(c *gin.Context) {

}
