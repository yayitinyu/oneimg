package models

import "time"

// InvitationCode stores only a hash of the secret. The plaintext code is
// returned once when an administrator creates it.
type InvitationCode struct {
	Id        int        `gorm:"primaryKey;column:id" json:"id"`
	CodeHash  string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	Hint      string     `gorm:"size:12;not null" json:"hint"`
	CreatedAt time.Time  `json:"created_at"`
	UsedAt    *time.Time `gorm:"index" json:"used_at,omitempty"`
	UsedBy    *int       `gorm:"index" json:"used_by,omitempty"`
}
