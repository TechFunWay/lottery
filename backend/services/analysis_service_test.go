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

func mkDraw(lt models.LotteryType, issue, numbers string) models.DrawResult {
	return models.DrawResult{LotteryType: lt, IssueNumber: issue, Numbers: numbers}
}

func TestAnalyzeDrawsSSQ(t *testing.T) {
	draws := []models.DrawResult{
		// 故意乱序，验证内部按期号升序
		mkDraw(models.ShuangSeQiu, "2026002", `{"red":[1,2,3,4,5,7],"blue":[9]}`),
		mkDraw(models.ShuangSeQiu, "2026001", `{"red":[1,2,3,4,5,6],"blue":[8]}`),
	}
	res, err := AnalyzeDraws(models.ShuangSeQiu, draws)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.IssueCount != 2 {
		t.Fatalf("issueCount = %d", res.IssueCount)
	}
	if res.Issues[0] != "2026001" || res.Issues[1] != "2026002" {
		t.Errorf("issues not ascending: %v", res.Issues)
	}
	if len(res.Zones) != 2 || res.Zones[0].Name != "红球" || res.Zones[1].Name != "蓝球" {
		t.Fatalf("zones = %+v", res.Zones)
	}
	// 红球 1 出现 2 次
	if res.Zones[0].Frequency[0].Num != 1 || res.Zones[0].Frequency[0].Count != 2 {
		t.Errorf("red 1 freq = %+v", res.Zones[0].Frequency[0])
	}
	// 走势与期号对齐：首期红球含 6
	if len(res.Zones[0].Trend) != 2 || res.Zones[0].Trend[0][5] != 6 {
		t.Errorf("trend = %v", res.Zones[0].Trend)
	}
	// metrics 首期和值 = 1+2+3+4+5+6 = 21
	if len(res.Metrics) != 2 || res.Metrics[0].Sum != 21 {
		t.Errorf("metrics = %+v", res.Metrics)
	}
}

func TestAnalyzeDrawsEmpty(t *testing.T) {
	res, err := AnalyzeDraws(models.DaLeTou, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.IssueCount != 0 || len(res.Issues) != 0 {
		t.Errorf("expected empty, got %+v", res)
	}
	if len(res.Zones) != 2 { // 前区 + 后区 仍按配置生成
		t.Errorf("zones = %d", len(res.Zones))
	}
	if len(res.Metrics) != 0 {
		t.Errorf("metrics should be empty, got %d", len(res.Metrics))
	}
}

func TestAnalyzeDrawsSkipsBadJSON(t *testing.T) {
	draws := []models.DrawResult{
		mkDraw(models.PaiLie3, "2026001", `{"numbers":[1,2,3]}`),
		mkDraw(models.PaiLie3, "2026002", `not-json`),
	}
	res, err := AnalyzeDraws(models.PaiLie3, draws)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.IssueCount != 1 || res.Issues[0] != "2026001" {
		t.Errorf("expected 1 valid issue, got %+v", res.Issues)
	}
}

func TestAnalyzeDrawsUnknownType(t *testing.T) {
	if _, err := AnalyzeDraws(models.LotteryType("不存在"), nil); err == nil {
		t.Fatal("expected error for unknown type")
	}
}
