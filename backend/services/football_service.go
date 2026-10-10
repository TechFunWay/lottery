package services

import (
	"lottery-backend/database"
	"lottery-backend/logger"
	"lottery-backend/models"
	"lottery-backend/rules"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type FootballService struct{}

func (s *FootballService) CreateMatch(match *models.FootballMatch) error {
	var existing models.FootballMatch
	err := database.DB.Where("match_id = ?", match.MatchID).First(&existing).Error
	if err == nil {
		return fmt.Errorf("比赛 %s 已存在", match.MatchID)
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return database.DB.Create(match).Error
}

func (s *FootballService) GetMatches(league, status string, page, size int) ([]models.FootballMatch, int64, error) {
	var matches []models.FootballMatch
	var total int64
	query := database.DB.Model(&models.FootballMatch{})
	if league != "" {
		query = query.Where("league = ?", league)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("match_time DESC").Offset((page - 1) * size).Limit(size).Find(&matches).Error
	return matches, total, err
}

func (s *FootballService) UpdateMatch(match *models.FootballMatch) error {
	return database.DB.Save(match).Error
}

func (s *FootballService) DeleteMatch(id uint) error {
	return database.DB.Delete(&models.FootballMatch{}, id).Error
}

func (s *FootballService) GetMatchByID(id uint) (*models.FootballMatch, error) {
	var match models.FootballMatch
	if err := database.DB.First(&match, id).Error; err != nil {
		return nil, err
	}
	return &match, nil
}

func (s *FootballService) GetFinishedMatches() ([]models.FootballMatch, error) {
	var matches []models.FootballMatch
	err := database.DB.Where("status = ?", models.MatchFinished).Find(&matches).Error
	return matches, err
}

func (s *FootballService) FetchMatches() ([]*models.FootballMatch, error) {
	url := "http://api.huiniao.top/interface/home/lotteryHistory?type=jczq&page=1&limit=20"

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("网络请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp huiNiaoResp
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析赛程数据失败: %v", err)
	}

	if apiResp.Code != 1 || len(apiResp.Data.Data.List) == 0 {
		return nil, fmt.Errorf("未找到赛程数据: %s", apiResp.Info)
	}

	var results []*models.FootballMatch
	for _, item := range apiResp.Data.Data.List {
		matchTime, _ := time.Parse("2006-01-02", item.Day)
		match := &models.FootballMatch{
			MatchID:     item.Code,
			IssueNumber: item.Code,
			League:      item.One + "",
			HomeTeam:    item.Two + "",
			AwayTeam:    item.Three + "",
			MatchTime:   matchTime,
			Status:      models.MatchNotStarted,
			Source:      "auto",
		}
		results = append(results, match)
	}
	return results, nil
}

func (s *FootballService) FetchMatchResults() ([]*models.FootballMatch, error) {
	url := "http://api.huiniao.top/interface/home/lotteryHistory?type=jczq&page=1&limit=20"

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("网络请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp huiNiaoResp
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析比赛结果失败: %v", err)
	}

	if apiResp.Code != 1 || len(apiResp.Data.Data.List) == 0 {
		return nil, fmt.Errorf("未找到比赛结果: %s", apiResp.Info)
	}

	var results []*models.FootballMatch
	for _, item := range apiResp.Data.Data.List {
		matchTime, _ := time.Parse("2006-01-02", item.Day)
		match := &models.FootballMatch{
			MatchID:     item.Code,
			IssueNumber: item.Code,
			MatchTime:   matchTime,
			Status:      models.MatchFinished,
			Source:      "auto",
		}
		results = append(results, match)
	}
	return results, nil
}

func (s *FootballService) CreateBet(bet *models.FootballBet) error {
	return database.DB.Create(bet).Error
}

func (s *FootballService) GetBets(userID uint, status string, page, size int) ([]models.FootballBet, int64, error) {
	var bets []models.FootballBet
	var total int64
	query := database.DB.Model(&models.FootballBet{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&bets).Error
	return bets, total, err
}

func (s *FootballService) UpdateBet(bet *models.FootballBet) error {
	return database.DB.Save(bet).Error
}

func (s *FootballService) DeleteBet(id uint) error {
	return database.DB.Delete(&models.FootballBet{}, id).Error
}

func (s *FootballService) GetBetByID(id uint) (*models.FootballBet, error) {
	var bet models.FootballBet
	if err := database.DB.First(&bet, id).Error; err != nil {
		return nil, err
	}
	return &bet, nil
}

func (s *FootballService) CheckBetWinning(bet *models.FootballBet, matches []models.FootballMatch) {
	result := rules.CalculateFootballBet(bet.Selections, matches)

	if result.Hit {
		bet.Status = models.BetWon
		bet.WinAmount = result.WinAmount * float64(bet.Multiple) * 2
	} else {
		allFinished := true
		var selections []models.FootballSelection
		json.Unmarshal([]byte(bet.Selections), &selections)

		matchMap := make(map[string]models.FootballMatch)
		for _, m := range matches {
			matchMap[m.MatchID] = m
		}

		for _, sel := range selections {
			m, ok := matchMap[sel.MatchID]
			if !ok || m.Status != models.MatchFinished {
				allFinished = false
				break
			}
		}

		if allFinished {
			bet.Status = models.BetLost
		}
	}
}

func (s *FootballService) RecheckAllBets() error {
	var bets []models.FootballBet
	if err := database.DB.Where("status = ?", models.BetPending).Find(&bets).Error; err != nil {
		return err
	}

	matches, err := s.GetFinishedMatches()
	if err != nil {
		return err
	}

	for i := range bets {
		s.CheckBetWinning(&bets[i], matches)
		database.DB.Model(&models.FootballBet{}).Where("id = ?", bets[i].ID).Updates(map[string]interface{}{
			"status":     bets[i].Status,
			"win_amount": bets[i].WinAmount,
		})
	}

	logger.GetSugarLogger().Infof("竞彩足球重新检查完成，共检查 %d 条投注记录", len(bets))
	return nil
}

func (s *FootballService) GetFootballOverview(userID uint) (map[string]interface{}, error) {
	var totalBets int64
	var totalAmount float64
	var totalWinAmount float64
	var winCount int64

	database.DB.Model(&models.FootballBet{}).Where("user_id = ?", userID).Count(&totalBets)
	database.DB.Model(&models.FootballBet{}).Where("user_id = ?", userID).Select("COALESCE(SUM(amount), 0)").Scan(&totalAmount)
	database.DB.Model(&models.FootballBet{}).Where("user_id = ? AND status = ?", userID, models.BetWon).Count(&winCount)
	database.DB.Model(&models.FootballBet{}).Where("user_id = ? AND status = ?", userID, models.BetWon).Select("COALESCE(SUM(win_amount), 0)").Scan(&totalWinAmount)

	var winRate float64
	if totalBets > 0 {
		winRate = float64(winCount) / float64(totalBets) * 100
	}

	return map[string]interface{}{
		"total_bets":     totalBets,
		"total_amount":   totalAmount,
		"total_win":      totalWinAmount,
		"net_profit":     totalWinAmount - totalAmount,
		"win_count":      winCount,
		"win_rate":       winRate,
	}, nil
}
