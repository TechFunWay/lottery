package rules

import "testing"

func ptr(v float64) *float64 { return &v }

func TestCalculateQiXingCai(t *testing.T) {
	draw := `{"red":[5,1,8,0,8,0],"blue":[2]}`
	cases := []struct {
		name  string
		bet   string
		level int
		amt   float64
	}{
		{"一等奖 全中", `{"red":[5,1,8,0,8,0],"blue":[2]}`, 1, 5000000},
		{"二等奖 前6全中特不中", `{"red":[5,1,8,0,8,0],"blue":[7]}`, 2, 30000},
		{"三等奖 5位+特", `{"red":[5,1,8,0,8,9],"blue":[2]}`, 3, 3000},
		{"四等奖 5位无特", `{"red":[5,1,8,0,8,9],"blue":[7]}`, 4, 500},
		{"四等奖 4位+特", `{"red":[5,1,8,0,9,9],"blue":[2]}`, 4, 500},
		{"五等奖 4位无特", `{"red":[5,1,8,0,9,9],"blue":[7]}`, 5, 30},
		{"五等奖 3位+特", `{"red":[5,1,8,9,9,9],"blue":[2]}`, 5, 30},
		{"六等奖 仅特", `{"red":[9,9,9,9,9,9],"blue":[2]}`, 6, 5},
		{"未中奖", `{"red":[9,9,9,9,9,9],"blue":[7]}`, 0, 0},
		{"旧格式裸数组", `[5,1,8,0,8,0,2]`, 1, 5000000},
	}
	for _, c := range cases {
		level, _, amount := CalculateQiXingCai(c.bet, draw, 1, nil, nil)
		if level != c.level || amount != c.amt {
			t.Errorf("%s: got level=%d amount=%.0f, want level=%d amount=%.0f", c.name, level, amount, c.level, c.amt)
		}
	}
	// 倍数与浮动奖：一等奖当期公布 625 万
	level, _, amount := CalculateQiXingCai(`{"red":[5,1,8,0,8,0],"blue":[2]}`, draw, 2, ptr(6250000), ptr(29001))
	if level != 1 || amount != 12500000 {
		t.Errorf("浮动奖×2倍: got level=%d amount=%.0f", level, amount)
	}
	// 二等奖用当期公布值
	_, _, amount = CalculateQiXingCai(`{"red":[5,1,8,0,8,0],"blue":[7]}`, draw, 1, ptr(6250000), ptr(29001))
	if amount != 29001 {
		t.Errorf("二等奖浮动: got %.0f want 29001", amount)
	}
	// 特别号缺失（red-only 旧数据）按不中处理
	level, _, _ = CalculateQiXingCai(`{"red":[5,1,8,0,8,0],"blue":[2]}`, `{"red":[5,1,8,0,8,0]}`, 1, nil, nil)
	if level != 2 {
		t.Errorf("缺特别号: got level=%d want 2", level)
	}
}
