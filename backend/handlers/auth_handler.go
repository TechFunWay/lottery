package handlers

import (
	"lottery-backend/models"
	"lottery-backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

var authService = services.AuthService{}
var registerConfigService = services.ConfigService{}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Status   string `json:"status"`
	Role     string `json:"role"`
	Password string `json:"password"`
}

// ChangePasswordRequest 修改自己密码的请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// Register POST /api/auth/register
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果已有管理员，检查是否允许注册
	if authService.CheckAdminExists() && !registerConfigService.IsAllowRegister() {
		c.JSON(http.StatusForbidden, gin.H{"error": "系统已关闭用户注册"})
		return
	}

	user, token, err := authService.Register(req.Username, req.Password, req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"user":  user,
			"token": token,
		},
		"message": "注册成功",
	})
}

// Login POST /api/auth/login
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"user":  user,
			"token": token,
		},
		"message": "登录成功",
	})
}

// GetCurrentUser GET /api/auth/me
func GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := authService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// GetAllUsers GET /api/users (仅管理员)
func GetAllUsers(c *gin.Context) {
	users, err := authService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

// UpdateUser PUT /api/users/:id (仅管理员)
func UpdateUser(c *gin.Context) {
	id, err := parseUintParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新状态
	if req.Status != "" {
		if err := authService.UpdateUserStatus(id, req.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// 更新角色
	if req.Role != "" {
		if err := authService.UpdateUserRole(id, req.Role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// 更新密码（重置密码）
	if req.Password != "" {
		if len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "密码长度至少为6位"})
			return
		}
		if err := authService.UpdateUserPassword(id, req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteUser DELETE /api/users/:id (仅管理员)
func DeleteUser(c *gin.Context) {
	id, err := parseUintParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := authService.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// CheckAdmin GET /api/auth/check-admin
func CheckAdmin(c *gin.Context) {
	exists := authService.CheckAdminExists()
	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

// ChangePassword PUT /api/auth/password 修改自己的密码
func ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码长度至少为6位"})
		return
	}

	if err := authService.ChangePassword(userID.(uint), req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// parseUintParam 解析uint参数
func parseUintParam(c *gin.Context) (uint, error) {
	var params struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&params); err != nil {
		return 0, err
	}
	return params.ID, nil
}

// GetUserFromContext 从上下文中获取用户信息
func GetUserFromContext(c *gin.Context) *models.User {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil
	}

	user, err := authService.GetUserByID(userID.(uint))
	if err != nil {
		return nil
	}

	return user
}
