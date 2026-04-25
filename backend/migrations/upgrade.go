package migrations

import (
	"lottery-backend/models"

	"gorm.io/gorm"
)

// UpgradeFunc 升级函数类型
type UpgradeFunc func(db *gorm.DB) error

// UpgradeScript 升级脚本信息
type UpgradeScript struct {
	Name     string      // 升级名称
	Func     UpgradeFunc // 升级函数
}

// UpgradeScripts 所有升级脚本的映射（按版本号）
var UpgradeScripts = map[string]UpgradeScript{
	"v1.0.0": {
		Name: "初始化数据库",
		Func: upgrade_v1_0_0,
	},
	"v1.1.0": {
		Name: "竞彩足球功能",
		Func: upgrade_v1_1_0,
	},
	"v1.2.0": {
		Name: "追加倍数多期投注",
		Func: upgrade_v1_2_0,
	},
}

// upgrade_v1_0_0 初始化数据库（当前版本）
func upgrade_v1_0_0(db *gorm.DB) error {
	// 自动创建所有表
	if err := db.AutoMigrate(
		&models.User{},
		&models.SystemConfig{},
		&models.PurchaseRecord{},
		&models.DrawResult{},
		&models.WinningRecord{},
		&models.SystemUpgrade{},
	); err != nil {
		return err
	}

	// 初始化系统配置数据
	// 检查是否已存在配置数据
	var configCount int64
	db.Model(&models.SystemConfig{}).Count(&configCount)
	if configCount == 0 {
		configs := []models.SystemConfig{
			{Key: "allow_register", Value: "true", Remark: "是否允许用户自主注册"},
		}
		if err := db.Create(&configs).Error; err != nil {
			return err
		}
	}

	return nil
}

func upgrade_v1_1_0(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.FootballMatch{},
		&models.FootballBet{},
	); err != nil {
		return err
	}
	return nil
}

func upgrade_v1_2_0(db *gorm.DB) error {
	// 为 purchase_records 表新增 multiple、append、periods 字段
	if err := db.AutoMigrate(&models.PurchaseRecord{}); err != nil {
		return err
	}
	return nil
}
