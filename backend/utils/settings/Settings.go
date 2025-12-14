package settings

import (
	"os"

	"oneimg/backend/database"
	"oneimg/backend/models"
)

func GetSettings() (models.Settings, error) {
	// 获取数据库实例
	db := database.GetDB()
	if db == nil {
		return models.Settings{}, nil
	}
	var settings models.Settings
	err := db.DB.First(&settings).Error

	// 环境变量作为后备：仅当数据库中的值为空时才使用环境变量
	// 这样用户可以在 UI 中删除密钥，不会被环境变量覆盖
	if settings.TurnstileSiteKey == "" {
		if siteKey := os.Getenv("TURNSTILE_SITE_KEY"); siteKey != "" {
			settings.TurnstileSiteKey = siteKey
		}
	}
	if settings.TurnstileSecretKey == "" {
		if secretKey := os.Getenv("TURNSTILE_SECRET_KEY"); secretKey != "" {
			settings.TurnstileSecretKey = secretKey
		}
	}

	// 支持通过环境变量禁用 Turnstile（紧急情况，优先级最高）
	if os.Getenv("TURNSTILE_DISABLED") == "true" {
		settings.Turnstile = false
	}

	return settings, err
}

