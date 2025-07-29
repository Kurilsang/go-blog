package service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"go_test/dto"
	"go_test/global"
	"go_test/model"
	"go_test/utils"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var (
	cacheKey        = "articles"
	articleCtxRedis = context.Background()
	cacheMutex      sync.RWMutex
)

type ArticleService struct{}

func NewArticleService() *ArticleService {
	return &ArticleService{}
}

// CreateArticle 创建文章业务逻辑
func (s *ArticleService) CreateArticle(req dto.ArticleRequest) (*dto.ArticleVO, error) {
	article := model.Article{
		Title:   req.Title,
		Content: req.Content,
		Preview: req.Preview,
	}

	if err := global.DB.Create(&article).Error; err != nil {
		return nil, err
	}

	// 清除缓存
	s.clearAllCache()

	return &dto.ArticleVO{
		ID:      article.ID,
		Title:   article.Title,
		Content: article.Content,
		Preview: article.Preview,
		Created: article.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetAllArticles 获取所有文章业务逻辑
func (s *ArticleService) GetAllArticles() ([]dto.ArticleVO, error) {
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
				return nil, err
			}

			// 转换为VO
			vos := make([]dto.ArticleVO, 0, len(articles))
			for _, a := range articles {
				vos = append(vos, dto.ArticleVO{
					ID:      a.ID,
					Title:   a.Title,
					Content: a.Content,
					Preview: a.Preview,
					Created: a.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}

			// 将数据存入缓存
			articleJSON, err := json.Marshal(vos)
			if err != nil {
				return nil, err
			}

			if err := global.RedisDB.Set(articleCtxRedis, cacheKey, articleJSON, time.Duration(global.CacheExpireArticles)*time.Second).Err(); err != nil {
				return nil, err
			}

			return vos, nil
		}
	} else if err != nil {
		return nil, err
	}

	// 缓存命中，直接返回缓存数据
	var vos []dto.ArticleVO
	if err := json.Unmarshal([]byte(cachedData), &vos); err != nil {
		return nil, err
	}
	return vos, nil
}

// GetArticlesWithPagination 分页查询文章业务逻辑
func (s *ArticleService) GetArticlesWithPagination(paginate *utils.Paginate, keyword string) (map[string]interface{}, error) {
	keyword = strings.TrimSpace(keyword)

	// 生成缓存键
	cacheKey := s.generatePaginationCacheKey(paginate.Page, paginate.PageSize, paginate.Order, keyword)

	// 先尝试读缓存
	cachedData, err := global.RedisDB.Get(articleCtxRedis, cacheKey).Result()

	if err == redis.Nil {
		// 缓存未命中，获取写锁防止缓存击穿
		cacheMutex.Lock()
		defer cacheMutex.Unlock()

		// 双重检查
		cachedData, err = global.RedisDB.Get(articleCtxRedis, cacheKey).Result()
		if err == redis.Nil {
			// 从数据库查询
			query := global.DB.Model(&model.Article{})

			// 如果有关键词，添加搜索条件
			if keyword != "" {
				searchPattern := "%" + keyword + "%"
				query = query.Where("title LIKE ? OR content LIKE ?", searchPattern, searchPattern)
			}

			// 执行分页查询
			var articles []model.Article
			if err := utils.PaginateWithCondition(query, paginate, &articles); err != nil {
				return nil, err
			}

			// 转换为VO
			vos := make([]dto.ArticleVO, 0, len(articles))
			for _, a := range articles {
				title := a.Title
				content := a.Content
				if keyword != "" {
					title = s.highlightKeyword(a.Title, keyword)
					content = s.highlightKeyword(a.Content, keyword)
				}

				vos = append(vos, dto.ArticleVO{
					ID:      a.ID,
					Title:   title,
					Content: content,
					Preview: a.Preview,
					Created: a.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}

			// 构建响应
			response := map[string]interface{}{
				"articles":   vos,
				"pagination": paginate.GetPaginationInfo(),
			}

			if keyword != "" {
				response["keyword"] = keyword
				response["is_search"] = true
			}

			// 将数据存入缓存
			responseJSON, err := json.Marshal(response)
			if err != nil {
				return nil, fmt.Errorf("序列化响应失败: %v", err)
			}

			if err := global.RedisDB.Set(articleCtxRedis, cacheKey, responseJSON, time.Duration(global.CacheExpireArticles)*time.Second).Err(); err != nil {
				fmt.Printf("写入缓存失败: %v\n", err)
			}

			return response, nil
		}
	} else if err != nil {
		// Redis连接错误，降级到数据库查询
		fmt.Printf("Redis连接错误: %v\n", err)
		return s.fallbackDatabaseQuery(paginate, keyword)
	}

	// 缓存命中
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(cachedData), &response); err != nil {
		return nil, fmt.Errorf("反序列化缓存数据失败: %v", err)
	}
	return response, nil
}

// BatchDeleteArticles 批量删除文章业务逻辑
func (s *ArticleService) BatchDeleteArticles(ids []uint, hardDelete bool) error {
	if len(ids) == 0 {
		return fmt.Errorf("删除ID列表不能为空")
	}

	// 验证ID是否都存在
	var count int64
	query := global.DB.Model(&model.Article{})
	if !hardDelete {
		query = query.Where("deleted_at IS NULL")
	} else {
		query = query.Unscoped()
	}

	if err := query.Where("id IN ?", ids).Count(&count).Error; err != nil {
		return err
	}

	if int(count) != len(ids) {
		deleteType := "软删除"
		if hardDelete {
			deleteType = "硬删除"
		}
		return fmt.Errorf("部分文章不存在或已被删除，请求%s %d个，实际可%s %d个", deleteType, len(ids), deleteType, count)
	}

	// 执行删除操作
	var deleteQuery *gorm.DB
	if hardDelete {
		deleteQuery = global.DB.Unscoped().Where("id IN ?", ids).Delete(&model.Article{})
	} else {
		deleteQuery = global.DB.Where("id IN ?", ids).Delete(&model.Article{})
	}

	if err := deleteQuery.Error; err != nil {
		deleteType := "硬删除"
		if !hardDelete {
			deleteType = "软删除"
		}
		return fmt.Errorf("%s失败: %s", deleteType, err.Error())
	}

	// 清除缓存
	s.clearAllCache()

	return nil
}

// GetArticleByID 根据ID获取文章业务逻辑
func (s *ArticleService) GetArticleByID(id string) (*dto.ArticleVO, error) {
	var article model.Article
	if err := global.DB.Where("id = ?", id).First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到该文章")
		}
		return nil, err
	}

	return &dto.ArticleVO{
		ID:      article.ID,
		Title:   article.Title,
		Content: article.Content,
		Preview: article.Preview,
		Created: article.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// generatePaginationCacheKey 生成分页查询的缓存键
func (s *ArticleService) generatePaginationCacheKey(page, pageSize int, order, keyword string) string {
	keyStr := fmt.Sprintf("page:%d_size:%d_order:%s_keyword:%s", page, pageSize, order, keyword)
	hash := md5.Sum([]byte(keyStr))
	return fmt.Sprintf("%s:%x", global.CacheKeyArticlesPagination, hash)
}

// clearAllCache 清除所有缓存
func (s *ArticleService) clearAllCache() {
	// 清除全量缓存
	if err := global.RedisDB.Del(articleCtxRedis, cacheKey).Err(); err != nil {
		fmt.Printf("警告: 清除全量缓存失败: %v\n", err)
	}
	// 清除分页缓存
	s.clearPaginationCache()
}

// clearPaginationCache 清除分页相关的所有缓存
func (s *ArticleService) clearPaginationCache() {
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

// highlightKeyword 高亮关键词
func (s *ArticleService) highlightKeyword(text, keyword string) string {
	if keyword == "" {
		return text
	}

	lowerText := strings.ToLower(text)
	lowerKeyword := strings.ToLower(keyword)

	if !strings.Contains(lowerText, lowerKeyword) {
		return text
	}

	index := strings.Index(lowerText, lowerKeyword)
	if index == -1 {
		return text
	}

	originalKeyword := text[index : index+len(keyword)]
	return strings.Replace(text, originalKeyword, "**"+originalKeyword+"**", -1)
}

// fallbackDatabaseQuery 数据库降级查询
func (s *ArticleService) fallbackDatabaseQuery(paginate *utils.Paginate, keyword string) (map[string]interface{}, error) {
	query := global.DB.Model(&model.Article{})
	if keyword != "" {
		searchPattern := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", searchPattern, searchPattern)
	}

	var articles []model.Article
	if err := utils.PaginateWithCondition(query, paginate, &articles); err != nil {
		return nil, err
	}

	vos := make([]dto.ArticleVO, 0, len(articles))
	for _, a := range articles {
		title := a.Title
		content := a.Content
		if keyword != "" {
			title = s.highlightKeyword(a.Title, keyword)
			content = s.highlightKeyword(a.Content, keyword)
		}

		vos = append(vos, dto.ArticleVO{
			ID:      a.ID,
			Title:   title,
			Content: content,
			Preview: a.Preview,
			Created: a.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response := map[string]interface{}{
		"articles":   vos,
		"pagination": paginate.GetPaginationInfo(),
	}

	if keyword != "" {
		response["keyword"] = keyword
		response["is_search"] = true
	}

	return response, nil
}
