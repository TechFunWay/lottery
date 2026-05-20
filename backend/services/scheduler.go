package services

import (
	"fmt"
	"lottery-backend/logger"
	"lottery-backend/models"
	"strings"
	"sync"
	"time"
)

type fetchTime struct {
	Hour   int
	Minute int
	Label  string
}

var defaultFetchTimes = []fetchTime{
	{6, 0, "早上"},
	{12, 0, "中午"},
	{21, 30, "晚上(开奖后)"},
}

// SchedulerService 定时任务服务
type SchedulerService struct {
	drawService    *DrawService
	winningService *WinningService
	ticker         *time.Ticker
	stopChan       chan struct{}
	running        bool
	mu             sync.Mutex
	fetchMu        sync.Mutex
	lastFetchDate  map[string]string
}

// NewSchedulerService 创建定时任务服务
func NewSchedulerService() *SchedulerService {
	return &SchedulerService{
		drawService:    &DrawService{},
		winningService: &WinningService{},
		stopChan:       make(chan struct{}),
		lastFetchDate:  make(map[string]string),
	}
}

// Start 启动定时任务（立即执行一次，后续按配置时段执行）
func (s *SchedulerService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		logger.GetSugarLogger().Warn("定时任务服务已在运行中")
		return
	}
	s.running = true
	s.mu.Unlock()

	logger.GetSugarLogger().Info("🕐 定时任务服务已启动，立即执行首次抓取")

	// 立即执行一次
	go s.runFetch()

	// 每分钟检查是否需要执行定时抓取
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		s.mu.Lock()
		s.ticker = ticker
		s.mu.Unlock()
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkScheduledTimes()
			case <-s.stopChan:
				logger.GetSugarLogger().Info("🛑 定时任务服务已停止")
				return
			}
		}
	}()
}

func (s *SchedulerService) checkScheduledTimes() {
	now := time.Now()
	today := now.Format("2006-01-02")

	for _, ft := range defaultFetchTimes {
		if now.Hour() == ft.Hour && now.Minute() == ft.Minute {
			key := fmt.Sprintf("%02d:%02d", ft.Hour, ft.Minute)
			if s.lastFetchDate[key] != today {
				s.lastFetchDate[key] = today
				logger.GetSugarLogger().Infof("⏰ 到达%s定时抓取时间 (%s)，开始抓取", ft.Label, key)
				go s.runFetch()
			}
			return
		}
	}
}

// Stop 停止定时任务
func (s *SchedulerService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopChan)
	if s.ticker != nil {
		s.ticker.Stop()
	}
}

// TriggerNow 手动触发一次抓取任务
func (s *SchedulerService) TriggerNow() {
	logger.GetSugarLogger().Info("👆 手动触发开奖号码抓取任务")
	go s.runFetch()
}

// IsRunning 返回定时任务是否正在运行
func (s *SchedulerService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// runFetch 执行抓取任务
func (s *SchedulerService) runFetch() {
	if !s.fetchMu.TryLock() {
		logger.GetSugarLogger().Warn("抓取任务正在执行中，跳过本次触发")
		return
	}
	defer s.fetchMu.Unlock()

	logger.GetSugarLogger().Info("🚀 开始执行抓取开奖号码任务")
	results := s.FetchLatestDraws()

	totalSaved := 0
	totalExist := 0
	for lotteryType, result := range results {
		if result.Error != nil {
			logger.GetSugarLogger().Errorf("❌ 抓取 %s 失败: %v", lotteryType, result.Error)
		} else {
			logger.GetSugarLogger().Infof("✅ %s 抓取完成，新增 %d 条，已存在 %d 条", lotteryType, result.SavedCount, result.ExistCount)
			totalSaved += result.SavedCount
			totalExist += result.ExistCount
		}
	}

	logger.GetSugarLogger().Infof("📊 抓取任务完成，总计新增 %d 条，已存在 %d 条", totalSaved, totalExist)
}

// FetchResult 单种彩票的抓取结果
type FetchResult struct {
	SavedCount int
	ExistCount int
	Error      error
}

// FetchLatestDraws 抓取所有支持彩票类型的最新开奖结果
func (s *SchedulerService) FetchLatestDraws() map[string]FetchResult {
	lotteryTypes := []models.LotteryType{
		models.ShuangSeQiu,
		models.DaLeTou,
		models.QiLeCai,
		models.PaiLie5,
	}

	results := make(map[string]FetchResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, lt := range lotteryTypes {
		wg.Add(1)
		go func(lotteryType models.LotteryType) {
			defer wg.Done()

			result := s.fetchSingleLottery(lotteryType)

			mu.Lock()
			results[string(lotteryType)] = result
			mu.Unlock()
		}(lt)
	}

	wg.Wait()
	return results
}

// fetchSingleLottery 抓取单种彩票的最新开奖结果
func (s *SchedulerService) fetchSingleLottery(lotteryType models.LotteryType) FetchResult {
	result := FetchResult{}

	draw, err := s.drawService.FetchDrawResult(lotteryType, "")
	if err != nil {
		result.Error = err
		return result
	}

	// 检查是否已存在
	existing, _ := s.drawService.GetDrawByIssue(string(lotteryType), draw.IssueNumber)
	if existing != nil {
		result.ExistCount = 1
		return result
	}

	if err := s.drawService.CreateDrawResult(draw); err != nil {
		if strings.Contains(err.Error(), "已存在") {
			result.ExistCount = 1
		} else {
			result.Error = err
		}
		return result
	}

	result.SavedCount = 1

	// 异步计算中奖
	go s.winningService.CheckAndSaveWinnings(draw.ID)

	return result
}
