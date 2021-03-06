package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/core/config"
	"github.com/mras-diplomarbeit/mras-api/core/db/mysql"
	errs "github.com/mras-diplomarbeit/mras-api/core/error"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"github.com/mras-diplomarbeit/mras-api/core/utils"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

//This function handles POST requests sent to the /api/v1/user/login endpoint
func (env *Env) LoginUser(c *gin.Context) {

	type loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		DeviceID string `json:"device_id"`
	}

	type loginResponse struct {
		AccessToken  string     `json:"access_token"`
		RefreshToken string     `json:"refresh_token"`
		User         mysql.User `json:"user"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request loginRequest
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	if request.Username == "" || request.Password == "" || request.DeviceID == "" {
		Log.WithField("module", "handler").Error("Empty Fields in Request Body")
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST002)
		return
	}

	user := mysql.User{}

	//lookup user in users database
	result := env.db.Where("upper(username) = upper(?) AND password = ?", request.Username, request.Password).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			Log.WithField("module", "handler").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH003)
			return
		} else {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}
	}

	err = env.db.Model(&user).Association("UserGroups").Find(&user.UserGroups)
	if err != nil {
		Log.WithField("module", "sql").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	//Add GroupIDs
	for _, group := range user.UserGroups {
		user.UserGroupIDs = append(user.UserGroupIDs, group.ID)
	}

	//Generate JWT AccessToken
	accessToken, err := utils.JWTAuthService(config.JWTAccessSecret).GenerateToken(user.ID, request.DeviceID, time.Hour*24)
	if err != nil {
		Log.WithField("module", "jwt").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.AUTH002)
		return
	}

	//Add AccessToken to Redis
	err = env.rdis.AddPair(fmt.Sprint(user.ID), accessToken, time.Hour*24)
	if err != nil {
		Log.WithField("module", "redis").WithError(err).Error("Error adding AccessToken to Redis.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ003)
		return
	}

	//Generate JWT RefreshToken
	refreshToken, err := utils.JWTAuthService(config.JWTRefreshSecret).GenerateToken(user.ID, request.DeviceID, time.Hour*4380)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.AUTH002)
		return
	}

	//Save RefreshToken to Database
	user.RefreshToken = refreshToken
	result = env.db.Save(&user)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ002)
		return
	}

	c.JSON(200, loginResponse{accessToken, refreshToken, user})
}

//This function handles POST requests sent to the /api/v1/user/register endpoint
func (env *Env) RegisterUser(c *gin.Context) {

	type registerRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		DeviceID string `json:"device_id"`
	}

	type registerResponse struct {
		AccessToken  string     `json:"access_token"`
		RefreshToken string     `json:"refresh_token"`
		User         mysql.User `json:"user"`
		ResetCode    string     `json:"reset_code"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request registerRequest
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	if request.Username == "" || request.Password == "" || request.DeviceID == "" {
		Log.WithField("module", "handler").Error("Empty Fields in Request Body")
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST002)
		return
	}

	var empty int64
	result := env.db.Model(&mysql.User{}).Count(&empty)
	if result.Error != nil {
		Log.WithField("module", "handler").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	user := mysql.User{}
	perms := mysql.Permissions{}
	defaultGroup := mysql.UserGroup{}

	if empty == 0 {

		perms.Admin = true
		perms.CanEdit = true

		defaultGroupPerms := mysql.Permissions{CanEdit: false, Admin: false}

		defaultGroup.Name = "default"

		result = env.db.Save(&defaultGroupPerms)
		if result.Error != nil {
			Log.WithField("module", "handler").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}

		defaultGroup.Permissions = defaultGroupPerms

		result = env.db.Save(&defaultGroup)
		if result.Error != nil {
			Log.WithField("module", "handler").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}

	} else {
		var exists int64
		//Check if Username already exists in Database
		result = env.db.Model(&user).Where("upper(username) = upper(?)", user.Username).Count(&exists)
		if result.Error != nil {
			Log.WithField("module", "handler").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}
		Log.WithField("module", "handler").Debug("Users found: ", exists)

		if exists != 0 {
			Log.WithField("module", "handler").Error("Username already exists in Database")
			c.AbortWithStatusJSON(http.StatusForbidden, errs.AUTH004)
			return
		}

		perms.Admin = false
		perms.CanEdit = false

		defaultGroup.Name = "default"
		result = env.db.Model(&defaultGroup).Find(&defaultGroup)
		if result.Error != nil {
			Log.WithField("module", "handler").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}

	}

	//Create permission entry for new user in permissions table
	result = env.db.Save(&perms)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	user.Username = request.Username
	user.Password = request.Password
	user.AvatarID = "default"
	user.PermID = perms.ID
	user.UserGroups = append(user.UserGroups, &defaultGroup)
	user.ResetCode = utils.GenerateCode()

	//Save new user to users database
	result = env.db.Save(&user)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	//Generate JWT AccessToken
	accessToken, err := utils.JWTAuthService(config.JWTAccessSecret).GenerateToken(user.ID, request.DeviceID, time.Hour*24)
	if err != nil {
		Log.WithField("module", "jwt").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.AUTH002)
		return
	}

	//Add AccessToken to Redis
	err = env.rdis.AddPair(fmt.Sprint(user.ID), accessToken, time.Hour*24)
	if err != nil {
		Log.WithField("module", "redis").WithError(err).Error("Error adding AccessToken to Redis.")
		err = nil
	}

	//Generate RefreshToken
	refreshToken, err := utils.JWTAuthService(config.JWTRefreshSecret).GenerateToken(user.ID, request.DeviceID, time.Hour*24)
	if err != nil {
		Log.WithField("module", "jwt").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.AUTH002)
		return
	}

	user.RefreshToken = refreshToken

	//Save RefreshToken to Database
	result = env.db.Save(&user)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ002)
		return
	}

	c.JSON(200, registerResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: user, ResetCode: user.ResetCode})
}

