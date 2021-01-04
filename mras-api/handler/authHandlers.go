package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/config"
	"github.com/mras-diplomarbeit/mras-api/db/mysql"
	"github.com/mras-diplomarbeit/mras-api/db/redis"
	. "github.com/mras-diplomarbeit/mras-api/logger"
	"github.com/mras-diplomarbeit/mras-api/service"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"time"
)

var rdis *redis.RedisServices
var db *mysql.SqlServices

func GenerateAccessToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Working"))
}

func RegisterUser(c *gin.Context) {
	if rdis == nil {
		connectRedis()
	}
	if db == nil {
		connectMySql()
	}

	type registerRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		DeviceID string `json:"device_id"`
	}

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		if err != nil {
			Log.WithField("module", "handler").WithError(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, config.Error{Code: "RQST001", Message: "Error decoding RequestBody" + fmt.Sprintf(err.Error())})
			return
		}
	}

	var request registerRequest
	json.Unmarshal(jsonData, &request)

	user := mysql.User{}
	perms := mysql.Permissions{}

	result := db.Con.Save(&perms)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	user.Username = request.Username
	user.Password = request.Password
	user.PermID = perms.ID

	for _, group := range user.UserGroups {
		user.UserGroupIDs = append(user.UserGroupIDs, group.ID)
	}

	result = db.Con.Save(&user)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	c.JSON(200, user)

}

func LoginUser(c *gin.Context) {

	if rdis == nil {
		connectRedis()
	}
	if db == nil {
		connectMySql()
	}

	type loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		DeviceID string `json:"device_id"`
	}

	type response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		if err != nil {
			Log.WithField("module", "handler").WithError(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, config.Error{Code: "RQST001", Message: "Error decoding RequestBody" + fmt.Sprintf(err.Error())})
			return
		}
	}

	var request loginRequest
	json.Unmarshal(jsonData, &request)

	user := mysql.User{}

	result := db.Con.Where("username = ? AND password = ?", request.Username, request.Password).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH003", Message: "User not found in Database"})
			return
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
			return
		}
	}

	accessToken, err := service.JWTAuthService(config.JWTAccessSecret).GenerateToken(user.ID, request.DeviceID, time.Hour*24)
	if err != nil {
		Log.WithField("module", "jwt").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "AUTH002", Message: "Error Generating JWT Token " + fmt.Sprintf(err.Error())})
		return
	}

	err = rdis.AddPair(accessToken, request.DeviceID, time.Hour*24)
	if err != nil {
		Log.WithField("module", "redis").WithError(err).Error("Error adding AccessToken to Redis.")
		err = nil
	}

	refreshToken, err := service.JWTAuthService(config.JWTRefreshSecret).GenerateToken(user.ID, request.DeviceID, time.Hour*4380)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "AUTH002", Message: "Error Generating JWT Token " + fmt.Sprintf(err.Error())})
		return
	}

	user.RefreshToken = refreshToken
	result = db.Con.Save(&user)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ002", Message: "Error Saving RefreshToken in Database"})
		return
	}

	c.JSON(200, response{accessToken, refreshToken})
}

func connectRedis() {
	var err error
	rdis, err = redis.RedisDBService().Initialize(config.Redis)
	if err != nil {
		Log.WithField("module", "redis").WithError(err)
	}

}

func connectMySql() {
	db = mysql.GormService().Connect(config.MySQL)
}
