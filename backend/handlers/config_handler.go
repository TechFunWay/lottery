package handlers

import (
	"lottery-backend/models"
	"lottery-backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

var configService = services.ConfigService{}

// GetConfigs GET /api/configs  (管理员)
func GetConfigs(c *gin.Context) {
	configs, err := configService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": configs})
}

// UpdateConfigs PUT /api/configs  (管理员)
// Body: [{"key":"allow_register","value":"true"}, ...]
func UpdateConfigs(c *gin.Context) {
	var items []models.SystemConfig
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]string, len(items))
	for _, item := range items {
		updates[item.Key] = item.Value
	}

	if err := configService.UpdateBatch(updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "配置更新成功"})
}

// GetPublicConfigs GET /api/configs/public  (无需登录，返回前端需要的公开配置)
func GetPublicConfigs(c *gin.Context) {
	allowRegister := configService.IsAllowRegister()
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"allow_register": allowRegister,
		},
	})
}
