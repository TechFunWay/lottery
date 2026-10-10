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

// combinations 生成从 n 个元素中选 k 个的所有组合
func combinations(arr []int, k int) [][]int {
	var result [][]int
	var current []int
	var backtrack func(start int)
	backtrack = func(start int) {
		if len(current) == k {
			tmp := make([]int, k)
			copy(tmp, current)
			result = append(result, tmp)
			return
		}
		for i := start; i < len(arr); i++ {
			current = append(current, arr[i])
			backtrack(i + 1)
			current = current[:len(current)-1]
		}
	}
	backtrack(0)
	return result
}

// CalculateShuangSeQiu 计算双色球中奖（支持复式）
func CalculateShuangSeQiu(purchaseJSON, drawJSON string, multiple int) (level int, name string, amount float64) {
	var purchase, draw ShuangSeQiuNumbers
	if err := json.Unmarshal([]byte(purchaseJSON), &purchase); err != nil {
		return 0, "未中奖", 0
	}
	if err := json.Unmarshal([]byte(drawJSON), &draw); err != nil {
		return 0, "未中奖", 0
	}

	// 单式投注：红球6个，蓝球1个
	isMultiple := len(purchase.Red) > 6 || len(purchase.Blue) > 1

	if !isMultiple {
		// 单式直接计算
		redMatch := countMatch(purchase.Red, draw.Red)
		blueMatch := 0
		if len(purchase.Blue) > 0 && len(draw.Blue) > 0 && purchase.Blue[0] == draw.Blue[0] {
			blueMatch = 1
		}
		key := formatKey(redMatch, blueMatch)
		if result, ok := shuangSeQiuPrizes[key]; ok {
			return result.Level, result.Name, result.Amount * float64(multiple)
		}
		return 0, "未中奖", 0
	}

	// 复式投注：计算所有组合，取最高奖级和总奖金
	redCombs := [][]int{purchase.Red}
	blueCombs := [][]int{purchase.Blue}

	if len(purchase.Red) > 6 {
		redCombs = combinations(purchase.Red, 6)
	}
	if len(purchase.Blue) > 1 {
		blueCombs = make([][]int, len(purchase.Blue))
		for i, b := range purchase.Blue {
			blueCombs[i] = []int{b}
		}
	}

	bestLevel := 0
	bestName := "未中奖"
	totalAmount := 0.0

	for _, reds := range redCombs {
		for _, blues := range blueCombs {
			redMatch := countMatch(reds, draw.Red)
			blueMatch := 0
			if len(blues) > 0 && len(draw.Blue) > 0 && blues[0] == draw.Blue[0] {
				blueMatch = 1
			}
			key := formatKey(redMatch, blueMatch)
			if result, ok := shuangSeQiuPrizes[key]; ok {
				totalAmount += result.Amount * float64(multiple)
				if bestLevel == 0 || result.Level < bestLevel {
					bestLevel = result.Level
					bestName = result.Name
				}
			}
		}
	}

	if bestLevel > 0 {
		return bestLevel, bestName, totalAmount
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
	{3, 2, 6, "六等奖", 200},
	{4, 0, 7, "七等奖", 100},
	{3, 1, 8, "八等奖", 15},
	{2, 2, 8, "八等奖", 15},
	{3, 0, 9, "九等奖", 5},
	{2, 1, 9, "九等奖", 5},
	{1, 2, 9, "九等奖", 5},
	{0, 2, 9, "九等奖", 5},
}

// CalculateDaLeTou 计算大乐透中奖（支持复式、追加）
func CalculateDaLeTou(purchaseJSON, drawJSON string, multiple int, append bool) (level int, name string, amount float64) {
	var purchase, draw DaLeTouNumbers
	if err := json.Unmarshal([]byte(purchaseJSON), &purchase); err != nil {
		return 0, "未中奖", 0
	}
	if err := json.Unmarshal([]byte(drawJSON), &draw); err != nil {
		return 0, "未中奖", 0
	}

	// 单式投注：前区5个，后区2个
	isMultiple := len(purchase.Front) > 5 || len(purchase.Back) > 2

	if !isMultiple {
		// 单式直接计算
		frontMatch := countMatch(purchase.Front, draw.Front)
		backMatch := countMatch(purchase.Back, draw.Back)

		for _, p := range daLeTouPrizes {
			if frontMatch == p.front && backMatch == p.back {
				prize := p.amount * float64(multiple)
				// 追加仅一至三等奖有效，奖金 × 1.6
				if append && p.level <= 3 {
					prize *= 1.6
				}
				return p.level, p.name, prize
			}
		}
		return 0, "未中奖", 0
	}

	// 复式投注：计算所有组合，取最高奖级和总奖金
	frontCombs := [][]int{purchase.Front}
	backCombs := [][]int{purchase.Back}

	if len(purchase.Front) > 5 {
		frontCombs = combinations(purchase.Front, 5)
	}
	if len(purchase.Back) > 2 {
		backCombs = combinations(purchase.Back, 2)
	}

	bestLevel := 0
	bestName := "未中奖"
	totalAmount := 0.0

	for _, fronts := range frontCombs {
		for _, backs := range backCombs {
			frontMatch := countMatch(fronts, draw.Front)
			backMatch := countMatch(backs, draw.Back)

			for _, p := range daLeTouPrizes {
				if frontMatch == p.front && backMatch == p.back {
					prize := p.amount * float64(multiple)
					// 追加仅一至三等奖有效，奖金 × 1.6
					if append && p.level <= 3 {
						prize *= 1.6
					}
					totalAmount += prize
					if bestLevel == 0 || p.level < bestLevel {
						bestLevel = p.level
						bestName = p.name
					}
					break
				}
			}
		}
	}

	if bestLevel > 0 {
		return bestLevel, bestName, totalAmount
	}
	return 0, "未中奖", 0
}

// FuCai3DNumbers 福彩3D号码
type FuCai3DNumbers struct {
	Numbers []int  `json:"numbers"` // 3个数字 0-9
	BetType string `json:"bet_type"` // 直选/组选6/组选3
}

func CalculateFuCai3D(purchaseJSON, drawJSON string, multiple int) (level int, name string, amount float64) {
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
			return 1, "直选奖", 1040 * float64(multiple)}
	case "组选6":
		p := make([]int, 3)
		d := make([]int, 3)
		copy(p, purchase.Numbers)
		copy(d, drawNums)
		sort.Ints(p)
		sort.Ints(d)
		if p[0] == d[0] && p[1] == d[1] && p[2] == d[2] {
			return 2, "组选6奖", 173 * float64(multiple)}
	case "组选3":
		p := make([]int, 3)
		d := make([]int, 3)
		copy(p, purchase.Numbers)
		copy(d, drawNums)
		sort.Ints(p)
		sort.Ints(d)
		if p[0] == d[0] && p[1] == d[1] && p[2] == d[2] {
			return 2, "组选3奖", 346 * float64(multiple)}
	}
	return 0, "未中奖", 0
}

