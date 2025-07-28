package controller

import (
	"go_test/global"
	"go_test/model"
	"go_test/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Register 用户注册
func Register(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 如果没有指定角色，默认为普通用户
	if user.Role == "" {
		user.Role = global.RoleUser
	}

	// 验证角色是否有效
	if user.Role != global.RoleAdmin && user.Role != global.RoleUser {
		c.JSON(http.StatusBadRequest, gin.H{"error": "角色参数无效，只能是admin或user"})
		return
	}

	// 对密码进行加密
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	user.Password = hashedPassword

	// 保存用户到数据库
	if err := global.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "注册失败，用户名可能已存在"})
		return
	}

	// 生成JWT令牌并返回
	generateAndReturnToken(c, user.Username, user.Role)
}

// Login 用户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 查询用户
	var user model.User
	if err := global.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户不存在"})
		return
	}

	// 校验密码
	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	// 生成JWT令牌并返回
	generateAndReturnToken(c, user.Username, user.Role)
}

// 生成JWT令牌并返回
func generateAndReturnToken(c *gin.Context, username, role string) {
	token, err := utils.GenerateJWT(username, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Username: username,
		Role:     role,
		Token:    token,
		Message:  "操作成功",
	})
}
