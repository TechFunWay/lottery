package handlers

import (
	"lottery-backend/models"
	"lottery-backend/services"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var drawService = &services.DrawService{}

type CreateDrawRequest struct {
	LotteryType string `json:"lottery_type" binding:"required"`
	IssueNumber string `json:"issue_number" binding:"required"`
	DrawDate    string `json:"draw_date" binding:"required"`
	Numbers     string `json:"numbers" binding:"required"`
}

// CreateDraw POST /api/draws
func CreateDraw(c *gin.Context) {
	var req CreateDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	drawDate, err := time.Parse("2006-01-02", req.DrawDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误"})
		return
	}

	draw := &models.DrawResult{
		LotteryType: models.LotteryType(req.LotteryType),
		IssueNumber: req.IssueNumber,
		DrawDate:    drawDate,
		Numbers:     req.Numbers,
		Source:      "manual",
	}

	if err := drawService.CreateDrawResult(draw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存开奖结果失败: " + err.Error()})
		return
	}

	// 自动计算中奖
	go winningService.CheckAndSaveWinnings(draw.ID)

	c.JSON(http.StatusCreated, gin.H{"data": draw, "message": "开奖结果录入成功"})
}

// GetDraws GET /api/draws
func GetDraws(c *gin.Context) {
	lotteryType := c.Query("lottery_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	draws, total, err := drawService.GetDrawResults(lotteryType, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  draws,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// FetchDraw GET /api/draws/fetch
func FetchDraw(c *gin.Context) {
	lotteryType := models.LotteryType(c.Query("lottery_type"))
	issue := c.Query("issue")

	if lotteryType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请指定彩票类型"})
		return
	}

	draw, err := drawService.FetchDrawResult(lotteryType, issue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "抓取开奖结果失败: " + err.Error()})
		return
	}

	// 检查是否已存在
	existing, _ := drawService.GetDrawByIssue(string(lotteryType), draw.IssueNumber)
	if existing != nil {
		c.JSON(http.StatusOK, gin.H{"data": existing, "message": "该期开奖结果已存在"})
		return
	}

	if err := drawService.CreateDrawResult(draw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存抓取结果失败: " + err.Error()})
		return
	}

	go winningService.CheckAndSaveWinnings(draw.ID)
	c.JSON(http.StatusCreated, gin.H{"data": draw, "message": "开奖结果抓取成功"})
}

// FetchBatchDraws POST /api/draws/fetch-batch
func FetchBatchDraws(c *gin.Context) {
	type BatchFetchRequest struct {
		LotteryType string `json:"lottery_type" binding:"required"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		Count       int    `json:"count"`
	}

	var req BatchFetchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Count <= 0 {
		req.Count = 10
	}
	if req.Count > 100 {
		req.Count = 100
	}

	lotteryType := models.LotteryType(req.LotteryType)
	draws, err := drawService.FetchBatchDrawResults(lotteryType, req.StartDate, req.EndDate, req.Count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量抓取失败: " + err.Error()})
		return
	}

	// 保存并统计
	savedCount := 0
	existCount := 0
	for _, draw := range draws {
		// 直接创建，CreateDrawResult 内部会处理已存在的情况
		if err := drawService.CreateDrawResult(draw); err != nil {
			if strings.Contains(err.Error(), "已存在") {
				existCount++
			} else {
				log.Printf("保存开奖结果失败: %v", err)
			}
			continue
		}
		savedCount++
		go winningService.CheckAndSaveWinnings(draw.ID)
	}

	msg := fmt.Sprintf("成功获取 %d 条开奖记录，新增 %d 条", len(draws), savedCount)
	if existCount > 0 {
		msg += fmt.Sprintf("，已存在 %d 条", existCount)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": msg,
		"data":    gin.H{"count": savedCount, "exist_count": existCount, "total": len(draws)},
	})
}

// UpdateDraw PUT /api/draws/:id
func UpdateDraw(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req CreateDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	drawDate, err := time.Parse("2006-01-02", req.DrawDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误"})
		return
	}

	// 获取原开奖记录
	var draw models.DrawResult
	if err := drawService.GetDB().First(&draw, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "开奖记录不存在"})
		return
	}

	// 更新字段
	draw.LotteryType = models.LotteryType(req.LotteryType)
	draw.IssueNumber = req.IssueNumber
	draw.DrawDate = drawDate
	draw.Numbers = req.Numbers

	if err := drawService.UpdateDrawResult(&draw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败: " + err.Error()})
		return
	}

	// 重新计算中奖
	go winningService.CheckAndSaveWinnings(draw.ID)

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteDraw DELETE /api/draws/:id
func DeleteDraw(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := drawService.DeleteDrawResult(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetWinnings GET /api/winnings
func GetWinnings(c *gin.Context) {
	lotteryType := c.Query("lottery_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	var winnings []models.WinningRecord
	var total int64
	db := services.GetDB()
	q := db.Model(&models.WinningRecord{}).
		Joins("JOIN purchase_records ON purchase_records.id = winning_records.purchase_id AND purchase_records.deleted_at IS NULL").
		Joins("JOIN draw_results ON draw_results.id = winning_records.draw_id AND draw_results.deleted_at IS NULL").
		Preload("Purchase").Preload("Draw")
	if lotteryType != "" {
		q = q.Where("winning_records.lottery_type = ?", lotteryType)
	}
	q.Count(&total)
	q.Order("winning_records.created_at DESC").Offset((page - 1) * size).Limit(size).Find(&winnings)

	c.JSON(http.StatusOK, gin.H{
		"data":  winnings,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// RecheckWinnings POST /api/winnings/recheck - 重新检查所有中奖
func RecheckWinnings(c *gin.Context) {
	// 获取所有开奖记录
	var draws []models.DrawResult
	if err := services.GetDB().Find(&draws).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询开奖记录失败"})
		return
	}

	// 逐个重新计算
	go func() {
		for _, draw := range draws {
			winningService.CheckAndSaveWinnings(draw.ID)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "正在重新计算中奖记录"})
}
