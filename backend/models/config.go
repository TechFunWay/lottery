package models

import (
	"time"

	"gorm.io/gorm"
)

// SystemConfig 系统配置表，key-value 格式
type SystemConfig struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Key       string         `json:"key" gorm:"uniqueIndex;not null;size:100"` // 配置键
	Value     string         `json:"value" gorm:"not null"`                    // 配置值
	Remark    string         `json:"remark" gorm:"size:255"`                   // 备注说明
}

// 预定义配置键
const (
	ConfigKeyAllowRegister  = "allow_register"  // 是否允许用户注册，值：true/false
	ConfigKeyFootballGlobal = "football_key_global" // 管理员全局 key,供未自配的用户降级使用
)

const ConfigKeyFootballUserPrefix = "football_key_user_"

// DefaultConfigs 系统默认配置
var DefaultConfigs = []SystemConfig{
	{Key: ConfigKeyAllowRegister, Value: "true", Remark: "是否允许用户自主注册"},
}
