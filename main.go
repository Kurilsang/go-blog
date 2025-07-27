package main

import (
	"fmt"
	"go_test/config"
	"go_test/model"
	"go_test/router"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitConfig()

	// 自动迁移数据库表
	model.AutoMigrate()

	ginServer := gin.Default()

	router.RegisterRoutes(ginServer)

	addr := fmt.Sprintf(":%d", config.GetAppConfig().Port)
	ginServer.Run(addr)
}
