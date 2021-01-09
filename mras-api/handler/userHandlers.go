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
	"strconv"
	"strings"
	"time"
)

var rdis *redis.RedisServices
var db *mysql.SqlServices

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
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, config.Error{Code: "RQST001", Message: "Error decoding RequestBody" + fmt.Sprintf(err.Error())})
		return
	}

	var request loginRequest
	json.Unmarshal(jsonData, &request)

	if request.Username == "" || request.Password == "" || request.DeviceID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, config.Error{Code: "RQST002", Message: "Request Body missing fields"})
		return
	}

	user := mysql.User{}

	//lookup user in users database
	result := db.Con.Where("upper(username) = upper(?) AND password = ?", request.Username, request.Password).First(&user)
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
	err = rdis.AddPair(fmt.Sprint(user.ID), accessToken, time.Hour*24)
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

	if request.Username == "" || request.Password == "" || request.DeviceID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, config.Error{Code: "RQST002", Message: "Request Body missing fields"})
		return
	}

	user := mysql.User{}
	user.Username = request.Username
	var exists int64

	//Check if Username already exists in Database
	result := db.Con.Model(&user).Where("upper(username) = upper(?)", user.Username).Count(&exists)
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
	perms.Admin = true
	perms.CanEdit = true
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
	user.ResetCode = strings.ToLower(utils.GenerateCode())

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
	err = rdis.AddPair(fmt.Sprint(user.ID), accessToken, time.Hour*24)
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

	c.JSON(200, registerResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: user, ResetCode: user.ResetCode})
}

