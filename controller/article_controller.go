package controller

import (
	"go_test/dto"
	"go_test/service"
	"go_test/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var articleService = service.NewArticleService()

// 创建文章
func CreateArticle(ctx *gin.Context) {
	var req dto.ArticleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article, err := articleService.CreateArticle(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, article)
}

// 获取所有文章
func GetArticles(ctx *gin.Context) {
	articles, err := articleService.GetAllArticles()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, articles)
}

// 分页查询文章 - 支持可选关键词搜索（带Redis缓存）
func GetArticlesWithPagination(ctx *gin.Context) {
	// 使用分页工具从上下文解析分页参数
	paginate := utils.PaginateFromContext(ctx)

	// 获取可选的关键词参数
	keyword := ctx.Query("keyword")

	response, err := articleService.GetArticlesWithPagination(paginate, keyword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// 批量删除文章
func BatchDeleteArticles(ctx *gin.Context) {
	var req struct {
		IDs        []uint `json:"ids" binding:"required"`
		HardDelete bool   `json:"hard_delete,omitempty"` // 是否硬删除，默认false（软删除）
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	err := articleService.BatchDeleteArticles(req.IDs, req.HardDelete)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deleteType := "软删除"
	if req.HardDelete {
		deleteType = "硬删除"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":       "成功" + deleteType + " " + strconv.Itoa(len(req.IDs)) + "篇文章",
		"delete_type":   deleteType,
		"deleted_ids":   req.IDs,
		"deleted_count": len(req.IDs),
		"note": map[string]string{
			"soft_delete": "软删除：数据仍在数据库中，只是标记为已删除，可以恢复",
			"hard_delete": "硬删除：数据从数据库中彻底删除，无法恢复",
		},
	})
}

// 根据ID获取文章
func GetArticleByID(ctx *gin.Context) {
	id := ctx.Param("id")

	article, err := articleService.GetArticleByID(id)
	if err != nil {
		if err.Error() == "未找到该文章" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, article)
}
