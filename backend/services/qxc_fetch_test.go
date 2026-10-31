package services

import (
	"encoding/json"
	"testing"

	"lottery-backend/models"
	"lottery-backend/rules"
)

func TestSportteryItemToQixingcai(t *testing.T) {
	// 真实接口样本：26123 期 5 1 8 0 8 0 2，二等奖单注 29,001，一等奖无人中（---）
	raw := `{"lotteryDrawNum":"26123","lotteryDrawResult":"5 1 8 0 8 0 2","lotteryDrawTime":"2026-10-27","totalSaleAmount":"15,663,882","prizeLevelList":[{"prizeLevel":"一等奖","stakeAmount":"---"},{"prizeLevel":"二等奖","stakeAmount":"29,001"},{"prizeLevel":"三等奖","stakeAmount":"3,000"}]}`
	var item sportteryDrawItem
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		t.Fatal(err)
	}
	draw, err := sportteryItemToQixingcai(item)
	if err != nil {
		t.Fatal(err)
	}
	if draw.LotteryType != models.QiXingCai || draw.IssueNumber != "26123" {
		t.Fatalf("basic fields: %+v", draw)
	}
	var nums rules.QiXingCaiNumbers
	if err := json.Unmarshal([]byte(draw.Numbers), &nums); err != nil {
		t.Fatal(err)
	}
	if len(nums.Red) != 6 || nums.Red[0] != 5 || len(nums.Blue) != 1 || nums.Blue[0] != 2 {
		t.Fatalf("numbers: %s", draw.Numbers)
	}
	if draw.FirstAmount != nil {
		t.Fatalf("一等奖无人中应为 nil, got %v", *draw.FirstAmount)
	}
	if draw.SecondAmount == nil || *draw.SecondAmount != 29001 {
		t.Fatalf("二等奖应为 29001, got %v", draw.SecondAmount)
	}
}
