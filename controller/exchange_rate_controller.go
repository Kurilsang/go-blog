package controller

import (
	"go_test/dto"
	"go_test/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

var exchangeRateService = service.NewExchangeRateService()

// 创建汇率
func CreateExchangeRate(ctx *gin.Context) {
	var req dto.ExchangeRateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rate, err := exchangeRateService.CreateExchangeRate(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, rate)
}

// 获取所有汇率
func GetExchangeRates(ctx *gin.Context) {
	rates, err := exchangeRateService.GetExchangeRates()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, rates)
}
