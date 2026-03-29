package handlers

import (
	"net/http"

	"lottery-backend/database"
	"lottery-backend/services"

	"github.com/gin-gonic/gin"
)

// GetVersion 获取系统版本信息
// 注意：这个函数现在在main.go中直接实现，以便访问Version等变量
// 保留这个函数声明以避免编译错误，但实际上不会使用

// GetCurrentVersion 获取当前数据库版本
func GetCurrentVersion(c *gin.Context) {
	// 数据目录路径
	dataDir := "./data"
	
	upgradeService := services.NewUpgradeService(database.DB, dataDir)
	version, err := upgradeService.GetCurrentVersion()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"version": "v1.0.0", // 默认版本
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"version": version,
	})
}

// GetUpgradeHistory 获取升级历史
func GetUpgradeHistory(c *gin.Context) {
	// 数据目录路径
	dataDir := "./data"
	
	upgradeService := services.NewUpgradeService(database.DB, dataDir)
	history, err := upgradeService.GetUpgradeHistory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取升级历史失败",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": history,
	})
}

// 版本信息（从main.go中复制，用于编译）
var (
	Version   = "v1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)