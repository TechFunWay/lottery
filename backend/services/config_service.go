package services

import (
	"lottery-backend/database"
	"lottery-backend/models"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
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

// FootballKeySource 数据源 key 的来源
type FootballKeySource string

const (
	FootballKeySourceNone    FootballKeySource = "none"
	FootballKeySourceUser    FootballKeySource = "user"
	FootballKeySourceAdmin   FootballKeySource = "admin"
	FootballKeySourceEnv     FootballKeySource = "env"
	FootballKeySourceBuiltin FootballKeySource = "builtin"
)

func (s *ConfigService) ResolveAPIFootballKey(userID uint) (string, FootballKeySource) {
	if userID > 0 {
		if v, err := s.Get(fmt.Sprintf("%s%d", models.ConfigKeyFootballUserPrefix, userID)); err == nil && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v), FootballKeySourceUser
		}
	}
	if v, err := s.Get(models.ConfigKeyFootballGlobal); err == nil && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v), FootballKeySourceAdmin
	}
	if env := strings.TrimSpace(os.Getenv("API_FOOTBALL_KEY")); env != "" {
		return env, FootballKeySourceEnv
	}
	if builtIn := strings.TrimSpace(BuiltInAPIFootballKey); builtIn != "" {
		return builtIn, FootballKeySourceBuiltin
	}
	return "", FootballKeySourceNone
}

func (s *ConfigService) SetAPIFootballKey(userID uint, key string) error {
	return s.Set(fmt.Sprintf("%s%d", models.ConfigKeyFootballUserPrefix, userID), strings.TrimSpace(key))
}

func (s *ConfigService) SetGlobalAPIFootballKey(key string) error {
	return s.Set(models.ConfigKeyFootballGlobal, strings.TrimSpace(key))
}

// MaskAPIFootballKey 前 4 后 4,中间 * 替代;短 key 一律 **** 防反推。
func MaskAPIFootballKey(key string) string {
	k := strings.TrimSpace(key)
	if len(k) <= 8 {
		return "****"
	}
	return k[:4] + strings.Repeat("*", len(k)-8) + k[len(k)-4:]
}

// TestAPIFootballKey 打一次 /timezone 验 key,200 即通过。
func TestAPIFootballKey(key string) error {
	k := strings.TrimSpace(key)
	if k == "" {
		return errors.New("key 为空")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, "https://v3.football.api-sports.io/timezone", nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-apisports-key", k)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("网络错误: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d(可能 key 无效或被限流)", resp.StatusCode)
	}
	return nil
}
