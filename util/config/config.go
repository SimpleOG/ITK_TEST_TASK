package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	LoggerLevel int8   `mapstructure:"LOGGER_LEVEL"`
	ServerAddr  string `mapstructure:"SERVER_ADDR"`
	DBHost      string `mapstructure:"DB_HOST"`
	DBPort      string `mapstructure:"DB_PORT"`
	DBUser      string `mapstructure:"DB_USER"`
	DBPassword  string `mapstructure:"DB_PASSWORD"`
	DBName      string `mapstructure:"DB_NAME"`
	DBSSLMode   string `mapstructure:"DB_SSL_MODE"`
	DBMaxConns  int32  `mapstructure:"DB_MAX_CONNS"`
}

func (c Config) DBURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func InitConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("config")
	viper.AutomaticEnv()
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
