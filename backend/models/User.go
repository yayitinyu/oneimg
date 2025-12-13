package models

// 用户模型
type User struct {
	Id       int    `gorm:"primarykey;column:id" json:"id"`
	Role     int    `gorm:"default:1;uniqueIndex:unique_idx" json:"role"`
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `gorm:"column:nickname;default:''" json:"nickname"` // 昵称
	Avatar   string `gorm:"column:avatar;default:''" json:"avatar"`     // 头像URL
}
