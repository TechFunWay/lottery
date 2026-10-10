package services

import (
	"lottery-backend/database"
	"lottery-backend/middleware"
	"lottery-backend/models"
	"lottery-backend/utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestChangePasswordInvalidatesExistingToken(t *testing.T) {
	previousDB := database.DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	database.DB = db
	t.Cleanup(func() { database.DB = previousDB })

	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("迁移用户表失败: %v", err)
	}

	const (
		oldPassword = "old-password-md5"
		newPassword = "new-password-md5"
	)
	user := models.User{
		Username: "admin",
		Password: utils.HashPassword(oldPassword),
		Role:     "admin",
		Status:   "active",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	oldToken, err := middleware.GenerateToken(&user)
	if err != nil {
		t.Fatalf("生成旧令牌失败: %v", err)
	}
	if status := authenticatedStatus(oldToken); status != http.StatusNoContent {
		t.Fatalf("修改密码前令牌状态码 = %d，期望 %d", status, http.StatusNoContent)
	}

	service := AuthService{}
	if err := service.ChangePassword(user.ID, oldPassword, newPassword); err != nil {
		t.Fatalf("修改密码失败: %v", err)
	}
	if status := authenticatedStatus(oldToken); status != http.StatusUnauthorized {
		t.Fatalf("修改密码后旧令牌状态码 = %d，期望 %d", status, http.StatusUnauthorized)
	}

	_, newToken, err := service.Login(user.Username, newPassword)
	if err != nil {
		t.Fatalf("使用新密码登录失败: %v", err)
	}
	if status := authenticatedStatus(newToken); status != http.StatusNoContent {
		t.Fatalf("新令牌状态码 = %d，期望 %d", status, http.StatusNoContent)
	}
}

func authenticatedStatus(token string) int {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/protected", middleware.AuthMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response.Code
}
