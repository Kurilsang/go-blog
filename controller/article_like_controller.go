package controller

import (
	"go_test/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

var articleLikeService = service.NewArticleLikeService()

// LikeArticle 给文章点赞
func LikeArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")

	err := articleLikeService.LikeArticle(articleID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully liked the article"})
}

// GetArticleLikes 获取文章点赞数量
func GetArticleLikes(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likes, err := articleLikeService.GetArticleLikes(articleID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"likes": likes})
}
