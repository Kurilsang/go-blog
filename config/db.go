package config

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	dbConfig := GetDBConfig()
	if dbConfig == nil {
		log.Fatalf("数据库配置未初始化")
	}

	host := dbConfig.Host
	port := dbConfig.Port
	user := dbConfig.User
	password := dbConfig.Password
	dbname := dbConfig.Name
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbname)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(dbConfig.ConnMaxLifetime) * time.Second)

	log.Printf("数据库连接池配置完成: MaxIdleConns=%d, MaxOpenConns=%d, ConnMaxLifetime=%ds",
		dbConfig.MaxIdleConns, dbConfig.MaxOpenConns, dbConfig.ConnMaxLifetime)

	return db
}
