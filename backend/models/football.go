package models

import (
	"time"

	"gorm.io/gorm"
)

type FootballPlayType string

const (
	PlayWinDrawLoss       FootballPlayType = "胜平负"
	PlayHandicapWinDraw   FootballPlayType = "让球胜平负"
	PlayScore             FootballPlayType = "比分"
	PlayTotalGoals        FootballPlayType = "总进球"
	PlayHalfFull          FootballPlayType = "半全场"
)

type FootballMatchStatus string

const (
	MatchNotStarted FootballMatchStatus = "未开赛"
	MatchInProgress FootballMatchStatus = "进行中"
	MatchFinished   FootballMatchStatus = "已完赛"
	MatchCancelled  FootballMatchStatus = "已取消"
	MatchPostponed  FootballMatchStatus = "延期"
)

type FootballBetType string

const (
	BetSingle FootballBetType = "单关"
	Bet2Parlay FootballBetType = "2串1"
	Bet3Parlay FootballBetType = "3串1"
	Bet4Parlay FootballBetType = "4串1"
	Bet5Parlay FootballBetType = "5串1"
	Bet6Parlay FootballBetType = "6串1"
	Bet7Parlay FootballBetType = "7串1"
	Bet8Parlay FootballBetType = "8串1"
)

type FootballBetStatus string

const (
	BetPending  FootballBetStatus = "待开奖"
	BetWon      FootballBetStatus = "已中奖"
	BetLost     FootballBetStatus = "未中奖"
	BetPartial  FootballBetStatus = "部分中奖"
)

type FootballMatch struct {
	ID            uint      `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	MatchID       string    `json:"match_id" gorm:"uniqueIndex:idx_football_match_id"`
	IssueNumber   string    `json:"issue_number"`
	League        string    `json:"league"`
	HomeTeam      string    `json:"home_team"`
	AwayTeam      string    `json:"away_team"`
	MatchTime     time.Time `json:"match_time"`
	Status        FootballMatchStatus `json:"status" gorm:"default:'未开赛'"`
	HomeScore     int       `json:"home_score"`
	AwayScore     int       `json:"away_score"`
	HalfHomeScore int       `json:"half_home_score"`
	HalfAwayScore int       `json:"half_away_score"`
	Handicap      float64   `json:"handicap"`
	Source        string    `json:"source" gorm:"default:'manual'"`
}

type FootballBet struct {
	ID         uint              `json:"id" gorm:"primarykey"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	DeletedAt  gorm.DeletedAt    `json:"-" gorm:"index"`
	UserID     uint              `json:"user_id" gorm:"not null;index"`
	BetType    FootballBetType   `json:"bet_type" gorm:"default:'单关'"`
	Amount     float64           `json:"amount"`
	Multiple   int               `json:"multiple" gorm:"default:1"`
	Status     FootballBetStatus `json:"status" gorm:"default:'待开奖'"`
	Selections string            `json:"selections" gorm:"type:text"`
	Remark     string            `json:"remark"`
	WinAmount  float64           `json:"win_amount"`
}

type FootballSelection struct {
	MatchID   string           `json:"match_id"`
	PlayType  FootballPlayType `json:"play_type"`
	Selection string           `json:"selection"`
	Odds      float64          `json:"odds"`
	Handicap  float64          `json:"handicap,omitempty"`
}
