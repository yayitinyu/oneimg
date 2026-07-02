package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"oneimg/backend/config"
	"oneimg/backend/database"
	"oneimg/backend/models"
	"oneimg/backend/utils/md5"
	"oneimg/backend/utils/settings"
	"oneimg/backend/utils/telegram"

	"github.com/gin-gonic/gin"
)

// TelegramUpdate Telegram Webhook 更新消息结构
type TelegramUpdate struct {
	UpdateID int `json:"update_id"`
	Message  *struct {
		MessageID int `json:"message_id"`
		From      *struct {
			ID        int64  `json:"id"`
			IsBot     bool   `json:"is_bot"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"from"`
		Chat *struct {
			ID   int64  `json:"id"`
			Type string `json:"type"`
		} `json:"chat"`
		Date int    `json:"date"`
		Text string `json:"text"`
	} `json:"message"`
}

// TelegramWebhook 处理 Telegram Bot 的 Webhook 消息
// 支持通过发送图片直链 URL 来上传图片
func TelegramWebhook(c *gin.Context) {
	setting, err := settings.GetSettings()
	if err != nil {
		log.Printf("Telegram Webhook: 获取配置失败: %v", err)
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	if !setting.TGWebhook || !telegram.ValidateWebhookSecret(setting.TGBotToken, c.GetHeader("X-Telegram-Bot-Api-Secret-Token")) {
		log.Printf("Telegram Webhook: 拒绝未认证请求")
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 解析 Telegram 更新消息
	var update TelegramUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		log.Printf("Telegram Webhook: 解析请求失败: %v", err)
		c.JSON(http.StatusOK, gin.H{"ok": true}) // 始终返回 200 避免 Telegram 重试
		return
	}

	// 忽略非文本消息
	if update.Message == nil || update.Message.Chat == nil || update.Message.Text == "" {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	text := strings.TrimSpace(update.Message.Text)
	chatID := update.Message.Chat.ID

	// 验证是否是授权的 Chat ID
	if !isAuthorizedChatID(setting.TGReceivers, chatID) {
		log.Printf("Telegram Webhook: 未授权的 Chat ID: %d", chatID)
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 检查是否是 URL
	if !strings.HasPrefix(text, "http://") && !strings.HasPrefix(text, "https://") {
		sendTelegramReply(setting.TGBotToken, chatID, "💡 发送图片直链 URL 即可上传图片\n支持格式: http:// 或 https:// 开头的图片链接")
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 发送处理中提示
	sendTelegramReply(setting.TGBotToken, chatID, "⏳ 正在下载并上传图片...")

	cfg := config.App
	maxSize := setting.MaxFileSize
	if maxSize <= 0 {
		maxSize = cfg.MaxFileSize
	}
	remote, err := downloadRemoteImage(c.Request.Context(), text, maxSize, cfg.AllowedTypes)
	if errors.Is(err, errRemoteImageTooLarge) {
		sendTelegramReply(setting.TGBotToken, chatID, fmt.Sprintf("❌ 图片大小超过限制 (最大 %d MB)", maxSize/1024/1024))
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	if err != nil {
		log.Printf("Telegram Webhook: 下载图片失败: %v", err)
		sendTelegramReply(setting.TGBotToken, chatID, "❌ 无法下载有效的图片，请检查 URL、格式和网络可达性")
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 从 URL 中提取文件名
	filename := path.Base(remote.URL.Path)
	if filename == "" || filename == "/" || filename == "." {
		filename = fmt.Sprintf("tg_upload_%d", time.Now().UnixMilli())
	}

	// 创建虚拟的 multipart.FileHeader
	fileHeader, err := createFileHeader(filename, remote.ContentType, remote.Data)
	if err != nil {
		sendTelegramReply(setting.TGBotToken, chatID, "❌ 准备图片数据失败")
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 获取存储上传器
	uploader, err := getStorageUploader(&setting)
	if err != nil {
		sendTelegramReply(setting.TGBotToken, chatID, fmt.Sprintf("❌ 存储配置错误: %v", err))
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 执行上传
	effectiveCfg := *cfg
	effectiveCfg.MaxFileSize = maxSize
	fileResult, err := uploader.Upload(c, &effectiveCfg, &setting, fileHeader)
	if err != nil {
		sendTelegramReply(setting.TGBotToken, chatID, fmt.Sprintf("❌ 上传失败: %v", err))
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 保存到数据库
	username := "TelegramBot"
	if update.Message.From != nil && update.Message.From.Username != "" {
		username = update.Message.From.Username
	}

	imageModel := models.Image{
		Url:       fileResult.URL,
		Thumbnail: fileResult.ThumbnailURL,
		FileName:  fileResult.FileName,
		FileSize:  fileResult.FileSize,
		MimeType:  fileResult.MimeType,
		Width:     fileResult.Width,
		Height:    fileResult.Height,
		Storage:   fileResult.Storage,
		UserId:    0, // Telegram 用户没有关联的系统用户 ID
		MD5:       md5.Md5(username + fileResult.FileName),
		UUID:      "",
	}

	db := database.GetDB()
	if db == nil || db.DB == nil {
		DeleteImageFile(imageModel)
		sendTelegramReply(setting.TGBotToken, chatID, "❌ 数据库连接失败，已回滚上传文件")
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	if err := db.DB.Create(&imageModel).Error; err != nil {
		DeleteImageFile(imageModel)
		log.Printf("Telegram Webhook: 保存图片记录失败: %v", err)
		sendTelegramReply(setting.TGBotToken, chatID, "❌ 保存图片记录失败，已回滚上传文件")
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 构建访问URL
	accessURL := formatNotificationURL(c.Request.Host, fileResult.URL)

	// 发送成功消息
	successMsg := fmt.Sprintf("✅ 上传成功！\n\n📁 文件名: %s\n📦 存储: %s\n🔗 链接: %s",
		fileResult.FileName,
		setting.StorageType,
		accessURL,
	)
	sendTelegramReply(setting.TGBotToken, chatID, successMsg)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// isAuthorizedChatID 检查 Chat ID 是否在授权列表中
func isAuthorizedChatID(receivers string, chatID int64) bool {
	chatIDStr := fmt.Sprintf("%d", chatID)
	// 支持多个接收者，用逗号分隔
	for _, r := range strings.Split(receivers, ",") {
		if strings.TrimSpace(r) == chatIDStr {
			return true
		}
	}
	return false
}

// sendTelegramReply 发送 Telegram 回复消息
func sendTelegramReply(botToken string, chatID int64, text string) {
	if botToken == "" {
		return
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Telegram Reply: 创建请求失败: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Telegram Reply: 发送失败: %v", err)
		return
	}
	defer resp.Body.Close()
}
