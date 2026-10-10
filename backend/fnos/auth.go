package fnos

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"lottery-backend/logger"
	"lottery-backend/middleware"
	"lottery-backend/models"
	"lottery-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	ErrNotBound     = errors.New("此飞牛 NAS 账号尚未绑定应用账号")
	ErrAlreadyBound = errors.New("此飞牛 NAS 账号已绑定其他应用账号")
)

type Identity struct {
	UserID   uint
	Username string
	IsAdmin  bool
}

type Service struct {
	DB *gorm.DB
}

// fnOSLoginTicket 把网关侧验证过的 NAS 身份安全地带到 TCP 直连端口：
// 网关头在直连上不可信（可伪造），票据短时效、一次性，泄露也无法重放。
// 时效需覆盖「确认弹窗 + 填写绑定/注册表单」的完整操作时长。
const fnOSTicketTTL = 10 * time.Minute

type fnOSLoginTicket struct {
	Identity  Identity
	ExpiresAt time.Time
}

var (
	fnOSTicketsMu sync.Mutex
	fnOSTickets   = make(map[string]fnOSLoginTicket)
)

func newFnOSTicketValue() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func storeFnOSTicket(identity Identity) (string, error) {
	ticket, err := newFnOSTicketValue()
	if err != nil {
		return "", err
	}
	now := time.Now()
	fnOSTicketsMu.Lock()
	for key, item := range fnOSTickets {
		if !item.ExpiresAt.After(now) {
			delete(fnOSTickets, key)
		}
	}
	fnOSTickets[ticket] = fnOSLoginTicket{Identity: identity, ExpiresAt: now.Add(fnOSTicketTTL)}
	fnOSTicketsMu.Unlock()
	return ticket, nil
}

func consumeFnOSTicket(ticket string) {
	if ticket == "" {
		return
	}
	fnOSTicketsMu.Lock()
	delete(fnOSTickets, ticket)
	fnOSTicketsMu.Unlock()
}

// identityFromGateway 从网关注入的 X-Trim-* 头解析 NAS 身份（仅网关连接可信任）。
func identityFromGateway(c *gin.Context) (Identity, bool) {
	uid, err := strconv.ParseUint(c.GetHeader("X-Trim-Userid"), 10, 32)
	username := c.GetHeader("X-Trim-Username")
	if err != nil || uid == 0 || username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未获取到飞牛 NAS 登录信息"})
		return Identity{}, false
	}
	return Identity{UserID: uint(uid), Username: username, IsAdmin: c.GetHeader("X-Trim-Isadmin") == "true"}, true
}

// IdentityFromRequest 解析 NAS 身份：网关连接直接读身份头；
// TCP 直连则只接受网关签发的一次性票据（X-FnOS-Ticket）。
func IdentityFromRequest(c *gin.Context) (Identity, bool) {
	if !IsGatewayRequest(c.Request.Context()) {
		ticket := c.GetHeader("X-FnOS-Ticket")
		if ticket == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请从飞牛桌面中的应用入口使用 NAS 登录"})
			return Identity{}, false
		}
		fnOSTicketsMu.Lock()
		item, ok := fnOSTickets[ticket]
		if ok && !item.ExpiresAt.After(time.Now()) {
			delete(fnOSTickets, ticket)
			ok = false
		}
		fnOSTicketsMu.Unlock()
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "飞牛登录凭证无效或已过期，请重新点击飞牛授权登录"})
			return Identity{}, false
		}
		return item.Identity, true
	}
	return identityFromGateway(c)
}

func (s Service) Login(identity Identity) (*models.User, string, error) {
	var binding models.FnOSBinding
	if err := s.DB.Where("fnos_user_id = ?", identity.UserID).First(&binding).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrNotBound
		}
		return nil, "", err
	}
	var user models.User
	if err := s.DB.First(&user, binding.UserID).Error; err != nil {
		return nil, "", err
	}
	if user.Status != "active" {
		return nil, "", fmt.Errorf("账号已被禁用")
	}
	now := time.Now()
	if err := s.DB.Model(&user).Updates(map[string]interface{}{"last_login": now}).Error; err != nil {
		return nil, "", err
	}
	if binding.FnOSUsername != identity.Username {
		if err := s.DB.Model(&binding).Update("fnos_username", identity.Username).Error; err != nil {
			return nil, "", err
		}
	}
	user.LastLogin = &now
	token, err := middleware.GenerateToken(&user)
	return &user, token, err
}

