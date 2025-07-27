package model

import (
	"go_test/global"
	"log"
)

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() {
	err := global.DB.AutoMigrate(&User{}, &ExchangeRate{}, &Article{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移成功")
}
