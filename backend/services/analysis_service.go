package services

import (
	"encoding/json"
	"fmt"
	"sort"

	"lottery-backend/logger"
	"lottery-backend/models"
)

// FreqItem 单个号码出现次数
type FreqItem struct {
	Num   int `json:"num"`
	Count int `json:"count"`
}

// OmissionItem 单个号码遗漏值（当前遗漏 + 历史最大遗漏）
type OmissionItem struct {
	Num     int `json:"num"`
	Current int `json:"current"`
	Max     int `json:"max"`
}

// ZoneResult 一个号码区的分析结果
type ZoneResult struct {
	Name      string         `json:"name"`
	Min       int            `json:"min"`
	Max       int            `json:"max"`
	Frequency []FreqItem     `json:"frequency"`
	Omission  []OmissionItem `json:"omission"`
	Trend     [][]int        `json:"trend"`
}

// MetricItem 单期主区指标
type MetricItem struct {
	Issue    string `json:"issue"`
	Sum      int    `json:"sum"`
	Span     int    `json:"span"`
	OddEven  string `json:"oddEven"`
	BigSmall string `json:"bigSmall"`
}

// AnalysisResult 开奖号码分析总结果
type AnalysisResult struct {
	LotteryType string       `json:"lottery_type"`
	IssueCount  int          `json:"issue_count"`
	Issues      []string     `json:"issues"`
	Zones       []ZoneResult `json:"zones"`
	Metrics     []MetricItem `json:"metrics"`
}

// rawNumbers 兼容所有彩种的号码 JSON 字段
type rawNumbers struct {
	Red     []int `json:"red"`
	Blue    []int `json:"blue"`
	Front   []int `json:"front"`
	Back    []int `json:"back"`
	Main    []int `json:"main"`
	Special []int `json:"special"`
	Numbers []int `json:"numbers"`
}

func (r rawNumbers) field(name string) []int {
	switch name {
	case "red":
		return r.Red
	case "blue":
		return r.Blue
	case "front":
		return r.Front
	case "back":
		return r.Back
	case "main":
		return r.Main
	case "special":
		return r.Special
	case "numbers":
		return r.Numbers
	}
	return nil
}

// zoneDef 号码区定义：index<0 取整段；index>=0 取该字段数组的指定位
type zoneDef struct {
	name  string
	min   int
	max   int
	field string
	index int
}

func (z zoneDef) values(r rawNumbers) []int {
	f := r.field(z.field)
	if z.index < 0 {
		return f
	}
	if z.index >= 0 && z.index < len(f) {
		return []int{f[z.index]}
	}
	return nil
}

// lotteryConfig 彩种分析配置
type lotteryConfig struct {
	zones             []zoneDef
	mainField         string // 用于和值/跨度/奇偶比的主区字段
	bigSmall          bool   // 是否计算大小比
	bigSmallThreshold int    // 号码 > 阈值 计为「大」
}

func digitZones(field string, n int) []zoneDef {
	names := []string{"第1位", "第2位", "第3位", "第4位", "第5位", "第6位", "第7位"}
	zones := make([]zoneDef, 0, n)
	for i := 0; i < n; i++ {
		zones = append(zones, zoneDef{name: names[i], min: 0, max: 9, field: field, index: i})
	}
	return zones
}

var lotteryConfigs = map[models.LotteryType]lotteryConfig{
	models.ShuangSeQiu: {
		zones: []zoneDef{
			{name: "红球", min: 1, max: 33, field: "red", index: -1},
			{name: "蓝球", min: 1, max: 16, field: "blue", index: -1},
		},
		mainField: "red", bigSmall: true, bigSmallThreshold: 16,
	},
	models.DaLeTou: {
		zones: []zoneDef{
			{name: "前区", min: 1, max: 35, field: "front", index: -1},
			{name: "后区", min: 1, max: 12, field: "back", index: -1},
		},
		mainField: "front", bigSmall: true, bigSmallThreshold: 17,
	},
	models.QiLeCai: {
		zones: []zoneDef{
			{name: "主号", min: 1, max: 30, field: "main", index: -1},
			{name: "特别号", min: 1, max: 30, field: "special", index: -1},
		},
		mainField: "main", bigSmall: true, bigSmallThreshold: 15,
	},
	models.QiXingCai: {
		zones: append(
			digitZones("red", 6),
			zoneDef{name: "第7位", min: 0, max: 9, field: "blue", index: 0},
		),
		mainField: "red", bigSmall: false,
	},
	models.FuCai3D: {
		zones:     []zoneDef{{name: "百位", min: 0, max: 9, field: "numbers", index: 0}, {name: "十位", min: 0, max: 9, field: "numbers", index: 1}, {name: "个位", min: 0, max: 9, field: "numbers", index: 2}},
		mainField: "numbers", bigSmall: false,
	},
	models.PaiLie3: {
		zones:     []zoneDef{{name: "百位", min: 0, max: 9, field: "numbers", index: 0}, {name: "十位", min: 0, max: 9, field: "numbers", index: 1}, {name: "个位", min: 0, max: 9, field: "numbers", index: 2}},
		mainField: "numbers", bigSmall: false,
	},
	models.PaiLie5: {
		zones:     digitZones("numbers", 5),
		mainField: "numbers", bigSmall: false,
	},
}

