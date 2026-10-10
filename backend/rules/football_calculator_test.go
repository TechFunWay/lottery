package rules

import (
	"encoding/json"
	"lottery-backend/models"
	"testing"
)

func selectionsJSON(t *testing.T, selections []models.FootballSelection) string {
	t.Helper()
	b, err := json.Marshal(selections)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestCalculateFootballBet_MultipleSelectionsInSameMatch(t *testing.T) {
	matches := []models.FootballMatch{
		{MatchID: "周一001", Status: models.MatchFinished, HomeScore: 2, AwayScore: 1},
		{MatchID: "周一002", Status: models.MatchFinished, HomeScore: 0, AwayScore: 0},
	}
	selections := []models.FootballSelection{
		{MatchID: "周一001", PlayType: models.PlayWinDrawLoss, Selection: "3", Odds: 1.80},
		{MatchID: "周一001", PlayType: models.PlayWinDrawLoss, Selection: "1", Odds: 3.20},
		{MatchID: "周一001", PlayType: models.PlayWinDrawLoss, Selection: "0", Odds: 4.10},
		{MatchID: "周一002", PlayType: models.PlayWinDrawLoss, Selection: "3", Odds: 2.10},
		{MatchID: "周一002", PlayType: models.PlayWinDrawLoss, Selection: "1", Odds: 3.00},
		{MatchID: "周一002", PlayType: models.PlayWinDrawLoss, Selection: "0", Odds: 2.80},
	}

	result := CalculateFootballBet(selectionsJSON(t, selections), matches)
	if !result.Hit {
		t.Fatal("两场复式均有一个命中选项，应判定中奖")
	}
	if result.WinAmount != 5.4 {
		t.Fatalf("奖金赔率 = 1.80 * 3.00 = 5.40，得到 %v", result.WinAmount)
	}
}

func TestCalculateFootballBet_OneMatchHasNoHit(t *testing.T) {
	matches := []models.FootballMatch{
		{MatchID: "周一001", Status: models.MatchFinished, HomeScore: 2, AwayScore: 1},
		{MatchID: "周一002", Status: models.MatchFinished, HomeScore: 0, AwayScore: 0},
	}
	selections := []models.FootballSelection{
		{MatchID: "周一001", PlayType: models.PlayWinDrawLoss, Selection: "3", Odds: 1.80},
		{MatchID: "周一002", PlayType: models.PlayWinDrawLoss, Selection: "0", Odds: 2.80},
	}

	result := CalculateFootballBet(selectionsJSON(t, selections), matches)
	if result.Hit || result.WinAmount != 0 {
		t.Fatalf("有一场未命中应判未中奖，得到 %+v", result)
	}
}
