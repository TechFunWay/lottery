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
	Multiple     int         `json:"multiple" gorm:"default:1"` // 倍数
	Append       bool        `json:"append" gorm:"default:false"` // 是否追加，仅大乐透有效
	Periods      int         `json:"periods" gorm:"default:1"` // 期数
	Remark       string      `json:"remark"`
	Status       string      `json:"status" gorm:"default:'待开奖'"` // 待开奖/已开奖/未中奖/已中奖
}

// DrawResult 开奖结果
type DrawResult struct {
	ID        uint       `json:"id" gorm:"primarykey"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	LotteryType LotteryType `json:"lottery_type" gorm:"not null;uniqueIndex:idx_lottery_issue"`
	IssueNumber string      `json:"issue_number" gorm:"not null;uniqueIndex:idx_lottery_issue"`
	DrawDate    time.Time   `json:"draw_date" gorm:"not null"`
	Numbers     string      `json:"numbers" gorm:"not null"` // JSON 格式存储开奖号码
	Source      string      `json:"source" gorm:"default:'manual'"` // manual/auto
	FuYunAward  bool        `json:"fu_yun_award" gorm:"default:false"` // 双色球福运奖：奖池≥15亿派奖期，中3红奖5元（按期手动标记）
	// 七星彩一、二等奖为浮动奖：自动抓取时把官方公布的单注奖金一并存下；
	// 手动录入或历史数据为空时，算奖回退到保底参考值
	FirstAmount  *float64 `json:"first_amount" gorm:"column:first_amount"`
	SecondAmount *float64 `json:"second_amount" gorm:"column:second_amount"`
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
	PrizeAmount float64     `json:"prize_amount"`             // 系统计算的奖金
	ManualAmount *float64   `json:"manual_amount"`            // 手动调整后的奖金（如活动翻倍），非空时以此为准，重算不覆盖
	Purchase    PurchaseRecord `json:"purchase" gorm:"foreignKey:PurchaseID"`
	Draw        DrawResult     `json:"draw" gorm:"foreignKey:DrawID"`
}
