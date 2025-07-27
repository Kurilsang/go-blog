package middleware

import (
	"net/http"
	"strings"

	"go_test/utils"

	"github.com/gin-gonic/gin"
)

// GlobalMiddleware Gin全局中间件，目前实现JWT验证
func GlobalMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供Token"})
			c.Abort()
			return
		}
		// 支持Bearer前缀
		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}
		username, err := utils.ParseJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token无效或已过期"})
			c.Abort()
			return
		}
		// 可将用户名存入上下文
		c.Set("username", username)
		c.Next()
	}
}
