package config

import (
	"github.com/spf13/viper"
)

var (
	MySQL map[string]interface{}
	Redis map[string]interface{}
	Port  int
)

func Init() error {
	v1, err := readConfig(".env", map[string]interface{}{
		"port": 3000,
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

	if err != nil {
		return err
	}

	Port = v1.GetInt("server.port")
	MySQL = v1.GetStringMap("server.mysql")
	Redis = v1.GetStringMap("server.redis")

	return nil
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
