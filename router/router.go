package router

import (
	"github.com/gin-gonic/gin"
	"go_test/config"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"name": config.AppConfig.Name,
				"port": config.AppConfig.Port,
			})
		})

		api.GET("/info", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"host": config.DbConfig.Host,
				"port": config.DbConfig.Port,
				"user": config.DbConfig.User,
				"password": config.DbConfig.Password,
				"name": config.DbConfig.Name,
			})
		})
	}
} 