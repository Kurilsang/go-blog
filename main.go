package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go_test/config"
	"go_test/model"
	"go_test/router"
)

func main() {
	config.InitConfig()

	// 自动迁移数据库表
	model.AutoMigrate()

	ginServer := gin.Default()

	router.RegisterRoutes(ginServer)

	addr := fmt.Sprintf(":%d", config.AppConfig.Port)
	ginServer.Run(addr)
}
