package models

import "time"

// AppSession 记录「用户在应用里主动退出登录」这件事，而不是记录登录本身。
//
// 为什么需要它：飞牛统一网关域上，服务端每个请求都能从网关注入的 X-Trim-Userid
// 认出 NAS 用户，于是只要 NAS 会话还在，就能一直隐式认回该用户已绑定的应用账号，
// 应用里的「退出登录」会被下一次请求立刻撤销。所以网关隐式认人之前先看这张表：
// 该用户主动登出后，网关身份不再自动换登录态，必须重新显式登录（点「使用飞牛
// NAS 登录」或输应用账号密码）。
type AppSession struct {
	UserID uint `gorm:"primarykey"`
	// Suppressed 为 true 表示该用户的网关隐式登录被抑制，直到他显式登录。
	Suppressed bool
	// SuppressedAt 是抑制发生的时间，仅用于排查。
	SuppressedAt time.Time
	UpdatedAt    time.Time
}
