package rules

import "testing"

// 双色球福运奖：奖池≥15亿活动期，中3个红球（蓝球未中）奖5元
func TestShuangSeQiu_FuYunAward(t *testing.T) {
	// 投注红 1-6，蓝 7；开奖红 1,2,3,30,31,32，蓝 8 → 中3红+0蓝
	purchase := `{"red":[1,2,3,4,5,6],"blue":[7]}`
	draw := `{"red":[1,2,3,30,31,32],"blue":[8]}`

	// 非活动期：中3红+0蓝 不中奖
	level, _, amount := CalculateShuangSeQiu(purchase, draw, 1, false)
	if level != 0 || amount != 0 {
		t.Fatalf("非活动期中3红应不中奖, 得到 level=%d amount=%v", level, amount)
	}

	// 活动期：中3红+0蓝 中福运奖5元
	level, name, amount := CalculateShuangSeQiu(purchase, draw, 1, true)
	if level == 0 || amount != 5 || name != "福运奖" {
		t.Fatalf("活动期中3红应中福运奖5元, 得到 level=%d name=%s amount=%v", level, name, amount)
	}

	// 活动期倍投：5元×2
	_, _, amount = CalculateShuangSeQiu(purchase, draw, 2, true)
	if amount != 10 {
		t.Fatalf("福运奖2倍应为10元, 得到 %v", amount)
	}
}

// 福运奖活动不影响其它常规奖级
func TestShuangSeQiu_FuYunDoesNotAffectNormalPrizes(t *testing.T) {
	purchase := `{"red":[1,2,3,4,5,6],"blue":[7]}`

	// 一等奖 6+1，活动开关不影响
	draw := `{"red":[1,2,3,4,5,6],"blue":[7]}`
	level, _, amount := CalculateShuangSeQiu(purchase, draw, 1, true)
	if level != 1 || amount != 5000000 {
		t.Fatalf("6+1应为一等奖500万, 得到 level=%d amount=%v", level, amount)
	}

	// 中3红+1蓝 仍为五等奖10元（不是福运奖）
	draw = `{"red":[1,2,3,30,31,32],"blue":[7]}`
	level, name, amount := CalculateShuangSeQiu(purchase, draw, 1, true)
	if level != 5 || amount != 10 || name != "五等奖" {
		t.Fatalf("3红+1蓝应为五等奖10元, 得到 level=%d name=%s amount=%v", level, name, amount)
	}
}
