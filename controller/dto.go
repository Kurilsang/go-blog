package controller

// Auth相关

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	Message  string `json:"message"`
}

// 汇率相关

type ExchangeRateRequest struct {
	FromCurrency string  `json:"fromCurrency" binding:"required"`
	ToCurrency   string  `json:"toCurrency" binding:"required"`
	Rate         float64 `json:"rate" binding:"required"`
}

type ExchangeRateVO struct {
	ID           uint    `json:"_id"`
	FromCurrency string  `json:"fromCurrency"`
	ToCurrency   string  `json:"toCurrency"`
	Rate         float64 `json:"rate"`
	Date         string  `json:"date"`
}

// 文章相关

type ArticleRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Preview string `json:"preview" binding:"required"`
}

type ArticleVO struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Preview string `json:"preview"`
	Created string `json:"created_at"`
}
