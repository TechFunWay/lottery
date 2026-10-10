package services

import (
	"lottery-backend/database"
	"lottery-backend/middleware"
	"lottery-backend/models"
	"lottery-backend/utils"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type AuthService struct{}

// GetDB 获取数据库连接
func (s *AuthService) GetDB() *gorm.DB {
	return database.DB
}

// Register 用户注册
func (s *AuthService) Register(username, password, email string) (*models.User, string, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	result := s.GetDB().Where("username = ?", username).First(&existingUser)
	if result.Error == nil {
		return nil, "", errors.New("用户名已存在")
	}

	// 检查是否已有管理员，如果有则只能创建普通用户
	var adminCount int64
	s.GetDB().Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)
	role := "user"
	if adminCount == 0 {
		role = "admin" // 第一个用户自动为管理员
	}

	// 对前端传来的 MD5 值再做一次加盐 MD5
	hashedPassword := utils.HashPassword(password)

	user := &models.User{
		Username: username,
		Password: hashedPassword,
		Email:    email,
		Role:     role,
		Status:   "active",
	}

	if err := s.GetDB().Create(user).Error; err != nil {
		return nil, "", err
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (*models.User, string, error) {
	var user models.User

	// 查找用户
	result := s.GetDB().Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, "", errors.New("账号已被禁用")
	}

	// 对前端传来的 MD5 值再做一次加盐 MD5
	hashedPassword := utils.HashPassword(password)
	if user.Password != hashedPassword {
		return nil, "", errors.New("用户名或密码错误")
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now
	s.GetDB().Save(&user)

	// 生成JWT令牌
	token, err := middleware.GenerateToken(&user)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	result := s.GetDB().First(&user, userID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetAllUsers 获取所有用户（仅管理员）
func (s *AuthService) GetAllUsers() ([]models.User, error) {
	var users []models.User
	result := s.GetDB().Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// UpdateUserStatus 更新用户状态（仅管理员）
func (s *AuthService) UpdateUserStatus(userID uint, status string) error {
	return s.GetDB().Model(&models.User{}).Where("id = ?", userID).Update("status", status).Error
}

// UpdateUserRole 更新用户角色（仅管理员）
func (s *AuthService) UpdateUserRole(userID uint, role string) error {
	// 检查目标用户是否是管理员
	var targetUser models.User
	if err := s.GetDB().First(&targetUser, userID).Error; err != nil {
		return err
	}

	// 如果要将管理员降级为普通用户，需要确保还有其他管理员
	if targetUser.Role == "admin" && role != "admin" {
		var adminCount int64
		s.GetDB().Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)
		if adminCount <= 1 {
			return errors.New("不能删除最后一个管理员")
		}
	}

	return s.GetDB().Model(&models.User{}).Where("id = ?", userID).Update("role", role).Error
}

// DeleteUser 删除用户（仅管理员）
func (s *AuthService) DeleteUser(userID uint) error {
	// 检查是否是管理员
	var targetUser models.User
	if err := s.GetDB().First(&targetUser, userID).Error; err != nil {
		return err
	}

	// 不能删除管理员
	if targetUser.Role == "admin" {
		return errors.New("不能删除管理员账号")
	}

	// 软删除
	return s.GetDB().Delete(&models.User{}, userID).Error
}

// CheckAdminExists 检查是否存在管理员
func (s *AuthService) CheckAdminExists() bool {
	var count int64
	s.GetDB().Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	return count > 0
}

// ResetAdminPassword 重置管理员密码（用于命令行）
// 命令行传入的是明文，需要先 MD5 再加盐 MD5（与注册/登录流程一致）
// 返回管理员用户名和错误信息
func (s *AuthService) ResetAdminPassword(password string) (string, error) {
	var admin models.User
	if err := s.GetDB().Where("role = ?", "admin").First(&admin).Error; err != nil {
		return "", errors.New("管理员不存在")
	}

	// 第一步：前端 MD5
	firstHash := utils.MD5Hash(password)
	// 第二步：加盐后端 MD5（与前端登录/注册流程一致）
	hashedPassword := utils.HashPassword(firstHash)

	if err := s.GetDB().Model(&admin).Update("password", hashedPassword).Error; err != nil {
		return "", err
	}

	return admin.Username, nil
}

// UpdateUserPassword 重置用户密码（仅管理员）
func (s *AuthService) UpdateUserPassword(userID uint, password string) error {
	fmt.Printf("[DEBUG] UpdateUserPassword: userID=%d, receivedPassword=%s\n", userID, password)

	// 检查目标用户是否存在
	var targetUser models.User
	if err := s.GetDB().First(&targetUser, userID).Error; err != nil {
		fmt.Printf("[DEBUG] User not found: %v\n", err)
		return err
	}

	// 不能重置管理员密码
	if targetUser.Role == "admin" {
		fmt.Println("[DEBUG] Cannot reset admin password")
		return errors.New("不能重置管理员密码")
	}

	// 对前端传来的 MD5 值再做一次加盐 MD5
	hashedPassword := utils.HashPassword(password)
	fmt.Printf("[DEBUG] Hashed password to save: %s\n", hashedPassword)

	result := s.GetDB().Model(&models.User{}).Where("id = ?", userID).Update("password", hashedPassword)
	fmt.Printf("[DEBUG] Update result: rowsAffected=%d, error=%v\n", result.RowsAffected, result.Error)
	return result.Error
}

// ChangePassword 用户修改自己的密码（需验证旧密码）
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	// 查找用户
	var user models.User
	if err := s.GetDB().First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码（前端已MD5，后端再做加盐MD5）
	oldHashed := utils.HashPassword(oldPassword)
	if user.Password != oldHashed {
		return errors.New("旧密码错误")
	}

	// 更新新密码（前端已MD5，后端再做加盐MD5）
	newHashed := utils.HashPassword(newPassword)
	return s.GetDB().Model(&models.User{}).Where("id = ?", userID).Update("password", newHashed).Error
}
