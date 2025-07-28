package global

// 缓存相关常量
const (
	// 缓存前缀
	CachePrefix = "goblog:"

	// 文章相关缓存键
	CacheKeyArticles           = CachePrefix + "articles"
	CacheKeyArticlesPagination = CachePrefix + "articles:pagination"
	CacheKeyArticleDetail      = CachePrefix + "article:detail"

	// 用户相关缓存键
	CacheKeyUser = CachePrefix + "user"

	// 汇率相关缓存键
	CacheKeyExchangeRate = CachePrefix + "exchange_rate"

	// 点赞相关缓存键
	CacheKeyArticleLikes = CachePrefix + "article:likes"
)

// 缓存过期时间（秒）
const (
	CacheExpireDefault      = 300  // 5分钟
	CacheExpireArticles     = 600  // 10分钟
	CacheExpireUserInfo     = 1800 // 30分钟
	CacheExpireExchangeRate = 3600 // 1小时
)

// 分页相关常量
const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// 其他业务常量
const (
	DefaultOrder = "created_at DESC"
)

// 用户角色常量
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// 用户状态常量
const (
	UserStatusActive   = "active"
	UserStatusDisabled = "disabled"
)
