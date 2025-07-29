package service

import (
	"go_test/dto"
	"go_test/global"
	"go_test/model"
	"time"
)

type ExchangeRateService struct{}

func NewExchangeRateService() *ExchangeRateService {
	return &ExchangeRateService{}
}

// CreateExchangeRate 创建汇率业务逻辑
func (s *ExchangeRateService) CreateExchangeRate(req dto.ExchangeRateRequest) (*dto.ExchangeRateVO, error) {
	rate := model.ExchangeRate{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Rate:         req.Rate,
		Date:         time.Now(),
	}

	if err := global.DB.Create(&rate).Error; err != nil {
		return nil, err
	}

	return &dto.ExchangeRateVO{
		ID:           rate.ID,
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		Rate:         rate.Rate,
		Date:         rate.Date.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetExchangeRates 获取所有汇率业务逻辑
func (s *ExchangeRateService) GetExchangeRates() ([]dto.ExchangeRateVO, error) {
	var rates []model.ExchangeRate
	if err := global.DB.Find(&rates).Error; err != nil {
		return nil, err
	}

	vos := make([]dto.ExchangeRateVO, 0, len(rates))
	for _, r := range rates {
		vos = append(vos, dto.ExchangeRateVO{
			ID:           r.ID,
			FromCurrency: r.FromCurrency,
			ToCurrency:   r.ToCurrency,
			Rate:         r.Rate,
			Date:         r.Date.Format("2006-01-02 15:04:05"),
		})
	}

	return vos, nil
}