// PaiLieNumbers 排列号码
type PaiLieNumbers struct {
	Numbers []int  `json:"numbers"`
	BetType string `json:"bet_type"` // 直选/组选
}

func CalculatePaiLie3(purchaseJSON, drawJSON string, multiple int) (level int, name string, amount float64) {
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
			return 1, "直选奖", 1040 * float64(multiple)}
	case "组选6":
		p := make([]int, 3)
		d := make([]int, 3)
		copy(p, purchase.Numbers)
		copy(d, drawNums)
		sort.Ints(p)
		sort.Ints(d)
		if p[0] == d[0] && p[1] == d[1] && p[2] == d[2] {
			return 2, "组选6奖", 173 * float64(multiple)}
	case "组选3":
		p := make([]int, 3)
		d := make([]int, 3)
		copy(p, purchase.Numbers)
		copy(d, drawNums)
		sort.Ints(p)
		sort.Ints(d)
		if p[0] == d[0] && p[1] == d[1] && p[2] == d[2] {
			return 2, "组选3奖", 346 * float64(multiple)}
	}
	return 0, "未中奖", 0
}

func CalculatePaiLie5(purchaseJSON, drawJSON string, multiple int) (level int, name string, amount float64) {
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
		return 1, "直选奖", 100000 * float64(multiple)
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
	{5, 1, 4, "四等奖", 200},
	{5, 0, 5, "五等奖", 50},
	{4, 1, 6, "六等奖", 10},
	{4, 0, 7, "七等奖", 5},
}

func CalculateQiLeCai(purchaseJSON, drawJSON string, multiple int) (level int, name string, amount float64) {
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
			return p.level, p.name, p.amount * float64(multiple)
		}
	}
	return 0, "未中奖", 0
}
