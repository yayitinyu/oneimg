package controllers

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"oneimg/backend/database"
	"oneimg/backend/models"
	"oneimg/backend/utils/ftp"
	"oneimg/backend/utils/result"
	"oneimg/backend/utils/s3"
	"oneimg/backend/utils/settings"
	"oneimg/backend/utils/telegram"
	"oneimg/backend/utils/webdav"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// DeleteImage 删除图片
func DeleteImage(c *gin.Context) {
	// 获取图片ID参数
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "图片ID不能为空",
		})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, result.Error(
			400,
			"图片ID无效",
		))
		return
	}

	db := database.GetDB().DB
	var image models.Image

	// 查询图片信息
	if err := db.First(&image, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "图片不存在",
		})
		return
	}

	// 校验权限
	if !CheckImageAccessPermission(c, image) {
		c.JSON(http.StatusForbidden, gin.H{
			"code": 403,
			"msg":  "无权访问",
		})
		return
	}

	// 删除存储文件
	deleteStatus := DeleteImageFile(image)

	// 删除数据库记录
	if err := db.Unscoped().Delete(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "删除图片记录失败",
		})
		return
	}

	if !deleteStatus {
		c.JSON(http.StatusOK, result.Success(
			"记录删除成功,物理删除失败",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, result.Success("删除成功", nil))
}

// DeleteImageRecord 仅删除图片记录（不删除存储文件）
// 用于首页"最近上传"批量删除记录，图片仍保留在画廊
func DeleteImageRecord(c *gin.Context) {
	// 获取图片ID参数
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "图片ID不能为空",
		})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, result.Error(
			400,
			"图片ID无效",
		))
		return
	}

	db := database.GetDB().DB
	var image models.Image

	// 查询图片信息
	if err := db.First(&image, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "图片不存在",
		})
		return
	}

	// 校验权限
	if !CheckImageAccessPermission(c, image) {
		c.JSON(http.StatusForbidden, gin.H{
			"code": 403,
			"msg":  "无权访问",
		})
		return
	}

	// 仅删除数据库记录（软删除/隐藏），不删除存储文件
	if err := db.Model(&image).Update("hidden", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "删除图片记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, result.Success("记录删除成功", nil))
}

// DismissImage 仅从"最近上传"中移除（不删除文件，不隐藏）
func DismissImage(c *gin.Context) {
	// 获取图片ID参数
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "图片ID不能为空",
		})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, result.Error(
			400,
			"图片ID无效",
		))
		return
	}

	db := database.GetDB().DB
	var image models.Image

	// 查询图片信息
	if err := db.First(&image, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "图片不存在",
		})
		return
	}

	// 校验权限
	if !CheckImageAccessPermission(c, image) {
		c.JSON(http.StatusForbidden, gin.H{
			"code": 403,
			"msg":  "无权访问",
		})
		return
	}

	// 仅更新 show_in_recent 字段
	if err := db.Model(&image).Update("show_in_recent", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "移除记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, result.Success("已从最近上传中移除", nil))
}

// 删除默认存储的图片
func DeleteDefaultStorageImage(image models.Image) (deleteStatus bool) {
	deleteFile := func(rawPath string) bool {
		filePath, ok := safeLocalUploadPath(rawPath)
		if !ok {
			return false
		}
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return false
		}
		return true
	}

	deleted := deleteFile(image.Url)
	if image.Thumbnail != "" && !deleteFile(image.Thumbnail) {
		deleted = false
	}
	return deleted
}

func safeLocalUploadPath(rawPath string) (string, bool) {
	normalized := strings.ReplaceAll(rawPath, "\\", "/")
	if !strings.HasPrefix(normalized, "/uploads/") {
		return "", false
	}
	relativePath := filepath.Clean(strings.TrimPrefix(normalized, "/uploads/"))
	if relativePath == "." || filepath.IsAbs(relativePath) || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return "", false
	}

	root, err := filepath.Abs("./uploads")
	if err != nil {
		return "", false
	}
	target, err := filepath.Abs(filepath.Join(root, relativePath))
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return target, true
}

// 删除S3存储的图片
func DeleteS3StorageImage(image models.Image) (deleteStatus bool) {
	// 获取系统配置
	setting, err := settings.GetSettings()
	if err != nil {
		return false
	}
	// 图片可能来自与当前选项不同的 S3 兼容后端，因此按记录中的
	// storage 字段选择对应凭据和 bucket。
	storageSetting := setting
	storageSetting.StorageType = image.Storage
	s3Client, err := s3.NewS3Client(storageSetting)
	if err != nil {
		return false
	}
	objectKey := storageObjectKey(image.Url)
	bucket := storageSetting.S3Bucket
	if image.Storage == "r2" {
		bucket = storageSetting.R2Bucket
	}
	}
	if bucket == "" || objectKey == "" {
		return false
	}

	// 构建删除请求
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, mainDeleteErr := s3Client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})

	// 检查是否存在缩略图
	var thumbnailDeleteErr error
	if image.Thumbnail != "" {
		objectKey = storageObjectKey(image.Thumbnail)
		_, thumbnailDeleteErr = s3Client.DeleteObject(ctx, &awss3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(objectKey),
		})
	}

	if mainDeleteErr != nil || thumbnailDeleteErr != nil {
		return !true
	}

	return true
}

