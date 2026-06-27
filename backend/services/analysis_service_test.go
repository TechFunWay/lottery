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
