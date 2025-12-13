package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"oneimg/backend/config"
	"oneimg/backend/database"
	"oneimg/backend/models"
	"oneimg/backend/utils/md5"
	"oneimg/backend/utils/settings"

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
	// 解析 Telegram 更新消息
	var update TelegramUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		log.Printf("Telegram Webhook: 解析请求失败: %v", err)
		c.JSON(http.StatusOK, gin.H{"ok": true}) // 始终返回 200 避免 Telegram 重试
		return
	}

	// 忽略非文本消息
	if update.Message == nil || update.Message.Text == "" {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	text := strings.TrimSpace(update.Message.Text)
	chatID := update.Message.Chat.ID

	// 获取系统配置
	setting, err := settings.GetSettings()
	if err != nil {
		log.Printf("Telegram Webhook: 获取配置失败: %v", err)
		sendTelegramReply(setting.TGBotToken, chatID, "❌ 系统配置错误")
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

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

	// 验证 URL 格式
	parsedURL, err := url.Parse(text)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		sendTelegramReply(setting.TGBotToken, chatID, "❌ URL 格式无效")
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 发送处理中提示
	sendTelegramReply(setting.TGBotToken, chatID, "⏳ 正在下载并上传图片...")

	// 下载图片
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(text)
	if err != nil {
		sendTelegramReply(setting.TGBotToken, chatID, fmt.Sprintf("❌ 下载图片失败: %v", err))
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		sendTelegramReply(setting.TGBotToken, chatID, fmt.Sprintf("❌ 下载图片失败，状态码: %d", resp.StatusCode))
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 检查 Content-Type 是否为图片
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		sendTelegramReply(setting.TGBotToken, chatID, "❌ URL 不是有效的图片资源")
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 读取图片内容
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		sendTelegramReply(setting.TGBotToken, chatID, "❌ 读取图片数据失败")
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 获取全局配置检查文件大小
	cfg := config.App
	if int64(len(imageData)) > cfg.MaxFileSize {
		sendTelegramReply(setting.TGBotToken, chatID, fmt.Sprintf("❌ 图片大小超过限制 (最大 %d MB)", cfg.MaxFileSize/1024/1024))
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 从 URL 中提取文件名
	filename := path.Base(parsedURL.Path)
	if filename == "" || filename == "/" || filename == "." {
		filename = fmt.Sprintf("tg_upload_%d", time.Now().UnixMilli())
	}

	// 创建虚拟的 multipart.FileHeader
	fileHeader := createTelegramFileHeader(filename, contentType, imageData)

	// 获取存储上传器
	uploader, err := getStorageUploader(&setting)
	if err != nil {
		sendTelegramReply(setting.TGBotToken, chatID, fmt.Sprintf("❌ 存储配置错误: %v", err))
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 执行上传
	fileResult, err := uploader.Upload(c, cfg, &setting, fileHeader)
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
	if db != nil {
		db.DB.Create(&imageModel)
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

// createTelegramFileHeader 为 Telegram 上传创建虚拟的 multipart.FileHeader
func createTelegramFileHeader(filename, contentType string, data []byte) *multipart.FileHeader {
	// 创建一个内存中的 multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建 form file
	part, _ := writer.CreateFormFile("file", filename)
	part.Write(data)
	writer.Close()

	// 解析 form 获取 FileHeader
	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(32 << 20)

	if files, ok := form.File["file"]; ok && len(files) > 0 {
		files[0].Header.Set("Content-Type", contentType)
		return files[0]
	}

	// 降级方案：手动构造
	return &multipart.FileHeader{
		Filename: filename,
		Size:     int64(len(data)),
		Header:   make(map[string][]string),
	}
}
