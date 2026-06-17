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
	"strconv"
	"time"

	"gorm.io/gorm"
)

var configService = ConfigService{}

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

// httpGetJSON 走 GET,自动设置 iPhone UA 以绕开阿里云 WAF,解析 JSON 到 v。
func httpGetJSON(url string, headers map[string]string, timeout time.Duration, v interface{}) error {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	for k, val := range headers {
		req.Header.Set(k, val)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("网络请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		snippet := string(body)
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, snippet)
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("解析 JSON 失败: %w (body 前 200 字节: %.200s)", err, string(body))
	}
	return nil
}

// sportteryScheduleResp sporttery 赛程接口响应
type sportteryScheduleResp struct {
	Success      bool   `json:"success"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	Value        struct {
		MatchInfoList []struct {
			BusinessDate string             `json:"businessDate"`
			SubMatchList []sportterySubMatch `json:"subMatchList"`
		} `json:"matchInfoList"`
	} `json:"value"`
}

type sportterySubMatch struct {
	MatchNumStr     string `json:"matchNumStr"`
	MatchStatus     string `json:"matchStatus"`
	LeagueAllName   string `json:"leagueAllName"`
	HomeTeamAbbName string `json:"homeTeamAbbName"`
	AwayTeamAbbName string `json:"awayTeamAbbName"`
	MatchDate       string `json:"matchDate"`
	MatchTime       string `json:"matchTime"`
	Hhad            struct {
		GoalLine string `json:"goalLine"`
	} `json:"hhad"`
}

type apiFootballFixture struct {
	Fixture struct {
		ID     int64  `json:"id"`
		Date   string `json:"date"`
		Status struct {
			Short string `json:"short"`
		} `json:"status"`
	} `json:"fixture"`
	League struct {
		Name string `json:"name"`
	} `json:"league"`
	Teams struct {
		Home struct {
			Name string `json:"name"`
		} `json:"home"`
		Away struct {
			Name string `json:"name"`
		} `json:"away"`
	} `json:"teams"`
	Goals struct {
		Home *int `json:"home"`
		Away *int `json:"away"`
	} `json:"goals"`
	Score struct {
		Halftime struct {
			Home *int `json:"home"`
			Away *int `json:"away"`
		} `json:"halftime"`
	} `json:"score"`
}

type apiFootballFixturesResp struct {
	Response []apiFootballFixture `json:"response"`
}

// sporttery 赛程接口:实测必须伪装 iPhone Safari UA + sporttery.cn Referer 才能过阿里云 WAF。
const (
	sportteryScheduleURL = "https://webapi.sporttery.cn/gateway/jc/football/getMatchCalculatorV1.qry?poolCode=hhad,had,crs,ttg,hafu&channel=c"
	sportteryReferer     = "https://www.sporttery.cn/"
)

func (s *FootballService) FetchMatches() ([]*models.FootballMatch, error) {
	headers := map[string]string{
		"Referer": sportteryReferer,
	}

	var resp sportteryScheduleResp
	if err := httpGetJSON(sportteryScheduleURL, headers, 12*time.Second, &resp); err != nil {
		return nil, fmt.Errorf("竞彩足球赛程接口不可达: %w", err)
	}

	if !resp.Success || resp.ErrorCode != "0" {
		return nil, fmt.Errorf("竞彩足球赛程接口业务错误: %s", resp.ErrorMessage)
	}

	if len(resp.Value.MatchInfoList) == 0 {
		logger.GetSugarLogger().Info("⚽ 竞彩足球官方赛程暂无可用数据")
		return nil, nil
	}

	loc, locErr := time.LoadLocation("Asia/Shanghai")
	if locErr != nil {
		loc = time.UTC
	}

	var results []*models.FootballMatch
	totalParsed := 0
	for _, day := range resp.Value.MatchInfoList {
		for _, m := range day.SubMatchList {
			matchTime, _ := time.ParseInLocation("2006-01-02 15:04:05", m.MatchDate+" "+m.MatchTime, loc)
			handicap, _ := strconv.ParseFloat(m.Hhad.GoalLine, 64)

			status := models.MatchNotStarted
			switch m.MatchStatus {
			case "Finished":
				status = models.MatchFinished
			}

			results = append(results, &models.FootballMatch{
				MatchID:     m.MatchNumStr,
				IssueNumber: m.MatchNumStr,
				League:      m.LeagueAllName,
				HomeTeam:    m.HomeTeamAbbName,
				AwayTeam:    m.AwayTeamAbbName,
				MatchTime:   matchTime,
				Status:      status,
				Handicap:    handicap,
				Source:      "auto-sporttery",
			})
			totalParsed++
		}
	}

	logger.GetSugarLogger().Infof("⚽ sporttery 赛程抓取完成,共解析 %d 场", totalParsed)
	return results, nil
}

func (s *FootballService) FetchMatchResults(userID uint) ([]*models.FootballMatch, error) {
	apiKey, source := configService.ResolveAPIFootballKey(userID)
	if apiKey == "" {
		logger.GetSugarLogger().Infof("⚽ 用户 #%d 未配置 API_FOOTBALL_KEY(per-user / admin / env 均空),赛果自动抓取暂不可用", userID)
		return nil, nil
	}
	logger.GetSugarLogger().Debugf("⚽ 使用 %s 来源的 API-Football Key", source)

	from := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	to := time.Now().Format("2006-01-02")
	url := fmt.Sprintf("https://v3.football.api-sports.io/fixtures?from=%s&to=%s&status=FT-AET-PEN", from, to)
	headers := map[string]string{
		"x-apisports-key": apiKey,
	}

	var resp apiFootballFixturesResp
	if err := httpGetJSON(url, headers, 12*time.Second, &resp); err != nil {
		return nil, fmt.Errorf("api-football 赛果接口不可达: %w", err)
	}

	if len(resp.Response) == 0 {
		logger.GetSugarLogger().Info("⚽ api-football 近 7 天无已完赛赛果")
		return nil, nil
	}

	var updated []*models.FootballMatch
	skippedNoMap, skippedNoMatch, skippedNoStatus := 0, 0, 0
	for _, f := range resp.Response {
		short := f.Fixture.Status.Short
		if short != "FT" && short != "AET" && short != "PEN" {
			skippedNoStatus++
			continue
		}

		homeCN := translateTeamName(f.Teams.Home.Name)
		awayCN := translateTeamName(f.Teams.Away.Name)
		if homeCN == "" || awayCN == "" {
			skippedNoMap++
			continue
		}

		matchTime, err := time.Parse(time.RFC3339, f.Fixture.Date)
		if err != nil {
			continue
		}

		var existing models.FootballMatch
		err = database.DB.Where(
			"home_team = ? AND away_team = ? AND status IN (?, ?) AND match_time BETWEEN ? AND ?",
			homeCN, awayCN,
			models.MatchNotStarted, models.MatchInProgress,
			matchTime.Add(-24*time.Hour), matchTime.Add(24*time.Hour),
		).First(&existing).Error
		if err != nil {
			skippedNoMatch++
			continue
		}

		if f.Goals.Home != nil {
			existing.HomeScore = *f.Goals.Home
		}
		if f.Goals.Away != nil {
			existing.AwayScore = *f.Goals.Away
		}
		if f.Score.Halftime.Home != nil {
			existing.HalfHomeScore = *f.Score.Halftime.Home
		}
		if f.Score.Halftime.Away != nil {
			existing.HalfAwayScore = *f.Score.Halftime.Away
		}
		existing.Status = models.MatchFinished
		if existing.Source == "" || existing.Source == "manual" {
			existing.Source = "auto-apifootball"
		}
		if err := database.DB.Save(&existing).Error; err != nil {
			continue
		}
		updated = append(updated, &existing)
	}

	logger.GetSugarLogger().Infof(
		"⚽ api-football 赛果回填完成,成功 %d 场(队名未翻译 %d,未匹配赛程 %d,状态不符 %d)",
		len(updated), skippedNoMap, skippedNoMatch, skippedNoStatus)
	return updated, nil
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
