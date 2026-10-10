package rules

import "testing"

func TestFuCai3D_ZhixuanSingle(t *testing.T) {
	// 直选单式命中
	level, _, amount := CalculateFuCai3D(`{"play":"直选","positions":[[1],[2],[3]]}`, `[1,2,3]`, 1)
	if level != 1 || amount != 1040 {
		t.Fatalf("直选单式命中 期望 level=1 amount=1040, 得到 level=%d amount=%v", level, amount)
	}
	// 顺序不同不中
	level, _, _ = CalculateFuCai3D(`{"play":"直选","positions":[[1],[2],[3]]}`, `[3,2,1]`, 1)
	if level != 0 {
		t.Fatalf("直选顺序错应不中, 得到 level=%d", level)
	}
}

func TestFuCai3D_ZhixuanMultiple(t *testing.T) {
	// 直选复式：第一位 1 或 5，命中 5,2,3
	level, _, amount := CalculateFuCai3D(`{"play":"直选","positions":[[1,5],[2],[3]]}`, `[5,2,3]`, 1)
	if level != 1 || amount != 1040 {
		t.Fatalf("直选复式命中 期望 level=1 amount=1040, 得到 level=%d amount=%v", level, amount)
	}
	// 复式倍数
	_, _, amount = CalculateFuCai3D(`{"play":"直选","positions":[[1,5],[2],[3]]}`, `[5,2,3]`, 2)
	if amount != 2080 {
		t.Fatalf("直选复式2倍 期望 amount=2080, 得到 %v", amount)
	}
}

func TestFuCai3D_Dingwei(t *testing.T) {
	// 定位胆：只投百位[1,5]和十位[2]，开奖 5,2,9 → 命中2位
	level, name, amount := CalculateFuCai3D(`{"play":"定位胆","positions":[[1,5],[2],[]]}`, `[5,2,9]`, 1)
	if level != 1 || amount != 36 {
		t.Fatalf("定位胆中2位 期望 amount=36(18*2), 得到 level=%d amount=%v name=%s", level, amount, name)
	}
	// 命中1位
	_, _, amount = CalculateFuCai3D(`{"play":"定位胆","positions":[[1,5],[2],[]]}`, `[5,9,9]`, 1)
	if amount != 18 {
		t.Fatalf("定位胆中1位 期望 amount=18, 得到 %v", amount)
	}
	// 全不中
	level, _, _ = CalculateFuCai3D(`{"play":"定位胆","positions":[[1,5],[2],[]]}`, `[9,9,9]`, 1)
	if level != 0 {
		t.Fatalf("定位胆全不中应 level=0, 得到 %d", level)
	}
}

func TestFuCai3D_Zu6(t *testing.T) {
	// 组选6：选 1,2,3,4，开奖 3,1,2（组六）→ 中
	level, _, amount := CalculateFuCai3D(`{"play":"组选6","group":[1,2,3,4]}`, `[3,1,2]`, 1)
	if level != 2 || amount != 173 {
		t.Fatalf("组选6命中 期望 level=2 amount=173, 得到 level=%d amount=%v", level, amount)
	}
	// 开奖为组三形态(有重复) → 组选6 不中
	level, _, _ = CalculateFuCai3D(`{"play":"组选6","group":[1,2,3,4]}`, `[1,1,2]`, 1)
	if level != 0 {
		t.Fatalf("组选6遇组三开奖应不中, 得到 level=%d", level)
	}
}

func TestFuCai3D_Zu3(t *testing.T) {
	// 组选3：选 1,2，开奖 1,1,2（组三 AAB）→ 中
	level, _, amount := CalculateFuCai3D(`{"play":"组选3","group":[1,2,3]}`, `[1,1,2]`, 1)
	if level != 2 || amount != 346 {
		t.Fatalf("组选3命中 期望 level=2 amount=346, 得到 level=%d amount=%v", level, amount)
	}
	// 开奖为组六(无重复) → 组选3 不中
	level, _, _ = CalculateFuCai3D(`{"play":"组选3","group":[1,2,3]}`, `[1,2,3]`, 1)
	if level != 0 {
		t.Fatalf("组选3遇组六开奖应不中, 得到 level=%d", level)
	}
}

func TestFuCai3D_LegacyFormat(t *testing.T) {
	// 旧格式：{numbers,bet_type}
	level, _, amount := CalculateFuCai3D(`{"numbers":[1,2,3],"bet_type":"直选"}`, `[1,2,3]`, 1)
	if level != 1 || amount != 1040 {
		t.Fatalf("旧格式直选命中 期望 level=1 amount=1040, 得到 level=%d amount=%v", level, amount)
	}
	// 开奖为对象格式 {numbers:[...]}
	level, _, _ = CalculateFuCai3D(`{"numbers":[1,2,3],"bet_type":"直选"}`, `{"numbers":[1,2,3]}`, 1)
	if level != 1 {
		t.Fatalf("开奖对象格式应能解析, 得到 level=%d", level)
	}
}

func TestPaiLie5_Zhixuan(t *testing.T) {
	level, _, amount := CalculatePaiLie5(`{"play":"直选","positions":[[1],[2],[3],[4],[5]]}`, `[1,2,3,4,5]`, 1)
	if level != 1 || amount != 100000 {
		t.Fatalf("排列5直选命中 期望 level=1 amount=100000, 得到 level=%d amount=%v", level, amount)
	}
	// 复式
	level, _, _ = CalculatePaiLie5(`{"play":"直选","positions":[[1,9],[2],[3],[4],[5]]}`, `[9,2,3,4,5]`, 1)
	if level != 1 {
		t.Fatalf("排列5直选复式命中应 level=1, 得到 %d", level)
	}
}