//This function handles POST requests sent to the /api/v1/user/refresh endpoint
func (env *Env) GenerateAccessToken(c *gin.Context) {

	type refreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	type refreshResponse struct {
		AccessToken string `json:"access_token"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request refreshRequest
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	user := mysql.User{}
	user.RefreshToken = request.RefreshToken

	token, _ := utils.JWTAuthService(config.JWTRefreshSecret).ValidateToken(user.RefreshToken)
	if token == nil || !token.Valid {
		Log.WithField("module", "handler").Error("Invalid JWT Token")
		c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH005)
		return
	}
	claims := token.Claims.(jwt.MapClaims)
	user.ID = int32(claims["userid"].(float64))
	Log.Debug(claims["userid"])

	var exists int64

	//Check if Refresh Token is valid
	result := env.db.Model(&user).Where("id = ? and refresh_token = ?", user.ID, request.RefreshToken).Count(&exists)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if exists == 0 {
		Log.WithField("module", "handler").Error("Invalid RefreshToken")
		c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH005)
		return
	}

	Log.WithField("model", "handler").Debug(user)

	//Generate JWT AccessToken
	accessToken, err := utils.JWTAuthService(config.JWTAccessSecret).GenerateToken(user.ID, claims["deviceid"].(string), time.Hour*24)
	if err != nil {
		Log.WithField("module", "jwt").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.AUTH002)
		return
	}

	//Add AccessToken to Redis
	err = env.rdis.AddPair(fmt.Sprint(user.ID), accessToken, time.Hour*24)
	if err != nil {
		Log.WithField("module", "redis").WithError(err).Error("Error adding AccessToken to Redis.")
		err = nil
	}

	c.JSON(http.StatusOK, refreshResponse{AccessToken: accessToken})

}

//This function handles POST requests sent to the /api/v1/user/password/reset/:username endpoint
func (env *Env) ResetUserPassword(c *gin.Context) {

	type resetUserPasswordRequest struct {
		ResetCode string `json:"reset_code"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request resetUserPasswordRequest
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var user mysql.User
	user.Username = c.Param("username")
	user.ResetCode = request.ResetCode

	var exists int64

	//Check if Username exists in Database
	result := env.db.Model(&user).Where("upper(username) = upper(?)", user.Username).Count(&exists)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if exists == 0 {
		Log.WithField("module", "handler").Error("Username not found in Database")
		c.AbortWithStatusJSON(http.StatusNotFound, errs.AUTH006)
		return
	}

	//Check Database if ResetCode is correct
	result = env.db.Where("upper(username) = upper(?) and reset_code = ?", user.Username, user.ResetCode).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			Log.WithField("module", "handler").Error("ResetCode Username combination not found (incorrect)")
			c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH007)
			return
		} else {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}
	}

	//Reset Password in Database
	user.Password = "RESET"
	user.PasswordReset = true
	result = env.db.Save(&user)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ004)
		return
	}
}

