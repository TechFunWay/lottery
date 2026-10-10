package rules

import (
	"encoding/json"
	"fmt"
	"lottery-backend/models"
)

type FootballSelectionResult struct {
	MatchID   string  `json:"match_id"`
	Hit       bool    `json:"hit"`
	PlayType  string  `json:"play_type"`
	Selection string  `json:"selection"`
	Odds      float64 `json:"odds"`
}

type FootballBetResult struct {
	Hit       bool                      `json:"hit"`
	WinAmount float64                   `json:"win_amount"`
	Details   []FootballSelectionResult `json:"details"`
}

func CalculateWinDrawLoss(homeScore, awayScore int) string {
	if homeScore > awayScore {
		return "3"
	} else if homeScore == awayScore {
		return "1"
	}
	return "0"
}

func CalculateHandicapWinDrawLoss(homeScore, awayScore int, handicap float64) string {
	adjustedHome := float64(homeScore) + handicap
	if adjustedHome > float64(awayScore) {
		return "3"
	} else if adjustedHome == float64(awayScore) {
		return "1"
	}
	return "0"
}

func CalculateScoreResult(homeScore, awayScore int) string {
	return fmt.Sprintf("%d:%d", homeScore, awayScore)
}

func CalculateTotalGoalsResult(homeScore, awayScore int) string {
	total := homeScore + awayScore
	if total >= 7 {
		return "7+"
	}
	return fmt.Sprintf("%d", total)
}

func CalculateHalfFullResult(halfHome, halfAway, fullHome, fullAway int) string {
	half := matchHalfResult(halfHome, halfAway)
	full := matchHalfResult(fullHome, fullAway)
	return half + full
}

func matchHalfResult(home, away int) string {
	if home > away {
		return "胜"
	} else if home == away {
		return "平"
	}
	return "负"
}

func CheckSelectionHit(selection models.FootballSelection, match models.FootballMatch) bool {
	if match.Status != models.MatchFinished {
		return false
	}

	switch selection.PlayType {
	case models.PlayWinDrawLoss:
		result := CalculateWinDrawLoss(match.HomeScore, match.AwayScore)
		return selection.Selection == result

	case models.PlayHandicapWinDraw:
		handicap := selection.Handicap
		if handicap == 0 {
			handicap = match.Handicap
		}
		result := CalculateHandicapWinDrawLoss(match.HomeScore, match.AwayScore, handicap)
		return selection.Selection == result

	case models.PlayScore:
		result := CalculateScoreResult(match.HomeScore, match.AwayScore)
		return selection.Selection == result

	case models.PlayTotalGoals:
		result := CalculateTotalGoalsResult(match.HomeScore, match.AwayScore)
		return selection.Selection == result

	case models.PlayHalfFull:
		result := CalculateHalfFullResult(match.HalfHomeScore, match.HalfAwayScore, match.HomeScore, match.AwayScore)
		return selection.Selection == result
	}

	return false
}

func CalculateFootballBet(selectionsJSON string, matches []models.FootballMatch) FootballBetResult {
	var selections []models.FootballSelection
	if err := json.Unmarshal([]byte(selectionsJSON), &selections); err != nil {
		return FootballBetResult{Hit: false, WinAmount: 0}
	}

	matchMap := make(map[string]models.FootballMatch)
	for _, m := range matches {
		matchMap[m.MatchID] = m
	}

	var details []FootballSelectionResult
	// 同一场比赛允许选择多个结果（复式）。同场选项是“或”的关系，
	// 不同场次才是串关的“且”关系。每场命中选项的赔率之和再参与串乘，
	// 等价于把复式拆成所有单独的串关组合后汇总奖金。
	matchHitOdds := make(map[string]float64)
	matchOrder := make([]string, 0)
	seenMatch := make(map[string]bool)

	for _, sel := range selections {
		if !seenMatch[sel.MatchID] {
			seenMatch[sel.MatchID] = true
			matchOrder = append(matchOrder, sel.MatchID)
		}
		match, ok := matchMap[sel.MatchID]
		if !ok {
			details = append(details, FootballSelectionResult{
				MatchID:   sel.MatchID,
				Hit:       false,
				PlayType:  string(sel.PlayType),
				Selection: sel.Selection,
				Odds:      sel.Odds,
			})
			continue
		}

		hit := CheckSelectionHit(sel, match)
		details = append(details, FootballSelectionResult{
			MatchID:   sel.MatchID,
			Hit:       hit,
			PlayType:  string(sel.PlayType),
			Selection: sel.Selection,
			Odds:      sel.Odds,
		})

		if hit && sel.Odds > 0 {
			matchHitOdds[sel.MatchID] += sel.Odds
		}
	}

	allHit := len(matchOrder) > 0
	totalOdds := 1.0
	for _, matchID := range matchOrder {
		odds := matchHitOdds[matchID]
		if odds <= 0 {
			allHit = false
			break
		}
		totalOdds *= odds
	}

	result := FootballBetResult{
		Hit:       allHit,
		WinAmount: 0,
		Details:   details,
	}

	if allHit && totalOdds > 0 {
		result.WinAmount = totalOdds
	}

	return result
}

func GetValidScoreOptions() []string {
	return []string{
		"1:0", "2:0", "2:1", "3:0", "3:1", "3:2",
		"4:0", "4:1", "4:2", "5:0", "5:1", "5:2",
		"0:1", "0:2", "1:2", "0:3", "1:3", "2:3",
		"0:4", "1:4", "2:4", "0:5", "1:5", "2:5",
		"0:0", "1:1", "2:2", "3:3", "胜其他", "平其他", "负其他",
	}
}

func GetValidTotalGoalsOptions() []string {
	return []string{"0", "1", "2", "3", "4", "5", "6", "7+"}
}

func GetValidHalfFullOptions() []string {
	return []string{"胜胜", "胜平", "胜负", "平胜", "平平", "平负", "负胜", "负平", "负负"}
}

func GetValidWinDrawLossOptions() []string {
	return []string{"3", "1", "0"}
}
