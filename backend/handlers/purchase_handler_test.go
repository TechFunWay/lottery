package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"lottery-backend/database"
	"lottery-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupPurchaseTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.PurchaseRecord{}, &models.DrawResult{}, &models.WinningRecord{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db.Exec("DELETE FROM purchase_records")
	database.DB = db

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.POST("/api/purchases", CreatePurchase)
	router.POST("/api/purchases/batch", CreatePurchaseBatch)
	return router, db
}

func postJSON(t *testing.T, router *gin.Engine, path string, body any) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	data, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return w, resp
}

func TestBatchCreateMultiPeriod(t *testing.T) {
	router, db := setupPurchaseTest(t)

	w, resp := postJSON(t, router, "/api/purchases/batch", gin.H{
		"items": []gin.H{
			gin.H{
				"lottery_type":  "双色球",
				"issue_number":  "2026001",
				"purchase_date": "2026-09-21",
				"numbers":       `{"red":[1,5,12,20,28,33],"blue":[6]}`,
				"bet_type":      "单式",
				"amount":        20,
				"periods":       3,
			},
			gin.H{
				"lottery_type":  "大乐透",
				"issue_number":  "26110",
				"purchase_date": "2026-09-21",
				"numbers":       `{"front":[3,9,17,25,31],"back":[4,11]}`,
				"bet_type":      "单式",
				"amount":        4,
			},
		},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if resp["total"].(float64) != 4 {
		t.Fatalf("期望 4 条（3期展开+1条），实际 %v", resp["total"])
	}

	var count int64
	db.Model(&models.PurchaseRecord{}).Count(&count)
	if count != 4 {
		t.Fatalf("落库条数=%d，期望 4", count)
	}

	var issues []string
	db.Model(&models.PurchaseRecord{}).Where("lottery_type = ?", "双色球").Order("issue_number").Pluck("issue_number", &issues)
	if len(issues) != 3 || issues[0] != "2026001" || issues[1] != "2026002" || issues[2] != "2026003" {
		t.Fatalf("期号递增异常: %v", issues)
	}
	var periods int
	db.Model(&models.PurchaseRecord{}).Where("lottery_type = ?", "双色球").Pluck("periods", &periods)
	if periods != 3 {
		t.Fatalf("periods=%d，期望 3", periods)
	}
}

func TestBatchCreateValidation(t *testing.T) {
	router, db := setupPurchaseTest(t)

	cases := []struct {
		name string
		item gin.H
	}{
		{"坏日期", gin.H{"lottery_type": "双色球", "issue_number": "1", "purchase_date": "09-21", "numbers": "[]", "amount": 2}},
		{"坏号码JSON", gin.H{"lottery_type": "双色球", "issue_number": "1", "purchase_date": "2026-09-21", "numbers": "not json", "amount": 2}},
		{"零金额", gin.H{"lottery_type": "双色球", "issue_number": "1", "purchase_date": "2026-09-21", "numbers": "[]", "amount": 0}},
	}
	for _, c := range cases {
		w, _ := postJSON(t, router, "/api/purchases/batch", gin.H{"items": []gin.H{c.item}})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s: status=%d body=%s", c.name, w.Code, w.Body.String())
		}
	}

	var count int64
	db.Model(&models.PurchaseRecord{}).Count(&count)
	if count != 0 {
		t.Fatalf("校验失败不应落库，实际 %d 条", count)
	}
}

func TestBatchCreateLimit(t *testing.T) {
	router, _ := setupPurchaseTest(t)

	items := make([]gin.H, 0, batchMaxItems+1)
	for i := 0; i < batchMaxItems+1; i++ {
		items = append(items, gin.H{
			"lottery_type":  "双色球",
			"issue_number":  "2026001",
			"purchase_date": "2026-09-21",
			"numbers":       `{"red":[1,5,12,20,28,33],"blue":[6]}`,
			"amount":        2,
		})
	}
	w, resp := postJSON(t, router, "/api/purchases/batch", gin.H{"items": items})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("超限应 400，实际 %d", w.Code)
	}
	if _, ok := resp["error"]; !ok {
		t.Fatalf("超限应返回 error: %s", w.Body.String())
	}
}

func TestBatchCreateAllOrNothing(t *testing.T) {
	router, db := setupPurchaseTest(t)

	// 第二条 amount 非法（binding required gt=0 之外还有金额为 0 的场景用 -1 绕过 gt 校验不行，
	// 这里用坏日期让第二条校验失败，验证整体不落库）
	w, _ := postJSON(t, router, "/api/purchases/batch", gin.H{
		"items": []gin.H{
			gin.H{"lottery_type": "双色球", "issue_number": "2026001", "purchase_date": "2026-09-21", "numbers": `{"red":[1],"blue":[2]}`, "amount": 2},
			gin.H{"lottery_type": "双色球", "issue_number": "2026002", "purchase_date": "bad-date", "numbers": `{"red":[1],"blue":[2]}`, "amount": 2},
		},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("应 400，实际 %d body=%s", w.Code, w.Body.String())
	}
	var count int64
	db.Model(&models.PurchaseRecord{}).Count(&count)
	if count != 0 {
		t.Fatalf("整体失败应 0 落库，实际 %d 条", count)
	}
}

func TestSingleCreateStillWorks(t *testing.T) {
	router, db := setupPurchaseTest(t)

	w, _ := postJSON(t, router, "/api/purchases", gin.H{
		"lottery_type":  "双色球",
		"issue_number":  "2026001",
		"purchase_date": "2026-09-21",
		"numbers":       `{"red":[1,5,12,20,28,33],"blue":[6]}`,
		"bet_type":      "单式",
		"amount":        2,
		"periods":       2,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var count int64
	db.Model(&models.PurchaseRecord{}).Count(&count)
	if count != 2 {
		t.Fatalf("多期展开后应为 2 条，实际 %d", count)
	}
}

func TestBatchIssueNonNumeric(t *testing.T) {
	router, db := setupPurchaseTest(t)

	w, resp := postJSON(t, router, "/api/purchases/batch", gin.H{
		"items": []gin.H{
			gin.H{
				"lottery_type":  "双色球",
				"issue_number":  "甲辰年第1期",
				"purchase_date": "2026-09-21",
				"numbers":       `{"red":[1,5,12,20,28,33],"blue":[6]}`,
				"amount":        2,
				"periods":       2,
			},
		},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	_ = resp
	var issues []string
	db.Model(&models.PurchaseRecord{}).Order("id").Pluck("issue_number", &issues)
	if len(issues) != 2 || issues[0] != "甲辰年第1期" || issues[1] != "甲辰年第1期" {
		t.Fatalf("非数字期号多期展开应保持原期号: %v", issues)
	}
}