//This function handles POST requests sent to the /api/v1/user/refresh endpoint
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
	if token == nil || !token.Valid {
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

//This function handles POST requests sent to the /api/v1/user/password/reset/:username endpoint
func ResetUserPassword(c *gin.Context) {

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	type resetUserPasswordRequest struct {
		ResetCode string `json:"reset_code"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, config.Error{Code: "RQST001", Message: "Error decoding RequestBody" + fmt.Sprintf(err.Error())})
		return
	}

	var request resetUserPasswordRequest
	json.Unmarshal(jsonData, &request)

	var user mysql.User
	user.Username = c.Param("username")
	user.ResetCode = request.ResetCode

	var exists int64

	//Check if Username exists in Database
	result := db.Con.Model(&user).Where("upper(username) = upper(?)", user.Username).Count(&exists)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	if exists == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, config.Error{Code: "AUTH006", Message: "User not found"})
		return
	}

	//Check Database if ResetCode is correct
	result = db.Con.Where("upper(username) = upper(?) and reset_code = ?", user.Username, user.ResetCode).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH007", Message: "Wrong Reset Code"})
			return
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
			return
		}
	}

	//Reset Password in Database
	user.Password = "RESET"
	user.PasswordReset = true
	result = db.Con.Save(&user)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ004", Message: "Error Reseting User Password"})
		return
	}
}

//This function handles POST requests sent to the /api/v1/user/password/new/:username endpoint
func NewUserPassword(c *gin.Context) {

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	type newUserPasswordRequest struct {
		Password string `json:"password"`
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

	var request newUserPasswordRequest
	json.Unmarshal(jsonData, &request)

	var user mysql.User
	user.Username = c.Param("username")

	var exists int64

	//Check if Username exists in Database
	result := db.Con.Model(&user).Where("upper(username) = upper(?)", user.Username).Count(&exists)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	if exists == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, config.Error{Code: "AUTH006", Message: "User not found"})
		return
	}

	//check if Password is reset
	result = db.Con.Where("upper(username) = upper(?) and password = ? and password_reset = ?", user.Username, "RESET", true).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH008", Message: "Password not Reset"})
			return
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
			return
		}
	}

	//Save new Password to Database
	user.Password = request.Password
	user.PasswordReset = false
	result = db.Con.Save(&user)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ005", Message: "Error Saving new Password"})
		return
	}

}

//This function handles GET requests sent to the /api/v1/user endpoint
func GetAllUsers(c *gin.Context) {

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	type getAllUsersResponse struct {
		Count int          `json:"count"`
		Users []mysql.User `json:"users"`
	}

	var users []mysql.User

	//Get all Users from Database
	result := db.Con.Find(&users)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	//Add GroupIDs
	for _, user := range users {
		err := db.Con.Model(&user).Association("UserGroups").Find(&user.UserGroups)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
			return
		}

		for _, group := range user.UserGroups {
			user.UserGroupIDs = append(user.UserGroupIDs, group.ID)
		}
	}

	c.JSON(http.StatusOK, getAllUsersResponse{Count: len(users), Users: users})
}

//This function handles GET requests sent to the /api/v1/user/:id endpoint
func GetUser(c *gin.Context) {

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	type getUserResponse struct {
		User mysql.User `json:"user"`
	}

	//Convert ID Parameter into int32
	var user mysql.User
	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "RQST001", Message: "Error decoding Request"})
		return
	}

	user.ID = int32(tmp)

	//Get User from Database
	result := db.Con.First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "DBSQ006", Message: "User ID not Found"})
			return
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
			return
		}
	}

	err = db.Con.Model(&user).Association("UserGroups").Find(&user.UserGroups)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	//Add GroupIDs
	for _, group := range user.UserGroups {
		user.UserGroupIDs = append(user.UserGroupIDs, group.ID)
	}

	c.JSON(http.StatusOK, getUserResponse{User: user})
}

//This function handles DELETE requests sent to the /api/v1/user/:id endpoint
func DeleteUser(c *gin.Context) {

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	//Convert ID Parameter into int32
	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "RQST001", Message: "Error decoding Request"})
		return
	}
	userid := int32(tmp)

	reqUserId, _ := c.Get("userid")

	//Check if UserID
	var exists int64
	result := db.Con.Model(mysql.User{}).Where("id = ?", userid).Count(&exists)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	if exists == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, config.Error{Code: "DBSQ006", Message: "User ID not Found"})
		return
	}

	if userid != reqUserId {
		var user mysql.User

		result := db.Con.Where("id = ?", reqUserId).First(&user)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
			return
		}

		Log.Debug(user)

		err = db.Con.Model(&user).Association("Permissions").Find(&user.Permissions)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
			return
		}

		if !user.Permissions.Admin {
			c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH009", Message: "User not Authorized for this Action"})
			return
		}
	}

	result = db.Con.Delete(mysql.User{}, userid)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "DBSQ006", Message: "User ID not Found"})
			return
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
			return
		}
	}

}

//This function handles GET requests sent to the /api/v1/user/:id/logout endpoint
func LogoutUser(c *gin.Context) {

	//Check if redis database connection is already established and create one if not
	if rdis == nil {
		connectRedis()
	}

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	//Convert ID Parameter into int32
	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "RQST001", Message: "Error decoding Request"})
		return
	}
	userid := int32(tmp)

	reqUserId, _ := c.Get("userid")

	//Check if UserID
	var exists int64
	result := db.Con.Model(mysql.User{}).Where("id = ?", userid).Count(&exists)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	if exists == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, config.Error{Code: "DBSQ006", Message: "User ID not Found"})
		return
	}

	if userid != reqUserId {
		var user mysql.User

		result := db.Con.Where("id = ?", reqUserId).First(&user)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
			return
		}

		Log.Debug(user)

		err = db.Con.Model(&user).Association("Permissions").Find(&user.Permissions)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
			return
		}

		if !user.Permissions.Admin {
			c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH009", Message: "User not Authorized for this Action"})
			return
		}
	}

	err = rdis.Remove(fmt.Sprint(userid))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ003", Message: "Error Accessing Redis"})
		return
	}

	result = db.Con.Model(&mysql.User{}).Where("id = ?", userid).Update("refresh_token", "")
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}
}

//This function handles GET requests sent to the /api/v1/user/:id/permissions endpoint
func GetPermissions(c *gin.Context) {

	//Check if mysql database connection is already established and create one if not
	if db == nil {
		connectMySql()
	}

	type getPermsResponse struct {
		Perms mysql.Permissions `json:"perms"`
	}

	//Convert ID Parameter into int32
	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "RQST001", Message: "Error decoding Request"})
		return
	}
	userid := int32(tmp)

	//Check if UserID exists
	var exists int64
	result := db.Con.Model(mysql.User{}).Where("id = ?", userid).Count(&exists)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	if exists == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, config.Error{Code: "DBSQ006", Message: "User ID not Found"})
		return
	}

	var perm mysql.Permissions

	err = db.Con.Model(mysql.User{}).Where("id = ?", userid).Association("Permissions").Find(&perm)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	err = db.Con.Model(&perm).Association("Speakers").Find(&perm.Speakers)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	err = db.Con.Model(&perm).Association("Rooms").Find(&perm.Rooms)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	err = db.Con.Model(&perm).Association("SpeakerGroups").Find(&perm.SpeakerGroups)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
		return
	}

	for _, speaker := range perm.Speakers {
		perm.SpeakerIDs = append(perm.SpeakerIDs, speaker.ID)
	}

	for _, room := range perm.Rooms {
		perm.RoomIDs = append(perm.RoomIDs, room.ID)
	}

	for _, speakergroup := range perm.SpeakerGroups {
		perm.SpeakerGroupIDs = append(perm.SpeakerGroupIDs, speakergroup.ID)
	}

	c.JSON(http.StatusOK, getPermsResponse{Perms: perm})
}

//This function handles PATCH requests sent to the /api/v1/user/:id/permissions endpoint
func UpdatePermissions(c *gin.Context) {

	////Check if mysql database connection is already established and create one if not
	//if db == nil {
	//	connectMySql()
	//}
	//
	//type updatePermsResponse struct{
	//	Perms mysql.Permissions `json:"perms"`
	//}
	//
	////Convert ID Parameter into int32
	//tmp, err := strconv.Atoi(c.Param("id"))
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusInternalServerError,config.Error{Code: "RQST001",Message: "Error decoding Request"})
	//	return
	//}
	//userid := int32(tmp)
	//
	////Check if UserID exists
	//var exists int64
	//result := db.Con.Model(mysql.User{}).Where("id = ?", userid).Count(&exists)
	//if result.Error != nil {
	//	c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
	//	return
	//}
	//
	//if exists == 0 {
	//	c.AbortWithStatusJSON(http.StatusNotFound, config.Error{Code: "DBSQ006", Message: "User ID not Found"})
	//	return
	//}
	//
	//var perm mysql.Permissions
	//
	//err = db.Con.Model(mysql.User{}).Where("id = ?", userid).Association("Permissions").Find(&perm)
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusInternalServerError, config.Error{Code: "DBSQ001", Message: "Error Accessing Database"})
	//	return
	//}

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
