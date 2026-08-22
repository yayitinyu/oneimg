package models

import "time"

const (
	StorageMigrationPending   = "pending"
	StorageMigrationRunning   = "running"
	StorageMigrationFailed    = "failed"
	StorageMigrationCompleted = "completed"

	StorageMigrationItemPending   = "pending"
	StorageMigrationItemCopying   = "copying"
	StorageMigrationItemFailed    = "failed"
	StorageMigrationItemCompleted = "completed"
)

// StorageMigration stores enough information to resume a copy after the
// application container restarts. Credential fields are deliberately omitted
// from JSON responses.
type StorageMigration struct {
	ID              int        `json:"id" gorm:"primaryKey"`
	Status          string     `json:"status" gorm:"not null;index"`
	SourceType      string     `json:"source_type" gorm:"not null"`
	SourceEndpoint  string     `json:"source_endpoint" gorm:"not null"`
	SourceRegion    string     `json:"source_region" gorm:"not null"`
	SourceAccessKey string     `json:"-" gorm:"not null"`
	SourceSecretKey string     `json:"-" gorm:"not null"`
	SourceBucket    string     `json:"source_bucket" gorm:"not null"`
	SourcePathStyle bool       `json:"source_path_style" gorm:"not null;default:false"`
	TargetType      string     `json:"target_type" gorm:"not null"`
	TargetEndpoint  string     `json:"target_endpoint" gorm:"not null"`
	TargetRegion    string     `json:"target_region" gorm:"not null"`
	TargetAccessKey string     `json:"-" gorm:"not null"`
	TargetSecretKey string     `json:"-" gorm:"not null"`
	TargetBucket    string     `json:"target_bucket" gorm:"not null"`
	TargetPathStyle bool       `json:"target_path_style" gorm:"not null;default:false"`
	TotalObjects    int        `json:"total_objects" gorm:"not null;default:0"`
	CopiedObjects   int        `json:"copied_objects" gorm:"not null;default:0"`
	FailedObjects   int        `json:"failed_objects" gorm:"not null;default:0"`
	CopiedBytes     int64      `json:"copied_bytes" gorm:"not null;default:0"`
	Error           string     `json:"error" gorm:"type:text"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type StorageMigrationItem struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	MigrationID int       `json:"migration_id" gorm:"not null;uniqueIndex:idx_migration_image_kind;index"`
	ImageID     int       `json:"image_id" gorm:"not null;uniqueIndex:idx_migration_image_kind;index"`
	Kind        string    `json:"kind" gorm:"not null;uniqueIndex:idx_migration_image_kind"`
	ObjectKey   string    `json:"object_key" gorm:"not null"`
	Status      string    `json:"status" gorm:"not null;index"`
	Attempts    int       `json:"attempts" gorm:"not null;default:0"`
	Size        int64     `json:"size" gorm:"not null;default:0"`
	SHA256      string    `json:"sha256,omitempty"`
	Error       string    `json:"error,omitempty" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
