package config

import (
	"github.com/BrainBuzzer/vpn/internal/redis"
	"github.com/spf13/viper"
)

type Config struct {
	RedisConfig redis.Config
}

func init() {
	// use viper to read config
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
}

func NewConfig() (*Config, error) {
	var c Config
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	c.RedisConfig = redis.Config{
		RedisURL: viper.GetString("REDIS_URL"),
	}

	return &c, nil
}
