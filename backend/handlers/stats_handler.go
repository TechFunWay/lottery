package handlers

import (
	"lottery-backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var statsService = &services.StatsService{}

// GetOverview GET /api/statistics/overview
func GetOverview(c *gin.Context) {
	userID, _ := c.Get("user_id")
	lotteryType := c.Query("lottery_type")
	stats, err := statsService.GetOverview(userID.(uint), lotteryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetPrizeDistribution GET /api/statistics/prizes
func GetPrizeDistribution(c *gin.Context) {
	lotteryType := c.Query("lottery_type")
	result, err := statsService.GetPrizeDistribution(lotteryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetNumberFrequency GET /api/statistics/numbers
func GetNumberFrequency(c *gin.Context) {
	lotteryType := c.Query("lottery_type")
	result, err := statsService.GetNumberFrequency(lotteryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetTrends GET /api/statistics/trends
func GetTrends(c *gin.Context) {
	userID, _ := c.Get("user_id")
	lotteryType := c.Query("lottery_type")
	months, _ := strconv.Atoi(c.DefaultQuery("months", "12"))
	if months < 1 || months > 24 {
		months = 12
	}
	result, err := statsService.GetTrends(userID.(uint), lotteryType, months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetRecentWinnings GET /api/statistics/recent-winnings
func GetRecentWinnings(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	result, err := statsService.GetRecentWinnings(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
