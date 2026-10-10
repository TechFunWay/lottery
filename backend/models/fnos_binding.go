package models

import "time"

// FnOSBinding associates one trusted NAS-local UID with one application user.
// It intentionally stores no fnOS API token or session credential.
type FnOSBinding struct {
	ID           uint   `json:"id" gorm:"primarykey"`
	UserID       uint   `json:"user_id" gorm:"not null;uniqueIndex"`
	FnOSUserID   uint   `json:"-" gorm:"column:fnos_user_id;not null;uniqueIndex"`
	FnOSUsername string `json:"fnos_username" gorm:"column:fnos_username;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
