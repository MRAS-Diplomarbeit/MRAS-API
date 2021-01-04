package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/config"
	"github.com/mras-diplomarbeit/mras-api/db/mysql"
	"github.com/mras-diplomarbeit/mras-api/db/redis"
	. "github.com/mras-diplomarbeit/mras-api/logger"
	"github.com/mras-diplomarbeit/mras-api/utils"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"time"
)

var rdis *redis.RedisServices
var db *mysql.SqlServices

func GenerateAccessToken(c *gin.Context) {

	//Check if redis database connection is already established and create one if not
	if rdis == nil {
		connectRedis()
	}

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	type refreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	type refreshResponse struct {
		AccessToken string `json:"access_token"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		if err != nil {
			Log.WithField("module", "handler").WithError(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, config.Error{Code: "RQST001", Message: "Error decoding RequestBody" + fmt.Sprintf(err.Error())})
			return
		}
	}

	var request refreshRequest
	json.Unmarshal(jsonData, &request)

	user := mysql.User{}
	user.RefreshToken = request.RefreshToken

	token, _ := utils.JWTAuthService(config.JWTAccessSecret).ValidateToken(user.RefreshToken)
	if !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH005", Message: "Invalid RefreshToken"})
		return
	}
	claims := token.Claims.(jwt.MapClaims)
	user.ID = int32(claims["userid"].(float64))
	Log.Debug(claims["userid"])

	var exists int64

	//Check if Username already exists in Database
	result := db.Con.Model(&user).Where("id = ? and refresh_token = ?", user.ID, request.RefreshToken).Count(&exists)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	if exists == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH005", Message: "Invalid RefreshToken"})
		return
	}

	Log.WithField("model", "handler").Debug(user)

	//Generate JWT AccessToken
	accessToken, err := utils.JWTAuthService(config.JWTAccessSecret).GenerateToken(user.ID, claims["deviceid"].(string), time.Hour*24)
	if err != nil {
		Log.WithField("module", "jwt").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "AUTH002", Message: "Error Generating JWT Token " + fmt.Sprintf(err.Error())})
		return
	}

	//Add AccessToken to Redis
	err = rdis.AddPair(accessToken, claims["deviceid"].(string), time.Hour*24)
	if err != nil {
		Log.WithField("module", "redis").WithError(err).Error("Error adding AccessToken to Redis.")
		err = nil
	}

	c.JSON(http.StatusOK, refreshResponse{AccessToken: accessToken})

}

//This function handles POST requests sent to the /api/v1/user/register endpoint
func RegisterUser(c *gin.Context) {

	//Check if redis database connection is already established and create one if not
	if rdis == nil {
		connectRedis()
	}

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	type registerRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		DeviceID string `json:"device_id"`
	}

	type registerResponse struct {
		AccessToken  string     `json:"access_token""`
		RefreshToken string     `json:"refresh_token""`
		User         mysql.User `json:"user"`
		ResetCode    string     `json:"reset_code"`
	}

	//decode request body
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
	user.Username = request.Username
	var exists int64

	//Check if Username already exists in Database
	result := db.Con.Model(&user).Where("username = ?", user.Username).Count(&exists)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}
	Log.WithField("module", "handler").Debug("Users found: ", exists)

	if exists != 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, config.Error{Code: "AUTH004", Message: "User already exists"})
		return
	}

	perms := mysql.Permissions{}

	//Create permission entry for new user in permissions table
	result = db.Con.Save(&perms)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	user.Username = request.Username
	user.Password = request.Password
	user.AvatarID = "default"
	user.PermID = perms.ID

	Log.WithField("model", "handler").Debug(user)

	//for _, group := range user.UserGroups {
	//	user.UserGroupIDs = append(user.UserGroupIDs, group.ID)
	//}

	//Save new user to users database
	result = db.Con.Save(&user)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	//Generate JWT AccessToken
	accessToken, err := utils.JWTAuthService(config.JWTAccessSecret).GenerateToken(user.ID, request.DeviceID, time.Hour*24)
	if err != nil {
		Log.WithField("module", "jwt").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "AUTH002", Message: "Error Generating JWT Token " + fmt.Sprintf(err.Error())})
		return
	}

	//Add AccessToken to Redis
	err = rdis.AddPair(accessToken, request.DeviceID, time.Hour*24)
	if err != nil {
		Log.WithField("module", "redis").WithError(err).Error("Error adding AccessToken to Redis.")
		err = nil
	}

	//Generate RefreshToken
	refreshToken, err := utils.JWTAuthService(config.JWTAccessSecret).GenerateToken(user.ID, request.DeviceID, time.Hour*24)
	if err != nil {
		Log.WithField("module", "jwt").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "AUTH002", Message: "Error Generating JWT Token " + fmt.Sprintf(err.Error())})
		return
	}

	user.RefreshToken = refreshToken

	//Save RefreshToken to Database
	result = db.Con.Save(&user)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ002", Message: "Error Saving RefreshToken in Database"})
		return
	}

	c.JSON(200, registerResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: user, ResetCode: utils.GenerateCode()})
}

//This function handles POST requests sent to the /api/v1/user/login endpoint
func LoginUser(c *gin.Context) {

	//Check if redis database connection is already established and create one if not
	if rdis == nil {
		connectRedis()
	}

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	type loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		DeviceID string `json:"device_id"`
	}

	type loginResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	//decode request body
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

	//lookup user in users database
	result := db.Con.Where("username = ? AND password = ?", request.Username, request.Password).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH003", Message: "User not found in Database (Wrong Username or Password)"})
			return
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
			return
		}
	}

	//Generate JWT AccessToken
	accessToken, err := utils.JWTAuthService(config.JWTAccessSecret).GenerateToken(user.ID, request.DeviceID, time.Hour*24)
	if err != nil {
		Log.WithField("module", "jwt").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "AUTH002", Message: "Error Generating JWT Token " + fmt.Sprintf(err.Error())})
		return
	}

	//Add AccessToken to Redis
	err = rdis.AddPair(accessToken, request.DeviceID, time.Hour*24)
	if err != nil {
		Log.WithField("module", "redis").WithError(err).Error("Error adding AccessToken to Redis.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ003", Message: "Error Accessing Redis"})
		return
	}

	//Generate JWT RefreshToken
	refreshToken, err := utils.JWTAuthService(config.JWTRefreshSecret).GenerateToken(user.ID, request.DeviceID, time.Hour*4380)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "AUTH002", Message: "Error Generating JWT Token " + fmt.Sprintf(err.Error())})
		return
	}

	//Save RefreshToken to Database
	user.RefreshToken = refreshToken
	result = db.Con.Save(&user)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ002", Message: "Error Saving RefreshToken in Database"})
		return
	}

	c.JSON(200, loginResponse{accessToken, refreshToken})
}

//Connect to redis database
func connectRedis() {
	var err error
	rdis, err = redis.RedisDBService().Initialize(config.Redis)
	if err != nil {
		Log.WithField("module", "redis").WithError(err)
	}

}

//create connections to mysql database
func connectMySql() {
	db = mysql.GormService().Connect(config.MySQL)
}
