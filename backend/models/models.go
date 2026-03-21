package models

import (
	"time"

	"gorm.io/gorm"
)

// LotteryType 彩票类型枚举
type LotteryType string

const (
	ShuangSeQiu LotteryType = "双色球"
	DaLeTou     LotteryType = "大乐透"
	FuCai3D     LotteryType = "福彩3D"
	PaiLie3     LotteryType = "排列3"
	PaiLie5     LotteryType = "排列5"
	QiLeCai     LotteryType = "七乐彩"
	QiXingCai   LotteryType = "七星彩"
)

// BetType 投注方式
type BetType string

const (
	DanShi  BetType = "单式"
	FuShi   BetType = "复式"
	DanTuo  BetType = "胆拖"
)

// PurchaseRecord 购买记录
type PurchaseRecord struct {
	ID        uint       `json:"id" gorm:"primarykey"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	UserID    uint       `json:"user_id" gorm:"not null;index"` // 用户ID,用于数据隔离
	LotteryType  LotteryType `json:"lottery_type" gorm:"not null"`
	IssueNumber  string      `json:"issue_number" gorm:"not null"` // 期号
	PurchaseDate time.Time   `json:"purchase_date" gorm:"not null"`
	Numbers      string      `json:"numbers" gorm:"not null"` // JSON 格式存储号码
	BetType      BetType     `json:"bet_type" gorm:"default:'单式'"`
	Amount       float64     `json:"amount" gorm:"not null"` // 购买金额（元）
	Remark       string      `json:"remark"`
	Status       string      `json:"status" gorm:"default:'待开奖'"` // 待开奖/已开奖/未中奖/已中奖
}

// DrawResult 开奖结果
type DrawResult struct {
	ID        uint       `json:"id" gorm:"primarykey"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	LotteryType LotteryType `json:"lottery_type" gorm:"not null"`
	IssueNumber string      `json:"issue_number" gorm:"not null;uniqueIndex:idx_lottery_issue"`
	DrawDate    time.Time   `json:"draw_date" gorm:"not null"`
	Numbers     string      `json:"numbers" gorm:"not null"` // JSON 格式存储开奖号码
	Source      string      `json:"source" gorm:"default:'manual'"` // manual/auto
}

// WinningRecord 中奖记录
type WinningRecord struct {
	ID        uint       `json:"id" gorm:"primarykey"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	PurchaseID  uint        `json:"purchase_id" gorm:"not null"`
	DrawID      uint        `json:"draw_id" gorm:"not null"`
	LotteryType LotteryType `json:"lottery_type" gorm:"not null"`
	IssueNumber string      `json:"issue_number" gorm:"not null"`
	PrizeLevel  int         `json:"prize_level"` // 0=未中奖, 1=一等奖, ...
	PrizeName   string      `json:"prize_name"`  // 如：一等奖、二等奖
	PrizeAmount float64     `json:"prize_amount"`
	Purchase    PurchaseRecord `json:"purchase" gorm:"foreignKey:PurchaseID"`
	Draw        DrawResult     `json:"draw" gorm:"foreignKey:DrawID"`
}
