package handlers

import (
	"lottery-backend/models"
	"lottery-backend/services"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var footballService = &services.FootballService{}

type CreateFootballMatchRequest struct {
	MatchID       string `json:"match_id" binding:"required"`
	IssueNumber   string `json:"issue_number"`
	League        string `json:"league"`
	HomeTeam      string `json:"home_team" binding:"required"`
	AwayTeam      string `json:"away_team" binding:"required"`
	MatchTime     string `json:"match_time" binding:"required"`
	HomeScore     int    `json:"home_score"`
	AwayScore     int    `json:"away_score"`
	HalfHomeScore int    `json:"half_home_score"`
	HalfAwayScore int    `json:"half_away_score"`
	Handicap      float64 `json:"handicap"`
	Status        string `json:"status"`
}

func CreateFootballMatch(c *gin.Context) {
	var req CreateFootballMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	matchTime, err := time.Parse("2006-01-02", req.MatchTime)
	if err != nil {
		matchTime, _ = time.Parse("2006-01-02T15:04:05Z07:00", req.MatchTime)
	}

	issueNumber := req.IssueNumber
	if issueNumber == "" {
		issueNumber = req.MatchID
	}

	status := models.FootballMatchStatus(req.Status)
	if status == "" {
		if req.HomeScore > 0 || req.AwayScore > 0 {
			status = models.MatchFinished
		} else {
			status = models.MatchNotStarted
		}
	}

	match := &models.FootballMatch{
		MatchID:       req.MatchID,
		IssueNumber:   issueNumber,
		League:        req.League,
		HomeTeam:      req.HomeTeam,
		AwayTeam:      req.AwayTeam,
		MatchTime:     matchTime,
		Status:        status,
		HomeScore:     req.HomeScore,
		AwayScore:     req.AwayScore,
		HalfHomeScore: req.HalfHomeScore,
		HalfAwayScore: req.HalfAwayScore,
		Handicap:      req.Handicap,
		Source:        "manual",
	}

	if err := footballService.CreateMatch(match); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go footballService.RecheckAllBets()

	c.JSON(http.StatusCreated, gin.H{"data": match, "message": "比赛创建成功"})
}

func GetFootballMatches(c *gin.Context) {
	league := c.Query("league")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	matches, total, err := footballService.GetMatches(league, status, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  matches,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func UpdateFootballMatch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	match, err := footballService.GetMatchByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "比赛不存在"})
		return
	}

	var req CreateFootballMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	matchTime, err := time.Parse("2006-01-02", req.MatchTime)
	if err != nil {
		matchTime = match.MatchTime
	}

	match.MatchID = req.MatchID
	match.IssueNumber = req.IssueNumber
	match.League = req.League
	match.HomeTeam = req.HomeTeam
	match.AwayTeam = req.AwayTeam
	match.MatchTime = matchTime
	match.HomeScore = req.HomeScore
	match.AwayScore = req.AwayScore
	match.HalfHomeScore = req.HalfHomeScore
	match.HalfAwayScore = req.HalfAwayScore
	match.Handicap = req.Handicap

	if req.Status != "" {
		match.Status = models.FootballMatchStatus(req.Status)
	} else if req.HomeScore > 0 || req.AwayScore > 0 {
		match.Status = models.MatchFinished
	}

	if err := footballService.UpdateMatch(match); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go footballService.RecheckAllBets()

	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": match})
}

func DeleteFootballMatch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := footballService.DeleteMatch(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func FetchFootballMatches(c *gin.Context) {
	matches, err := footballService.FetchMatches()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "抓取赛程失败: " + err.Error()})
		return
	}

	if len(matches) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "当前数据源暂无可用赛程数据,请手动录入或稍后重试",
			"data": gin.H{
				"total":       0,
				"saved_count": 0,
				"exist_count": 0,
				"empty":       true,
			},
		})
		return
	}

	savedCount := 0
	existCount := 0
	for _, match := range matches {
		if err := footballService.CreateMatch(match); err != nil {
			existCount++
			continue
		}
		savedCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "赛程抓取完成",
		"data": gin.H{
			"total":       len(matches),
			"saved_count": savedCount,
			"exist_count": existCount,
			"empty":       false,
		},
	})
}

func FetchFootballResults(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(uint)

	matches, err := footballService.FetchMatchResults(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "抓取比赛结果失败: " + err.Error()})
		return
	}

	if len(matches) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "当前数据源暂无可用比赛结果,请稍后重试",
			"data": gin.H{
				"total":         0,
				"updated_count": 0,
				"empty":         true,
			},
		})
		return
	}

	updatedCount := 0
	for _, fetchedMatch := range matches {
		var existing models.FootballMatch
		if err := services.GetDB().Where("match_id = ?", fetchedMatch.MatchID).First(&existing).Error; err != nil {
			continue
		}

		existing.HomeScore = fetchedMatch.HomeScore
		existing.AwayScore = fetchedMatch.AwayScore
		existing.HalfHomeScore = fetchedMatch.HalfHomeScore
		existing.HalfAwayScore = fetchedMatch.HalfAwayScore
		existing.Status = models.MatchFinished

		if err := footballService.UpdateMatch(&existing); err != nil {
			continue
		}
		updatedCount++
	}

	go footballService.RecheckAllBets()

	c.JSON(http.StatusOK, gin.H{
		"message": "比赛结果更新完成",
		"data": gin.H{
			"total":         len(matches),
			"updated_count": updatedCount,
			"empty":         false,
		},
	})
}

