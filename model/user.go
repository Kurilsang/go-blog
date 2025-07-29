package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null" json:"username"`
	Password string `gorm:"not null" json:"password"`
	Role     string `gorm:"not null;default:'user'" json:"role"`     // admin 或 user
	Status   string `gorm:"not null;default:'active'" json:"status"` // active 或 disabled

	// 新增用户资料字段
	Email    string `gorm:"unique;size:100" json:"email"` // 邮箱，唯一
	Avatar   string `gorm:"size:255" json:"avatar"`       // 头像URL
	Nickname string `gorm:"size:50" json:"nickname"`      // 昵称
	Bio      string `gorm:"type:text" json:"bio"`         // 个人简介
	Phone    string `gorm:"size:20" json:"phone"`         // 电话号码
}
