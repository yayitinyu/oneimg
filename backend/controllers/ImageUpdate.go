package controllers

import (
	"errors"
	"log"
	"oneimg/backend/config"
	"oneimg/backend/database"
	"oneimg/backend/interfaces"
	"oneimg/backend/models"
	"oneimg/backend/utils/md5"
	"oneimg/backend/utils/settings"
	"oneimg/backend/utils/telegram"
	"oneimg/backend/utils/uploads"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadImages 图片上传主入口
func UploadImages(c *gin.Context) {
	// 初始化上传上下文
	uc := uploads.NewUploadContext(c)

	// 获取全局配置
	cfg, ok := c.MustGet("config").(*config.Config)
	if !ok {
		uc.Fail(500, "全局配置获取失败")
		return
	}

	// 获取系统配置
	setting, err := settings.GetSettings()
	if err != nil {
		uc.Fail(500, "获取上传配置失败：%v", err)
		return
	}

	expiresAt, err := parseImageExpiration(c.PostForm("expires_in"), time.Now())
	if err != nil {
		uc.Fail(400, "%s", err.Error())
		return
	}
	if rawSaveWebP := c.PostForm("save_webp"); rawSaveWebP != "" {
		saveWebP, parseErr := strconv.ParseBool(rawSaveWebP)
		if parseErr != nil {
			uc.Fail(400, "WebP 选项无效")
			return
		}
		setting.SaveWebp = saveWebP
	}

	// 解析并校验上传文件
	// 优先使用数据库配置的最大文件大小
	maxSize := setting.MaxFileSize
	if maxSize <= 0 {
		maxSize = cfg.MaxFileSize
	}

	files, err := uc.ParseAndValidateFiles(maxSize)
	if err != nil {
		uc.Fail(400, "文件解析失败: %v", err)
		return
	}
	effectiveCfg := *cfg
	effectiveCfg.MaxFileSize = maxSize

	// 获取存储上传器
	uploader, err := uc.GetStorageUploader(&setting)
	if err != nil {
		uc.Fail(400, "%s", err.Error())
		return
	}

	// 先完成存储，再在同一事务中校验配额并写入全部记录，避免批量
	// 上传只落库一部分。
	pendingImages := make([]*models.Image, 0, len(files))
	pendingResults := make([]*interfaces.ImageUploadResult, 0, len(files))

	for _, file := range files {
		fileResult, err := uploader.Upload(c, &effectiveCfg, &setting, file)
		if err != nil {
			cleanupUploadedImages(pendingImages)
			uc.Fail(500, "文件[%s]上传失败：%v", file.Filename, err)
			return
		}

		isHidden := c.Query("hidden") == "true"
		imageModel := &models.Image{
			Url:       fileResult.URL,
			Thumbnail: fileResult.ThumbnailURL,
			FileName:  fileResult.FileName,
			FileSize:  fileResult.FileSize,
			MimeType:  fileResult.MimeType,
			Width:     fileResult.Width,
			Height:    fileResult.Height,
			Storage:   fileResult.Storage,
			UserId:    c.GetInt("user_id"),
			MD5:       md5.Md5(c.GetString("username") + fileResult.FileName),
			UUID:      GetUUID(c),
			Hidden:    isHidden,
			ExpiresAt: expiresAt,
		}
		pendingImages = append(pendingImages, imageModel)
		pendingResults = append(pendingResults, fileResult)
	}

	db := database.GetDB()
	if db == nil || db.DB == nil {
		cleanupUploadedImages(pendingImages)
		uc.Fail(500, "数据库连接失败，已回滚上传文件")
		return
	}
	if err := persistUploadedImages(db.DB, c.GetInt("user_id"), c.GetInt("user_role"), setting.UserStorageQuota, pendingImages); err != nil {
		cleanupUploadedImages(pendingImages)
		if errors.Is(err, errUserStorageQuotaExceeded) {
			uc.Fail(413, "%s", storageQuotaMessage(setting.UserStorageQuota))
			return
		}
		log.Printf("保存图片记录失败: %v", err)
		uc.Fail(500, "保存图片记录失败，已回滚上传文件")
		return
	}

	uploadResults := make([]interfaces.ImageUploadResult, 0, len(pendingResults))
	for index, fileResult := range pendingResults {
		fileResult.ID = pendingImages[index].Id
		fileResult.ExpiresAt = expiresAt
		uploadResults = append(uploadResults, *fileResult)

		if setting.TGNotice {
			placeholderData := telegram.PlaceholderData{
				Username:    c.GetString("username"),
				Date:        time.Now().Format("2006-01-02 15:04:05"),
				Filename:    fileResult.FileName,
				StorageType: setting.StorageType,
				URL:         formatNotificationURL(c.Request.Host, fileResult.URL),
			}

			err := telegram.SendSimpleMsg(
				setting.TGBotToken,   // 机器人Token
				setting.TGReceivers,  // 接收者ChatID
				setting.TGNoticeText, // 模板文本
				placeholderData,      // 占位符数据
			)
			if err != nil {
				log.Println(err)
				// 忽略错误
			}
		}
	}

	// 返回上传结果
	uc.Success("上传成功", map[string]any{
		"files": uploadResults,
		"count": len(uploadResults),
	})
}

// UploadImage 单文件上传
func UploadImage(c *gin.Context) {
	UploadImages(c)
}
