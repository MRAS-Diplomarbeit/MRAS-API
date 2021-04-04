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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (env *Env) GetAllSpeakers(c *gin.Context) {

	type resAllSpeakers struct {
		Count    int             `json:"count"`
		Speakers []mysql.Speaker `json:"speakers"`
	}

	var speakers []mysql.Speaker
	userid, _ := c.Get("userid")

	//Get all Speakers from Database
	result := env.db.Where("speakers.id in (select speaker_id from speaker_user_perms where user_id = @userid) and speakers.alive = true", sql.Named("userid", userid)).Find(&speakers)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, resAllSpeakers{Count: len(speakers), Speakers: speakers})

}

func (env *Env) UpdateSpeaker(c *gin.Context) {

	type updateSpeaker struct {
		ID          null.Int       `json:"id"`
		Name        null.String    `json:"name"`
		Description null.String    `json:"description"`
		Position    mysql.Position `json:"position"`
		RoomID      null.Int       `json:"room_id"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var updtSpeaker updateSpeaker
	err = json.Unmarshal(jsonData, &updtSpeaker)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	if !updtSpeaker.ID.Valid {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST002)
		return
	}

	reqUserId, _ := c.Get("userid")

	var rights int64
	result := env.db.Model(&mysql.Speaker{}).Where("speakers.id = @speakerid and (speakers.id in (select speaker_id from speaker_user_perms where user_id = @userid))",
		sql.Named("userid", reqUserId), sql.Named("speakerid", updtSpeaker.ID.Int64)).Count(&rights)
	if rights == 0 {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH009)
		return
	}

	var ogSpeaker mysql.Speaker
	ogSpeaker.ID = int32(updtSpeaker.ID.Int64)

	result = env.db.Find(&ogSpeaker)
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

	if updtSpeaker.Name.Valid {
		ogSpeaker.Name = updtSpeaker.Name.String
	}
	if updtSpeaker.Description.Valid {
		ogSpeaker.Description = updtSpeaker.Description.String
	}
	if updtSpeaker.Position.PosY.Valid {
		ogSpeaker.Position.PosY = updtSpeaker.Position.PosY
	}
	if updtSpeaker.Position.PosX.Valid {
		ogSpeaker.Position.PosX = updtSpeaker.Position.PosX
	}
	if updtSpeaker.RoomID.Valid {
		ogSpeaker.RoomID = updtSpeaker.RoomID
	}

	result = env.db.Save(&ogSpeaker)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ007)
		return
	}

	c.JSON(http.StatusOK, ogSpeaker)
}

func (env *Env) GetSpeaker(c *gin.Context) {

	var speaker mysql.Speaker

	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	speaker.ID = int32(tmp)

	result := env.db.Find(&speaker)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, speaker)
}

func (env *Env) RemoveSpeaker(c *gin.Context) {

	var speaker mysql.Speaker

	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	speaker.ID = int32(tmp)

	//err = db.Model(&mysql.Permissions{}).Association("Speakers").Delete(&speaker)
	//if err != nil {
	//	Log.WithField("module", "sql").WithError(err)
	//	c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
	//	return
	//}

	result := env.db.Delete(&speaker)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}
}

func (env *Env) EnablePlaybackSpeaker(c *gin.Context) {

	type playbackClientReq struct {
		Method      string   `json:"method"`
		DisplayName string   `json:"displayname"`
		DeviceIPs   []string `json:"device_ips"`
		MulticastIP string   `json:"multicast_ip"`
	}

	type playbackClientRes struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		DeadIps []string `json:"dead_ips"`
	}

	type playbackReq struct {
		DisplayName string `json:"displayname"`
		Method      string `json:"method"`
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

	var speaker mysql.Speaker
	result := env.db.Where("id = ?", c.Param("id")).First(&speaker)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if !speaker.Alive {
		Log.WithField("module", "handler").Warnf("Speaker ID:%d inactive", speaker.ID)
		c.AbortWithStatusJSON(http.StatusNotFound, errs.CLIE001)
		return
	}

	clientreq := playbackClientReq{Method: request.Method, DisplayName: request.DisplayName, DeviceIPs: []string{}, MulticastIP: ""}

	res, err := utils.PostRequest(config.ClientBackendProtocol+"://"+speaker.IPAddress+":"+strconv.Itoa(config.ClientBackendPort)+config.ClientBackendPlaybackPath, "application/json", clientreq)
	if err != nil {
		Log.WithField("module", "client").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE002)
		return
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {

		var session mysql.Sessions

		session.Speaker = speaker
		session.DisplayName = request.DisplayName
		session.Method = request.Method
		session.Speakers = append(session.Speakers, &speaker)

		result = env.db.Save(&session)
		if result.Error != nil {

			type stopPlayback struct {
				IPs []string `json:"ips"`
			}

			stopreq := stopPlayback{}
			stopreq.IPs = append(stopreq.IPs, speaker.IPAddress)

			res, err := utils.DeleteRequest(config.ClientBackendProtocol+"://"+speaker.IPAddress+":"+strconv.Itoa(config.ClientBackendPort)+config.ClientBackendPlaybackPath, "application/json", stopreq)
			if err != nil {
				Log.WithField("module", "client").WithError(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE002)
				return
			}

			res.Body.Close()
		} else {
			result = env.db.Model(&speaker).Update("active", true)
			if result.Error != nil {
				Log.WithField("module", "sql").WithError(result.Error)
				c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
				return
			}
			return
		}
		c.JSON(http.StatusInternalServerError, errs.CLIE003)
	}

	var response playbackClientRes
	json.NewDecoder(res.Body).Decode(&response)
	Log.WithField("module", "handler").Debug(response)

	if res.StatusCode == 404 {
		for _, ip := range response.DeadIps {
			result = env.db.Model(&mysql.Speaker{}).Where("ip_address = ?", ip).Update("alive", false)
			if result.Error != nil {
				Log.WithField("module", "sql").WithError(result.Error)
				c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
				return
			}
		}
	}
	c.JSON(res.StatusCode, errs.Error{Code: strconv.Itoa(response.Code), Message: response.Message})
}

func (env *Env) StopPlaybackSpeaker(c *gin.Context) {

	type stopPlaybackReq struct {
		IPs []string `json:"ips"`
	}

	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}

	speakerid := int32(tmp)

	var session mysql.Sessions

	result := env.db.Model(&session).Where("speaker_id = @speakerid or id = (select sessions_id from session_speakers where session_speakers.speaker_id = @speakerid)", sql.Named("speakerid", speakerid)).Preload(clause.Associations).Find(&session)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	for _, speaker := range session.Speakers {
		session.SpeakerIPs = append(session.SpeakerIPs, speaker.IPAddress)
	}

	stopPlaybackReqBody := stopPlaybackReq{session.SpeakerIPs}

	res, err := utils.DeleteRequest(config.ClientBackendProtocol+"://"+session.Speaker.IPAddress+":"+strconv.Itoa(config.ClientBackendPort)+"/api/v1/playback", "application/json", stopPlaybackReqBody)
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

		result = env.db.Model(&session.Speaker).Update("active", false)
		if result.Error != nil {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}
	}

}

func (env *Env) ActiveSpeaker(c *gin.Context) {

	type activeRes struct {
		Active string `json:"active"`
	}

	var speaker mysql.Speaker

	result := env.db.Where("id = ?", c.Param("id")).Find(&speaker)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if speaker.Active {
		c.JSON(http.StatusOK, activeRes{Active: "active"})
		return
	}

	var exists int64

	result = env.db.Model(&mysql.SpeakerGroup{}).Where("(active = true) and id in (select speaker_group_id from speakergroup_speakers where speaker_id = ?)", speaker.ID).Count(&exists)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}
	if exists > 0 {
		c.JSON(http.StatusOK, activeRes{Active: "inuse"})
		return
	}

	c.JSON(http.StatusOK, activeRes{Active: "inactive"})
}

func (env *Env) GetSpeakerPlaybackMethods(c *gin.Context) {

	type speakerMethodsRes struct {
		Methods []string `json:"methods"`
	}

	var speaker mysql.Speaker

	result := env.db.Where("id = ?", c.Param("id")).Find(&speaker)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	res, err := utils.GetRequest(config.ClientBackendProtocol + "://" + speaker.IPAddress + ":" + strconv.Itoa(config.ClientBackendPort) + config.ClientBackendMethodPath)
	if err != nil {
		Log.WithField("module", "client").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE002)
		return
	}
	defer res.Body.Close()

	jsonData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE004)
		return
	}

	if res.StatusCode == 200 {

		var playbackMethods speakerMethodsRes

		err = json.Unmarshal(jsonData, &playbackMethods)
		if err != nil {
			Log.WithField("module", "handler").WithError(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE004)
			return
		}

		c.JSON(http.StatusOK, &playbackMethods)
	} else {

		var clientError errs.Error

		err = json.Unmarshal(jsonData, &clientError)
		if err != nil {
			Log.WithField("module", "handler").WithError(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE004)
			return
		}

		c.JSON(http.StatusInternalServerError, clientError)

	}

}

func (env *Env) SetSpeakerPlaybackMethod(c *gin.Context) {

	type setMethodReq struct {
		Method string `json:"method"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var method setMethodReq
	err = json.Unmarshal(jsonData, &method)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var speaker mysql.Speaker

	result := env.db.Where("id = ?", c.Param("id")).Find(&speaker)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	res, err := utils.PutRequest(config.ClientBackendProtocol+"://"+speaker.IPAddress+":"+strconv.Itoa(config.ClientBackendPort)+config.ClientBackendMethodPath, "application/json", method)
	if err != nil {
		Log.WithField("module", "client").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE002)
		return
	}
	defer res.Body.Close()

	clientJsonData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE004)
		return
	}

	if res.StatusCode != 200 {

		var clientError errs.Error

		err = json.Unmarshal(clientJsonData, &clientError)
		if err != nil {
			Log.WithField("module", "handler").WithError(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE004)
			return
		}

		c.JSON(http.StatusInternalServerError, clientError)
	}
}
