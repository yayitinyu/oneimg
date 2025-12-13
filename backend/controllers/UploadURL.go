package controllers

import (
	"bytes"
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
	URL string `json:"url" binding:"required"`
}

// UploadImageByURL 通过URL上传图片
func UploadImageByURL(c *gin.Context) {
	var req UploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, "请提供有效的图片URL"))
		return
	}

	// 验证URL格式
	parsedURL, err := url.Parse(req.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		c.JSON(http.StatusBadRequest, result.Error(400, "URL格式无效，仅支持http和https"))
		return
	}

	// 获取系统配置
	setting, err := settings.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "获取上传配置失败"))
		return
	}

	// 获取全局配置
	cfg, ok := c.MustGet("config").(*config.Config)
	if !ok {
		c.JSON(http.StatusInternalServerError, result.Error(500, "获取全局配置失败"))
		return
	}

	// 下载远程图片
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, "无法下载图片: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadRequest, result.Error(400, fmt.Sprintf("下载图片失败，状态码: %d", resp.StatusCode)))
		return
	}

	// 检查Content-Type是否为图片
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, result.Error(400, "URL不是有效的图片资源"))
		return
	}

	// 读取图片内容
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "读取图片数据失败"))
		return
	}

	// 检查文件大小
	if int64(len(imageData)) > cfg.MaxFileSize {
		c.JSON(http.StatusBadRequest, result.Error(400, fmt.Sprintf("图片大小超过限制 (最大 %d MB)", cfg.MaxFileSize/1024/1024)))
		return
	}

	// 从URL中提取文件名
	filename := path.Base(parsedURL.Path)
	if filename == "" || filename == "/" || filename == "." {
		filename = fmt.Sprintf("url_image_%d", time.Now().UnixMilli())
	}

	// 创建一个虚拟的 multipart.FileHeader
	fileHeader := createFileHeader(filename, contentType, imageData)

	// 获取存储上传器
	uploader, err := getStorageUploader(&setting)
	if err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, err.Error()))
		return
	}

	// 执行上传
	fileResult, err := uploader.Upload(c, cfg, &setting, fileHeader)
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
	}

	db := database.GetDB()
	if db != nil {
		db.DB.Create(&imageModel)
	}

	// TG通知
	if setting.TGNotice {
		placeholderData := telegram.PlaceholderData{
			Username:    c.GetString("username"),
			Date:        time.Now().Format("2006-01-02 15:04:05"),
			Filename:    fileResult.FileName,
			StorageType: setting.StorageType,
			URL:         c.Request.Host + fileResult.URL,
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
func createFileHeader(filename, contentType string, data []byte) *multipart.FileHeader {
	// 创建一个内存中的multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建form file
	part, _ := writer.CreateFormFile("file", filename)
	part.Write(data)
	writer.Close()

	// 解析form获取FileHeader
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
