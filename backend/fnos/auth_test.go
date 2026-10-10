package fnos

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"lottery-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupFnOSDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.SystemConfig{}, &models.FnOSBinding{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Create(&models.SystemConfig{Key: "allow_register", Value: "true"}).Error; err != nil {
		t.Fatalf("seed config: %v", err)
	}
	return db
}

func TestBindThenLoginWithFnOS(t *testing.T) {
	service := Service{DB: setupFnOSDB(t)}
	identity := Identity{UserID: 1000, Username: "nas-admin", IsAdmin: true}
	user, token, err := service.Bind(identity, "register", "admin", "client-md5", "admin@example.com")
	if err != nil || user.Role != "admin" || token == "" {
		t.Fatalf("bind result: user=%#v token=%q err=%v", user, token, err)
	}
	if user.Email != "admin@example.com" {
		t.Fatalf("bind register lost email: %q", user.Email)
	}
	user, token, err = service.Login(Identity{UserID: 1000, Username: "renamed-admin"})
	if err != nil || user.Username != "admin" || token == "" {
		t.Fatalf("login result: user=%#v token=%q err=%v", user, token, err)
	}
	if _, _, err := service.Bind(Identity{UserID: 2000, Username: "other"}, "bind", "admin", "client-md5", ""); err == nil {
		t.Fatal("same application account was bound to a second NAS user")
	}
}

func TestTicketFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1) 网关连接凭身份头签发票据 → 直连凭票据拿到身份
	gatewayReq := httptest.NewRequest(http.MethodPost, "/api/auth/fnos/ticket", nil)
	gatewayReq.Header.Set("X-Trim-Userid", "1000")
	gatewayReq.Header.Set("X-Trim-Username", "nas-admin")
	gc, _ := gin.CreateTestContext(httptest.NewRecorder())
	gc.Request = gatewayReq.WithContext(context.WithValue(gatewayReq.Context(), gatewayContextKey{}, true))
	identity, ok := identityFromGateway(gc)
	if !ok {
		t.Fatal("gateway identity rejected")
	}
	ticket, err := storeFnOSTicket(identity)
	if err != nil {
		t.Fatalf("store ticket: %v", err)
	}

	directReq := httptest.NewRequest(http.MethodPost, "/api/auth/fnos/login", nil)
	directReq.Header.Set("X-FnOS-Ticket", ticket)
	dc, _ := gin.CreateTestContext(httptest.NewRecorder())
	dc.Request = directReq
	got, ok := IdentityFromRequest(dc)
	if !ok || got.UserID != 1000 || got.Username != "nas-admin" {
		t.Fatalf("ticket identity = %#v, ok=%v", got, ok)
	}

	// 2) 票据在成功登录/绑定前可重试（如绑定密码填错后重试），
	//    由 handler 在成功后调用 consumeFnOSTicket 完成一次性消费
	consumeFnOSTicket(ticket)
	directReq2 := httptest.NewRequest(http.MethodPost, "/api/auth/fnos/login", nil)
	directReq2.Header.Set("X-FnOS-Ticket", ticket)
	dc2, _ := gin.CreateTestContext(httptest.NewRecorder())
	dc2.Request = directReq2
	if _, ok := IdentityFromRequest(dc2); ok {
		t.Fatal("consumed ticket was accepted again")
	}
}

func TestGatewayHeadersRequireTrustedConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/fnos/login", nil)
	req.Header.Set("X-Trim-Userid", "1000")
	req.Header.Set("X-Trim-Username", "nas-admin")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	if _, ok := IdentityFromRequest(c); ok {
		t.Fatal("trusted client headers from a normal TCP request")
	}

	c.Request = req.WithContext(context.WithValue(req.Context(), gatewayContextKey{}, true))
	identity, ok := IdentityFromRequest(c)
	if !ok || identity.UserID != 1000 || identity.Username != "nas-admin" {
		t.Fatalf("trusted identity = %#v, ok=%v", identity, ok)
	}
}
