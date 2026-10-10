package handlers

import (
	"encoding/json"
	"fmt"
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
	Multiple     int     `json:"multiple"`
	Append       bool    `json:"append"`
	Periods      int     `json:"periods"`
	Remark       string  `json:"remark"`
}

// buildPurchaseRecords 校验并构建一条请求对应的购买记录；periods>1 时按期号递增展开为多条。
// 返回错误信息为空串表示校验通过。
func buildPurchaseRecords(userID uint, req CreatePurchaseRequest) ([]*models.PurchaseRecord, string) {
	purchaseDate, err := time.Parse("2006-01-02", req.PurchaseDate)
	if err != nil {
		return nil, "日期格式错误，请使用 YYYY-MM-DD"
	}
	if !json.Valid([]byte(req.Numbers)) {
		return nil, "号码格式错误"
	}
	if req.Amount <= 0 {
		return nil, "金额必须大于 0"
	}
	if req.LotteryType == "" {
		return nil, "彩票类型不能为空"
	}

	betType := models.BetType(req.BetType)
	if betType == "" {
		betType = models.DanShi
	}

	// 参数默认值与校验
	multiple := req.Multiple
	if multiple < 1 {
		multiple = 1
	}
	if multiple > 99 {
		multiple = 99
	}

	periods := req.Periods
	if periods < 1 {
		periods = 1
	}
	if periods > 10 {
		periods = 10
	}

	records := make([]*models.PurchaseRecord, 0, periods)
	for i := 0; i < periods; i++ {
		issueNum := req.IssueNumber
		if i > 0 {
			// 尝试将期号解析为数字并递增
			if n, err := strconv.Atoi(req.IssueNumber); err == nil {
				issueNum = fmt.Sprintf("%d", n+i)
			}
		}
		records = append(records, &models.PurchaseRecord{
			UserID:       userID,
			LotteryType:  models.LotteryType(req.LotteryType),
			IssueNumber:  issueNum,
			PurchaseDate: purchaseDate,
			Numbers:      req.Numbers,
			BetType:      betType,
			Amount:       req.Amount,
			Multiple:     multiple,
			Append:       req.Append,
			Periods:      periods,
			Remark:       req.Remark,
			Status:       "待开奖",
		})
	}
	return records, ""
}

// CreatePurchase POST /api/purchases
func CreatePurchase(c *gin.Context) {
	var req CreatePurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	records, errMsg := buildPurchaseRecords(userID.(uint), req)
	if errMsg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	if len(records) == 1 {
		if err := purchaseService.CreatePurchase(records[0]); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建记录失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": records[0], "message": "创建成功"})
		return
	}

	// 多期投注
	if err := purchaseService.CreatePurchasesBatch(records); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建多期记录失败: " + err.Error()})
		return
	}

	created := make([]models.PurchaseRecord, 0, len(records))
	for _, r := range records {
		created = append(created, *r)
	}
	c.JSON(http.StatusCreated, gin.H{"data": created, "message": "创建成功，共 " + strconv.Itoa(len(records)) + " 期"})
}

// 单次批量导入的上限：条目数与展开后期数后的总记录数
const (
	batchMaxItems   = 100
	batchMaxRecords = 500
)

type BatchPurchaseRequest struct {
	Items []CreatePurchaseRequest `json:"items" binding:"required,min=1"`
}

// CreatePurchaseBatch POST /api/purchases/batch
// 一次导入多注号码，每条支持 periods 多期展开，整体在一个事务内落库。
func CreatePurchaseBatch(c *gin.Context) {
	var req BatchPurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Items) > batchMaxItems {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("单次最多导入 %d 条记录", batchMaxItems)})
		return
	}

	userID, _ := c.Get("user_id")

	records := make([]*models.PurchaseRecord, 0, len(req.Items))
	for i, item := range req.Items {
		itemRecords, errMsg := buildPurchaseRecords(userID.(uint), item)
		if errMsg != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第 %d 条：%s", i+1, errMsg)})
			return
		}
		records = append(records, itemRecords...)
	}
	if len(records) > batchMaxRecords {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("展开后期数后共 %d 条，超过单次上限 %d 条", len(records), batchMaxRecords)})
		return
	}

	if err := purchaseService.CreatePurchasesBatch(records); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量创建记录失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    records,
		"total":   len(records),
		"message": "导入成功，共 " + strconv.Itoa(len(records)) + " 条",
	})
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
