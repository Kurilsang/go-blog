package global

import (
	"sync"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var (
	DB        *gorm.DB
	RedisDB   *redis.Client
	dbOnce    sync.Once
	redisOnce sync.Once
)

// InitDB 使用sync.Once确保数据库连接只初始化一次
func InitDB(db *gorm.DB) {
	dbOnce.Do(func() {
		DB = db
	})
}

// InitRedis 使用sync.Once确保Redis连接只初始化一次
func InitRedis(redis *redis.Client) {
	redisOnce.Do(func() {
		RedisDB = redis
	})
}
