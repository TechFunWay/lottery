package database

import (
	"lottery-backend/logger"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		logger.GetSugarLogger().Fatalf("failed to connect database: %v", err)
	}
	logger.GetSugarLogger().Info("Database connected successfully")
}

func AutoMigrate(models ...interface{}) {
	if err := DB.AutoMigrate(models...); err != nil {
		logger.GetSugarLogger().Fatalf("failed to migrate database: %v", err)
	}
	logger.GetSugarLogger().Info("Database migration completed")
}