func (s Service) Bind(identity Identity, mode, username, password, email string) (*models.User, string, error) {
	var user models.User
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var bindingCount int64
		if err := tx.Model(&models.FnOSBinding{}).Where("fnos_user_id = ?", identity.UserID).Count(&bindingCount).Error; err != nil {
			return err
		}
		if bindingCount > 0 {
			return ErrAlreadyBound
		}

		switch mode {
		case "register":
			var existing models.User
			if err := tx.Where("username = ?", username).First(&existing).Error; err == nil {
				return fmt.Errorf("用户名已存在")
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			var userCount int64
			if err := tx.Model(&models.User{}).Count(&userCount).Error; err != nil {
				return err
			}
			if userCount > 0 {
				var cfg models.SystemConfig
				if err := tx.Where("key = ?", "allow_register").First(&cfg).Error; err != nil || cfg.Value != "true" {
					return fmt.Errorf("系统已关闭用户注册")
				}
			}
			role := "user"
			if userCount == 0 {
				role = "admin"
			}
			user = models.User{Username: username, Password: utils.HashPassword(password), Email: email, Role: role, Status: "active"}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
		case "bind":
			if err := tx.Where("username = ?", username).First(&user).Error; err != nil || user.Status != "active" || user.Password != utils.HashPassword(password) {
				return fmt.Errorf("用户名或密码错误")
			}
		default:
			return fmt.Errorf("无效的绑定方式")
		}

		var existing models.FnOSBinding
		if err := tx.Where("user_id = ?", user.ID).First(&existing).Error; err == nil {
			return fmt.Errorf("该应用账号已绑定其他飞牛 NAS 账号")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return tx.Create(&models.FnOSBinding{UserID: user.ID, FnOSUserID: identity.UserID, FnOSUsername: identity.Username}).Error
	})
	if err != nil {
		return nil, "", err
	}
	token, err := middleware.GenerateToken(&user)
	return &user, token, err
}

// clearGatewaySuppression 飞牛授权登录/绑定成功即解除「已主动登出」抑制：
// 用户显式确认过的登录必须立刻生效，不能被上一次登出的标记挡住。
func clearGatewaySuppression(userID uint) {
	if err := middleware.ClearSessionSuppression(userID); err != nil {
		logger.GetSugarLogger().Warnf("⚠️  解除网关登录抑制失败: %v", err)
	}
}

func RegisterRoutes(group *gin.RouterGroup, db *gorm.DB) {
	service := Service{DB: db}
	// 一次性登录票据：仅网关连接可签发（浏览器会话已被 fnOS 校验，身份头不可伪造），
	// 供直连端口上的 SPA 换取登录态，整个过程不接触网关身份头。
	group.POST("/ticket", func(c *gin.Context) {
		if !IsGatewayRequest(c.Request.Context()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请从飞牛桌面中的应用入口使用 NAS 登录"})
			return
		}
		identity, ok := identityFromGateway(c)
		if !ok {
			return
		}
		ticket, err := storeFnOSTicket(identity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成飞牛登录凭证失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"ticket": ticket, "fnos_username": identity.Username}})
	})
	group.POST("/login", func(c *gin.Context) {
		identity, ok := IdentityFromRequest(c)
		if !ok {
			return
		}
		var req struct {
			Remember *bool `json:"remember"`
		}
		if c.Request.ContentLength != 0 {
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "登录参数无效"})
				return
			}
		}
		user, token, err := service.Login(identity)
		if errors.Is(err, ErrNotBound) {
			var accountCount int64
			if err := db.Model(&models.User{}).Count(&accountCount).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "检查应用账号状态失败"})
				return
			}
			var matchingCount int64
			if err := db.Model(&models.User{}).Where("username = ?", identity.Username).Count(&matchingCount).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "检查飞牛账号绑定状态失败"})
				return
			}
			suggestedMode := "register"
			if accountCount > 0 {
				suggestedMode = "bind"
			}
			suggestedUsername := ""
			if matchingCount > 0 {
				suggestedUsername = identity.Username
			}
			c.JSON(http.StatusOK, gin.H{"data": gin.H{
				"binding_required":   true,
				"fnos_username":      identity.Username,
				"has_accounts":       accountCount > 0,
				"suggested_mode":     suggestedMode,
				"suggested_username": suggestedUsername,
			}})
			return
		}
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "飞牛 NAS 登录失败"})
			return
		}
		consumeFnOSTicket(c.GetHeader("X-FnOS-Ticket"))
		clearGatewaySuppression(user.ID)
		middleware.SetAppSessionCookie(c, token)
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"user": user, "token": token}, "message": "登录成功"})
	})
	group.POST("/bind", func(c *gin.Context) {
		identity, ok := IdentityFromRequest(c)
		if !ok {
			return
		}
		var req struct {
			Mode     string `json:"mode" binding:"required"`
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
			Email    string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请输入应用账号和密码"})
			return
		}
		user, token, err := service.Bind(identity, req.Mode, req.Username, req.Password, req.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		consumeFnOSTicket(c.GetHeader("X-FnOS-Ticket"))
		clearGatewaySuppression(user.ID)
		middleware.SetAppSessionCookie(c, token)
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"user": user, "token": token}, "message": "绑定成功"})
	})
}
