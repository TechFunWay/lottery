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

// DrawService 开奖结果服务
type DrawService struct{}

// 惠鸟彩票API响应结构
type huiNiaoResp struct {
	Code int    `json:"code"`
	Info string `json:"info"`
	Data struct {
		Last struct {
			Code          string `json:"code"`
			Day           string `json:"day"`
			One           string `json:"one"`
			Two           string `json:"two"`
			Three         string `json:"three"`
			Four          string `json:"four"`
			Five          string `json:"five"`
			Six           string `json:"six"`
			Seven         string `json:"seven"`
			Eight         string `json:"eight"`
			OpenTime      string `json:"open_time"`
			NextOpenTime  string `json:"next_open_time"`
			NextCode      string `json:"next_code"`
		} `json:"last"`
		Data struct {
			List []struct {
				Code  string `json:"code"`
				Day   string `json:"day"`
				One   string `json:"one"`
				Two   string `json:"two"`
				Three string `json:"three"`
				Four  string `json:"four"`
				Five  string `json:"five"`
				Six   string `json:"six"`
				Seven string `json:"seven"`
				Eight string `json:"eight"`
			} `json:"list"`
		} `json:"data"`
	} `json:"data"`
}

// CreateDrawResult 手动创建开奖结果
func (s *DrawService) CreateDrawResult(draw *models.DrawResult) error {
	// 先检查是否存在未删除的记录
	var existing models.DrawResult
	err := database.DB.Where("lottery_type = ? AND issue_number = ?", draw.LotteryType, draw.IssueNumber).First(&existing).Error
	if err == nil {
		// 记录已存在，不处理
		logger.GetSugarLogger().Warnf("期号 %s 已存在（未删除），跳过", draw.IssueNumber)
		return fmt.Errorf("该期号已存在")
	}
	if err != gorm.ErrRecordNotFound {
		logger.GetSugarLogger().Errorf("检查期号 %s 出错: %v", draw.IssueNumber, err)
		return err
	}

	// 检查是否存在软删除的记录
	var softDeleted models.DrawResult
	err = database.DB.Unscoped().
		Where("lottery_type = ? AND issue_number = ? AND deleted_at IS NOT NULL", draw.LotteryType, draw.IssueNumber).
		First(&softDeleted).Error

	if err == nil {
		logger.GetSugarLogger().Infof("期号 %s 存在软删除记录，恢复并更新", draw.IssueNumber)
		// 存在软删除的记录，使用原始记录恢复
		softDeleted.DrawDate = draw.DrawDate
		softDeleted.Numbers = draw.Numbers
		softDeleted.Source = draw.Source
		softDeleted.DeletedAt = gorm.DeletedAt{}
		return database.DB.Unscoped().Save(&softDeleted).Error
	}

	if err != gorm.ErrRecordNotFound {
		logger.GetSugarLogger().Errorf("检查软删除期号 %s 出错: %v", draw.IssueNumber, err)
		return err
	}

	// 全新记录，直接创建
	logger.GetSugarLogger().Infof("期号 %s 不存在，创建新记录", draw.IssueNumber)
	return database.DB.Create(draw).Error
}

// UpdateDrawResult 更新开奖结果
func (s *DrawService) UpdateDrawResult(draw *models.DrawResult) error {
	return database.DB.Save(draw).Error
}

// DeleteDrawResult 删除开奖结果
func (s *DrawService) DeleteDrawResult(id uint) error {
	return database.DB.Delete(&models.DrawResult{}, id).Error
}

// GetDB 获取数据库实例
func (s *DrawService) GetDB() *gorm.DB {
	return database.DB
}

// GetDrawResults 查询开奖结果
func (s *DrawService) GetDrawResults(lotteryType string, page, size int) ([]models.DrawResult, int64, error) {
	var draws []models.DrawResult
	var total int64
	query := database.DB.Model(&models.DrawResult{})
	if lotteryType != "" {
		query = query.Where("lottery_type = ?", lotteryType)
	}
	query.Count(&total)
	err := query.Order("draw_date DESC").Offset((page - 1) * size).Limit(size).Find(&draws).Error
	return draws, total, err
}

// GetDrawByIssue 通过期号查询开奖结果
func (s *DrawService) GetDrawByIssue(lotteryType, issue string) (*models.DrawResult, error) {
	var draw models.DrawResult
	// 使用Unscoped查询包括软删除的记录
	err := database.DB.Unscoped().Where("lottery_type = ? AND issue_number = ?", lotteryType, issue).First(&draw).Error
	if err != nil {
		return nil, err
	}
	return &draw, nil
}

