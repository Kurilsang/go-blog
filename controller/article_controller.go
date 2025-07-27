package controller

import (
	"context"
	"encoding/json"
	"go_test/global"
	"go_test/model"
	"net/http"
	"time"

	"go_test/config"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var cacheKey = "articles"
var articleCtxRedis = context.Background()

// 创建文章
func CreateArticle(ctx *gin.Context) {
	var req ArticleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article := model.Article{
		Title:   req.Title,
		Content: req.Content,
		Preview: req.Preview,
	}

	if err := global.DB.Create(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 清除缓存
	if err := global.RedisDB.Del(articleCtxRedis, cacheKey).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, ArticleVO{
		ID:      article.ID,
		Title:   article.Title,
		Content: article.Content,
		Preview: article.Preview,
		Created: article.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// 获取所有文章
func GetArticles(ctx *gin.Context) {
	cachedData, err := global.RedisDB.Get(articleCtxRedis, cacheKey).Result()

	if err == redis.Nil {
		// 缓存未命中，从数据库查询
		var articles []model.Article
		if err := global.DB.Find(&articles).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 转换为VO
		vos := make([]ArticleVO, 0, len(articles))
		for _, a := range articles {
			vos = append(vos, ArticleVO{
				ID:      a.ID,
				Title:   a.Title,
				Content: a.Content,
				Preview: a.Preview,
				Created: a.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		// 将数据存入缓存，设置过期时间
		articleJSON, err := json.Marshal(vos)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := global.RedisDB.Set(articleCtxRedis, cacheKey, articleJSON, time.Duration(config.CacheConf.ArticleExpire)*time.Second).Err(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, vos)
	} else if err != nil {
		// Redis连接错误
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		// 缓存命中，直接返回缓存数据
		var vos []ArticleVO
		if err := json.Unmarshal([]byte(cachedData), &vos); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, vos)
	}
}

// 根据ID获取文章
func GetArticleByID(ctx *gin.Context) {
	id := ctx.Param("id")
	var article model.Article
	if err := global.DB.Where("id = ?", id).First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "未找到该文章"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	ctx.JSON(http.StatusOK, ArticleVO{
		ID:      article.ID,
		Title:   article.Title,
		Content: article.Content,
		Preview: article.Preview,
		Created: article.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}
