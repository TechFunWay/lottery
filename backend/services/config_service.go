package services

import (
	"lottery-backend/database"
	"lottery-backend/models"
	"errors"
)

type ConfigService struct{}

// InitDefaultConfigs 初始化默认配置（仅插入不存在的）
func (s *ConfigService) InitDefaultConfigs() {
	for _, cfg := range models.DefaultConfigs {
		var existing models.SystemConfig
		result := database.DB.Where("key = ?", cfg.Key).First(&existing)
		if result.Error != nil {
			database.DB.Create(&models.SystemConfig{
				Key:    cfg.Key,
				Value:  cfg.Value,
				Remark: cfg.Remark,
			})
		}
	}
}

// GetAll 获取所有配置
func (s *ConfigService) GetAll() ([]models.SystemConfig, error) {
	var configs []models.SystemConfig
	if err := database.DB.Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// Get 获取单个配置值
func (s *ConfigService) Get(key string) (string, error) {
	var cfg models.SystemConfig
	if err := database.DB.Where("key = ?", key).First(&cfg).Error; err != nil {
		return "", err
	}
	return cfg.Value, nil
}

// Set 设置配置值（不存在则创建）
func (s *ConfigService) Set(key, value string) error {
	var cfg models.SystemConfig
	result := database.DB.Where("key = ?", key).First(&cfg)
	if result.Error != nil {
		return database.DB.Create(&models.SystemConfig{Key: key, Value: value}).Error
	}
	return database.DB.Model(&cfg).Update("value", value).Error
}

// IsAllowRegister 是否允许注册
func (s *ConfigService) IsAllowRegister() bool {
	val, err := s.Get(models.ConfigKeyAllowRegister)
	if err != nil {
		return true // 查不到默认允许
	}
	return val == "true"
}

// UpdateBatch 批量更新配置
func (s *ConfigService) UpdateBatch(updates map[string]string) error {
	for key, value := range updates {
		if err := s.Set(key, value); err != nil {
			return errors.New("更新配置失败: " + key)
		}
	}
	return nil
}
