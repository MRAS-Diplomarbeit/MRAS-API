package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/core/config"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
	errs "github.com/mras-diplomarbeit/mras-api/core/error"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"github.com/mras-diplomarbeit/mras-api/core/utils"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"strconv"
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
	result := db.Con.Where("(speakers.id in (select speaker_id from perm_speakers where permissions_id ="+
		"(select perm_id from users where users.id = ?)) or "+
		"speakers.id = any (select speaker_id from perm_speakers where permissions_id = any"+
		"(select perm_id from user_groups where user_groups.id = any"+
		"(select user_group_id from user_usergroups where user_id = ?))) or"+
		"(select admin from permissions where id = "+
		"(select perm_id from users where users.id = ?)) = true or"+
		"(select admin from permissions where permissions.id = any"+
		"(select perm_id from user_groups where user_groups.id = any"+
		"(select user_group_id from user_usergroups where user_id = ?))) = true)"+
		"and speakers.alive = true", userid, userid, userid, userid).Find(&speakers)

	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, resAllSpeakers{Count: len(speakers), Speakers: speakers})

}

func UpdateSpeaker(c *gin.Context) {

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

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

	result := db.Con.Where("((speakers.id in (select speaker_id from perm_speakers"+
		"where permissions_id = (select perm_id from users where users.id = ?))) or"+
		"(speakers.id in (select speaker_id from perm_speakers where permissions_id = any"+
		"(select perm_id from user_groups where user_groups.id = any"+
		"(select user_group_id from user_usergroups where user_id = ?)))) or"+
		"((select admin from permissions where id = (select perm_id from users where users.id = ?)) = true) or"+
		"((select admin from permissions where permissions.id = any "+
		"(select perm_id from user_groups where user_groups.id = any"+
		"(select user_group_id from user_usergroups where user_id = ?))) = true))"+
		"and speakers.id = ?", reqUserId, reqUserId, reqUserId, reqUserId, updtSpeaker.ID.Int64).Count(&rights)

	if rights == 0 {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH009)
		return
	}

	var ogSpeaker mysql.Speaker
	ogSpeaker.ID = int32(updtSpeaker.ID.Int64)

	result = db.Con.Find(&ogSpeaker)
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

	result = db.Con.Save(&ogSpeaker)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ007)
		return
	}

	c.JSON(http.StatusOK, ogSpeaker)
}

func GetSpeaker(c *gin.Context) {

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	var speaker mysql.Speaker

	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	speaker.ID = int32(tmp)

	result := db.Con.Find(&speaker)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	c.JSON(http.StatusOK, speaker)
}

func EnablePlaybackSpeaker(c *gin.Context) {

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	type playbackClientReq struct {
		Method      string   `json:"method"`
		DisplayName string   `json:"displayname"`
		DeviceIPs   []string `json:"device_ips"`
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
	result := db.Con.Where("id = ?", c.Param("id")).First(&speaker)
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

	clientreq := playbackClientReq{Method: request.Method, DisplayName: request.DisplayName, DeviceIPs: []string{}}

	res, err := utils.DispatchRequest("http://"+speaker.IPAddress+":"+strconv.Itoa(config.ClientBackendPort)+config.ClientBackendPath, "application/json", "POST", clientreq)
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

		result = db.Con.Save(&session)
		if result.Error != nil {

			type stopPlayback struct {
				IPs []string `json:"ips"`
			}

			res, err := utils.DispatchRequest("http://"+speaker.IPAddress+":"+strconv.Itoa(config.ClientBackendPort)+config.ClientBackendPath, "application/json", "DELETE", stopPlayback{})
			if err != nil {
				Log.WithField("module", "client").WithError(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE002)
				return
			}

			res.Body.Close()
		} else {
			return
		}
		c.JSON(http.StatusInternalServerError, errs.CLIE003)
	}

	var response playbackClientRes
	json.NewDecoder(res.Body).Decode(&response)
	Log.WithField("module", "handler").Debug(response)

	if res.StatusCode == 404 {
		for _, ip := range response.DeadIps {
			result = db.Con.Model(&mysql.Speaker{}).Where("ip_address = ?", ip).Update("alive", false)
			if result.Error != nil {
				Log.WithField("module", "sql").WithError(result.Error)
				c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
				return
			}
		}
	}
	c.JSON(res.StatusCode, errs.Error{Code: strconv.Itoa(response.Code), Message: response.Message})
}

func StopPlaybackSpeaker(c *gin.Context) {

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

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

	result := db.Con.Model(&session).Where("speaker_id = ? or id = (select sessions_id from session_speakers where session_speakers.speaker_id = ?)", speakerid, speakerid).Find(&session)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	err = db.Con.Model(&mysql.Speaker{}).Association("Speakers").Find(&session.Speaker)
	if err != nil {
		Log.WithField("module", "sql").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	err = db.Con.Model(&mysql.Speaker{}).Association("Speakers").Find(&session.Speakers)
	if err != nil {
		Log.WithField("module", "sql").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	for _, speaker := range session.Speakers {
		session.SpeakerIPs = append(session.SpeakerIPs, speaker.IPAddress)
	}

	stopPlaybackReqBody := stopPlaybackReq{session.SpeakerIPs}

	res, err := utils.DispatchRequest("http://"+session.Speaker.IPAddress+":"+strconv.Itoa(config.ClientBackendPort)+config.ClientBackendPath, "application/json", "DELETE", stopPlaybackReqBody)
	if err != nil {
		Log.WithField("module", "client").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE002)
		return
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		result = db.Con.Delete(&session)
		if result.Error != nil {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}
	}

}

func RemoveSpeaker(c *gin.Context) {

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	var speaker mysql.Speaker

	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	speaker.ID = int32(tmp)

	//err = db.Con.Model(&mysql.Permissions{}).Association("Speakers").Delete(&speaker)
	//if err != nil {
	//	Log.WithField("module", "sql").WithError(err)
	//	c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
	//	return
	//}

	result := db.Con.Delete(&speaker)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}
}
