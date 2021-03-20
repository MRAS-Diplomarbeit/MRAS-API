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
	"strconv"
)

func (env *Env) GetAllRooms(c *gin.Context) {

	type resAllRooms struct {
		Count int          `json:"count"`
		Rooms []mysql.Room `json:"rooms"`
	}

	var rooms []mysql.Room
	userid, _ := c.Get("userid")

	//Get all Rooms from Database
	result := env.db.Where("rooms.id in (select room_id from room_user_perms where user_id = @userid)",
		sql.Named("userid", userid)).Find(&rooms)

	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, resAllRooms{Count: len(rooms), Rooms: rooms})
}

func (env *Env) CreateRoom(c *gin.Context) {

	type reqCreateRoom struct {
		Name        string           `json:"name"`
		Description string           `json:"description"`
		Dimensions  mysql.Dimensions `json:"dimensions"`
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

	dimensions := request.Dimensions
	room := &mysql.Room{Name: request.Name, Description: request.Description}
	room.Dimensions = dimensions

	result := env.db.Create(&room)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, &room)
}

func (env *Env) UpdateRoom(c *gin.Context) {

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request mysql.Room
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	roomid := int32(tmp)

	if roomid != request.ID {
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	result := env.db.Save(&request)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, &request)

}

func (env *Env) GetRoom(c *gin.Context) {

	var room mysql.Room

	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	room.ID = int32(tmp)

	result := env.db.Find(&room)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, &room)

}

func (env *Env) DeleteRoom(c *gin.Context) {

	var room mysql.Room

	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	room.ID = int32(tmp)

	result := env.db.Delete(&room)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.DBSQ001)
		return
	}

}

func (env *Env) EnablePlaybackRoom(c *gin.Context) {
	//
	//type playbackReq struct {
	//	DisplayName string `json:"displayname"`
	//	Method      string `json:"method"`
	//}
	//
	////decode request body
	//jsonData, err := ioutil.ReadAll(c.Request.Body)
	//if err != nil {
	//	Log.WithField("module", "handler").WithError(err)
	//	c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
	//	return
	//}
	//
	//var request playbackReq
	//err = json.Unmarshal(jsonData, &request)
	//if err != nil {
	//	Log.WithField("module", "handler").WithError(err)
	//	c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
	//	return
	//}
	//
	//var room mysql.Room
	//var speaker

}

func (env *Env) StopPlaybackRoom(c *gin.Context) {

}
