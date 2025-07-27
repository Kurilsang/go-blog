#  GoBlog - 高性能博客系统

基于 Go 语言和 Gin 框架构建的高性能博客系统，采用现代化技术栈，支持文章管理、用户认证、汇率查询等功能。

## ✨ 技术特性

- 🔐 **JWT 认证** - 基于 JWT 的无状态用户认证
- 🚀 **高性能缓存** - Redis 缓存策略，防止缓存击穿
- 🛡️ **并发安全** - 原子操作和双重检查锁定模式
- 🎯 **RESTful API** - 标准化的 REST API 设计
- 🔧 **配置管理** - 基于 Viper 的灵活配置系统
- 🗄️ **数据库 ORM** - GORM 数据库操作和自动迁移
- 🌐 **跨域支持** - CORS 中间件支持前端跨域请求

## 🏗️ 技术架构

### 核心技术栈

| 技术 | 版本      | 用途 |
|------|---------|------|
| **Go** | 1.24.5  | 后端开发语言 |
| **Gin** | v1.10.1 | Web 框架 |
| **GORM** | v1.30.0 | ORM 框架 |
| **Redis** | v8.11.5 | 缓存数据库 |
| **MySQL** | v5.9+   | 主数据库 |
| **JWT** | v4.5.2  | 用户认证 |
| **Viper** | v1.20.1 | 配置管理 |

## 🔧 项目结构

```
go_test/
├── 📁 config/           # 配置管理
├── 📁 controller/       # 控制器层
├── 📁 model/           # 数据模型
├── 📁 middleware/      # 中间件
├── 📁 router/          # 路由管理
├── 📁 utils/           # 工具函数
├── 📁 global/          # 全局变量
├── 📁 sql/             # SQL脚本
└── main.go            # 应用入口
```

## 🚀 核心功能

### 🔐 用户认证系统
- 基于 JWT 的无状态认证
- Bcrypt 密码加密
- 中间件级别的权限控制

### 📝 文章管理系统
- 文章的 CRUD 操作
- Redis 缓存策略
- 防缓存击穿机制

### 👍 点赞系统
- Redis 原子操作
- 实时点赞计数
- 高性能点赞处理

### 💱 汇率查询系统
- 汇率数据管理
- 实时汇率查询
- 数据持久化存储

## 🔒 并发安全设计

### 原子操作配置
```go
// 使用 atomic.Value 避免互斥锁性能损失
var appConfig atomic.Value
func GetAppConfig() *Config {
    return appConfig.Load().(*Config)
}
```

### 缓存击穿防护
```go
// 双重检查锁定模式
cacheMutex.Lock()
defer cacheMutex.Unlock()
// 双重检查防止重复查询数据库
```

### 单次初始化保证
```go
// sync.Once 确保全局资源只初始化一次
var dbOnce sync.Once
func InitDB(db *gorm.DB) {
    dbOnce.Do(func() {
        DB = db
    })
}
```

## 🛠️ 快速开始

1. **克隆项目**
```bash
git clone https://github.com/Kurilsang/go-blog.git
cd goblog
```

2. **安装依赖**
```bash
go mod tidy
```

3**启动数据库**
```bash
启动mysql及redis
```

3. **配置环境**
```bash
# 修改 config/config.yml 配置文件

```

4. **启动服务**
```bash
go run main.go
```

## 📊 性能优化

- **缓存策略**: Redis 缓存减少数据库查询
- **连接池**: 数据库和 Redis 连接池优化
- **原子操作**: 无锁配置读取，高性能并发
- **中间件优化**: 最小化中间件开销

## 📄 许可证

本项目采用 MIT 许可证

---

⭐ 如果这个项目对你有帮助，请给它一个星标！ 
