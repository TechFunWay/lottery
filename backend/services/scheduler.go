package services

import (
	"lottery-backend/logger"
	"lottery-backend/models"
	"strings"
	"sync"
	"time"
)

// SchedulerService 定时任务服务
type SchedulerService struct {
	drawService    *DrawService
	winningService *WinningService
	ticker         *time.Ticker
	stopChan       chan struct{}
	running        bool
	mu             sync.Mutex
}

// NewSchedulerService 创建定时任务服务
func NewSchedulerService() *SchedulerService {
	return &SchedulerService{
		drawService:    &DrawService{},
		winningService: &WinningService{},
		stopChan:       make(chan struct{}),
	}
}

// Start 启动定时任务（每天凌晨2点执行）
func (s *SchedulerService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		logger.GetSugarLogger().Warn("定时任务服务已在运行中")
		return
	}

	s.running = true

	// 计算下一个凌晨2点的时间
	nextRun := s.nextRunTime()
	initialDelay := time.Until(nextRun)

	logger.GetSugarLogger().Infof("🕐 定时任务服务已启动，下次执行时间: %s", nextRun.Format("2006-01-02 15:04:05"))

	// 使用单次定时器等待到第一个执行点，然后启动24小时周期的ticker
	go func() {
		select {
		case <-time.After(initialDelay):
			s.runFetch()
		case <-s.stopChan:
			return
		}

		// 之后每24小时执行一次
		s.ticker = time.NewTicker(24 * time.Hour)
		defer s.ticker.Stop()

		for {
			select {
			case <-s.ticker.C:
				s.runFetch()
			case <-s.stopChan:
				logger.GetSugarLogger().Info("🛑 定时任务服务已停止")
				return
			}
		}
	}()
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

// nextRunTime 计算下一个凌晨2点的时间
func (s *SchedulerService) nextRunTime() time.Time {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if !now.Before(next) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

// runFetch 执行抓取任务
func (s *SchedulerService) runFetch() {
	logger.GetSugarLogger().Info("🚀 开始执行定时抓取开奖号码任务")
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

	logger.GetSugarLogger().Infof("📊 定时抓取任务完成，总计新增 %d 条，已存在 %d 条", totalSaved, totalExist)
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