// FetchDrawResult 从网络抓取开奖结果
func (s *DrawService) FetchDrawResult(lotteryType models.LotteryType, issue string) (*models.DrawResult, error) {
	switch lotteryType {
	case models.ShuangSeQiu:
		return s.fetchShuangSeQiu(issue)
	case models.DaLeTou:
		return s.fetchDaLeTou(issue)
	case models.QiLeCai:
		return s.fetchQiLeCai(issue)
	case models.PaiLie5:
		return s.fetchPaiLie5(issue)
	default:
		return nil, fmt.Errorf("暂不支持自动抓取 %s 开奖结果", lotteryType)
	}
}

type shuangSeQiuAPIResp struct {
	State   int    `json:"state"`
	Message string `json:"message"`
	Data    struct {
		List []struct {
			Code      string `json:"code"`
			Date      string `json:"date"`
			Red       string `json:"red"`
			Blue      string `json:"blue"`
		} `json:"list"`
	} `json:"data"`
}

func (s *DrawService) fetchShuangSeQiu(issue string) (*models.DrawResult, error) {
	url := "http://api.huiniao.top/interface/home/lotteryHistory?type=ssq&page=1&limit=1"

	client := &http.Client{Timeout: 10 * time.Second}
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
		return nil, fmt.Errorf("解析开奖数据失败: %v", err)
	}

	if apiResp.Code != 1 || len(apiResp.Data.Data.List) == 0 {
		return nil, fmt.Errorf("未找到开奖数据: %s", apiResp.Info)
	}

	item := apiResp.Data.Data.List[0]
	// 解析号码
	red := []int{parseInt(item.One), parseInt(item.Two), parseInt(item.Three), parseInt(item.Four), parseInt(item.Five), parseInt(item.Six)}
	blue := []int{parseInt(item.Seven)}
	numbersJSON, _ := json.Marshal(rules.ShuangSeQiuNumbers{
		Red:  red,
		Blue: blue,
	})

	drawDate, _ := time.Parse("2006-01-02", item.Day)
	draw := &models.DrawResult{
		LotteryType: models.ShuangSeQiu,
		IssueNumber: item.Code,
		DrawDate:    drawDate,
		Numbers:     string(numbersJSON),
		Source:      "auto",
	}
	return draw, nil
}

type daLeTouAPIResp struct {
	Value struct {
		List []struct {
			LotteryDrawNum     string `json:"lotteryDrawNum"`
			LotteryDrawTime    string `json:"lotteryDrawTime"`
			LotteryDrawResult  string `json:"lotteryDrawResult"`
		} `json:"list"`
	} `json:"value"`
}

func (s *DrawService) fetchDaLeTou(issue string) (*models.DrawResult, error) {
	url := "http://api.huiniao.top/interface/home/lotteryHistory?type=dlt&page=1&limit=1"

	client := &http.Client{Timeout: 10 * time.Second}
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
		return nil, fmt.Errorf("解析大乐透开奖数据失败: %v", err)
	}

	if apiResp.Code != 1 || len(apiResp.Data.Data.List) == 0 {
		return nil, fmt.Errorf("未找到开奖数据: %s", apiResp.Info)
	}

	item := apiResp.Data.Data.List[0]
	front := []int{parseInt(item.One), parseInt(item.Two), parseInt(item.Three), parseInt(item.Four), parseInt(item.Five)}
	back := []int{parseInt(item.Six), parseInt(item.Seven)}
	numbersJSON, _ := json.Marshal(rules.DaLeTouNumbers{Front: front, Back: back})
	drawDate, _ := time.Parse("2006-01-02", item.Day)

	return &models.DrawResult{
		LotteryType: models.DaLeTou,
		IssueNumber: item.Code,
		DrawDate:    drawDate,
		Numbers:     string(numbersJSON),
		Source:      "auto",
	}, nil
}

// parseNumbers 解析逗号分隔的数字字符串
func parseNumbers(s string) []int {
	var nums []int
	if s == "" {
		return nums
	}
	for i := 0; i < len(s); {
		j := i
		for j < len(s) && s[j] != ',' {
			j++
		}
		var n int
		fmt.Sscanf(s[i:j], "%d", &n)
		nums = append(nums, n)
		i = j + 1
	}
	return nums
}

// parseInt 解析字符串为整数
func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// parseSpaceNumbers 解析空格分隔的数字字符串
func parseSpaceNumbers(s string) []int {
	var nums []int
	var n int
	for _, part := range splitString(s, ' ') {
		if _, err := fmt.Sscanf(part, "%d", &n); err == nil {
			nums = append(nums, n)
		}
	}
	return nums
}

