package config

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var RedisConfig *RedisConf

// RedisConf Redis配置结构体
type RedisConf struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func InitRedis() *redis.Client {
	RedisConfig = &RedisConf{}
	_ = viper.UnmarshalKey("redis", RedisConfig)
	addr := fmt.Sprintf("%s:%d", RedisConfig.Host, RedisConfig.Port)
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: RedisConfig.Password,
		DB:       RedisConfig.DB,
	})
}