func storageObjectKey(rawURL string) string {
	if parsed, err := url.Parse(rawURL); err == nil && parsed.Path != "" {
		rawURL = parsed.Path
	}
	return strings.TrimPrefix(rawURL, "/")
}

// 删除WebDAV存储的图片
func DeleteWebDavStorageImage(image models.Image) (deleteStatus bool) {
	// 获取系统配置
	setting, err := settings.GetSettings()
	if err != nil {
		return false
	}
	// 获取WebDav客户端
	client := webdav.Client(webdav.Config{
		BaseURL:  setting.WebdavURL,
		Username: setting.WebdavUser,
		Password: setting.WebdavPass,
		Timeout:  30 * time.Second,
	})

	var deleteFile = func(filePath string) bool {
		if filePath == "" {
			return false
		}
		err := client.WebDAVDelete(context.TODO(), filePath)
		if err != nil {
			return !true
		}
		return true
	}

	// 检查是否存在缩略图
	if image.Thumbnail != "" {
		deleteFile(image.Thumbnail)
	}
	return deleteFile(image.Url)
}

// 删除FTP存储的图片
func DeleteFtpStorageImage(image models.Image) (deleteStatus bool) {
	// 获取系统配置
	setting, err := settings.GetSettings()
	if err != nil {
		return false
	}
	// 初始化FTP客户端
	ftpUtil := ftp.NewFTPUtil(ftp.FTPConfig{
		Host:     setting.FTPHost,
		Port:     setting.FTPPort,
		User:     setting.FTPUser,
		Password: setting.FTPPass,
		Timeout:  60,
	})
	defer ftpUtil.Close()

	// 删除图片
	if err := ftpUtil.DeleteImage(image.Url); err != nil {
		return !true
	}

	// 检查是否存在缩略图
	if image.Thumbnail != "" {
		// 删除缩略图
		if err := ftpUtil.DeleteImage(image.Thumbnail); err != nil {
			return !true
		}
	}
	return true
}

// 删除TG存储的图片
func DeleteTelegramStorageImage(image models.Image) (deleteStatus bool) {
	// 获取系统配置
	setting, err := settings.GetSettings()
	if err != nil {
		return false
	}

	// 查询图片ID
	db := database.GetDB()
	if db == nil || db.DB == nil {
		return false
	}
	var telegramModel models.ImageTeleGram
	if err := db.DB.Where("file_name = ?", image.FileName).First(&telegramModel).Error; err != nil {
		return false
	}

	tgClient := telegram.NewClient(setting.TGBotToken)
	tgClient.Timeout = 20 * time.Second
	tgClient.Retry = 3

	uploader := telegram.NewTelegramUploader(tgClient)

	// 直接删除，不检查是否成功
	storageTarget := setting.GetTGStorageTarget()
	if storageTarget == "" {
		return false
	}
	if err := uploader.DeletePhoto(storageTarget, telegramModel.TGMessageId); err != nil {
		return false
	}

	// 检查是否存在缩略图
	if image.Thumbnail != "" {
		// 删除缩略图，不检查是否成功
		if telegramModel.TGThumbnailMessageId > 0 {
			if err := uploader.DeletePhoto(storageTarget, telegramModel.TGThumbnailMessageId); err != nil {
				return false
			}
		}
	}
	if err := db.DB.Delete(&telegramModel).Error; err != nil {
		return false
	}
	return true
}

// DeleteImageFile 删除图片文件（根据存储类型分发）
func DeleteImageFile(image models.Image) bool {
	var deleteStatus bool
	switch image.Storage {
	case "default":
		deleteStatus = DeleteDefaultStorageImage(image)
	case "s3":
		deleteStatus = DeleteS3StorageImage(image)
	case "r2":
		deleteStatus = DeleteS3StorageImage(image)
	case "webdav":
		deleteStatus = DeleteWebDavStorageImage(image)
	case "ftp":
		deleteStatus = DeleteFtpStorageImage(image)
	case "telegram":
		deleteStatus = DeleteTelegramStorageImage(image)
	default:
		deleteStatus = false
	}
	return deleteStatus
}
