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

var (
	ClientBackendPort int
	ClientBackendPath string
)

func LoadConfig(path string) {
	v1, _ := readConfig(".env", path, map[string]interface{}{
		"port":             3000,
		"jwtAccessSecret":  "ABCDEF",
		"jwtRefreshSecret": "ABCDEFG",
		"logfile":          "log.txt",
		"logLevel":         "info",
		"app": map[string]interface{}{
			"port": 3000,
		},
		"client": map[string]interface{}{
			"port": 3001,
		},
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

	ClientBackendPort = v1.GetInt("client.client-backend.port")
	ClientBackendPath = v1.GetString("client.client-backend.path")
}

func readConfig(filename string, path string, defaults map[string]interface{}) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	v.SetConfigName(filename)
	v.AddConfigPath(path)
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}
