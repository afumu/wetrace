package model

// MemberActivity 成员活跃度统计
type MemberActivity struct {
	MemberID     int64  `json:"memberId"`
	PlatformID   string `json:"platformId"`
	Name         string `json:"name"`
	MessageCount int    `json:"messageCount"`
	Avatar       string `json:"avatar,omitempty"`
}
