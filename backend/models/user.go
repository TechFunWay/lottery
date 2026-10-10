package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint       `json:"id" gorm:"primarykey"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Username  string     `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Password  string     `json:"-" gorm:"not null"` // 不返回给前端
	Email     string     `json:"email" gorm:"size:100"`
	Role      string     `json:"role" gorm:"default:'user'"` // admin/user
	Status    string     `json:"status" gorm:"default:'active'"` // active/disabled
	LastLogin *time.Time `json:"last_login"`
}
