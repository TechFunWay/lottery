package rules

import (
	"encoding/json"
	"sort"
)

// NumberSet 双色球号码结构
type ShuangSeQiuNumbers struct {
	Red  []int `json:"red"`  // 6个红球 01-33
	Blue []int `json:"blue"` // 1个蓝球 01-16
}

// ShuangSeQiuResult 双色球中奖结果
type ShuangSeQiuResult struct {
	Level  int
	Name   string
	Amount float64
}

var shuangSeQiuPrizes = map[string]ShuangSeQiuResult{
	"6+1": {1, "一等奖", 5000000},
	"6+0": {2, "二等奖", 200000},
	"5+1": {3, "三等奖", 3000},
	"5+0": {4, "四等奖", 200},
	"4+1": {4, "四等奖", 200},
	"4+0": {5, "五等奖", 10},
	"3+1": {5, "五等奖", 10},
	"2+1": {6, "六等奖", 5},
	"1+1": {6, "六等奖", 5},
	"0+1": {6, "六等奖", 5},
}

// CalculateShuangSeQiu 计算双色球中奖
func CalculateShuangSeQiu(purchaseJSON, drawJSON string) (level int, name string, amount float64) {
	var purchase, draw ShuangSeQiuNumbers
	if err := json.Unmarshal([]byte(purchaseJSON), &purchase); err != nil {
		return 0, "未中奖", 0
	}
	if err := json.Unmarshal([]byte(drawJSON), &draw); err != nil {
		return 0, "未中奖", 0
	}

	redMatch := countMatch(purchase.Red, draw.Red)
	blueMatch := 0
	if len(purchase.Blue) > 0 && len(draw.Blue) > 0 && purchase.Blue[0] == draw.Blue[0] {
		blueMatch = 1
	}

	key := formatKey(redMatch, blueMatch)
	if result, ok := shuangSeQiuPrizes[key]; ok {
		return result.Level, result.Name, result.Amount
	}
	return 0, "未中奖", 0
}

func countMatch(a, b []int) int {
	set := make(map[int]bool)
	for _, v := range b {
		set[v] = true
	}
	count := 0
	for _, v := range a {
		if set[v] {
			count++
		}
	}
	return count
}

func formatKey(red, blue int) string {
	return string(rune('0'+red)) + "+" + string(rune('0'+blue))
}

// DaLeTouNumbers 大乐透号码
type DaLeTouNumbers struct {
	Front []int `json:"front"` // 5个前区 01-35
	Back  []int `json:"back"`  // 2个后区 01-12
}

var daLeTouPrizes = []struct {
	front, back int
	level       int
	name        string
	amount      float64
}{
	{5, 2, 1, "一等奖", 10000000},
	{5, 1, 2, "二等奖", 500000},
	{5, 0, 3, "三等奖", 10000},
	{4, 2, 4, "四等奖", 3000},
	{4, 1, 5, "五等奖", 300},
	{3, 2, 5, "五等奖", 300},
	{4, 0, 6, "六等奖", 100},
	{3, 1, 6, "六等奖", 100},
	{2, 2, 6, "六等奖", 100},
	{3, 0, 7, "七等奖", 15},
	{2, 1, 7, "七等奖", 15},
	{1, 2, 7, "七等奖", 15},
	{0, 2, 7, "七等奖", 15},
}

func CalculateDaLeTou(purchaseJSON, drawJSON string) (level int, name string, amount float64) {
	var purchase, draw DaLeTouNumbers
	if err := json.Unmarshal([]byte(purchaseJSON), &purchase); err != nil {
		return 0, "未中奖", 0
	}
	if err := json.Unmarshal([]byte(drawJSON), &draw); err != nil {
		return 0, "未中奖", 0
	}

	frontMatch := countMatch(purchase.Front, draw.Front)
	backMatch := countMatch(purchase.Back, draw.Back)

	for _, p := range daLeTouPrizes {
		if frontMatch == p.front && backMatch == p.back {
			return p.level, p.name, p.amount
		}
	}
	return 0, "未中奖", 0
}

// FuCai3DNumbers 福彩3D号码
type FuCai3DNumbers struct {
	Numbers []int  `json:"numbers"` // 3个数字 0-9
	BetType string `json:"bet_type"` // 直选/组选6/组选3
}

