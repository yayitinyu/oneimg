package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"oneimg/backend/database"
	"oneimg/backend/models"
	"oneimg/backend/utils/result"
	"oneimg/backend/utils/settings"
	"oneimg/backend/utils/telegram"
)

// 定义请求参数
type UpdateSettingsRequest struct {
	Key   string `json:"key" binding:"required"`
	Value any    `json:"value"`
}

// 自定义查询参数
type GetSettingsRequest struct {
	Keys []string `json:"keys"`
}

func GetSettings(c *gin.Context) {
	settings, err := settings.GetSettings()
	if err != nil {
		c.JSON(500, result.Error(500, "获取设置失败"))
		return
	}

	// Simplify: directly return all settings to ensure no data loss in filtering
	// log.Printf("Returning Settings: %+v", settings)
	c.JSON(200, result.Success("ok", settings))
}

// 返回登录配置
func GetLoginSettings(c *gin.Context) {
	settings, err := settings.GetSettings()
	if err != nil {
		c.JSON(500, result.Error(500, "获取设置失败"))
		return
	}

	c.JSON(200, result.Success("ok",
		map[string]any{
			"turnstile":          settings.Turnstile,
			"turnstile_site_key": settings.TurnstileSiteKey,
			"tourist":            settings.Tourist,
			"registration_mode":  normalizedRegistrationMode(settings.RegistrationMode),
			"save_webp":          settings.SaveWebp,
			"site_logo":          settings.SiteLogo,
		},
	))
}

func UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, result.Error(400, "请求参数错误: "+err.Error()))
		return
	}
	// 查询是否有该设置项
	currentSettings, err := settings.GetSettings()
	if err != nil {
		c.JSON(500, result.Error(500, "获取设置失败"))
		return
	}

	// 校验设置数据
	if err := validateSettingData(req.Key, req.Value); err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, err.Error()))
		return
	}

	// 如果是更新 logo，尝试删除旧 logo
	if req.Key == "site_logo" {
		oldLogo := currentSettings.SiteLogo
		newLogo, ok := req.Value.(string)
		// 如果 oldLogo 存在，且 newLogo 有效（或者是空字符串用于清除），且两者不相等
		if oldLogo != "" && ok && oldLogo != newLogo {
			db := database.GetDB().DB
			var oldImage models.Image
			// 查找旧图片记录 (包括隐藏的)
			if err := db.Unscoped().Where("url = ?", oldLogo).First(&oldImage).Error; err == nil {
				DeleteImageFile(oldImage)
				db.Unscoped().Delete(&oldImage)
			}
		}
	}

	if err := updateSettingsField(&currentSettings, req.Key, req.Value); err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, err.Error()))
		return
	}

	// 更新设置项
	db := database.GetDB().DB

	// 使用 Save 而不是 Update，避免 JSON unmarshal 带来的类型问题 (如 float64 vs int)
	// currentSettings 已经被 updateSettingsField 正确更新了类型
	if err := db.Save(&currentSettings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "更新失败"))
		log.Println("UpdateSettings Error:", err)
		return
	}

	// 自动设置/删除 Telegram Webhook
	if req.Key == "tg_webhook" {
		go handleTelegramWebhookUpdate(&currentSettings)
	}

	c.JSON(200, result.Success("更新成功", nil))
}

// 辅助函数，筛选设置项
func filterSettings(settings *models.Settings, keys []string) *models.Settings {
	if len(keys) == 0 {
		return settings
	}

	filteredSettings := &models.Settings{}
	srcVal := reflect.ValueOf(settings).Elem()
	dstVal := reflect.ValueOf(filteredSettings).Elem()
	srcTyp := srcVal.Type()
	for i := 0; i < srcTyp.NumField(); i++ {
		srcField := srcTyp.Field(i)
		srcFieldVal := srcVal.Field(i)
		jsonTag := srcField.Tag.Get("json")
		if jsonTag == "" {
			continue
		}
		for _, key := range keys {
			if jsonTag == key {
				dstField := dstVal.FieldByName(srcField.Name)
				if dstField.IsValid() && dstField.CanSet() {
					dstField.Set(srcFieldVal)
				}
				break
			}
		}
	}
	return filteredSettings
}

func updateSettingsField(settings *models.Settings, key string, value any) error {
	// 获取结构体反射值（指针解引用）
	val := reflect.ValueOf(settings).Elem()
	typ := val.Type()

	// 1. 遍历结构体字段，匹配JSON Tag或字段名
	var targetField reflect.Value
	var fieldType reflect.Type
	found := false

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		// 优先匹配JSON Tag（如 json:"tourist"）
		jsonTag := field.Tag.Get("json")
		if jsonTag == key || field.Name == key {
			targetField = val.Field(i)
			fieldType = field.Type
			found = true
			break
		}
	}

	// 校验字段是否存在
	if !found {
		return fmt.Errorf("设置项 %s 不存在", key)
	}

	// 2. 校验字段是否可修改（必须是导出字段）
	if !targetField.CanSet() {
		return fmt.Errorf("设置项 %s 不可修改", key)
	}

	// 3. 处理nil值（避免panic）
	if value == nil {
		return fmt.Errorf("设置项 %s 的值不能为空", key)
	}

	// 4. 转换value类型为字段实际类型
	convertedValue, err := convertValueToTargetType(key, value, fieldType)
	if err != nil {
		return err
	}

	valueVal := reflect.ValueOf(convertedValue)

	// 5. 设置字段值
	targetField.Set(valueVal)
	return nil
}

