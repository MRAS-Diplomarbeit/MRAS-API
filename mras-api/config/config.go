package config

import (
	"github.com/spf13/viper"
)

var (
	MySQL            map[string]interface{}
	Redis            map[string]interface{}
	Port             int
	JWTAccessSecret  string
	JWTRefreshSecret string
	Loglevel         string
	LogLocation      string
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func init() {
	v1, _ := readConfig(".env", map[string]interface{}{
		"port":             3000,
		"jwtAccessSecret":  "ABCDEF",
		"jwtRefreshSecret": "ABCDEFG",
		"logLocation":      "log.txt",
		"logLevel":         "info",
		"mysql": map[string]interface{}{
			"host":     "localhost",
			"port":     3306,
			"user":     "root",
			"password": "Passwort123!",
			"dbname":   "mras",
		}, "redis": map[string]interface{}{
			"host":     "localhost",
			"port":     6379,
			"password": "",
			"db":       0,
		},
	})

	Port = v1.GetInt("server.port")
	MySQL = v1.GetStringMap("server.mysql")
	Redis = v1.GetStringMap("server.redis")
	JWTAccessSecret = v1.GetString("server.jwtAccessSecret")
	JWTRefreshSecret = v1.GetString("server.jwtRefreshSecret")
	Loglevel = v1.GetString("server.logLevel")
	LogLocation = v1.GetString("server.logLocation")
}

func readConfig(filename string, defaults map[string]interface{}) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	v.SetConfigName(filename)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}
