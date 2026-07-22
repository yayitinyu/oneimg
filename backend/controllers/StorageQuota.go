package controllers

import (
	"errors"
	"fmt"

	"oneimg/backend/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errUserStorageQuotaExceeded = errors.New("user storage quota exceeded")

// persistUploadedImages creates image records atomically. Locking the owning
// user serializes quota checks for concurrent uploads on databases that support
// row locks; SQLite already serializes writes at the database level.
func persistUploadedImages(db *gorm.DB, userID, role int, quota int64, images []*models.Image) error {
	if db == nil {
		return errors.New("database is unavailable")
	}
	if len(images) == 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if role == models.RoleUser && quota > 0 {
			var user models.User
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&user, userID).Error; err != nil {
				return err
			}

			used, err := userStorageUsage(tx, userID)
			if err != nil {
				return err
			}
			var incoming int64
			for _, image := range images {
				if image == nil || image.FileSize < 0 || image.FileSize > quota-incoming {
					return errUserStorageQuotaExceeded
				}
				incoming += image.FileSize
			}
			if storageQuotaExceeded(used, incoming, quota) {
				return errUserStorageQuotaExceeded
			}
		}

		return tx.Create(&images).Error
	})
}

func userStorageUsage(db *gorm.DB, userID int) (int64, error) {
	var used int64
	err := db.Model(&models.Image{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&used).Error
	return used, err
}

func storageQuotaExceeded(used, incoming, quota int64) bool {
	if used < 0 || incoming < 0 {
		return true
	}
	if quota <= 0 {
		return false
	}
	if used >= quota {
		return incoming > 0 || used > quota
	}
	return incoming > quota-used
}

func storageQuotaMessage(quota int64) string {
	return fmt.Sprintf("普通用户存储空间不足（每个账号上限 %s）", formatStorageSize(quota))
}

func formatStorageSize(size int64) string {
	const (
		mb = int64(1024 * 1024)
		gb = int64(1024 * 1024 * 1024)
	)
	if size >= gb && size%gb == 0 {
		return fmt.Sprintf("%d GB", size/gb)
	}
	if size >= mb {
		return fmt.Sprintf("%.1f GB", float64(size)/float64(gb))
	}
	return fmt.Sprintf("%d MB", size/mb)
}

func cleanupUploadedImages(images []*models.Image) {
	for _, image := range images {
		if image != nil {
			DeleteImageFile(*image)
		}
	}
}
