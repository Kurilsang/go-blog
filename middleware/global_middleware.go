package middleware

import (
	"errors"
	"go_test/global"
	"go_test/model"
	"go_test/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserInfo 用户信息结构
type UserInfo struct {
	Username string
	Role     string
	UserID   uint
}

// AuthContext 认证上下文
type AuthContext struct {
	UserClaims *utils.UserClaims
	UserInfo   *UserInfo
}

// Validator 验证器接口
type Validator interface {
	Validate(ctx *AuthContext) error
}

// AdminRoleValidator 管理员角色验证器
type AdminRoleValidator struct{}

func (v *AdminRoleValidator) Validate(ctx *AuthContext) error {
	if ctx.UserClaims.Role != global.RoleAdmin {
		return errors.New("权限不足，仅管理员可访问")
	}
	return nil
}

// DatabaseStatusValidator 数据库状态验证器
type DatabaseStatusValidator struct{}

func (v *DatabaseStatusValidator) Validate(ctx *AuthContext) error {
	// 从数据库查询最新用户信息
	var user model.User
	if err := global.DB.Where("username = ?", ctx.UserClaims.Username).First(&user).Error; err != nil {
		return errors.New("用户不存在或已被禁用")
	}

	// 检查用户状态
	if user.Status != global.UserStatusActive {
		return errors.New("用户账户已被禁用")
	}

	// 更新上下文中的用户信息为数据库中的最新信息
	ctx.UserInfo = &UserInfo{
		Username: user.Username,
		Role:     user.Role,
		UserID:   user.ID,
	}
	return nil
}

// DatabaseAdminValidator 数据库管理员验证器（状态+角色双重验证）
type DatabaseAdminValidator struct{}

func (v *DatabaseAdminValidator) Validate(ctx *AuthContext) error {
	// 从数据库查询最新用户信息
	var user model.User
	if err := global.DB.Where("username = ?", ctx.UserClaims.Username).First(&user).Error; err != nil {
		return errors.New("用户不存在或已被禁用")
	}

	// 检查用户状态
	if user.Status != global.UserStatusActive {
		return errors.New("用户账户已被禁用")
	}

	// 验证管理员权限（使用数据库中的最新角色）
	if user.Role != global.RoleAdmin {
		return errors.New("权限不足，您不是管理员或管理员权限已被撤销")
	}

	// 更新上下文中的用户信息为数据库中的最新信息
	ctx.UserInfo = &UserInfo{
		Username: user.Username,
		Role:     user.Role,
		UserID:   user.ID,
	}
	return nil
}

// parseTokenFromRequest 从请求中解析JWT token
func parseTokenFromRequest(c *gin.Context) (*utils.UserClaims, error) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		return nil, errors.New("未提供Token")
	}

	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	}

	return utils.ParseJWT(tokenString)
}

func authMiddleware(validators ...Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 基础JWT认证
		userClaims, err := parseTokenFromRequest(c)
		if err != nil {
			if err.Error() == "未提供Token" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供Token"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token无效或已过期"})
			}
			c.Abort()
			return
		}

		// 2. 验证JWT中必须包含用户ID
		if userClaims.UserID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token格式无效，请重新登录"})
			c.Abort()
			return
		}

		// 3. 初始化认证上下文
		authCtx := &AuthContext{
			UserClaims: userClaims,
			UserInfo: &UserInfo{
				Username: userClaims.Username,
				Role:     userClaims.Role,
				UserID:   userClaims.UserID,
			},
		}

		// 4. 依次执行所有验证器
		for _, validator := range validators {
			if err := validator.Validate(authCtx); err != nil {
				// 根据错误类型返回不同的状态码
				statusCode := http.StatusForbidden
				if strings.Contains(err.Error(), "不存在") || strings.Contains(err.Error(), "禁用") {
					statusCode = http.StatusUnauthorized
				}
				c.JSON(statusCode, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
		}

		// 5. 将最终的用户信息存入上下文
		c.Set("username", authCtx.UserInfo.Username)
		c.Set("userRole", authCtx.UserInfo.Role)
		c.Set("userID", authCtx.UserInfo.UserID)
		c.Next()
	}
}

// AuthMiddleware 基础JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return authMiddleware() // 只有基础JWT认证，无额外验证器
}

// AdminOnlyMiddleware 管理员中间件（JWT + 角色验证）
func AdminOnlyMiddleware() gin.HandlerFunc {
	return authMiddleware(&AdminRoleValidator{})
}

// SensitiveAdminMiddleware 敏感操作中间件（数据库实时验证管理员权限）
func SensitiveAdminMiddleware() gin.HandlerFunc {
	return authMiddleware(&DatabaseAdminValidator{})
}

// GlobalMiddleware 保持向后兼容
func GlobalMiddleware() gin.HandlerFunc {
	return AuthMiddleware()
}
