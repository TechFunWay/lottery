package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"lottery-backend/database"
	"lottery-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// 网关域会话的回归测试：飞牛接入层会拦掉应用自己的 Authorization，会话 cookie 也可能
// 到不了应用；这两条路都走不通时必须退到网关注入的 NAS 身份（已绑定账号），否则网关域
// 上任何请求都 401，前端会被打回登录页（用户报的「确认登录后又跳回登录页」）。

func setupAuthTest(t *testing.T) (*gin.Engine, *models.User) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.FnOSBinding{}, &models.AppSession{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db

	user := models.User{Username: "tester", Password: "x", Role: "admin", Status: "active"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&models.FnOSBinding{UserID: user.ID, FnOSUserID: 1001, FnOSUsername: "nasuser"}).Error; err != nil {
		t.Fatalf("create binding: %v", err)
	}

	router := gin.New()
	router.GET("/me", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.GetUint("user_id"),
			"auth_via": c.GetString("auth_via"),
		})
	})
	return router, &user
}

func getMe(router *gin.Engine, mutate func(*http.Request)) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	if mutate != nil {
		mutate(req)
	}
	router.ServeHTTP(w, req)
	return w
}

func asGateway(on bool) {
	if on {
		SetGatewayDetector(func(context.Context) bool { return true })
		return
	}
	SetGatewayDetector(func(context.Context) bool { return false })
}

func withGatewayIdentity(req *http.Request) {
	req.Header.Set("X-Trim-Userid", "1001")
	req.Header.Set("X-Trim-Username", "nasuser")
}

// 网关 socket 上：没有 Authorization、没有会话 cookie，只有网关注入身份 → 必须认人。
func TestGatewayIdentityFallback(t *testing.T) {
	router, _ := setupAuthTest(t)
	asGateway(true)

	w := getMe(router, withGatewayIdentity)
	if w.Code != http.StatusOK {
		t.Fatalf("gateway request without app credentials: got %d, want 200 (%s)", w.Code, w.Body.String())
	}
	if body := w.Body.String(); body == "" || !strings.Contains(body, `"auth_via":"gateway"`) {
		t.Fatalf("expected auth_via=gateway, got %s", body)
	}
}

// TCP 直连端口可以伪造同名请求头：只有网关注入的身份才可信。
func TestGatewayIdentityRejectedOnTCP(t *testing.T) {
	router, _ := setupAuthTest(t)
	asGateway(false)

	if w := getMe(router, withGatewayIdentity); w.Code != http.StatusUnauthorized {
		t.Fatalf("forged X-Trim header on direct port: got %d, want 401", w.Code)
	}
}

// 未绑定的 NAS 账号不给登录态。
func TestGatewayIdentityWithoutBinding(t *testing.T) {
	router, _ := setupAuthTest(t)
	asGateway(true)

	if w := getMe(router, func(req *http.Request) {
		req.Header.Set("X-Trim-Userid", "9999")
		req.Header.Set("X-Trim-Username", "stranger")
	}); w.Code != http.StatusUnauthorized {
		t.Fatalf("unbound NAS account: got %d, want 401", w.Code)
	}
}

// 应用自己的凭证优先于网关隐式身份；无效凭证仍可被网关身份兜住（cookie 被接入层丢掉时）。
func TestTokenPreferredAndInvalidTokenFallsBack(t *testing.T) {
	router, user := setupAuthTest(t)
	asGateway(true)

	token, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	w := getMe(router, func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+token)
		withGatewayIdentity(req)
	})
	if !strings.Contains(w.Body.String(), `"auth_via":"token"`) {
		t.Fatalf("valid token should win, got %s", w.Body.String())
	}

	w = getMe(router, func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer not-a-jwt")
		withGatewayIdentity(req)
	})
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"auth_via":"gateway"`) {
		t.Fatalf("invalid token should fall back to gateway identity, got %d %s", w.Code, w.Body.String())
	}
}

// 用户主动登出后，网关身份不再自动换登录态；显式登录解除抑制。
func TestLogoutSuppressionBlocksGatewayIdentity(t *testing.T) {
	router, user := setupAuthTest(t)
	asGateway(true)

	if err := SuppressGatewaySession(user.ID); err != nil {
		t.Fatalf("suppress: %v", err)
	}
	if w := getMe(router, withGatewayIdentity); w.Code != http.StatusUnauthorized {
		t.Fatalf("suppressed user must not be recognised: got %d, want 401", w.Code)
	}

	if err := ClearSessionSuppression(user.ID); err != nil {
		t.Fatalf("clear suppression: %v", err)
	}
	if w := getMe(router, withGatewayIdentity); w.Code != http.StatusOK {
		t.Fatalf("explicit login should re-enable gateway identity: got %d", w.Code)
	}
}

// 会话 cookie 是网关域上应用自有的通道，解析成功时按 token 记来源。
func TestSessionCookieAuthenticates(t *testing.T) {
	router, user := setupAuthTest(t)
	asGateway(true)

	token, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	w := getMe(router, func(req *http.Request) {
		req.AddCookie(&http.Cookie{Name: AppSessionCookieName, Value: token})
	})
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"auth_via":"token"`) {
		t.Fatalf("session cookie should authenticate: got %d %s", w.Code, w.Body.String())
	}
}

