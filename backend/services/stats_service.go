package services

import (
	"lottery-backend/database"
	"lottery-backend/models"
	"encoding/json"
	"time"
)

type StatsService struct{}

// OverviewStats 盈亏总览
type OverviewStats struct {
	TotalInvestment float64 `json:"total_investment"` // 总投入
	TotalWinning    float64 `json:"total_winning"`    // 总中奖
	NetProfit       float64 `json:"net_profit"`       // 净盈亏
	TotalBets       int64   `json:"total_bets"`       // 总购买次数
	WinCount        int64   `json:"win_count"`        // 中奖次数
	WinRate         float64 `json:"win_rate"`         // 中奖率
}

func (s *StatsService) GetOverview(userID uint, lotteryType string) (*OverviewStats, error) {
	stats := &OverviewStats{}

	query := database.DB.Model(&models.PurchaseRecord{}).Where("user_id = ?", userID)
	if lotteryType != "" {
		query = query.Where("lottery_type = ?", lotteryType)
	}
	query.Count(&stats.TotalBets)
	query.Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalInvestment)

	winQuery := database.DB.Model(&models.WinningRecord{}).
		Joins("JOIN purchase_records ON winning_records.purchase_id = purchase_records.id").
		Where("winning_records.prize_level > 0 AND purchase_records.user_id = ?", userID)
	if lotteryType != "" {
		winQuery = winQuery.Where("winning_records.lottery_type = ?", lotteryType)
	}
	winQuery.Count(&stats.WinCount)
	winQuery.Select("COALESCE(SUM(winning_records.prize_amount), 0)").Scan(&stats.TotalWinning)

	stats.NetProfit = stats.TotalWinning - stats.TotalInvestment
	if stats.TotalBets > 0 {
		stats.WinRate = float64(stats.WinCount) / float64(stats.TotalBets) * 100
	}
	return stats, nil
}

// PrizeDistribution 奖级分布
type PrizeDistribution struct {
	PrizeName string  `json:"prize_name"`
	Count     int64   `json:"count"`
	Amount    float64 `json:"amount"`
}

func (s *StatsService) GetPrizeDistribution(lotteryType string) ([]PrizeDistribution, error) {
	results := make([]PrizeDistribution, 0) // 保证返回空数组而非 null
	query := database.DB.Model(&models.WinningRecord{}).
		Where("prize_level > 0").
		Select("prize_name, COUNT(*) as count, SUM(prize_amount) as amount").
		Group("prize_name").
		Order("count DESC")
	if lotteryType != "" {
		query = query.Where("lottery_type = ?", lotteryType)
	}
	err := query.Scan(&results).Error
	return results, err
}

// NumberFrequency 号码频率
type NumberFrequency struct {
	Number    int    `json:"number"`
	LotteryType string `json:"lottery_type"`
	Position  string `json:"position"` // red/blue/front/back/main/special
	Count     int    `json:"count"`
}

func (s *StatsService) GetNumberFrequency(lotteryType string) (map[string][]NumberFrequency, error) {
	result := make(map[string][]NumberFrequency)

	var purchases []models.PurchaseRecord
	query := database.DB.Model(&models.PurchaseRecord{})
	if lotteryType != "" {
		query = query.Where("lottery_type = ?", lotteryType)
	}
	query.Find(&purchases)

	// 统计开奖号码频率
	var draws []models.DrawResult
	drawQuery := database.DB.Model(&models.DrawResult{})
	if lotteryType != "" {
		drawQuery = drawQuery.Where("lottery_type = ?", lotteryType)
	}
	drawQuery.Find(&draws)

	// 统计开奖号码
	for _, draw := range draws {
		freqs := parseNumberFrequency(string(draw.LotteryType), draw.Numbers)
		for pos, nums := range freqs {
			result[pos] = mergeFrequency(result[pos], nums)
		}
	}

	return result, nil
}

func parseNumberFrequency(lotteryType, numbersJSON string) map[string][]int {
	result := make(map[string][]int)
	switch models.LotteryType(lotteryType) {
	case models.ShuangSeQiu:
		var nums struct {
			Red  []int `json:"red"`
			Blue []int `json:"blue"`
		}
		if json.Unmarshal([]byte(numbersJSON), &nums) == nil {
			result["red"] = nums.Red
			result["blue"] = nums.Blue
		}
	case models.DaLeTou:
		var nums struct {
			Front []int `json:"front"`
			Back  []int `json:"back"`
		}
		if json.Unmarshal([]byte(numbersJSON), &nums) == nil {
			result["front"] = nums.Front
			result["back"] = nums.Back
		}
	default:
		var nums []int
		if json.Unmarshal([]byte(numbersJSON), &nums) == nil {
			result["main"] = nums
		}
	}
	return result
}

func mergeFrequency(existing []NumberFrequency, newNums []int) []NumberFrequency {
	countMap := make(map[int]int)
	for _, f := range existing {
		countMap[f.Number] = f.Count
	}
	for _, n := range newNums {
		countMap[n]++
	}
	result := make([]NumberFrequency, 0, len(countMap))
	for num, count := range countMap {
		result = append(result, NumberFrequency{Number: num, Count: count})
	}
	return result
}

// TrendData 中奖率趋势
type TrendData struct {
	Month       string  `json:"month"`
	TotalBets   int64   `json:"total_bets"`
	WinCount    int64   `json:"win_count"`
	Investment  float64 `json:"investment"`
	WinAmount   float64 `json:"win_amount"`
}

func (s *StatsService) GetTrends(userID uint, lotteryType string, months int) ([]TrendData, error) {
	var trends []TrendData
	now := time.Now()

	for i := months - 1; i >= 0; i-- {
		month := now.AddDate(0, -i, 0)
		start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
		end := start.AddDate(0, 1, 0)

		var totalBets int64
		var investment float64
		purchaseQuery := database.DB.Model(&models.PurchaseRecord{}).
			Where("purchase_date >= ? AND purchase_date < ? AND user_id = ?", start, end, userID)
		if lotteryType != "" {
			purchaseQuery = purchaseQuery.Where("lottery_type = ?", lotteryType)
		}
		purchaseQuery.Count(&totalBets)
		purchaseQuery.Select("COALESCE(SUM(amount), 0)").Scan(&investment)

		var winCount int64
		var winAmount float64
		winQuery := database.DB.Model(&models.WinningRecord{}).
			Joins("JOIN purchase_records ON winning_records.purchase_id = purchase_records.id").
			Where("purchase_records.purchase_date >= ? AND purchase_records.purchase_date < ? AND winning_records.prize_level > 0 AND purchase_records.user_id = ?", start, end, userID)
		if lotteryType != "" {
			winQuery = winQuery.Where("winning_records.lottery_type = ?", lotteryType)
		}
		winQuery.Count(&winCount)
		winQuery.Select("COALESCE(SUM(winning_records.prize_amount), 0)").Scan(&winAmount)

		trends = append(trends, TrendData{
			Month:      month.Format("2006-01"),
			TotalBets:  totalBets,
			WinCount:   winCount,
			Investment: investment,
			WinAmount:  winAmount,
		})
	}
	return trends, nil
}

// RecentWinnings 最近中奖
func (s *StatsService) GetRecentWinnings(limit int) ([]models.WinningRecord, error) {
	winnings := make([]models.WinningRecord, 0)
	err := database.DB.Preload("Purchase").Where("prize_level > 0").
		Order("created_at DESC").Limit(limit).Find(&winnings).Error
	return winnings, err
}