type CreateFootballBetRequest struct {
	BetType    string  `json:"bet_type"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Multiple   int     `json:"multiple"`
	Selections string  `json:"selections" binding:"required"`
	Remark     string  `json:"remark"`
}

func CreateFootballBet(c *gin.Context) {
	var req CreateFootballBetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	betType := models.FootballBetType(req.BetType)
	if betType == "" {
		betType = models.BetSingle
	}

	multiple := req.Multiple
	if multiple <= 0 {
		multiple = 1
	}

	bet := &models.FootballBet{
		UserID:     userID.(uint),
		BetType:    betType,
		Amount:     req.Amount,
		Multiple:   multiple,
		Status:     models.BetPending,
		Selections: req.Selections,
		Remark:     req.Remark,
	}

	if err := footballService.CreateBet(bet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建投注记录失败: " + err.Error()})
		return
	}

	go footballService.RecheckAllBets()

	c.JSON(http.StatusCreated, gin.H{"data": bet, "message": "投注记录创建成功"})
}

func GetFootballBets(c *gin.Context) {
	userID, _ := c.Get("user_id")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	bets, total, err := footballService.GetBets(userID.(uint), status, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  bets,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func UpdateFootballBet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	bet, err := footballService.GetBetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "投注记录不存在"})
		return
	}

	var req CreateFootballBetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bet.BetType = models.FootballBetType(req.BetType)
	bet.Amount = req.Amount
	bet.Selections = req.Selections
	bet.Remark = req.Remark
	if req.Multiple > 0 {
		bet.Multiple = req.Multiple
	}

	if err := footballService.UpdateBet(bet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go footballService.RecheckAllBets()

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func DeleteFootballBet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := footballService.DeleteBet(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func RecheckFootballBets(c *gin.Context) {
	go footballService.RecheckAllBets()
	c.JSON(http.StatusOK, gin.H{"message": "正在重新检查竞彩足球中奖记录"})
}

func GetFootballOverview(c *gin.Context) {
	userID, _ := c.Get("user_id")
	overview, err := footballService.GetFootballOverview(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": overview})
}

const apiFootballRegistrationURL = "https://dashboard.api-football.com/prod/register"

func GetFootballConfigStatus(c *gin.Context) {
	key, source := configService.ResolveAPIFootballKey(0)
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"configured":       key != "",
			"source":           string(source),
			"masked_key":       services.MaskAPIFootballKey(key),
			"registration_url": apiFootballRegistrationURL,
		},
	})
}

func GetMyFootballConfig(c *gin.Context) {
	uid, _ := c.Get("user_id")
	key, source := configService.ResolveAPIFootballKey(uid.(uint))
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"configured": key != "",
			"source":     string(source),
			"masked_key": services.MaskAPIFootballKey(key),
		},
	})
}

type setFootballKeyRequest struct {
	Key string `json:"key"`
}

func SetMyFootballConfig(c *gin.Context) {
	uid, _ := c.Get("user_id")
	var req setFootballKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := configService.SetAPIFootballKey(uid.(uint), req.Key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败: " + err.Error()})
		return
	}
	msg := "已保存个人 API-Football Key"
	if strings.TrimSpace(req.Key) == "" {
		msg = "已清除个人 API-Football Key,将降级使用管理员/环境变量配置"
	}
	c.JSON(http.StatusOK, gin.H{"message": msg})
}

func SetGlobalFootballConfig(c *gin.Context) {
	var req setFootballKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := configService.SetGlobalAPIFootballKey(req.Key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败: " + err.Error()})
		return
	}
	msg := "已保存全局 API-Football Key,所有未自配用户均可使用"
	if strings.TrimSpace(req.Key) == "" {
		msg = "已清除全局 API-Football Key,赛果抓取将仅对自配用户可用"
	}
	c.JSON(http.StatusOK, gin.H{"message": msg})
}

func TestFootballKey(c *gin.Context) {
	uid, _ := c.Get("user_id")
	var req setFootballKeyRequest
	_ = c.ShouldBindJSON(&req)

	testKey := strings.TrimSpace(req.Key)
	if testKey == "" {
		k, _ := configService.ResolveAPIFootballKey(uid.(uint))
		testKey = k
	}

	if err := services.TestAPIFootballKey(testKey); err != nil {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{
			"success": false,
			"message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"success": true,
		"message": "Key 有效,可正常抓取赛果",
	}})
}
