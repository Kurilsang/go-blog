package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null" json:"username"`
	Password string `gorm:"not null" json:"password"`
	Role     string `gorm:"not null;default:'user'" json:"role"`     // admin 或 user
	Status   string `gorm:"not null;default:'active'" json:"status"` // active 或 disabled
}
