package middleware

import (
	"context"
	"errors"
	"lottery-backend/database"
	"lottery-backend/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("lottery-assistant-secret-key-2024")

// AppSessionCookieName 是网关域上应用自有会话的 cookie 名。飞牛统一网关会拦截
// 应用自己的 Authorization 头（带该头的请求根本到不了应用），网关域上账密/飞牛
// 登录换来的 JWT 改由 HttpOnly cookie（Path=网关前缀）携带，见 SetAppSessionCookie。
// middleware 里 c.Cookie 与它对应，两边不能各改各的。
const AppSessionCookieName = "token"

// 认人成功时记下这次靠的是什么凭证。AuthViaGateway 表示这次登录态来自网关注入的
// NAS 身份（会话归 NAS 管），其余都是应用自己签发的会话。
const (
	AuthViaToken   = "token"
	AuthViaGateway = "gateway"
)

// gatewayDetector 由 main 启动时注入（fnos.IsGatewayRequest）：判断请求是否来自
// 飞牛统一网关的 unix socket 监听器。TCP 直连端口可以伪造同名请求头，所以网关注入
// 的 X-Trim-* 身份只在 socket 连接上可信。
var gatewayDetector func(context.Context) bool

// SetGatewayDetector 注入网关连接判定函数；未注入（非飞牛部署）时不做网关隐式认人。
func SetGatewayDetector(f func(context.Context) bool) {
	gatewayDetector = f
}

func onGatewayConn(ctx context.Context) bool {
	return gatewayDetector != nil && gatewayDetector(ctx)
}

// Claims JWT claims
type Claims struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	TokenVersion uint   `json:"token_version"`
	jwt.RegisteredClaims
}

// gatewayPrefix 由 main 启动时注入；为空表示非飞牛网关部署，不种/清会话 cookie。
var gatewayPrefix string

// SetGatewayPrefix 注入网关前缀（如 /app/techfunway-lottery），登录接口据此种 cookie。
func SetGatewayPrefix(prefix string) {
	gatewayPrefix = strings.TrimSuffix(prefix, "/")
}

// SetAppSessionCookie 登录/注册/飞牛登录成功后种下应用会话 cookie；未启用网关时不种。
// 直连端口访问时请求路径不带网关前缀，浏览器不会回发该 cookie，仍走 Authorization。
func SetAppSessionCookie(c *gin.Context, token string) {
	if gatewayPrefix == "" || token == "" {
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     AppSessionCookieName,
		Value:    token,
		Path:     gatewayPrefix,
		MaxAge:   30 * 24 * 3600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// 飞牛网关在局域网走 http，Secure 会导致浏览器拒发 cookie。
		Secure: false,
	})
}

