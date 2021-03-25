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
		for _, speaker := range groups[i].Speakers {
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

	if len(speakergroup.SpeakerIds) > 0 {
		result := env.db.Find(&speakergroup.Speakers, speakergroup.SpeakerIds)
		if result.Error != nil {
			Log.WithField("module", "sql").WithError(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}
	}

	result := env.db.Save(&speakergroup)
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

	result := env.db.Preload(clause.Associations).Find(&orgGroup)
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
			orgGroup.Speakers = nil
		} else {
			orgGroup.Speakers = nil
			orgGroup.SpeakerIds = request.SpeakerIds
			result = env.db.Find(&orgGroup.Speakers, orgGroup.SpeakerIds)
			if result.Error != nil {
				Log.WithField("module", "sql").WithError(result.Error)
				c.AbortWithStatusJSON(http.StatusBadRequest, errs.DBSQ001)
				return
			}
		}
	}

	if request.SpeakerIds != nil {
		err = env.db.Model(&orgGroup).Association("Speakers").Replace(&orgGroup.Speakers)
		if err != nil {
			Log.WithField("module", "sql").WithError(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}
	}

	result = env.db.Save(&orgGroup)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
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

	for _, speaker := range group.Speakers {
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

	var speakerGroup mysql.SpeakerGroup
	result := env.db.Where("id = ?", c.Param("id")).Preload(clause.Associations).First(&speakerGroup)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	var clientReq playbackClientReq

	clientReq.Method = request.Method
	clientReq.DisplayName = request.DisplayName
	clientReq.MulticastIP = utils.GenerateMulticastIP()

	for _, speaker := range speakerGroup.Speakers {
		if speaker.Alive {
			clientReq.DeviceIPs = append(clientReq.DeviceIPs, speaker.IPAddress)
		}
	}

	res, err := utils.PostRequest(config.ClientBackendProtocol+"://"+clientReq.DeviceIPs[0]+":"+strconv.Itoa(config.ClientBackendPort)+config.ClientBackendPlaybackPath, "application/json", clientReq)
	if err != nil {
		Log.WithField("module", "client").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE002)
		return
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {

		var session mysql.Sessions

		session.Speaker = *speakerGroup.Speakers[0]
		session.DisplayName = request.DisplayName
		session.Method = request.Method
		session.MulticastIP = clientReq.MulticastIP
		session.Speakers = speakerGroup.Speakers

		result = env.db.Save(&session)
		if result.Error != nil {

			type stopPlayback struct {
				IPs []string `json:"ips"`
			}

			res2, err := utils.DeleteRequest(config.ClientBackendProtocol+"://"+clientReq.DeviceIPs[0]+":"+strconv.Itoa(config.ClientBackendPort)+config.ClientBackendPlaybackPath, "application/json", stopPlayback{})
			if err != nil {
				Log.WithField("module", "client").WithError(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, errs.CLIE002)
				return
			}

			_ = res2.Body.Close()
		} else {
			result = env.db.Model(&speakerGroup).Update("active", true)
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
	_ = json.NewDecoder(res.Body).Decode(&response)
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

func (env *Env) StopPlaybackSpeakerGroup(c *gin.Context) {

	type stopPlaybackReq struct {
		IPs []string `json:"ips"`
	}

	var speakerGroup mysql.SpeakerGroup
	result := env.db.Where("id = ?", c.Param("id")).Preload(clause.Associations).First(&speakerGroup)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	var session mysql.Sessions

	var speakerIds []int32

	for _, speaker := range speakerGroup.Speakers {
		speakerIds = append(speakerIds, speaker.ID)
	}

	result = env.db.Where("id in (select sessions_id from session_speakers where speaker_id in ?)", speakerIds).Preload(clause.Associations).Find(&session)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	for _, speaker := range session.Speakers {
		session.SpeakerIPs = append(session.SpeakerIPs, speaker.IPAddress)
	}

	stopPlaybackReqBody := stopPlaybackReq{session.SpeakerIPs}

	res, err := utils.DeleteRequest(config.ClientBackendProtocol+"://"+session.Speaker.IPAddress+":"+strconv.Itoa(config.ClientBackendPort)+config.ClientBackendPlaybackPath, "application/json", stopPlaybackReqBody)
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

		result = env.db.Model(&speakerGroup).Update("active", false)
		if result.Error != nil {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}
	}
}

func (env *Env) ActiveSpeakerGroup(c *gin.Context) {

	type activeRes struct {
		Active string `json:"active"`
	}

	var speakerGroup mysql.SpeakerGroup

	result := env.db.Where("id = ?", c.Param("id")).Find(&speakerGroup)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if speakerGroup.Active {
		c.JSON(http.StatusOK, activeRes{Active: "active"})
		return
	}

	var active int64

	result = env.db.Model(&mysql.Speaker{}).Where("active = true and id in (select speaker_id from speakergroup_speakers where speaker_group_id = ?)", c.Param("id")).Count(&active)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if active > 0 {
		c.JSON(http.StatusOK, activeRes{Active: "inuse"})
		return
	}

	c.JSON(http.StatusOK, activeRes{Active: "inactive"})
}
