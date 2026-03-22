package models

import (
	"time"

	"gorm.io/gorm"
)

// SystemUpgrade 系统升级记录
type SystemUpgrade struct {
	ID        uint           `gorm:"primaryKey"`
	Version   string         `gorm:"type:varchar(20);uniqueIndex;not null;comment:版本号"`
	Name      string         `gorm:"type:varchar(100);not null;comment:升级名称"`
	Status    int            `gorm:"type:int;default:0;comment:升级状态:0-待执行,1-成功,2-失败"`
	StartTime time.Time      `comment:开始时间"`
	EndTime   time.Time      `comment:结束时间"`
	Remark    string         `gorm:"type:text;comment:备注/错误信息"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName 指定表名
func (SystemUpgrade) TableName() string {
	return "system_upgrade"
}
