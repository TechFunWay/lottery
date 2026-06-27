package services

import (
	"encoding/json"
	"testing"

	"lottery-backend/models"
)

func TestRawNumbersField(t *testing.T) {
	var r rawNumbers
	if err := json.Unmarshal([]byte(`{"red":[1,7,15],"blue":[8]}`), &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := r.field("red"); len(got) != 3 || got[0] != 1 || got[2] != 15 {
		t.Fatalf("field red = %v", got)
	}
	if got := r.field("blue"); len(got) != 1 || got[0] != 8 {
		t.Fatalf("field blue = %v", got)
	}
	if got := r.field("front"); got != nil {
		t.Fatalf("field front expected nil, got %v", got)
	}
}

func TestZoneValues(t *testing.T) {
	var r rawNumbers
	_ = json.Unmarshal([]byte(`{"numbers":[3,5,9]}`), &r)
	whole := zoneDef{field: "numbers", index: -1}
	if got := whole.values(r); len(got) != 3 || got[1] != 5 {
		t.Fatalf("whole zone = %v", got)
	}
	pos := zoneDef{field: "numbers", index: 2}
	if got := pos.values(r); len(got) != 1 || got[0] != 9 {
		t.Fatalf("pos zone = %v", got)
	}
	oob := zoneDef{field: "numbers", index: 9}
	if got := oob.values(r); got != nil {
		t.Fatalf("oob zone expected nil, got %v", got)
	}
}

func TestLotteryConfigsCoverAllTypes(t *testing.T) {
	for _, lt := range []models.LotteryType{
		models.ShuangSeQiu, models.DaLeTou, models.FuCai3D,
		models.PaiLie3, models.PaiLie5, models.QiLeCai, models.QiXingCai,
	} {
		if _, ok := lotteryConfigs[lt]; !ok {
			t.Errorf("缺少彩种配置: %s", lt)
		}
	}
}

func TestComputeFrequency(t *testing.T) {
	// 3 期，单位号码区 0-9
	per := [][]int{{1}, {2}, {1}}
	freq := computeFrequency(0, 9, per)
	if len(freq) != 10 {
		t.Fatalf("expected 10 items, got %d", len(freq))
	}
	if freq[0].Num != 0 || freq[0].Count != 0 {
		t.Errorf("num 0 = %+v", freq[0])
	}
	if freq[1].Num != 1 || freq[1].Count != 2 {
		t.Errorf("num 1 = %+v", freq[1])
	}
	if freq[2].Count != 1 {
		t.Errorf("num 2 count = %d", freq[2].Count)
	}
}

func TestComputeOmission(t *testing.T) {
	per := [][]int{{1}, {2}, {1}} // 时间升序
	om := computeOmission(0, 9, per)
	get := func(n int) OmissionItem {
		for _, o := range om {
			if o.Num == n {
				return o
			}
		}
		t.Fatalf("num %d not found", n)
		return OmissionItem{}
	}
	// 1：最近一期(末期)开出 → 当前遗漏 0；最大遗漏 1（中间漏 1 期）
	if o := get(1); o.Current != 0 || o.Max != 1 {
		t.Errorf("num 1 = %+v", o)
	}
	// 2：第 2 期开出，之后漏 1 期 → 当前遗漏 1；最大遗漏 1
	if o := get(2); o.Current != 1 || o.Max != 1 {
		t.Errorf("num 2 = %+v", o)
	}
	// 0：从未开出 → 当前遗漏 3；最大遗漏 3
	if o := get(0); o.Current != 3 || o.Max != 3 {
		t.Errorf("num 0 = %+v", o)
	}
}

func TestComputeMetrics(t *testing.T) {
	issues := []string{"2026001"}
	main := [][]int{{1, 2, 17, 33}} // 和=53 跨度=32 奇偶: 1,17,33奇=3, 2偶=1 → 3:1 大小(阈值16): 17,33>16=2, 1,2=2 → 2:2
	m := computeMetrics(issues, main, true, 16)
	if len(m) != 1 {
		t.Fatalf("expected 1, got %d", len(m))
	}
	if m[0].Issue != "2026001" || m[0].Sum != 53 || m[0].Span != 32 {
		t.Errorf("sum/span = %+v", m[0])
	}
	if m[0].OddEven != "3:1" {
		t.Errorf("oddEven = %q", m[0].OddEven)
	}
	if m[0].BigSmall != "2:2" {
		t.Errorf("bigSmall = %q", m[0].BigSmall)
	}
}

func TestComputeMetricsNoBigSmall(t *testing.T) {
	m := computeMetrics([]string{"x"}, [][]int{{3, 5, 9}}, false, 0)
	if m[0].Sum != 17 || m[0].Span != 6 {
		t.Errorf("sum/span = %+v", m[0])
	}
	if m[0].BigSmall != "" {
		t.Errorf("expected empty bigSmall, got %q", m[0].BigSmall)
	}
}
