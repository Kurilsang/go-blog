package router

import (
	"go_test/config"
	"go_test/controller"
	"go_test/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// 认证相关接口（不需要JWT拦截器）
		auth := api.Group("/auth")
		// POST http://localhost:8080/api/auth/register
		auth.POST("/register", controller.Register)
		// POST http://localhost:8080/api/auth/login
		auth.POST("/login", controller.Login)

		// 业务相关接口（需要JWT拦截器）
		biz := api.Group("/biz", middleware.GlobalMiddleware())
		// GET http://localhost:8080/api/biz/hello
		biz.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"name": config.AppConfig.Name,
				"port": config.AppConfig.Port,
			})
		})
		// GET http://localhost:8080/api/biz/info
		biz.GET("/info", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"host":     config.DbConfig.Host,
				"port":     config.DbConfig.Port,
				"user":     config.DbConfig.User,
				"password": config.DbConfig.Password,
				"name":     config.DbConfig.Name,
			})
		})
		// 汇率相关接口
		// POST http://localhost:8080/api/biz/rate
		biz.POST("/rate", controller.CreateExchangeRate)
		// GET http://localhost:8080/api/biz/rate
		biz.GET("/rate", controller.GetExchangeRates)
		// 文章相关接口
		// POST http://localhost:8080/api/biz/article
		biz.POST("/article", controller.CreateArticle)
		// GET http://localhost:8080/api/biz/article
		biz.GET("/article", controller.GetArticles)
		// GET http://localhost:8080/api/biz/article/:id
		biz.GET("/article/:id", controller.GetArticleByID)
		// 文章点赞相关接口
		// POST http://localhost:8080/api/biz/article/:id/like
		biz.POST("/article/:id/like", controller.LikeArticle)
		// GET http://localhost:8080/api/biz/article/:id/like
		biz.GET("/article/:id/like", controller.GetArticleLikes)
	}
}
