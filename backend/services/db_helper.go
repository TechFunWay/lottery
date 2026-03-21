package services

import (
	"lottery-backend/database"

	"gorm.io/gorm"
)

// GetDB 获取数据库实例（供 handlers 使用）
func GetDB() *gorm.DB {
	return database.DB
}
