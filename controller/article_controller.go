package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"go_test/global"
	"go_test/model"
	"go_test/utils"
	"net/http"
	"strings"
	"sync"
	"time"

	"go_test/config"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var (
	cacheKey        = "articles"
	articleCtxRedis = context.Background()
	cacheMutex      sync.RWMutex
)

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
	// 先尝试读缓存
	cachedData, err := global.RedisDB.Get(articleCtxRedis, cacheKey).Result()

	if err == redis.Nil {
		// 缓存未命中，获取写锁防止缓存击穿
		cacheMutex.Lock()
		defer cacheMutex.Unlock()

		// 双重检查，防止在获取锁期间其他goroutine已经更新了缓存
		cachedData, err = global.RedisDB.Get(articleCtxRedis, cacheKey).Result()
		if err == redis.Nil {
			// 缓存仍未命中，从数据库查询
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

			if err := global.RedisDB.Set(articleCtxRedis, cacheKey, articleJSON, time.Duration(config.GetCacheConfig().ArticleExpire)*time.Second).Err(); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			ctx.JSON(http.StatusOK, vos)
			return
		}
	} else if err != nil {
		// Redis连接错误
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 缓存命中，直接返回缓存数据
	var vos []ArticleVO
	if err := json.Unmarshal([]byte(cachedData), &vos); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, vos)
}

// 分页查询文章 - 使用新的分页工具
func GetArticlesWithPagination(ctx *gin.Context) {
	// 使用分页工具从上下文解析分页参数
	paginate := utils.PaginateFromContext(ctx)

	// 查询分页数据
	var articles []model.Article
	if err := utils.PaginateWithTotal(global.DB, &model.Article{}, paginate, &articles); err != nil {
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

	ctx.JSON(http.StatusOK, gin.H{
		"articles":   vos,
		"pagination": paginate.GetPaginationInfo(),
	})
}

// 搜索文章
func SearchArticles(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	if keyword == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	// 去除关键词首尾空格
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	// 构建搜索条件，使用LIKE进行模糊搜索
	searchPattern := "%" + keyword + "%"

	// 查询标题匹配的文章（优先级高）
	var titleMatchedArticles []model.Article
	if err := global.DB.Where("title LIKE ?", searchPattern).Find(&titleMatchedArticles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查询内容匹配但标题不匹配的文章（优先级低）
	var contentMatchedArticles []model.Article
	if err := global.DB.Where("content LIKE ? AND title NOT LIKE ?", searchPattern, searchPattern).Find(&contentMatchedArticles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 合并结果，标题匹配的排在前面
	allArticles := make([]model.Article, 0, len(titleMatchedArticles)+len(contentMatchedArticles))
	allArticles = append(allArticles, titleMatchedArticles...)
	allArticles = append(allArticles, contentMatchedArticles...)

	// 转换为VO
	vos := make([]ArticleVO, 0, len(allArticles))
	for _, a := range allArticles {
		// 高亮关键词（简单实现）
		highlightedTitle := highlightKeyword(a.Title, keyword)
		highlightedContent := highlightKeyword(a.Content, keyword)

		vos = append(vos, ArticleVO{
			ID:      a.ID,
			Title:   highlightedTitle,
			Content: highlightedContent,
			Preview: a.Preview,
			Created: a.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"keyword":         keyword,
		"total":           len(vos),
		"title_matched":   len(titleMatchedArticles),
		"content_matched": len(contentMatchedArticles),
		"articles":        vos,
	})
}

// 批量删除文章
func BatchDeleteArticles(ctx *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	if len(req.IDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "删除ID列表不能为空"})
		return
	}

	// 验证ID是否都存在
	var count int64
	if err := global.DB.Model(&model.Article{}).Where("id IN ?", req.IDs).Count(&count).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if int(count) != len(req.IDs) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("部分文章不存在，请求删除%d个，实际存在%d个", len(req.IDs), count)})
		return
	}

	// 执行批量删除
	if err := global.DB.Where("id IN ?", req.IDs).Delete(&model.Article{}).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}

	// 清除缓存以确保数据一致性
	if err := global.RedisDB.Del(articleCtxRedis, cacheKey).Err(); err != nil {
		// 删除成功但清除缓存失败，记录错误但不影响删除操作
		fmt.Printf("警告: 删除文章成功但清除缓存失败: %v\n", err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("成功删除%d篇文章", len(req.IDs)),
		"deleted_ids":   req.IDs,
		"deleted_count": len(req.IDs),
	})
}

// 高亮关键词（简单实现，用**包围关键词）
func highlightKeyword(text, keyword string) string {
	if keyword == "" {
		return text
	}

	// 不区分大小写的替换
	lowerText := strings.ToLower(text)
	lowerKeyword := strings.ToLower(keyword)

	if !strings.Contains(lowerText, lowerKeyword) {
		return text
	}

	// 找到关键词在原文中的位置
	index := strings.Index(lowerText, lowerKeyword)
	if index == -1 {
		return text
	}

	// 提取原文中的关键词（保持原大小写）
	originalKeyword := text[index : index+len(keyword)]

	// 替换为高亮格式
	return strings.Replace(text, originalKeyword, "**"+originalKeyword+"**", -1)
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
