package config

import (
	"go_test/global"
	"log"
	"sync/atomic"

	"github.com/spf13/viper"
)

var (
	appConfig   atomic.Value // *Config
	dbConfig    atomic.Value // *DBConfig
	cacheConfig atomic.Value // *CacheConfig
	jwtConfig   atomic.Value // *JWTConfig
)

type Config struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
}

type DBConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type CacheConfig struct {
	ArticleExpire int `mapstructure:"article_expire"`
	LikeExpire    int `mapstructure:"like_expire"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// GetAppConfig 原子读取应用配置
func GetAppConfig() *Config {
	if config := appConfig.Load(); config != nil {
		return config.(*Config)
	}
	return nil
}

// GetDBConfig 原子读取数据库配置
func GetDBConfig() *DBConfig {
	if config := dbConfig.Load(); config != nil {
		return config.(*DBConfig)
	}
	return nil
}

// GetCacheConfig 原子读取缓存配置
func GetCacheConfig() *CacheConfig {
	if config := cacheConfig.Load(); config != nil {
		return config.(*CacheConfig)
	}
	return nil
}

// GetJWTConfig 原子读取JWT配置
func GetJWTConfig() *JWTConfig {
	if config := jwtConfig.Load(); config != nil {
		return config.(*JWTConfig)
	}
	return nil
}

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// 解析并原子存储配置
	app := &Config{}
	if err := viper.UnmarshalKey("app", app); err != nil {
		log.Fatalf("解析应用配置失败: %v", err)
	}
	appConfig.Store(app)

	db := &DBConfig{}
	if err := viper.UnmarshalKey("db", db); err != nil {
		log.Fatalf("解析数据库配置失败: %v", err)
	}
	dbConfig.Store(db)

	cache := &CacheConfig{}
	if err := viper.UnmarshalKey("cache", cache); err != nil {
		log.Fatalf("解析缓存配置失败: %v", err)
	}
	cacheConfig.Store(cache)

	jwt := &JWTConfig{}
	if err := viper.UnmarshalKey("jwt", jwt); err != nil {
		log.Fatalf("解析JWT配置失败: %v", err)
	}
	jwtConfig.Store(jwt)

	global.InitDB(InitDB())
	global.InitRedis(InitRedis())
}
