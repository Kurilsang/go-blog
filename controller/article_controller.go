package controller

import (
	"go_test/global"
	"go_test/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	var articles []model.Article
	if err := global.DB.Find(&articles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
	ctx.JSON(http.StatusOK, vos)
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
