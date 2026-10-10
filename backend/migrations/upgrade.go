package migrations

import (
	"lottery-backend/models"

	"gorm.io/gorm"
)

// UpgradeFunc 升级函数类型
type UpgradeFunc func(db *gorm.DB) error

// UpgradeScript 升级脚本信息
type UpgradeScript struct {
	Name string      // 升级名称
	Func UpgradeFunc // 升级函数
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
	"v1.3.0": {
		Name: "中奖金额手动调整",
		Func: upgrade_v1_3_0,
	},
	"v1.4.0": {
		Name: "双色球福运奖标记",
		Func: upgrade_v1_4_0,
	},
	"v1.4.1": {
		Name: "登录令牌失效机制",
		Func: upgrade_v1_4_1,
	},
	"v1.4.2": {
		Name: "飞牛 NAS 账号绑定",
		Func: upgrade_v1_4_2,
	},
	"v1.4.3": {
		Name: "开奖结果按彩种+期号复合唯一",
		Func: upgrade_v1_4_3,
	},
	"v1.4.8": {
		Name: "网关域会话与登出抑制",
		Func: upgrade_v1_4_8,
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
		&models.FnOSBinding{},
		&models.AppSession{},
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

func upgrade_v1_3_0(db *gorm.DB) error {
	// 为 winning_records 表新增 manual_amount 字段（手动调整奖金）
	if err := db.AutoMigrate(&models.WinningRecord{}); err != nil {
		return err
	}
	return nil
}

func upgrade_v1_4_0(db *gorm.DB) error {
	// 为 draw_results 表新增 fu_yun_award 字段（双色球福运奖标记）
	if err := db.AutoMigrate(&models.DrawResult{}); err != nil {
		return err
	}
	return nil
}

func upgrade_v1_4_1(db *gorm.DB) error {
	// 为用户表新增令牌版本；修改密码或账号权限后递增，使旧 JWT 立即失效。
	return db.AutoMigrate(&models.User{})
}

func upgrade_v1_4_2(db *gorm.DB) error {
	return db.AutoMigrate(&models.FnOSBinding{})
}

// upgrade_v1_4_3 重建开奖结果期号唯一索引。原索引只按期号单列唯一，而
// 双色球/七乐彩、排列3/排列5 两两期号相同，先入库的会把后入库的挡死
// （UNIQUE constraint failed: draw_results.issue_number），改为彩种+期号复合唯一。
func upgrade_v1_4_3(db *gorm.DB) error {
	// AutoMigrate 对已存在的同名索引不做重建，必须先删旧的单列唯一索引
	if db.Migrator().HasIndex(&models.DrawResult{}, "idx_lottery_issue") {
		if err := db.Migrator().DropIndex(&models.DrawResult{}, "idx_lottery_issue"); err != nil {
			return err
		}
	}
	return db.AutoMigrate(&models.DrawResult{})
}

// upgrade_v1_4_8 新增 app_sessions 表：记录「用户在应用里主动退出登录」，
// 网关隐式认人（X-Trim-Userid）之前先查它，保证飞牛授权登录的会话也退得掉。
func upgrade_v1_4_8(db *gorm.DB) error {
	return db.AutoMigrate(&models.AppSession{})
}