func CalculateFuCai3D(purchaseJSON, drawJSON string) (level int, name string, amount float64) {
	var purchase FuCai3DNumbers
	var drawNums []int
	if err := json.Unmarshal([]byte(purchaseJSON), &purchase); err != nil {
		return 0, "未中奖", 0
	}
	if err := json.Unmarshal([]byte(drawJSON), &drawNums); err != nil {
		return 0, "未中奖", 0
	}

	if len(purchase.Numbers) != 3 || len(drawNums) != 3 {
		return 0, "未中奖", 0
	}

	switch purchase.BetType {
	case "直选":
		if purchase.Numbers[0] == drawNums[0] && purchase.Numbers[1] == drawNums[1] && purchase.Numbers[2] == drawNums[2] {
			return 1, "直选奖", 1040}
	case "组选6":
		p := make([]int, 3)
		d := make([]int, 3)
		copy(p, purchase.Numbers)
		copy(d, drawNums)
		sort.Ints(p)
		sort.Ints(d)
		if p[0] == d[0] && p[1] == d[1] && p[2] == d[2] {
			return 2, "组选6奖", 173}
	case "组选3":
		p := make([]int, 3)
		d := make([]int, 3)
		copy(p, purchase.Numbers)
		copy(d, drawNums)
		sort.Ints(p)
		sort.Ints(d)
		if p[0] == d[0] && p[1] == d[1] && p[2] == d[2] {
			return 2, "组选3奖", 346}
	}
	return 0, "未中奖", 0
}

// PaiLieNumbers 排列号码
type PaiLieNumbers struct {
	Numbers []int  `json:"numbers"`
	BetType string `json:"bet_type"` // 直选/组选
}

func CalculatePaiLie3(purchaseJSON, drawJSON string) (level int, name string, amount float64) {
	var purchase PaiLieNumbers
	var drawNums []int
	if err := json.Unmarshal([]byte(purchaseJSON), &purchase); err != nil {
		return 0, "未中奖", 0
	}
	if err := json.Unmarshal([]byte(drawJSON), &drawNums); err != nil {
		return 0, "未中奖", 0
	}
	if len(purchase.Numbers) != 3 || len(drawNums) != 3 {
		return 0, "未中奖", 0
	}
	if purchase.BetType == "直选" {
		if purchase.Numbers[0] == drawNums[0] && purchase.Numbers[1] == drawNums[1] && purchase.Numbers[2] == drawNums[2] {
			return 1, "直选奖", 1000
		}
	} else {
		p := make([]int, 3)
		d := make([]int, 3)
		copy(p, purchase.Numbers)
		copy(d, drawNums)
		sort.Ints(p)
		sort.Ints(d)
		if p[0] == d[0] && p[1] == d[1] && p[2] == d[2] {
			return 2, "组选奖", 167
		}
	}
	return 0, "未中奖", 0
}

func CalculatePaiLie5(purchaseJSON, drawJSON string) (level int, name string, amount float64) {
	var purchase PaiLieNumbers
	var drawNums []int
	if err := json.Unmarshal([]byte(purchaseJSON), &purchase); err != nil {
		return 0, "未中奖", 0
	}
	if err := json.Unmarshal([]byte(drawJSON), &drawNums); err != nil {
		return 0, "未中奖", 0
	}
	if len(purchase.Numbers) != 5 || len(drawNums) != 5 {
		return 0, "未中奖", 0
	}
	match := true
	for i := range purchase.Numbers {
		if purchase.Numbers[i] != drawNums[i] {
			match = false
			break
		}
	}
	if match {
		return 1, "直选奖", 100000
	}
	return 0, "未中奖", 0
}

// QiLeCaiNumbers 七乐彩号码
type QiLeCaiNumbers struct {
	Main    []int `json:"main"`    // 7个主号 01-30
	Special []int `json:"special"` // 1个特别号
}

// PaiLie5Numbers 排列5号码
type PaiLie5Numbers struct {
	Numbers []int `json:"numbers"` // 5个数字 0-9
}

var qiLeCaiPrizes = []struct {
	main, special int
	level         int
	name          string
	amount        float64
}{
	{7, 0, 1, "一等奖", 5000000},
	{6, 1, 2, "二等奖", 10000},
	{6, 0, 3, "三等奖", 1000},
	{5, 1, 4, "四等奖", 100},
	{5, 0, 4, "四等奖", 100},
	{4, 1, 5, "五等奖", 30},
	{4, 0, 5, "五等奖", 30},
	{3, 1, 6, "六等奖", 10},
	{7, 0, 7, "七等奖", 5}, // 仅特别号
}

func CalculateQiLeCai(purchaseJSON, drawJSON string) (level int, name string, amount float64) {
	var purchase, draw QiLeCaiNumbers
	if err := json.Unmarshal([]byte(purchaseJSON), &purchase); err != nil {
		return 0, "未中奖", 0
	}
	if err := json.Unmarshal([]byte(drawJSON), &draw); err != nil {
		return 0, "未中奖", 0
	}

	mainMatch := countMatch(purchase.Main, draw.Main)
	specialMatch := 0
	if len(purchase.Special) > 0 && len(draw.Special) > 0 && purchase.Special[0] == draw.Special[0] {
		specialMatch = 1
	}

	for _, p := range qiLeCaiPrizes {
		if mainMatch == p.main && specialMatch == p.special {
			return p.level, p.name, p.amount
		}
	}
	return 0, "未中奖", 0
}
