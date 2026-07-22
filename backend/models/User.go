package models

import "time"

const (
	RoleAdmin = 1
	RoleUser  = 2
	RoleGuest = 3
)

// User is a persisted account. Guest identities only live in a session and
// therefore never create rows in this table.
type User struct {
	Id        int       `gorm:"primarykey;column:id" json:"id"`
	Role      int       `gorm:"not null;default:2;index" json:"role"`
	Username  string    `gorm:"size:64;not null;uniqueIndex" json:"username"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Nickname  string    `gorm:"column:nickname;default:''" json:"nickname"`
	Avatar    string    `gorm:"column:avatar;default:''" json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
}
