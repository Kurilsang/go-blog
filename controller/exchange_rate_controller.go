package controller

import (
	"go_test/global"
	"go_test/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 创建汇率
func CreateExchangeRate(ctx *gin.Context) {
	var req ExchangeRateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rate := model.ExchangeRate{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Rate:         req.Rate,
		Date:         time.Now(),
	}

	if err := global.DB.Create(&rate).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, ExchangeRateVO{
		ID:           rate.ID,
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		Rate:         rate.Rate,
		Date:         rate.Date.Format("2006-01-02 15:04:05"),
	})
}

// 获取所有汇率
func GetExchangeRates(ctx *gin.Context) {
	var rates []model.ExchangeRate
	if err := global.DB.Find(&rates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	vos := make([]ExchangeRateVO, 0, len(rates))
	for _, r := range rates {
		vos = append(vos, ExchangeRateVO{
			ID:           r.ID,
			FromCurrency: r.FromCurrency,
			ToCurrency:   r.ToCurrency,
			Rate:         r.Rate,
			Date:         r.Date.Format("2006-01-02 15:04:05"),
		})
	}
	ctx.JSON(http.StatusOK, vos)
}
