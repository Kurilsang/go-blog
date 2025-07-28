package controller

import (
	"context"
	"crypto/md5"
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

// generatePaginationCacheKey 生成分页查询的缓存键
func generatePaginationCacheKey(page, pageSize int, order, keyword string) string {
	// 构建缓存键字符串
	keyStr := fmt.Sprintf("page:%d_size:%d_order:%s_keyword:%s", page, pageSize, order, keyword)
	// 使用MD5哈希生成短键名
	hash := md5.Sum([]byte(keyStr))
	return fmt.Sprintf("%s:%x", global.CacheKeyArticlesPagination, hash)
}

// clearPaginationCache 清除分页相关的所有缓存
func clearPaginationCache() {
	// 使用模式匹配删除所有分页缓存
	pattern := global.CacheKeyArticlesPagination + ":*"
	keys, err := global.RedisDB.Keys(articleCtxRedis, pattern).Result()
	if err != nil {
		fmt.Printf("获取缓存键失败: %v\n", err)
		return
	}
	if len(keys) > 0 {
		if err := global.RedisDB.Del(articleCtxRedis, keys...).Err(); err != nil {
			fmt.Printf("清除分页缓存失败: %v\n", err)
		}
	}
}

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

	// 清除原有的全量文章缓存和分页缓存
	if err := global.RedisDB.Del(articleCtxRedis, cacheKey).Err(); err != nil {
		fmt.Printf("警告: 创建文章成功但清除全量缓存失败: %v\n", err)
	}
	// 清除所有分页相关缓存
	clearPaginationCache()

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

// 分页查询文章 - 支持可选关键词搜索（带Redis缓存）
func GetArticlesWithPagination(ctx *gin.Context) {
	// 使用分页工具从上下文解析分页参数
	paginate := utils.PaginateFromContext(ctx)

	// 获取可选的关键词参数
	keyword := ctx.Query("keyword")
	keyword = strings.TrimSpace(keyword)

	// 生成缓存键
	cacheKey := generatePaginationCacheKey(paginate.Page, paginate.PageSize, paginate.Order, keyword)

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
			// 构建查询
			query := global.DB.Model(&model.Article{})

			// 如果有关键词，添加搜索条件
			if keyword != "" {
				searchPattern := "%" + keyword + "%"
				query = query.Where("title LIKE ? OR content LIKE ?", searchPattern, searchPattern)
			}

			// 执行分页查询
			var articles []model.Article
			if err := utils.PaginateWithCondition(query, paginate, &articles); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// 转换为VO
			vos := make([]ArticleVO, 0, len(articles))
			for _, a := range articles {
				// 如果有关键词，进行高亮处理
				title := a.Title
				content := a.Content
				if keyword != "" {
					title = highlightKeyword(a.Title, keyword)
					content = highlightKeyword(a.Content, keyword)
				}

				vos = append(vos, ArticleVO{
					ID:      a.ID,
					Title:   title,
					Content: content,
					Preview: a.Preview,
					Created: a.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}

			// 构建响应
			response := gin.H{
				"articles":   vos,
				"pagination": paginate.GetPaginationInfo(),
			}

			// 如果有关键词，添加搜索相关信息
			if keyword != "" {
				response["keyword"] = keyword
				response["is_search"] = true
			}

			// 将数据存入缓存
			responseJSON, err := json.Marshal(response)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "序列化响应失败: " + err.Error()})
				return
			}

			// 设置缓存，使用全局配置的过期时间
			if err := global.RedisDB.Set(articleCtxRedis, cacheKey, responseJSON, time.Duration(global.CacheExpireArticles)*time.Second).Err(); err != nil {
				// 缓存写入失败不影响业务，只记录日志
				fmt.Printf("写入缓存失败: %v\n", err)
			}

			ctx.JSON(http.StatusOK, response)
			return
		}
	} else if err != nil {
		// Redis连接错误，直接查数据库
		fmt.Printf("Redis连接错误: %v\n", err)
		// 降级到数据库查询
		query := global.DB.Model(&model.Article{})
		if keyword != "" {
			searchPattern := "%" + keyword + "%"
			query = query.Where("title LIKE ? OR content LIKE ?", searchPattern, searchPattern)
		}

		var articles []model.Article
		if err := utils.PaginateWithCondition(query, paginate, &articles); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		vos := make([]ArticleVO, 0, len(articles))
		for _, a := range articles {
			title := a.Title
			content := a.Content
			if keyword != "" {
				title = highlightKeyword(a.Title, keyword)
				content = highlightKeyword(a.Content, keyword)
			}

			vos = append(vos, ArticleVO{
				ID:      a.ID,
				Title:   title,
				Content: content,
				Preview: a.Preview,
				Created: a.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		response := gin.H{
			"articles":   vos,
			"pagination": paginate.GetPaginationInfo(),
		}

		if keyword != "" {
			response["keyword"] = keyword
			response["is_search"] = true
		}

		ctx.JSON(http.StatusOK, response)
		return
	}

	// 缓存命中，直接返回缓存数据
	var response gin.H
	if err := json.Unmarshal([]byte(cachedData), &response); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "反序列化缓存数据失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
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
		fmt.Printf("警告: 删除文章成功但清除全量缓存失败: %v\n", err)
	}
	// 清除所有分页相关缓存
	clearPaginationCache()

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
