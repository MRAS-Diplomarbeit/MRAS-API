package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/core/config"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
	errs "github.com/mras-diplomarbeit/mras-api/core/error"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"github.com/mras-diplomarbeit/mras-api/core/utils"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm/clause"
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

	type reqUpdtRoom struct {
		ID          int32            `json:"id"`
		Name        null.String      `json:"name"`
		Description null.String      `json:"description"`
		Dimensions  mysql.Dimensions `json:"dimensions"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request reqUpdtRoom
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

	var orgRoom mysql.Room
	orgRoom.ID = roomid
	result := env.db.Model(&orgRoom).Find(&orgRoom)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.DBSQ001)
		return
	}

	if request.Name.Valid {
		orgRoom.Name = request.Name.String
	}
	if request.Description.Valid {
		orgRoom.Description = request.Description.String
	}
	orgRoom.Dimensions = request.Dimensions

	result = env.db.Omit("created_at").Omit("active").Save(&orgRoom)
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

	type playbackReq struct {
		DisplayName string `json:"displayname"`
		Method      string `json:"method"`
	}

	type playbackClientReq struct {
		Method      string   `json:"method"`
		DisplayName string   `json:"displayname"`
		DeviceIPs   []string `json:"device_ips"`
		MulticastIP string   `json:"multicast_ip"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request playbackReq
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var room mysql.Room
	var speakers []*mysql.Speaker

	result := env.db.Where("id = ?", c.Param("id")).Find(&room)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	result = env.db.Where("room_id = ?", room.ID).Find(&speakers)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	var clientReq playbackClientReq

	clientReq.Method = request.Method
	clientReq.DisplayName = request.DisplayName
	clientReq.MulticastIP = utils.GenerateMulticastIP()

	for _, speaker := range speakers {
		if speaker.Alive {
			clientReq.DeviceIPs = append(clientReq.DeviceIPs, speaker.IPAddress)
		}
	}

	res, err := utils.PostRequest("http://"+clientReq.DeviceIPs[0]+":"+strconv.Itoa(config.ClientBackendPort)+config.ClientBackendPath, "application/json", clientReq)
	if err != nil {
		Log.WithField("module", "client").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE002)
		return
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {

		var session mysql.Sessions

		session.Speaker = *speakers[0]
		session.DisplayName = request.DisplayName
		session.Method = request.Method
		session.MulticastIP = clientReq.MulticastIP
		session.Speakers = speakers

		result = env.db.Save(&session)
		if result.Error != nil {

			type stopPlayback struct {
				IPs []string `json:"ips"`
			}

			res2, err := utils.DeleteRequest("http://"+clientReq.DeviceIPs[0]+":"+strconv.Itoa(config.ClientBackendPort)+config.ClientBackendPath, "application/json", stopPlayback{})
			if err != nil {
				Log.WithField("module", "client").WithError(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE002)
				return
			}

			_ = res2.Body.Close()
		} else {
			result = env.db.Model(&room).Update("active", true)
			if result.Error != nil {
				Log.WithField("module", "sql").WithError(result.Error)
				c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
				return
			}
			return
		}
		c.JSON(http.StatusInternalServerError, errs.CLIE003)

	}
}

func (env *Env) StopPlaybackRoom(c *gin.Context) {

	type stopPlaybackReq struct {
		IPs []string `json:"ips"`
	}

	var room mysql.Room
	var speakers []*mysql.Speaker

	result := env.db.Where("id = ?", c.Param("id")).Find(&room)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	result = env.db.Where("room_id = ?", room.ID).Find(&speakers)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	var session mysql.Sessions

	var speakerIds []int32

	for _, speaker := range speakers {
		speakerIds = append(speakerIds, speaker.ID)
	}

	result = env.db.Where("id = (select sessions_id from session_speakers where speaker_id in ?)", speakerIds).Preload(clause.Associations).Find(&session)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	for _, speaker := range session.Speakers {
		session.SpeakerIPs = append(session.SpeakerIPs, speaker.IPAddress)
	}

	Log.Debug(session)

	stopPlaybackReqBody := stopPlaybackReq{session.SpeakerIPs}

	res, err := utils.DeleteRequest("http://"+session.Speaker.IPAddress+":"+strconv.Itoa(config.ClientBackendPort)+config.ClientBackendPath,
		"application/json", stopPlaybackReqBody)
	if err != nil {
		Log.WithField("module", "client").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE002)
		return
	}
	defer res.Body.Close()

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	Log.Debug(jsonData)

	if res.StatusCode == 200 {

		err = env.db.Model(&session).Association("Speakers").Clear()
		if err != nil {
			Log.WithField("module", "sql").WithError(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}

		result = env.db.Delete(&session)
		if result.Error != nil {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}

		result = env.db.Model(&room).Update("active", false)
		if result.Error != nil {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}

	}
}

func (env *Env) ActiveRoom(c *gin.Context) {

	type activeRes struct {
		Active string `json:"active"`
	}

	var room mysql.Room

	result := env.db.Where("id = ?", c.Param("id")).Find(&room)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if room.Active {
		c.JSON(http.StatusOK, activeRes{Active: "active"})
		return
	}

	var active int64

	result = env.db.Model(&mysql.Speaker{}).Where("active = true and room_id = ?", c.Param("id")).Count(&active)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if active > 0 {
		c.JSON(http.StatusOK, activeRes{Active: "inuse" +
			""})
		return
	}

	c.JSON(http.StatusOK, activeRes{Active: "inactive"})
}