// FetchBatchDrawResults 批量抓取开奖结果
func (s *DrawService) FetchBatchDrawResults(lotteryType models.LotteryType, startDate, endDate string, count int) ([]*models.DrawResult, error) {
	switch lotteryType {
	case models.ShuangSeQiu:
		return s.fetchBatchShuangSeQiu(startDate, endDate, count)
	case models.DaLeTou:
		return s.fetchBatchDaLeTou(startDate, endDate, count)
	case models.QiLeCai:
		return s.fetchBatchQiLeCai(startDate, endDate, count)
	case models.PaiLie5:
		return s.fetchBatchPaiLie5(startDate, endDate, count)
	default:
		return nil, fmt.Errorf("暂不支持自动抓取 %s 开奖结果", lotteryType)
	}
}

// fetchBatchShuangSeQiu 批量抓取双色球
func (s *DrawService) fetchBatchShuangSeQiu(startDate, endDate string, count int) ([]*models.DrawResult, error) {
	url := fmt.Sprintf("http://api.huiniao.top/interface/home/lotteryHistory?type=ssq&page=1&limit=%d", count)

	client := &http.Client{Timeout: 30 * time.Second}
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
		return nil, fmt.Errorf("解析开奖数据失败: %v", err)
	}

	if apiResp.Code != 1 || len(apiResp.Data.Data.List) == 0 {
		return nil, fmt.Errorf("获取双色球开奖数据失败: %s", apiResp.Info)
	}

	var results []*models.DrawResult
	for _, item := range apiResp.Data.Data.List {
		red := []int{parseInt(item.One), parseInt(item.Two), parseInt(item.Three), parseInt(item.Four), parseInt(item.Five), parseInt(item.Six)}
		blue := []int{parseInt(item.Seven)}
		numbersJSON, _ := json.Marshal(rules.ShuangSeQiuNumbers{
			Red:  red,
			Blue: blue,
		})
		drawDate, _ := time.Parse("2006-01-02", item.Day)
		draw := &models.DrawResult{
			LotteryType: models.ShuangSeQiu,
			IssueNumber: item.Code,
			DrawDate:    drawDate,
			Numbers:     string(numbersJSON),
			Source:      "auto",
		}
		results = append(results, draw)
	}
	return results, nil
}

// fetchBatchDaLeTou 批量抓取大乐透
func (s *DrawService) fetchBatchDaLeTou(startDate, endDate string, count int) ([]*models.DrawResult, error) {
	url := fmt.Sprintf("http://api.huiniao.top/interface/home/lotteryHistory?type=dlt&page=1&limit=%d", count)

	client := &http.Client{Timeout: 30 * time.Second}
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
		return nil, fmt.Errorf("解析大乐透开奖数据失败: %v", err)
	}

	if apiResp.Code != 1 || len(apiResp.Data.Data.List) == 0 {
		return nil, fmt.Errorf("获取大乐透开奖数据失败: %s", apiResp.Info)
	}

	var results []*models.DrawResult
	for _, item := range apiResp.Data.Data.List {
		front := []int{parseInt(item.One), parseInt(item.Two), parseInt(item.Three), parseInt(item.Four), parseInt(item.Five)}
		back := []int{parseInt(item.Six), parseInt(item.Seven)}
		numbersJSON, _ := json.Marshal(rules.DaLeTouNumbers{Front: front, Back: back})
		drawDate, _ := time.Parse("2006-01-02", item.Day)
		draw := &models.DrawResult{
			LotteryType: models.DaLeTou,
			IssueNumber: item.Code,
			DrawDate:    drawDate,
			Numbers:     string(numbersJSON),
			Source:      "auto",
		}
		results = append(results, draw)
	}
	return results, nil
}

// fetchQiLeCai 抓取七乐彩开奖结果
func (s *DrawService) fetchQiLeCai(issue string) (*models.DrawResult, error) {
	url := "http://api.huiniao.top/interface/home/lotteryHistory?type=qlc&page=1&limit=1"

	client := &http.Client{Timeout: 10 * time.Second}
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
		return nil, fmt.Errorf("解析开奖数据失败: %v", err)
	}

	if apiResp.Code != 1 || len(apiResp.Data.Data.List) == 0 {
		return nil, fmt.Errorf("获取七乐彩开奖数据失败: %s", apiResp.Info)
	}

	item := apiResp.Data.Data.List[0]
	main := []int{parseInt(item.One), parseInt(item.Two), parseInt(item.Three), parseInt(item.Four), parseInt(item.Five), parseInt(item.Six), parseInt(item.Seven)}
	special := []int{parseInt(item.Eight)}
	numbersJSON, _ := json.Marshal(rules.QiLeCaiNumbers{
		Main:    main,
		Special: special,
	})
	drawDate, _ := time.Parse("2006-01-02", item.Day)

	return &models.DrawResult{
		LotteryType: models.QiLeCai,
		IssueNumber: item.Code,
		DrawDate:    drawDate,
		Numbers:     string(numbersJSON),
		Source:      "auto",
	}, nil
}