// computeFrequency 统计 [min,max] 内每个号码在 perIssue 中出现的次数
func computeFrequency(min, max int, perIssue [][]int) []FreqItem {
	counts := make(map[int]int)
	for _, vals := range perIssue {
		for _, v := range vals {
			counts[v]++
		}
	}
	out := make([]FreqItem, 0, max-min+1)
	for n := min; n <= max; n++ {
		out = append(out, FreqItem{Num: n, Count: counts[n]})
	}
	return out
}

// computeOmission 计算每个号码的当前遗漏与历史最大遗漏。perIssue 须时间升序。
func computeOmission(min, max int, perIssue [][]int) []OmissionItem {
	total := len(perIssue)
	out := make([]OmissionItem, 0, max-min+1)
	for n := min; n <= max; n++ {
		lastSeen := -1
		maxGap := 0
		gap := 0
		for i := 0; i < total; i++ {
			present := false
			for _, v := range perIssue[i] {
				if v == n {
					present = true
					break
				}
			}
			if present {
				if gap > maxGap {
					maxGap = gap
				}
				gap = 0
				lastSeen = i
			} else {
				gap++
			}
		}
		if gap > maxGap { // 末尾连续遗漏
			maxGap = gap
		}
		current := total
		if lastSeen >= 0 {
			current = total - 1 - lastSeen
		}
		out = append(out, OmissionItem{Num: n, Current: current, Max: maxGap})
	}
	return out
}

// computeMetrics 按期计算主区的和值、跨度、奇偶比、大小比
func computeMetrics(issues []string, mainPerIssue [][]int, bigSmall bool, threshold int) []MetricItem {
	out := make([]MetricItem, 0, len(issues))
	for i, nums := range mainPerIssue {
		if i >= len(issues) {
			break
		}
		sum, odd, even, big, small := 0, 0, 0, 0, 0
		mn, mx := 0, 0
		for j, v := range nums {
			sum += v
			if v%2 == 0 {
				even++
			} else {
				odd++
			}
			if v > threshold {
				big++
			} else {
				small++
			}
			if j == 0 || v < mn {
				mn = v
			}
			if j == 0 || v > mx {
				mx = v
			}
		}
		item := MetricItem{
			Issue:   issues[i],
			Sum:     sum,
			Span:    mx - mn,
			OddEven: fmt.Sprintf("%d:%d", odd, even),
		}
		if bigSmall {
			item.BigSmall = fmt.Sprintf("%d:%d", big, small)
		}
		out = append(out, item)
	}
	return out
}

// AnalyzeDraws 对给定彩种的开奖记录做号码分析。draws 顺序不限，内部按期号升序处理。
func AnalyzeDraws(lt models.LotteryType, draws []models.DrawResult) (*AnalysisResult, error) {
	cfg, ok := lotteryConfigs[lt]
	if !ok {
		return nil, fmt.Errorf("不支持的彩种: %s", lt)
	}

	sorted := make([]models.DrawResult, len(draws))
	copy(sorted, draws)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].IssueNumber < sorted[j].IssueNumber
	})

	issues := make([]string, 0, len(sorted))
	parsed := make([]rawNumbers, 0, len(sorted))
	for _, d := range sorted {
		var r rawNumbers
		if err := json.Unmarshal([]byte(d.Numbers), &r); err != nil {
			logger.GetSugarLogger().Warnf("号码分析跳过无法解析的开奖记录 id=%d issue=%s: %v", d.ID, d.IssueNumber, err)
			continue
		}
		issues = append(issues, d.IssueNumber)
		parsed = append(parsed, r)
	}

	result := &AnalysisResult{
		LotteryType: string(lt),
		IssueCount:  len(issues),
		Issues:      issues,
		Zones:       make([]ZoneResult, 0, len(cfg.zones)),
		Metrics:     []MetricItem{},
	}

	for _, z := range cfg.zones {
		perIssue := make([][]int, len(parsed))
		for i, r := range parsed {
			perIssue[i] = z.values(r)
		}
		result.Zones = append(result.Zones, ZoneResult{
			Name:      z.name,
			Min:       z.min,
			Max:       z.max,
			Frequency: computeFrequency(z.min, z.max, perIssue),
			Omission:  computeOmission(z.min, z.max, perIssue),
			Trend:     perIssue,
		})
	}

	mainPerIssue := make([][]int, len(parsed))
	for i, r := range parsed {
		mainPerIssue[i] = r.field(cfg.mainField)
	}
	result.Metrics = computeMetrics(issues, mainPerIssue, cfg.bigSmall, cfg.bigSmallThreshold)

	return result, nil
}

// AnalysisService 号码分析服务
type AnalysisService struct{}

// GetAnalysis 取该彩种最近 count 期（count<=0 表示全部），返回分析结果
func (s *AnalysisService) GetAnalysis(lotteryType string, count int) (*AnalysisResult, error) {
	lt := models.LotteryType(lotteryType)
	if _, ok := lotteryConfigs[lt]; !ok {
		return nil, fmt.Errorf("不支持的彩种: %s", lotteryType)
	}
	q := GetDB().Model(&models.DrawResult{}).
		Where("lottery_type = ?", lotteryType).
		Order("issue_number DESC")
	if count > 0 {
		q = q.Limit(count)
	}
	var draws []models.DrawResult
	if err := q.Find(&draws).Error; err != nil {
		return nil, err
	}
	return AnalyzeDraws(lt, draws)
}
