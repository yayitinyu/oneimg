package controllers

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"oneimg/backend/database"
	"oneimg/backend/models"
)

var lifecycleWorkerOnce sync.Once

var imageLifetimes = map[string]time.Duration{
	"1h":  time.Hour,
	"1d":  24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
	"90d": 90 * 24 * time.Hour,
}

func parseImageExpiration(value string, now time.Time) (*time.Time, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "never" || value == "permanent" {
		return nil, nil
	}
	duration, ok := imageLifetimes[value]
	if !ok {
		return nil, fmt.Errorf("不支持的图片生命周期")
	}
	expiresAt := now.Add(duration).UTC()
	return &expiresAt, nil
}

func StartImageLifecycleWorker() {
	lifecycleWorkerOnce.Do(func() {
		go func() {
			purgeExpiredImages(time.Now())
			ticker := time.NewTicker(15 * time.Minute)
			defer ticker.Stop()
			for now := range ticker.C {
				purgeExpiredImages(now)
			}
		}()
	})
}

func purgeExpiredImages(now time.Time) int {
	storageMutationMu.RLock()
	defer storageMutationMu.RUnlock()
	if active, err := hasActiveStorageMigration(database.GetDB()); err == nil && active {
		return 0
	}

	db := database.GetDB()
	if db == nil || db.DB == nil {
		return 0
	}
	var expired []models.Image
	if err := db.DB.Where("expires_at IS NOT NULL AND expires_at <= ?", now).Find(&expired).Error; err != nil {
		log.Printf("清理过期图片查询失败: %v", err)
		return 0
	}

	removed := 0
	for _, image := range expired {
		if !DeleteImageFile(image) {
			log.Printf("过期图片物理删除失败，将在下次重试: %s", image.FileName)
			continue
		}
		if err := db.DB.Unscoped().Delete(&image).Error; err != nil {
			log.Printf("删除过期图片记录失败: %v", err)
			continue
		}
		removed++
	}
	return removed
}