// fetchBatchQiLeCai 批量抓取七乐彩
func (s *DrawService) fetchBatchQiLeCai(startDate, endDate string, count int) ([]*models.DrawResult, error) {
	url := fmt.Sprintf("http://api.huiniao.top/interface/home/lotteryHistory?type=qlc&page=1&limit=%d", count)

	client := &http.Client{Timeout: 30 * time.Second}
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
		return nil, fmt.Errorf("解析开奖数据失败: %v", err)
	}

	if apiResp.Code != 1 || len(apiResp.Data.Data.List) == 0 {
		return nil, fmt.Errorf("获取七乐彩开奖数据失败: %s", apiResp.Info)
	}

	var results []*models.DrawResult
	for _, item := range apiResp.Data.Data.List {
		main := []int{parseInt(item.One), parseInt(item.Two), parseInt(item.Three), parseInt(item.Four), parseInt(item.Five), parseInt(item.Six), parseInt(item.Seven)}
		special := []int{parseInt(item.Eight)}
		numbersJSON, _ := json.Marshal(rules.QiLeCaiNumbers{
			Main:    main,
			Special: special,
		})
		drawDate, _ := time.Parse("2006-01-02", item.Day)
		draw := &models.DrawResult{
			LotteryType: models.QiLeCai,
			IssueNumber: item.Code,
			DrawDate:    drawDate,
			Numbers:     string(numbersJSON),
			Source:      "auto",
		}
		results = append(results, draw)
	}
	return results, nil
}

// fetchPaiLie5 抓取排列5开奖结果
func (s *DrawService) fetchPaiLie5(issue string) (*models.DrawResult, error) {
	url := "http://api.huiniao.top/interface/home/lotteryHistory?type=pl5&page=1&limit=1"

	client := &http.Client{Timeout: 10 * time.Second}
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
		return nil, fmt.Errorf("解析开奖数据失败: %v", err)
	}

	if apiResp.Code != 1 || len(apiResp.Data.Data.List) == 0 {
		return nil, fmt.Errorf("获取排列5开奖数据失败: %s", apiResp.Info)
	}

	item := apiResp.Data.Data.List[0]
	// 排列5是5位数字
	numbers := []int{parseInt(item.One), parseInt(item.Two), parseInt(item.Three), parseInt(item.Four), parseInt(item.Five)}
	numbersJSON, _ := json.Marshal(rules.PaiLie5Numbers{
		Numbers: numbers,
	})
	drawDate, _ := time.Parse("2006-01-02", item.Day)

	return &models.DrawResult{
		LotteryType: models.PaiLie5,
		IssueNumber: item.Code,
		DrawDate:    drawDate,
		Numbers:     string(numbersJSON),
		Source:      "auto",
	}, nil
}

// fetchBatchPaiLie5 批量抓取排列5
func (s *DrawService) fetchBatchPaiLie5(startDate, endDate string, count int) ([]*models.DrawResult, error) {
	url := fmt.Sprintf("http://api.huiniao.top/interface/home/lotteryHistory?type=pl5&page=1&limit=%d", count)

	client := &http.Client{Timeout: 30 * time.Second}
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
		return nil, fmt.Errorf("解析开奖数据失败: %v", err)
	}

	if apiResp.Code != 1 || len(apiResp.Data.Data.List) == 0 {
		return nil, fmt.Errorf("获取排列5开奖数据失败: %s", apiResp.Info)
	}

	var results []*models.DrawResult
	for _, item := range apiResp.Data.Data.List {
		numbers := []int{parseInt(item.One), parseInt(item.Two), parseInt(item.Three), parseInt(item.Four), parseInt(item.Five)}
		numbersJSON, _ := json.Marshal(rules.PaiLie5Numbers{
			Numbers: numbers,
		})
		drawDate, _ := time.Parse("2006-01-02", item.Day)
		draw := &models.DrawResult{
			LotteryType: models.PaiLie5,
			IssueNumber: item.Code,
			DrawDate:    drawDate,
			Numbers:     string(numbersJSON),
			Source:      "auto",
		}
		results = append(results, draw)
	}
	return results, nil
}