//This function handles POST requests sent to the /api/v1/user/password/new/:username endpoint
func (env *Env) NewUserPassword(c *gin.Context) {

	type newUserPasswordRequest struct {
		Password string `json:"password"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request newUserPasswordRequest
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var user mysql.User
	user.Username = c.Param("username")

	var exists int64

	//Check if Username exists in Database
	result := env.db.Model(&user).Where("upper(username) = upper(?)", user.Username).Count(&exists)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if exists == 0 {
		Log.WithField("module", "sql").Error("Username not Found in Database")
		c.AbortWithStatusJSON(http.StatusNotFound, errs.AUTH006)
		return
	}

	//check if Password is reset
	result = env.db.Where("upper(username) = upper(?) and password = ? and password_reset = ?", user.Username, "RESET", true).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			Log.WithField("module", "handler").Error("Username and ResetCode combination not found!")
			c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH008)
			return
		} else {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}
	}

	//Save new Password to Database
	user.Password = request.Password
	user.PasswordReset = false
	result = env.db.Save(&user)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ005)
		return
	}

}

//This function handles GET requests sent to the /api/v1/user endpoint
func (env *Env) GetAllUsers(c *gin.Context) {

	type getAllUsersResponse struct {
		Count int          `json:"count"`
		Users []mysql.User `json:"users"`
	}

	var users []mysql.User

	//Get all Users from Database
	result := env.db.Find(&users)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	//Add GroupIDs
	for i := 0; i < len(users); i++ {
		err := env.db.Model(&users[i]).Association("UserGroups").Find(&users[i].UserGroups)
		if err != nil {
			Log.WithField("module", "sql").WithError(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}

		for _, group := range users[i].UserGroups {
			users[i].UserGroupIDs = append(users[i].UserGroupIDs, group.ID)
		}
	}

	c.JSON(http.StatusOK, getAllUsersResponse{Count: len(users), Users: users})
}

//This function handles GET requests sent to the /api/v1/user/:id endpoint
func (env *Env) GetUser(c *gin.Context) {

	type getUserResponse struct {
		User mysql.User `json:"user"`
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

	//Get User from Database
	result := env.db.First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			Log.WithField("module", "handler").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusUnauthorized, errs.DBSQ006)
			return
		} else {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}
	}

	err = env.db.Model(&user).Association("UserGroups").Find(&user.UserGroups)
	if err != nil {
		Log.WithField("module", "sql").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	//Add GroupIDs
	for _, group := range user.UserGroups {
		user.UserGroupIDs = append(user.UserGroupIDs, group.ID)
	}

	c.JSON(http.StatusOK, getUserResponse{User: user})
}

//This function handles DELETE requests sent to the /api/v1/user/:id endpoint
func (env *Env) DeleteUser(c *gin.Context) {

	//Convert ID Parameter into int32
	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	userid := int32(tmp)

	reqUserId, _ := c.Get("userid")

	//Check if UserID
	var exists int64
	result := env.db.Model(mysql.User{}).Where("id = ?", userid).Count(&exists)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	if exists == 0 {
		Log.WithField("module", "handler").Error("User not Found in Database")
		c.AbortWithStatusJSON(http.StatusNotFound, errs.DBSQ006)
		return
	}

	if userid != reqUserId {
		var user mysql.User

		result := env.db.Where("id = ?", reqUserId).First(&user)
		if result.Error != nil {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}

		Log.Debug(user)

		err = env.db.Model(&user).Association("Permissions").Find(&user.Permissions)
		if err != nil {
			Log.WithField("module", "sql").WithError(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}

		if !user.Permissions.Admin {
			Log.WithField("module", "handler").Error("User not Authorized for this Action")
			c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH009)
			return
		}
	}

	result = env.db.Delete(mysql.User{}, userid)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			Log.WithField("module", "handler").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusNotFound, errs.DBSQ006)
			return
		} else {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}
	}

}

//This function handles GET requests sent to the /api/v1/user/:id/logout endpoint
func (env *Env) LogoutUser(c *gin.Context) {

	//Convert ID Parameter into int32
	tmp, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.RQST001)
		return
	}
	userid := int32(tmp)

	reqUserId, _ := c.Get("userid")

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
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.DBSQ006)
		return
	}

	if userid != reqUserId {
		var user mysql.User

		result := env.db.Where("id = ?", reqUserId).First(&user)
		if result.Error != nil {
			Log.WithField("module", "sql").WithError(result.Error)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}

		Log.Debug(user)

		err = env.db.Model(&user).Association("Permissions").Find(&user.Permissions)
		if err != nil {
			Log.WithField("module", "sql").WithError(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
			return
		}

		if !user.Permissions.Admin {
			Log.WithField("module", "handler").Error("User not Authorized for this Action")
			c.AbortWithStatusJSON(http.StatusUnauthorized, errs.AUTH009)
			return
		}
	}

	err = env.rdis.Remove(fmt.Sprint(userid))
	if err != nil {
		Log.WithField("module", "redis").WithError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ003)
		return
	}

	result = env.db.Model(&mysql.User{}).Where("id = ?", userid).Update("refresh_token", "")
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}
}

func (env *Env) UpdatePassword(c *gin.Context) {

	type newUserPasswordRequest struct {
		Password string `json:"password"`
	}

	//decode request body
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var request newUserPasswordRequest
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		Log.WithField("module", "handler").WithError(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errs.RQST001)
		return
	}

	var user mysql.User
	tmp, _ := c.Get("userid")
	user.ID = tmp.(int32)
	Log.Debug(user.ID)

	result := env.db.Find(&user)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ001)
		return
	}

	//Save new Password to Database
	user.Password = request.Password

	result = env.db.Save(&user)
	if result.Error != nil {
		Log.WithField("module", "sql").WithError(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errs.DBSQ005)
		return
	}
}
