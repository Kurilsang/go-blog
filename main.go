package main

import (
	"github.com/gin-gonic/gin"
	"go_test/config"
	"go_test/router"
)

func main() {
	config.InitConfig()
	ginServer := gin.Default()

	router.RegisterRoutes(ginServer)

	ginServer.Run(":8080")
}
