package config

import (
	"fmt"
	"log"

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
	return db
}
