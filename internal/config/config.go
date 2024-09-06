package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Address string
}

type DatabaseConfig struct {
	DNS string
}

type Queueconfig struct {
	Address string
}

type RedisConfig struct {
	Address  string
	Username string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret         string
	ExpireDuration string
}
type RateLimitConfig struct {
	RequestsPerMinute int64
}

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	Queue     Queueconfig
	JWT       JWTConfig
	RateLimit RateLimitConfig
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./")

	err := viper.ReadInConfig()

	if err != nil {
		return nil, err
	}

	var config Config

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	fmt.Println("Config:", &config)

	return &config, nil
}