// ClearAppSessionCookie 清掉会话 cookie。cookie 不分端口，直连端口与网关域共用
// 同一张（Path=网关前缀），两个监听器上退出登录都要清。
func ClearAppSessionCookie(c *gin.Context) {
	if gatewayPrefix == "" {
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     AppSessionCookieName,
		Value:    "",
		Path:     gatewayPrefix,
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// SessionSuppressed 报告该应用账号是否处于「已主动登出」状态。网关隐式认人前必须
// 查它，这是「飞牛授权登录进来的会话也退得掉」的关键。查询出错时按未抑制处理：
// 一次数据库抖动不该把所有网关用户挡在登录页外。
func SessionSuppressed(userID uint) bool {
	if userID == 0 || database.DB == nil {
		return false
	}
	var session models.AppSession
	if err := database.DB.Where("user_id = ?", userID).First(&session).Error; err != nil {
		return false
	}
	return session.Suppressed
}

// SuppressGatewaySession 记录「用户在应用里主动退出登录」。网关域上光清前端没用，
// 下一个请求就能被网关注入的 NAS 身份重新认回来，必须落一条持久化标记。
func SuppressGatewaySession(userID uint) error {
	if userID == 0 || database.DB == nil {
		return nil
	}
	now := time.Now()
	return database.DB.Save(&models.AppSession{
		UserID:       userID,
		Suppressed:   true,
		SuppressedAt: now,
		UpdatedAt:    now,
	}).Error
}

// ClearSessionSuppression 在用户显式登录（账密、注册、飞牛授权、绑定）成功后解除
// 抑制：从这一刻起，只要飞牛那侧还登录着，网关身份就能继续维持应用登录态。
func ClearSessionSuppression(userID uint) error {
	if userID == 0 || database.DB == nil {
		return nil
	}
	return database.DB.Model(&models.AppSession{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{"suppressed": false, "updated_at": time.Now()}).Error
}

// gatewayIdentityUser 仅凭网关注入的 NAS 身份解析出已绑定的应用账号。它是网关域上
// 应用自有凭证之外唯一的登录态来源：飞牛接入层在转发前已经校验过 NAS 会话，并按
// 官方约定注入 X-Trim-Userid / X-Trim-Username；应用自己的 Authorization 在网关域
// 不保证能到达应用，会话 cookie 也可能被接入层丢掉，所以这里必须兜住。
// 只认已存在的绑定，绝不自动建立/修改绑定（那是 /auth/fnos/* 正门的事）。
func gatewayIdentityUser(c *gin.Context) (*models.User, bool) {
	if !onGatewayConn(c.Request.Context()) || database.DB == nil {
		return nil, false
	}
	raw := strings.TrimSpace(c.GetHeader("X-Trim-Userid"))
	if raw == "" {
		return nil, false
	}
	nasUserID, err := strconv.ParseUint(raw, 10, 32)
	if err != nil || nasUserID == 0 {
		return nil, false
	}
	var binding models.FnOSBinding
	if err := database.DB.Where("fnos_user_id = ?", uint(nasUserID)).First(&binding).Error; err != nil {
		return nil, false
	}
	var user models.User
	if err := database.DB.Select("id", "username", "role", "status", "token_version").
		First(&user, binding.UserID).Error; err != nil {
		return nil, false
	}
	if user.Status != "active" {
		return nil, false
	}
	if SessionSuppressed(user.ID) {
		return nil, false
	}
	return &user, true
}

// appTokenFromRequest 取出应用自己的会话凭证：Authorization 优先（直连端口），
// 其次会话 cookie（网关域）。Authorization 头存在但不合法时不再退回 cookie，
// 与网关接入层「认证头不合法即拒绝」的口径一致。
func appTokenFromRequest(c *gin.Context) (string, string) {
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return "", "认证令牌格式错误"
		}
		return parts[1], ""
	}
	if cookie, err := c.Cookie(AppSessionCookieName); err == nil {
		return cookie, ""
	}
	return "", ""
}

// userFromToken 解析应用自己的 JWT 并核对账号状态。
func userFromToken(tokenString string) (*models.User, string) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) { return jwtSecret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !token.Valid {
		return nil, "认证令牌无效"
	}

	var user models.User
	if err := database.DB.Select("id", "username", "role", "status", "token_version").First(&user, claims.UserID).Error; err != nil {
		return nil, "用户不存在或登录已失效"
	}
	if user.Status != "active" {
		return nil, "账号已被禁用"
	}
	if user.TokenVersion != claims.TokenVersion {
		return nil, "登录已失效，请重新登录"
	}
	if user.Username != claims.Username || user.Role != claims.Role {
		return nil, "账号信息已变更，请重新登录"
	}
	return &user, ""
}

// ResolveUser 依次尝试应用自有凭证与网关隐式身份，返回账号与凭证来源
// （AuthViaToken / AuthViaGateway）。不需要强制登录的处理器（如退出登录、
// 会话自检）复用它，保证与 AuthMiddleware 同一套认人口径。
func ResolveUser(c *gin.Context) (*models.User, string) {
	if tokenString, _ := appTokenFromRequest(c); tokenString != "" {
		if user, _ := userFromToken(tokenString); user != nil {
			return user, AuthViaToken
		}
	}

	// 网关域兜底：NAS 会话已由飞牛接入层校验，按注入身份认已绑定的应用账号。
	if user, ok := gatewayIdentityUser(c); ok {
		return user, AuthViaGateway
	}

	return nil, ""
}

// AuthMiddleware JWT认证中间件：Authorization 优先（直连端口），其次会话 cookie
// （网关域），都认不出来时退到网关注入的 NAS 身份（已绑定的应用账号）。
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, tokenErr := appTokenFromRequest(c)
		if tokenString != "" {
			user, errMsg := userFromToken(tokenString)
			if user == nil {
				tokenErr = errMsg
			} else {
				setUserContext(c, user, AuthViaToken)
				c.Next()
				return
			}
		}

		if user, ok := gatewayIdentityUser(c); ok {
			setUserContext(c, user, AuthViaGateway)
			c.Next()
			return
		}

		if tokenErr == "" {
			tokenErr = "未提供认证令牌"
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": tokenErr})
		c.Abort()
	}
}

func setUserContext(c *gin.Context, user *models.User, authVia string) {
	c.Set("auth_via", authVia)
	c.Set("user_id", user.ID)
	c.Set("username", user.Username)
	c.Set("role", user.Role)
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			c.Abort()
			return
		}

		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateToken 生成JWT令牌
func GenerateToken(user *models.User) (string, error) {
	if user == nil {
		return "", errors.New("用户不能为空")
	}

	claims := Claims{
		UserID:       user.ID,
		Username:     user.Username,
		Role:         user.Role,
		TokenVersion: user.TokenVersion,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
