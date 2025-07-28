package utils

import (
	"errors"
	"go_test/config"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// GenerateJWT 生成JWT令牌
func GenerateJWT(username string) (string, error) {
	jwtConfig := config.GetJWTConfig()
	if jwtConfig == nil {
		return "", errors.New("JWT配置未初始化")
	}

	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Duration(jwtConfig.ExpireHours) * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.Secret))
}

// ParseJWT 校验并解析JWT令牌，返回用户名和错误信息
func ParseJWT(tokenString string) (string, error) {
	jwtConfig := config.GetJWTConfig()
	if jwtConfig == nil {
		return "", errors.New("JWT配置未初始化")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("签名方法不正确")
		}
		return []byte(jwtConfig.Secret), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if username, ok := claims["username"].(string); ok {
			return username, nil
		}
		return "", errors.New("用户名不存在")
	}
	return "", errors.New("无效的token")
}
