package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	LoggerLevel int8 `mapstructure:"LOGGER_LEVEL"`
	ServerAddr  string
}

func InitConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("config")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return config, err

}
