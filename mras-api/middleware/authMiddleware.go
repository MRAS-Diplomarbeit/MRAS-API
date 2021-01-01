package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mras-diplomarbeit/mras-api/config"
	. "github.com/mras-diplomarbeit/mras-api/logger"
	"github.com/mras-diplomarbeit/mras-api/service"
	"net/http"
)

//
//type Claims struct {
//	UserID int `json:"userID"`
//	DeviceID string `json:"deviceID"`
//	jwt.StandardClaims
//}
//
//func AuthMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		authHeader := strings.Split(r.Header.Get("Authorization")," ")
//		if len(authHeader)!=2 {
//			w.WriteHeader(http.StatusUnauthorized)
//			json.NewEncoder(w).Encode(config.Error{Code: "AU001", Message: "Invalid Header"})
//			return
//		} else {
//			jwtToken := authHeader[1]
//
//			claims := &Claims{}
//
//			tkn, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
//				return []byte(config.JWTAccessSecret), nil
//			})
//
//			s := strconv.FormatInt(int64(claims.UserID), 10)
//			if claims.UserID != 0 && tkn.Valid{
//				r.Header.Set("userID",s)
//				next.ServeHTTP(w, r)
//			} else {
//				fmt.Println(err)
//				w.WriteHeader(http.StatusUnauthorized)
//				w.Write([]byte("Unauthorized"))
//				return
//			}
//		}
//	})
//}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		Log.Debug("Authenticating JWT Token")
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH001", Message: "Missing or Invalid Authorization Header"})
			return
		}
		tokenString := authHeader[len("Bearer "):]

		token, _ := service.JWTAuthService(config.JWTAccessSecret).ValidateToken(tokenString)
		if token.Valid {
			claims := token.Claims.(jwt.MapClaims)
			c.Set("userid", claims["userid"])
			c.Set("deviceid", claims["deviceid"])
		} else {
			Log.Warn("Invalid JWT")
			c.AbortWithStatusJSON(http.StatusUnauthorized, config.Error{Code: "AUTH001", Message: "Missing or Invalid Authorization Header"})
			return
		}
		c.Next()
	}
}
