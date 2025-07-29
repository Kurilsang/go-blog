package utils

import (
	"errors"
	"go_test/config"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// UserClaims JWT用户信息结构
type UserClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	UserID   uint   `json:"user_id"`
}

// GenerateJWT 生成JWT令牌（包含用户ID）
func GenerateJWT(username, role string, userID uint) (string, error) {
	jwtConfig := config.GetJWTConfig()
	if jwtConfig == nil {
		return "", errors.New("JWT配置未初始化")
	}

	claims := jwt.MapClaims{
		"username": username,
		"role":     role,
		"user_id":  userID,
		"exp":      time.Now().Add(time.Duration(jwtConfig.ExpireHours) * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.Secret))
}

// ParseJWT 校验并解析JWT令牌，返回用户信息和错误信息
func ParseJWT(tokenString string) (*UserClaims, error) {
	jwtConfig := config.GetJWTConfig()
	if jwtConfig == nil {
		return nil, errors.New("JWT配置未初始化")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("签名方法不正确")
		}
		return []byte(jwtConfig.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, usernameOk := claims["username"].(string)
		role, roleOk := claims["role"].(string)
		userIDFloat, userIDOk := claims["user_id"].(float64)

		if !usernameOk || !roleOk || !userIDOk {
			return nil, errors.New("token中缺少必要的用户信息")
		}

		return &UserClaims{
			Username: username,
			Role:     role,
			UserID:   uint(userIDFloat),
		}, nil
	}
	return nil, errors.New("无效的token")
}
