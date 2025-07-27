package router

import (
	"go_test/config"
	"go_test/controller"
	"go_test/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	// 添加CORS中间件，允许所有跨域请求
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
				"name": config.GetAppConfig().Name,
				"port": config.GetAppConfig().Port,
			})
		})
		// GET http://localhost:8080/api/biz/info
		biz.GET("/info", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"host":     config.GetDBConfig().Host,
				"port":     config.GetDBConfig().Port,
				"user":     config.GetDBConfig().User,
				"password": config.GetDBConfig().Password,
				"name":     config.GetDBConfig().Name,
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
