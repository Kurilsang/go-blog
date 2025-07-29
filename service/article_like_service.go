package service

import (
	"context"
	"go_test/global"

	"github.com/go-redis/redis/v8"
)

var likeCtxRedis = context.Background()

type ArticleLikeService struct{}

func NewArticleLikeService() *ArticleLikeService {
	return &ArticleLikeService{}
}

// LikeArticle 给文章点赞业务逻辑
func (s *ArticleLikeService) LikeArticle(articleID string) error {
	likeKey := "article:" + articleID + ":likes"

	if err := global.RedisDB.Incr(likeCtxRedis, likeKey).Err(); err != nil {
		return err
	}

	return nil
}

// GetArticleLikes 获取文章点赞数量业务逻辑
func (s *ArticleLikeService) GetArticleLikes(articleID string) (string, error) {
	likeKey := "article:" + articleID + ":likes"

	likes, err := global.RedisDB.Get(likeCtxRedis, likeKey).Result()

	if err == redis.Nil {
		return "0", nil
	} else if err != nil {
		return "", err
	}

	return likes, nil
}
