package services

import (
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
