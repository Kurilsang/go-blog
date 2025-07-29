package dto

// Auth相关

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	Token    string `json:"token"`
	Message  string `json:"message"`
}

// 用户资料相关

// UpdateProfileRequest 更新用户资料请求DTO
type UpdateProfileRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`     // 邮箱，支持验证格式
	Avatar   string `json:"avatar"`                              // 头像URL
	Nickname string `json:"nickname" binding:"omitempty,max=50"` // 昵称，最大50字符
	Bio      string `json:"bio" binding:"omitempty,max=500"`     // 个人简介，最大500字符
	Phone    string `json:"phone" binding:"omitempty,max=20"`    // 电话号码，最大20字符
}

// ChangePasswordRequest 修改密码请求DTO
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6"` // 旧密码，必填，最少6位
	NewPassword string `json:"new_password" binding:"required,min=6"` // 新密码，必填，最少6位
}

// UserProfileVO 用户资料响应DTO
type UserProfileVO struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Created  string `json:"created_at"`
	Updated  string `json:"updated_at"`
}

// AdminUpdateUserRequest 管理员更新用户请求DTO（包含状态和角色）
type AdminUpdateUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname" binding:"omitempty,max=50"`
	Bio      string `json:"bio" binding:"omitempty,max=500"`
	Phone    string `json:"phone" binding:"omitempty,max=20"`
	Role     string `json:"role" binding:"omitempty,oneof=admin user"`        // 管理员可以修改角色
	Status   string `json:"status" binding:"omitempty,oneof=active disabled"` // 管理员可以修改状态
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
