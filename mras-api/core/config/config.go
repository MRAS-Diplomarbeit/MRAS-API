package config

import (
	"github.com/spf13/viper"
)

var (
	MySQL            map[string]interface{}
	Redis            map[string]interface{}
	AppPort          int
	JWTAccessSecret  string
	JWTRefreshSecret string
	Loglevel         string
	LogLocation      string
	ClientPort       int
)

func init() {
	v1, _ := readConfig(".env", map[string]interface{}{
		"port":             3000,
		"jwtAccessSecret":  "ABCDEF",
		"jwtRefreshSecret": "ABCDEFG",
		"logfile":      "log.txt",
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

	AppPort = v1.GetInt("server.app.port")
	ClientPort = v1.GetInt("server.client.port")
	MySQL = v1.GetStringMap("server.mysql")
	Redis = v1.GetStringMap("server.redis")
	JWTAccessSecret = v1.GetString("server.jwtAccessSecret")
	JWTRefreshSecret = v1.GetString("server.jwtRefreshSecret")
	Loglevel = v1.GetString("server.logLevel")
	LogLocation = v1.GetString("server.logfile")
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
