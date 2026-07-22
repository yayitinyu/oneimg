package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"time"

	"oneimg/backend/config"
	"oneimg/backend/database"
	"oneimg/backend/interfaces"
	"oneimg/backend/models"
	"oneimg/backend/utils/md5"
	"oneimg/backend/utils/result"
	"oneimg/backend/utils/settings"
	"oneimg/backend/utils/telegram"
	"oneimg/backend/utils/uploads"

	"github.com/gin-gonic/gin"
)

// UploadURLRequest URL上传请求
type UploadURLRequest struct {
	URL       string `json:"url" binding:"required"`
	ExpiresIn string `json:"expires_in"`
	SaveWebp  *bool  `json:"save_webp"`
}

// UploadImageByURL 通过URL上传图片
func UploadImageByURL(c *gin.Context) {
	var req UploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, "请提供有效的图片URL"))
		return
	}

	// 获取系统配置
	setting, err := settings.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "获取上传配置失败"))
		return
	}
	expiresAt, err := parseImageExpiration(req.ExpiresIn, time.Now())
	if err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, err.Error()))
		return
	}
	if req.SaveWebp != nil {
		setting.SaveWebp = *req.SaveWebp
	}

	// 获取全局配置
	cfg, ok := c.MustGet("config").(*config.Config)
	if !ok {
		c.JSON(http.StatusInternalServerError, result.Error(500, "获取全局配置失败"))
		return
	}

	// 下载时即限制响应体大小，并拒绝内网、环回和链路本地地址，避免 SSRF 与内存耗尽。
	maxSize := setting.MaxFileSize
	if maxSize <= 0 {
		maxSize = cfg.MaxFileSize
	}
	remote, err := downloadRemoteImage(c.Request.Context(), req.URL, maxSize, cfg.AllowedTypes)
	if errors.Is(err, errRemoteImageTooLarge) {
		c.JSON(http.StatusBadRequest, result.Error(400, fmt.Sprintf("图片大小超过限制 (最大 %d MB)", maxSize/1024/1024)))
		return
	}
	if err != nil {
		log.Printf("URL 图片下载失败: %v", err)
		c.JSON(http.StatusBadRequest, result.Error(400, "无法下载有效的图片，请检查 URL、格式和网络可达性"))
		return
	}

	// 从URL中提取文件名
	filename := path.Base(remote.URL.Path)
	if filename == "" || filename == "/" || filename == "." {
		filename = fmt.Sprintf("url_image_%d", time.Now().UnixMilli())
	}

	// 创建一个虚拟的 multipart.FileHeader
	fileHeader, err := createFileHeader(filename, remote.ContentType, remote.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "准备图片数据失败"))
		return
	}

	// 获取存储上传器
	uploader, err := getStorageUploader(&setting)
	if err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, err.Error()))
		return
	}

	// 执行上传
	effectiveCfg := *cfg
	effectiveCfg.MaxFileSize = maxSize
	fileResult, err := uploader.Upload(c, &effectiveCfg, &setting, fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "上传失败: "+err.Error()))
		return
	}

	// 保存到数据库
	imageModel := models.Image{
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
		ExpiresAt: expiresAt,
	}

	db := database.GetDB()
	if db == nil || db.DB == nil {
		DeleteImageFile(imageModel)
		c.JSON(http.StatusInternalServerError, result.Error(500, "数据库连接失败，已回滚上传文件"))
		return
	}
	if err := persistUploadedImages(db.DB, c.GetInt("user_id"), c.GetInt("user_role"), setting.UserStorageQuota, []*models.Image{&imageModel}); err != nil {
		DeleteImageFile(imageModel)
		if errors.Is(err, errUserStorageQuotaExceeded) {
			c.JSON(http.StatusRequestEntityTooLarge, result.Error(413, storageQuotaMessage(setting.UserStorageQuota)))
			return
		}
		log.Printf("保存 URL 上传记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, result.Error(500, "保存图片记录失败，已回滚上传文件"))
		return
	}
	fileResult.ID = imageModel.Id
	fileResult.ExpiresAt = expiresAt

	// TG通知
	if setting.TGNotice {
		placeholderData := telegram.PlaceholderData{
			Username:    c.GetString("username"),
			Date:        time.Now().Format("2006-01-02 15:04:05"),
			Filename:    fileResult.FileName,
			StorageType: setting.StorageType,
			URL:         formatNotificationURL(c.Request.Host, fileResult.URL),
		}

		err := telegram.SendSimpleMsg(
			setting.TGBotToken,
			setting.TGReceivers,
			setting.TGNoticeText,
			placeholderData,
		)
		if err != nil {
			log.Println(err)
		}
	}

	// 返回结果
	c.JSON(http.StatusOK, result.Success("上传成功", map[string]any{
		"files": []interfaces.ImageUploadResult{*fileResult},
		"count": 1,
	}))
}

// createFileHeader 创建虚拟的 multipart.FileHeader
func createFileHeader(filename, contentType string, data []byte) (*multipart.FileHeader, error) {
	// 创建一个内存中的multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建form file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(data); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	// 解析form获取FileHeader
	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(int64(len(data)) + 1024)
	if err != nil {
		return nil, err
	}

	if files, ok := form.File["file"]; ok && len(files) > 0 {
		files[0].Header.Set("Content-Type", contentType)
		return files[0], nil
	}

	return nil, errors.New("multipart image file is missing")
}

// getStorageUploader 获取存储上传器
func getStorageUploader(setting *models.Settings) (interfaces.StorageUploader, error) {
	storageType := strings.ToLower(setting.StorageType)
	switch storageType {
	case "s3", "r2":
		return &uploads.S3R2Uploader{}, nil
	case "webdav":
		return &uploads.WebDAVUploader{}, nil
	case "ftp":
		return &uploads.FTPUploader{}, nil
	case "telegram":
		return &uploads.TelegramUploader{}, nil
	case "default", "":
		return &uploads.DefaultUploader{}, nil

	default:
		return nil, fmt.Errorf("不支持的存储类型: %s", storageType)
	}
}
