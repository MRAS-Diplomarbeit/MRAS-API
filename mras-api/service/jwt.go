package service

import (
	"fmt"
	. "github.com/mras-diplomarbeit/mras-api/logger"
	"time"

	"github.com/dgrijalva/jwt-go"
)

//jwt service
type JWTService interface {
	GenerateToken(userID string, deviceID string) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}
type AuthCustomClaims struct {
	UserID   string `json:"userid"`
	DeviceID string `json:"deviceid"`
	jwt.StandardClaims
}

type jwtServices struct {
	secretKey string
	issure    string
}

func JWTAuthService(secret string) JWTService {
	return &jwtServices{
		secretKey: secret,
		issure:    "mras-api",
	}
}

func (service *jwtServices) GenerateToken(userID string, deviceID string) (string, error) {
	claims := &AuthCustomClaims{
		userID,
		deviceID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    service.issure,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//encoded string
	t, err := token.SignedString([]byte(service.secretKey))
	if err != nil {
		Log.Error("Error Generating JWT Token ", err)
		return "", err
	}
	return t, nil
}

func (service *jwtServices) ValidateToken(encodedToken string) (*jwt.Token, error) {
	Log.Debug("Validating JWT")
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			err := fmt.Errorf("Invalid token", token.Header["alg"])
			Log.Error(err)
			return nil, err
		}
		Log.Debug("Successfully Validated JWT")
		return []byte(service.secretKey), nil
	})
}
