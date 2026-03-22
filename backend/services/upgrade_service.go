package services

import (
	"fmt"
	"lottery-backend/logger"
	"lottery-backend/migrations"
	"lottery-backend/models"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UpgradeService 升级服务
type UpgradeService struct {
	db       *gorm.DB
	dataDir  string // 数据目录（用于日志）
}

// NewUpgradeService 创建升级服务
func NewUpgradeService(db *gorm.DB, dataDir string) *UpgradeService {
	return &UpgradeService{
		db:      db,
		dataDir: dataDir,
	}
}

// ParseVersion 解析版本号 v1.2.3 -> [1, 2, 3]
func ParseVersion(version string) []int {
	// 去掉 'v' 前缀
	version = strings.TrimPrefix(version, "v")

	// 按 '.' 分割
	parts := strings.Split(version, ".")

	// 转换为数字数组
	result := make([]int, len(parts))
	for i, part := range parts {
		num, _ := strconv.Atoi(part)
		result[i] = num
	}

	return result
}

// CompareVersions 比较两个版本号
// 返回: -1 表示 v1 < v2, 0 表示 v1 == v2, 1 表示 v1 > v2
func CompareVersions(v1, v2 string) int {
	p1 := ParseVersion(v1)
	p2 := ParseVersion(v2)

	maxLen := len(p1)
	if len(p2) > maxLen {
		maxLen = len(p2)
	}

	// 补齐长度，不足的补0
	for len(p1) < maxLen {
		p1 = append(p1, 0)
	}
	for len(p2) < maxLen {
		p2 = append(p2, 0)
	}

	// 逐位比较
	for i := 0; i < maxLen; i++ {
		if p1[i] < p2[i] {
			return -1
		} else if p1[i] > p2[i] {
			return 1
		}
	}

	return 0
}

// RunUpgrades 执行所有待执行的升级
func (s *UpgradeService) RunUpgrades() error {
	// 写入升级开始日志
	logger.WriteUpgradeLog(s.dataDir, "========================================")
	logger.WriteUpgradeLog(s.dataDir, "开始检查数据库升级")
	logger.WriteUpgradeLog(s.dataDir, "========================================")

	// 确保表存在
	if err := s.db.AutoMigrate(&models.SystemUpgrade{}); err != nil {
		logger.WriteUpgradeLog(s.dataDir, fmt.Sprintf("❌ 创建system_upgrade表失败: %v", err))
		return fmt.Errorf("创建system_upgrade表失败: %w", err)
	}

	// 获取所有已执行成功的版本
	var executedVersions []models.SystemUpgrade
	if err := s.db.Where("status = ?", 1).Order("version ASC").Find(&executedVersions).Error; err != nil {
		logger.WriteUpgradeLog(s.dataDir, fmt.Sprintf("❌ 查询已执行版本失败: %v", err))
		return fmt.Errorf("查询已执行版本失败: %w", err)
	}

	// 获取当前最新版本号（已执行且成功的）
	var currentVersion string
	if len(executedVersions) > 0 {
		currentVersion = executedVersions[len(executedVersions)-1].Version
	}

	logger.WriteUpgradeLog(s.dataDir, fmt.Sprintf("当前数据库版本: %s", currentVersion))

	// 收集所有待执行的升级脚本
	var versionsToRun []string
	for version := range migrations.UpgradeScripts {
		if CompareVersions(version, currentVersion) > 0 {
			versionsToRun = append(versionsToRun, version)
		}
	}

	// 按版本号排序
	sort.Slice(versionsToRun, func(i, j int) bool {
		return CompareVersions(versionsToRun[i], versionsToRun[j]) < 0
	})

	// 如果没有需要升级的版本，直接返回
	if len(versionsToRun) == 0 {
		logger.WriteUpgradeLog(s.dataDir, fmt.Sprintf("✅ 当前已是最新版本: %s", currentVersion))
		fmt.Printf("✅ 当前已是最新版本: %s\n", currentVersion)
		return nil
	}

	logger.WriteUpgradeLog(s.dataDir, fmt.Sprintf("发现 %d 个待升级版本", len(versionsToRun)))
	for _, v := range versionsToRun {
		logger.WriteUpgradeLog(s.dataDir, fmt.Sprintf("  - %s", v))
	}

	// 执行升级
	for _, version := range versionsToRun {
		script := migrations.UpgradeScripts[version]
		logger.WriteUpgradeLog(s.dataDir, fmt.Sprintf("🔨 开始升级到 %s: %s", version, script.Name))
		fmt.Printf("🔨 正在升级到 %s: %s\n", version, script.Name)

		// 检查是否已执行过
		var existing models.SystemUpgrade
		err := s.db.Where("version = ?", version).First(&existing).Error
		if err == nil {
			// 如果已执行且成功，跳过
			if existing.Status == 1 {
				logger.WriteUpgradeLog(s.dataDir, fmt.Sprintf("⏭️  %s 已执行，跳过", version))
				fmt.Printf("⏭️  %s 已执行，跳过\n", version)
				continue
			}
			// 如果之前失败，更新记录重新执行
			logger.WriteUpgradeLog(s.dataDir, fmt.Sprintf("⚠️  %s 之前执行失败，重新执行", version))
		}

		// 创建升级记录
		upgrade := models.SystemUpgrade{
			Version:   version,
			Name:      script.Name,
			Status:    0, // 待执行
			StartTime: time.Now(),
		}
		s.db.Save(&upgrade)

		// 执行升级函数
		startTime := time.Now()
		err = script.Func(s.db)
		duration := time.Since(startTime)

		if err != nil {
			// 升级失败，记录错误
			upgrade.EndTime = time.Now()
			upgrade.Status = 2 // 失败
			upgrade.Remark = err.Error()
			s.db.Save(&upgrade)

			logger.WriteUpgradeLog(s.dataDir, fmt.Sprintf("❌ 升级 %s 失败，耗时: %v, 错误: %v", version, duration, err))
			return fmt.Errorf("升级 %s 失败: %w", version, err)
		}

		// 升级成功
		upgrade.EndTime = time.Now()
		upgrade.Status = 1 // 成功
		upgrade.Remark = "升级成功"
		s.db.Save(&upgrade)

		logger.WriteUpgradeLog(s.dataDir, fmt.Sprintf("✅ 升级 %s 成功，耗时: %v", version, duration))
		fmt.Printf("✅ 升级 %s 成功\n", version)
	}

	logger.WriteUpgradeLog(s.dataDir, fmt.Sprintf("🎉 所有升级完成，当前版本: %s", versionsToRun[len(versionsToRun)-1]))
	fmt.Printf("🎉 所有升级完成，当前版本: %s\n", versionsToRun[len(versionsToRun)-1])

	logger.WriteUpgradeLog(s.dataDir, "========================================")
	logger.WriteUpgradeLog(s.dataDir, "升级检查结束")
	logger.WriteUpgradeLog(s.dataDir, "========================================")

	return nil
}

// GetCurrentVersion 获取当前数据库版本
func (s *UpgradeService) GetCurrentVersion() (string, error) {
	var upgrade models.SystemUpgrade
	err := s.db.Where("status = ?", 1).Order("version DESC").First(&upgrade).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return upgrade.Version, nil
}

// GetUpgradeHistory 获取升级历史
func (s *UpgradeService) GetUpgradeHistory() ([]models.SystemUpgrade, error) {
	var upgrades []models.SystemUpgrade
	err := s.db.Order("version ASC").Find(&upgrades).Error
	return upgrades, err
}