func splitString(s string, sep rune) []string {
	var parts []string
	start := 0
	for i, c := range s {
		if c == sep {
			if i > start {
				parts = append(parts, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	return parts
}

// WinningService 中奖计算服务
type WinningService struct{}

// CalculateWinning 计算中奖情况
func (s *WinningService) CalculateWinning(purchase *models.PurchaseRecord, draw *models.DrawResult) *models.WinningRecord {
	var level int
	var name string
	var amount float64

	switch purchase.LotteryType {
	case models.ShuangSeQiu:
		level, name, amount = rules.CalculateShuangSeQiu(purchase.Numbers, draw.Numbers)
	case models.DaLeTou:
		level, name, amount = rules.CalculateDaLeTou(purchase.Numbers, draw.Numbers)
	case models.FuCai3D:
		level, name, amount = rules.CalculateFuCai3D(purchase.Numbers, draw.Numbers)
	case models.PaiLie3:
		level, name, amount = rules.CalculatePaiLie3(purchase.Numbers, draw.Numbers)
	case models.PaiLie5:
		level, name, amount = rules.CalculatePaiLie5(purchase.Numbers, draw.Numbers)
	case models.QiLeCai:
		level, name, amount = rules.CalculateQiLeCai(purchase.Numbers, draw.Numbers)
	}

	return &models.WinningRecord{
		PurchaseID:  purchase.ID,
		DrawID:      draw.ID,
		LotteryType: purchase.LotteryType,
		IssueNumber: purchase.IssueNumber,
		PrizeLevel:  level,
		PrizeName:   name,
		PrizeAmount: amount,
	}
}

// CheckAndSaveWinnings 检查并保存中奖记录
func (s *WinningService) CheckAndSaveWinnings(drawID uint) error {
	var draw models.DrawResult
	if err := database.DB.First(&draw, drawID).Error; err != nil {
		logger.GetSugarLogger().Errorf("查找开奖记录失败: %v", err)
		return err
	}

	logger.GetSugarLogger().Infof("开始检查中奖: 彩票类型=%s, 期号=%s", draw.LotteryType, draw.IssueNumber)

	var purchases []models.PurchaseRecord
	database.DB.Where("lottery_type = ? AND issue_number = ?", draw.LotteryType, draw.IssueNumber).Find(&purchases)

	logger.GetSugarLogger().Infof("找到 %d 条匹配的购买记录", len(purchases))

	for _, purchase := range purchases {
		winning := s.CalculateWinning(&purchase, &draw)
		// 保存中奖记录（先硬删除旧的）
		database.DB.Unscoped().Where("purchase_id = ? AND draw_id = ?", purchase.ID, drawID).Delete(&models.WinningRecord{})

		// 只保存中奖的记录（PrizeLevel > 0）
		if winning.PrizeLevel > 0 {
			if err := database.DB.Create(winning).Error; err != nil {
				logger.GetSugarLogger().Errorf("保存中奖记录失败: %v", err)
			}
		}

		// 更新购买记录状态
		var status string
		if winning.PrizeLevel > 0 {
			status = "已中奖"
		} else {
			status = "未中奖"
		}

		result := database.DB.Model(&models.PurchaseRecord{}).Where("id = ?", purchase.ID).Update("status", status)
		logger.GetSugarLogger().Infof("更新购买记录状态: ID=%d, 期号=%s, 奖级=%d, 状态=%s, 影响行数=%d",
			purchase.ID, purchase.IssueNumber, winning.PrizeLevel, status, result.RowsAffected)
	}
	return nil
}

// PurchaseService 购买记录服务
type PurchaseService struct{}

func (s *PurchaseService) CreatePurchase(p *models.PurchaseRecord) error {
	return database.DB.Create(p).Error
}

func (s *PurchaseService) GetPurchases(userID uint, lotteryType, status string, page, size int) ([]models.PurchaseRecord, int64, error) {
	var purchases []models.PurchaseRecord
	var total int64
	query := database.DB.Model(&models.PurchaseRecord{}).Where("user_id = ?", userID)
	if lotteryType != "" {
		query = query.Where("lottery_type = ?", lotteryType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("purchase_date DESC").Offset((page - 1) * size).Limit(size).Find(&purchases).Error
	return purchases, total, err
}

func (s *PurchaseService) DeletePurchase(id uint) error {
	return database.DB.Delete(&models.PurchaseRecord{}, id).Error
}

func (s *PurchaseService) UpdatePurchase(id uint, p *models.PurchaseRecord) error {
	return database.DB.Model(&models.PurchaseRecord{}).Where("id = ?", id).Updates(p).Error
}
