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

	// 支持环境变量覆盖 Turnstile 配置
	// 环境变量优先级高于数据库设置
	if siteKey := os.Getenv("TURNSTILE_SITE_KEY"); siteKey != "" {
		settings.TurnstileSiteKey = siteKey
	}
	if secretKey := os.Getenv("TURNSTILE_SECRET_KEY"); secretKey != "" {
		settings.TurnstileSecretKey = secretKey
	}
	// 如果环境变量设置了密钥，自动启用 Turnstile
	if settings.TurnstileSiteKey != "" && settings.TurnstileSecretKey != "" {
		// 仅当两个密钥都配置时才考虑启用
		// 但不强制覆盖数据库的 turnstile 开关设置
	}
	// 支持通过环境变量禁用 Turnstile（紧急情况）
	if os.Getenv("TURNSTILE_DISABLED") == "true" {
		settings.Turnstile = false
	}

	return settings, err
}
