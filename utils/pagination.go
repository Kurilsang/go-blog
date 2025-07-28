package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Paginate GORM分页scope
type Paginate struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Total    int64  `json:"total"`
	Order    string `json:"order,omitempty"` // 排序字段
}

// NewPaginate 创建分页参数
func NewPaginate(page, pageSize int) *Paginate {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	return &Paginate{
		Page:     page,
		PageSize: pageSize,
	}
}

// PaginateFromContext 从gin.Context中解析分页参数
func PaginateFromContext(ctx *gin.Context) *Paginate {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	order := ctx.DefaultQuery("order", "created_at DESC")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return &Paginate{
		Page:     page,
		PageSize: pageSize,
		Order:    order,
	}
}

// Scope 返回GORM scope函数
func (p *Paginate) Scope() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (p.Page - 1) * p.PageSize
		if p.Order != "" {
			db = db.Order(p.Order)
		}
		return db.Offset(offset).Limit(p.PageSize)
	}
}

// GetPaginationInfo 获取分页信息
func (p *Paginate) GetPaginationInfo() map[string]interface{} {
	totalPages := (int(p.Total) + p.PageSize - 1) / p.PageSize
	hasNext := p.Page < totalPages
	hasPrev := p.Page > 1

	return map[string]interface{}{
		"page":        p.Page,
		"page_size":   p.PageSize,
		"total":       p.Total,
		"total_pages": totalPages,
		"has_next":    hasNext,
		"has_prev":    hasPrev,
	}
}

// PaginateWithTotal 带总数查询的分页
func PaginateWithTotal(db *gorm.DB, model interface{}, paginate *Paginate, dest interface{}) error {
	// 计算总数
	if err := db.Model(model).Count(&paginate.Total).Error; err != nil {
		return err
	}

	// 查询分页数据
	return db.Scopes(paginate.Scope()).Find(dest).Error
}

// PaginateWithCondition 带条件的分页查询
func PaginateWithCondition(db *gorm.DB, paginate *Paginate, dest interface{}) error {
	// 先计算总数（保持原有的where条件）
	countDB := db.Session(&gorm.Session{})
	if err := countDB.Count(&paginate.Total).Error; err != nil {
		return err
	}

	// 查询分页数据
	return db.Scopes(paginate.Scope()).Find(dest).Error
}
