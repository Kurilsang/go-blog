package service

import (
	"fmt"
	"go_test/dto"
	"go_test/global"
	"go_test/model"
	"go_test/utils"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// Register 用户注册业务逻辑
func (s *AuthService) Register(user model.User) (*dto.AuthResponse, error) {
	// 如果没有指定角色，默认为普通用户
	if user.Role == "" {
		user.Role = global.RoleUser
	}

	// 验证角色是否有效
	if user.Role != global.RoleAdmin && user.Role != global.RoleUser {
		return nil, fmt.Errorf("角色参数无效，只能是admin或user")
	}

	// 对密码进行加密
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败")
	}
	user.Password = hashedPassword

	// 保存用户到数据库
	if err := global.DB.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("注册失败，用户名可能已存在")
	}

	// 生成JWT令牌（包含用户ID）
	token, err := utils.GenerateJWT(user.Username, user.Role, user.ID)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败")
	}

	return &dto.AuthResponse{
		Username: user.Username,
		Role:     user.Role,
		Token:    token,
		Message:  "注册成功",
	}, nil
}

// Login 用户登录业务逻辑
func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	// 查询用户
	var user model.User
	if err := global.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 校验密码
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, fmt.Errorf("密码错误")
	}

	// 生成JWT令牌（包含用户ID）
	token, err := utils.GenerateJWT(user.Username, user.Role, user.ID)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败")
	}

	return &dto.AuthResponse{
		Username: user.Username,
		Role:     user.Role,
		Token:    token,
		Message:  "登录成功",
	}, nil
}
