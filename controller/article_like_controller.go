package controller

import (
	"go_test/global"
	"net/http"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var ctxRedis = context.Background()

// LikeArticle 给文章点赞
func LikeArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")
	likeKey := "article:" + articleID + ":likes"

	if err := global.RedisDB.Incr(ctxRedis, likeKey).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully liked the article"})
}

// GetArticleLikes 获取文章点赞数量
func GetArticleLikes(ctx *gin.Context) {
	articleID := ctx.Param("id")
	likeKey := "article:" + articleID + ":likes"

	likes, err := global.RedisDB.Get(ctxRedis, likeKey).Result()

	if err == redis.Nil {
		likes = "0"
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"likes": likes})
}