func convertValueToTargetType(key string, value any, targetType reflect.Type) (any, error) {
	valueVal := reflect.ValueOf(value)
	valueType := valueVal.Type()

	// 类型已匹配，直接返回
	if valueType == targetType {
		return value, nil
	}

	// 场景1：反射支持直接转换（如 int→float64、bool→int 等）
	if valueType.ConvertibleTo(targetType) {
		return valueVal.Convert(targetType).Interface(), nil
	}

	// 场景2：反射不支持直接转换，手动处理常见类型解析
	switch targetType.Kind() {
	// 处理 string → float64
	case reflect.Float64:
		if valueType.Kind() == reflect.String {
			strVal := valueVal.String()
			floatVal, err := strconv.ParseFloat(strVal, 64)
			if err != nil {
				return nil, fmt.Errorf("设置项 %s 类型转换失败，期望 float64，实际 string（值：%s），错误：%v",
					key, strVal, err)
			}
			return floatVal, nil
		}

	// 处理 string → int/int64
	case reflect.Int:
		if valueType.Kind() == reflect.String {
			strVal := valueVal.String()
			intVal, err := strconv.Atoi(strVal)
			if err != nil {
				return nil, fmt.Errorf("设置项 %s 类型转换失败，期望 int，实际 string（值：%s），错误：%v",
					key, strVal, err)
			}
			return intVal, nil
		}
	case reflect.Int64:
		if valueType.Kind() == reflect.String {
			strVal := valueVal.String()
			int64Val, err := strconv.ParseInt(strVal, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("设置项 %s 类型转换失败，期望 int64，实际 string（值：%s），错误：%v",
					key, strVal, err)
			}
			return int64Val, nil
		}

	// 处理 string → bool
	case reflect.Bool:
		if valueType.Kind() == reflect.String {
			strVal := valueVal.String()
			boolVal, err := strconv.ParseBool(strVal)
			if err != nil {
				return nil, fmt.Errorf("设置项 %s 类型转换失败，期望 bool，实际 string（值：%s），错误：%v",
					key, strVal, err)
			}
			return boolVal, nil
		}
	case reflect.String:
		// 所有基础类型都可以转为string
		str := fmt.Sprintf("%v", value)
		return strings.TrimSpace(str), nil
	}

	// 不支持的转换类型
	return nil, fmt.Errorf("设置项 %s 类型不匹配，期望 %s，实际 %T",
		key, targetType, value)
}

func validateSettingData(key string, value any) error {
	switch key {
	case "site_logo":
		logo, ok := value.(string)
		if !ok {
			return fmt.Errorf("站点 Logo 必须是字符串")
		}
		if len(logo) > 255 {
			return fmt.Errorf("站点 Logo URL 长度不能超过 255 个字符")
		}
	case "max_file_size":
		size, err := numberToInt64(value)
		if err != nil || size < 1 {
			return errors.New("最大文件大小必须大于 0")
		}
	case "r2_endpoint", "r2_access_key", "r2_secret_key", "r2_bucket":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s 必须是字符串", key)
		}
	case "webp_quality":
		quality, err := numberToInt64(value)
		if err != nil || quality < 1 || quality > 100 {
			return errors.New("WebP 质量必须在 1–100 之间")
		}
	case "registration_mode":
		mode, ok := value.(string)
		if !ok {
			return errors.New("注册方式必须是字符串")
		}
		switch normalizedRegistrationMode(mode) {
		case models.RegistrationOpen, models.RegistrationInvite, models.RegistrationClosed:
		default:
			return errors.New("注册方式无效")
		}
	}
	return nil
}

func numberToInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	default:
		return 0, fmt.Errorf("unsupported number type %T", value)
	}
}

func normalizedRegistrationMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		return models.RegistrationOpen
	}
	return mode
}

// handleTelegramWebhookUpdate 处理 Telegram Webhook 自动设置/删除
func handleTelegramWebhookUpdate(s *models.Settings) {
	if s.TGWebhook {
		// 启用 Webhook：校验必要配置后自动设置
		if s.TGBotToken == "" {
			log.Println("Telegram Webhook: Bot Token 未配置，跳过设置")
			return
		}
		if s.SiteDomain == "" {
			log.Println("Telegram Webhook: 网站域名未配置，跳过设置")
			return
		}

		err := telegram.SetWebhook(s.TGBotToken, s.SiteDomain, "/api/telegram/webhook")
		if err != nil {
			log.Printf("Telegram Webhook 设置失败: %v", err)
		} else {
			log.Printf("Telegram Webhook 设置成功: https://%s/api/telegram/webhook", s.SiteDomain)
		}
	} else {
		// 禁用 Webhook：删除已设置的 Webhook
		if s.TGBotToken == "" {
			return
		}

		err := telegram.DeleteWebhook(s.TGBotToken)
		if err != nil {
			log.Printf("Telegram Webhook 删除失败: %v", err)
		} else {
			log.Println("Telegram Webhook 已删除")
		}
	}
}
