package config

import (
	"github.com/spf13/viper"
	"log"
	"go_test/global"
)

var AppConfig *Config
var DbConfig *DBConfig

type Config struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	AppConfig = &Config{}
	if err := viper.UnmarshalKey("app", AppConfig); err != nil {
		log.Fatalf("解析应用配置失败: %v", err)
	}

	DbConfig = &DBConfig{}
	if err := viper.UnmarshalKey("db", DbConfig); err != nil {
		log.Fatalf("解析数据库配置失败: %v", err)
	}

	global.DB = InitDB()
} 