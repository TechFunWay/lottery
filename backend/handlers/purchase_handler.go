package handlers

import (
	"lottery-backend/models"
	"lottery-backend/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var purchaseService = &services.PurchaseService{}
var winningService = &services.WinningService{}

type CreatePurchaseRequest struct {
	LotteryType  string  `json:"lottery_type" binding:"required"`
	IssueNumber  string  `json:"issue_number" binding:"required"`
	PurchaseDate string  `json:"purchase_date" binding:"required"`
	Numbers      string  `json:"numbers" binding:"required"`
	BetType      string  `json:"bet_type"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	Remark       string  `json:"remark"`
}

// CreatePurchase POST /api/purchases
func CreatePurchase(c *gin.Context) {
	var req CreatePurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	purchaseDate, err := time.Parse("2006-01-02", req.PurchaseDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误，请使用 YYYY-MM-DD"})
		return
	}

	betType := models.BetType(req.BetType)
	if betType == "" {
		betType = models.DanShi
	}

	// 获取当前用户ID
	userID, _ := c.Get("user_id")

	purchase := &models.PurchaseRecord{
		UserID:       userID.(uint),
		LotteryType:  models.LotteryType(req.LotteryType),
		IssueNumber:  req.IssueNumber,
		PurchaseDate: purchaseDate,
		Numbers:      req.Numbers,
		BetType:      betType,
		Amount:       req.Amount,
		Remark:       req.Remark,
		Status:       "待开奖",
	}

	if err := purchaseService.CreatePurchase(purchase); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建记录失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": purchase, "message": "创建成功"})
}

// GetPurchases GET /api/purchases
func GetPurchases(c *gin.Context) {
	userID, _ := c.Get("user_id")
	lotteryType := c.Query("lottery_type")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	purchases, total, err := purchaseService.GetPurchases(userID.(uint), lotteryType, status, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  purchases,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// DeletePurchase DELETE /api/purchases/:id
func DeletePurchase(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	if err := purchaseService.DeletePurchase(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// UpdatePurchase PUT /api/purchases/:id
func UpdatePurchase(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	var req CreatePurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	purchaseDate, _ := time.Parse("2006-01-02", req.PurchaseDate)
	purchase := &models.PurchaseRecord{
		LotteryType:  models.LotteryType(req.LotteryType),
		IssueNumber:  req.IssueNumber,
		PurchaseDate: purchaseDate,
		Numbers:      req.Numbers,
		BetType:      models.BetType(req.BetType),
		Amount:       req.Amount,
		Remark:       req.Remark,
	}
	if err := purchaseService.UpdatePurchase(uint(id), purchase); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}
