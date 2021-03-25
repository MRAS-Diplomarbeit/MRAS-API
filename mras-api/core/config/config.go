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
	ClientPort       int
	Loglevel         string
	LogLocation      string
	LogMaxSize       int
	LogMaxAge        int
)

var (
	ClientBackendPort         int
	ClientBackendPlaybackPath string
	ClientBackendMethodPath   string
	ClientBackendProtocol     string
)

func LoadConfig(path string) {
	v1, _ := readConfig("config", path, nil)

	AppPort = v1.GetInt("server.app.port")
	ClientPort = v1.GetInt("server.client.port")
	MySQL = v1.GetStringMap("server.mysql")
	Redis = v1.GetStringMap("server.redis")
	JWTAccessSecret = v1.GetString("server.jwtAccessSecret")
	JWTRefreshSecret = v1.GetString("server.jwtRefreshSecret")

	LogLocation = v1.GetString("logs.path")
	Loglevel = v1.GetString("logs.level")
	LogMaxSize = v1.GetInt("logs.maxSize")
	LogMaxAge = v1.GetInt("logs.maxAge")

	ClientBackendPort = v1.GetInt("client.client-backend.port")
	ClientBackendPlaybackPath = v1.GetString("client.client-backend.path-playback")
	ClientBackendMethodPath = v1.GetString("client.client-backend.path-method")
	ClientBackendProtocol = v1.GetString("client.client-backend.protocol")
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
