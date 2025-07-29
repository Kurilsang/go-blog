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
		{
			// POST http://localhost:8080/api/auth/register
			auth.POST("/register", controller.Register)
			// POST http://localhost:8080/api/auth/login
			auth.POST("/login", controller.Login)
		}

		// 普通用户可访问的接口（只需要基础认证）
		user := api.Group("/user", middleware.AuthMiddleware())
		{
			// 文章查看相关接口
			// GET http://localhost:8080/api/user/article
			user.GET("/article", controller.GetArticles)
			// GET http://localhost:8080/api/user/article/pagination - 支持可选关键词搜索
			user.GET("/article/pagination", controller.GetArticlesWithPagination)
			// GET http://localhost:8080/api/user/article/:id
			user.GET("/article/:id", controller.GetArticleByID)

			// 文章点赞相关接口
			// POST http://localhost:8080/api/user/article/:id/like
			user.POST("/article/:id/like", controller.LikeArticle)
			// GET http://localhost:8080/api/user/article/:id/like
			user.GET("/article/:id/like", controller.GetArticleLikes)

			// 汇率查看接口
			// GET http://localhost:8080/api/user/rate
			user.GET("/rate", controller.GetExchangeRates)

			// 用户个人中心接口
			// GET http://localhost:8080/api/user/profile - 获取自己的资料
			user.GET("/profile", controller.GetMyProfile)
			// PUT http://localhost:8080/api/user/profile - 更新自己的资料
			user.PUT("/profile", controller.UpdateMyProfile)
			// PUT http://localhost:8080/api/user/password - 修改自己的密码
			user.PUT("/password", controller.ChangeMyPassword)
			// GET http://localhost:8080/api/user/profile/:id - 查看指定用户资料（需要权限验证）
			user.GET("/profile/:id", controller.GetUserProfile)
		}

		// 管理员专用接口（需要管理员权限）
		admin := api.Group("/admin", middleware.AdminOnlyMiddleware())
		{
			// 普通管理操作（使用JWT验证）
			// POST http://localhost:8080/api/admin/article
			admin.POST("/article", controller.CreateArticle)
			// POST http://localhost:8080/api/admin/rate
			admin.POST("/rate", controller.CreateExchangeRate)

			// 用户管理接口
			// GET http://localhost:8080/api/admin/users - 获取所有用户列表
			admin.GET("/users", controller.GetAllUsers)
			// GET http://localhost:8080/api/admin/user/:id - 查看指定用户资料
			admin.GET("/user/:id", controller.GetUserProfile)
			// PUT http://localhost:8080/api/admin/user/:id - 更新指定用户资料（包含角色和状态）
			admin.PUT("/user/:id", controller.UpdateUserProfile)
		}

		// 敏感操作接口（需要数据库实时验证）
		sensitive := api.Group("/admin/sensitive", middleware.SensitiveAdminMiddleware())
		{
			// 敏感操作：删除数据（查询数据库验证权限）
			// DELETE http://localhost:8080/api/admin/sensitive/article/batch
			sensitive.DELETE("/article/batch", controller.BatchDeleteArticles)
		}

		// 保持向后兼容的业务接口（使用原有的全局中间件）
		biz := api.Group("/biz", middleware.GlobalMiddleware())
		{
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
		}
	}
}
