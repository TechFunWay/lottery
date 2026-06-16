package rules

import (
	"encoding/json"
	"fmt"
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

// 定位胆每命中一个位置的单位奖金（非官方标准玩法，可按需调整）
const dingweiUnitPrize = 18.0

// DigitBet 数字型彩票投注（福彩3D / 排列3 / 排列5）
// 新结构：
//   直选/定位胆 → positions（每位的候选数字数组，定位胆未投的位为空数组）
//   组选3/组选6 → group（选中的数字集合）
// 兼容旧结构：{ numbers:[...], bet_type:"直选" } 或裸数组 [..]
type DigitBet struct {
	Play      string  `json:"play"`      // 直选/组选3/组选6/定位胆
	Positions [][]int `json:"positions"` // 直选/定位胆
	Group     []int   `json:"group"`     // 组选
	// 兼容旧格式
	Numbers []int  `json:"numbers"`
	BetType string `json:"bet_type"`
}

// parseDigitBet 解析投注号码，兼容旧格式
func parseDigitBet(purchaseJSON string) (*DigitBet, bool) {
	var b DigitBet
	if err := json.Unmarshal([]byte(purchaseJSON), &b); err != nil {
		// 可能是裸数组
		var arr []int
		if err2 := json.Unmarshal([]byte(purchaseJSON), &arr); err2 != nil {
			return nil, false
		}
		b.Numbers = arr
	}

	// 归一化玩法
	if b.Play == "" {
		if b.BetType != "" {
			b.Play = b.BetType
		} else {
			b.Play = "直选"
		}
	}

	// 旧格式 numbers → 新结构
	if len(b.Positions) == 0 && len(b.Group) == 0 && len(b.Numbers) > 0 {
		if b.Play == "组选3" || b.Play == "组选6" {
			b.Group = b.Numbers
		} else {
			for _, n := range b.Numbers {
				b.Positions = append(b.Positions, []int{n})
			}
		}
	}
	return &b, true
}

// parseDrawDigits 解析开奖号码，兼容裸数组 / {numbers:[...]} / {numbers:[...],bet_type}
func parseDrawDigits(drawJSON string) ([]int, bool) {
	var arr []int
	if err := json.Unmarshal([]byte(drawJSON), &arr); err == nil {
		return arr, true
	}
	var obj struct {
		Numbers []int `json:"numbers"`
	}
	if err := json.Unmarshal([]byte(drawJSON), &obj); err == nil && len(obj.Numbers) > 0 {
		return obj.Numbers, true
	}
	return nil, false
}

func containsInt(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// isZu6 判断开奖号是否为组六形态（3 个互不相同）
func isZu6(d []int) bool {
	return len(d) == 3 && d[0] != d[1] && d[1] != d[2] && d[0] != d[2]
}

// zu3Pair 判断开奖号是否为组三形态（AAB，恰好一对），返回成对值与单只值
func zu3Pair(d []int) (pair int, single int, ok bool) {
	if len(d) != 3 {
		return 0, 0, false
	}
	switch {
	case d[0] == d[1] && d[1] != d[2]:
		return d[0], d[2], true
	case d[0] == d[2] && d[0] != d[1]:
		return d[0], d[1], true
	case d[1] == d[2] && d[1] != d[0]:
		return d[1], d[0], true
	}
	return 0, 0, false
}

// allIn 判断 vals 中每个数字都在 set 内
func allIn(vals, set []int) bool {
	for _, v := range vals {
		if !containsInt(set, v) {
			return false
		}
	}
	return true
}

// calcDigit 数字型彩票通用中奖计算
// zu6Prize/zu3Prize/dingweiUnit 为 0 时表示该玩法不适用（如排列5 仅直选）
func calcDigit(purchaseJSON, drawJSON string, multiple int, digitCount int, directPrize, zu6Prize, zu3Prize, dingweiUnit float64) (int, string, float64) {
	bet, ok := parseDigitBet(purchaseJSON)
	if !ok {
		return 0, "未中奖", 0
	}
	draw, ok := parseDrawDigits(drawJSON)
	if !ok || len(draw) != digitCount {
		return 0, "未中奖", 0
	}
	m := float64(multiple)

	switch bet.Play {
	case "直选":
		if len(bet.Positions) != digitCount {
			return 0, "未中奖", 0
		}
		for i := 0; i < digitCount; i++ {
			if !containsInt(bet.Positions[i], draw[i]) {
				return 0, "未中奖", 0
			}
		}
		return 1, "直选奖", directPrize * m
	case "定位胆":
		if dingweiUnit <= 0 {
			return 0, "未中奖", 0
		}
		matched := 0
		for i := 0; i < digitCount && i < len(bet.Positions); i++ {
			if len(bet.Positions[i]) > 0 && containsInt(bet.Positions[i], draw[i]) {
				matched++
			}
		}
		if matched > 0 {
			return 1, fmt.Sprintf("定位胆(中%d位)", matched), dingweiUnit * float64(matched) * m
		}
	case "组选6":
		if zu6Prize > 0 && isZu6(draw) && allIn(draw, bet.Group) {
			return 2, "组选6奖", zu6Prize * m
		}
	case "组选3":
		if zu3Prize > 0 {
			if pair, single, isZu3 := zu3Pair(draw); isZu3 && containsInt(bet.Group, pair) && containsInt(bet.Group, single) {
				return 2, "组选3奖", zu3Prize * m
			}
		}
	}
	return 0, "未中奖", 0
}

func CalculateFuCai3D(purchaseJSON, drawJSON string, multiple int) (level int, name string, amount float64) {
	return calcDigit(purchaseJSON, drawJSON, multiple, 3, 1040, 173, 346, dingweiUnitPrize)
}

func CalculatePaiLie3(purchaseJSON, drawJSON string, multiple int) (level int, name string, amount float64) {
	return calcDigit(purchaseJSON, drawJSON, multiple, 3, 1040, 173, 346, dingweiUnitPrize)
}

func CalculatePaiLie5(purchaseJSON, drawJSON string, multiple int) (level int, name string, amount float64) {
	return calcDigit(purchaseJSON, drawJSON, multiple, 5, 100000, 0, 0, 0)
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
