package uploads

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"

	"oneimg/backend/config"
	"oneimg/backend/database"
	"oneimg/backend/interfaces"
	"oneimg/backend/models"
	"oneimg/backend/utils/ftp"
	"oneimg/backend/utils/images"
	"oneimg/backend/utils/s3"
	"oneimg/backend/utils/telegram"
    "oneimg/backend/utils/customapi"
	"oneimg/backend/utils/webdav"
)

// 所有上传器实现统一接口
type S3R2Uploader struct{}
type WebDAVUploader struct{}
type DefaultUploader struct{}
type FTPUploader struct{}
type TelegramUploader struct{}
type CustomApiUploader struct{}

// S3/R2上传实现
func (u *S3R2Uploader) Upload(c *gin.Context, cfg *config.Config, setting *models.Settings, fileHeader *multipart.FileHeader) (*interfaces.ImageUploadResult, error) {
	// 验证图片
	if err := images.ValidateImageFile(fileHeader, cfg); err != nil {
		return nil, fmt.Errorf("图片验证失败: %v", err)
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 处理图片
	processedImage, err := images.ImageSvc.ProcessImage(file, fileHeader, *setting)
	if err != nil {
		return nil, fmt.Errorf("图片处理失败: %v", err)
	}

	uniqueFileName := processedImage.UniqueFileName

	// 创建年/月子目录
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	subDir := PathJoin("uploads", year, month)
	objectKey := PathJoin(subDir, uniqueFileName)

	// 获取S3/R2客户端

	client, err := s3.NewS3Client(*setting)
	if err != nil {
		return nil, fmt.Errorf("创建S3/R2客户端失败：%v", err)
	}

	// 上传文件到S3/R2
	contentType := "image/webp"
	if !setting.SaveWebp {
		contentType = fileHeader.Header.Get("Content-Type")
	}

	_, err = client.PutObject(context.TODO(), &awss3.PutObjectInput{
		Bucket:      aws.String(setting.S3Bucket),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(processedImage.CompressedBytes),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("S3/R2上传失败：%v", err)
	}

	thumbnailURL := ""
	// 检查是否上传缩略图
	if setting.Thumbnail {
		_, err = client.PutObject(context.TODO(), &awss3.PutObjectInput{
			Bucket:      aws.String(setting.S3Bucket),
			Key:         aws.String(PathJoin("uploads", year, month, "thumbnails", uniqueFileName)), // 缩略图存放路径
			Body:        bytes.NewReader(processedImage.ThumbnailBytes),
			ContentType: aws.String("image/webp"),
		})
		if err == nil {
			thumbnailURL = "/" + PathJoin("uploads", year, month, "thumbnails", uniqueFileName)
		}
	}

	url := "/" + PathJoin("uploads", year, month, uniqueFileName)

	return &interfaces.ImageUploadResult{
		Success:      true,
		Message:      "上传成功",
		FileName:     uniqueFileName,
		FileSize:     int64(len(processedImage.CompressedBytes)),
		MimeType:     contentType,
		URL:          url,
		ThumbnailURL: thumbnailURL,
		Storage:      setting.StorageType,
		CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
		Width:        processedImage.Width,
		Height:       processedImage.Height,
	}, nil
}

// WebDAV上传实现
func (u *WebDAVUploader) Upload(c *gin.Context, cfg *config.Config, setting *models.Settings, fileHeader *multipart.FileHeader) (*interfaces.ImageUploadResult, error) {
	// 验证图片
	if err := images.ValidateImageFile(fileHeader, cfg); err != nil {
		return nil, fmt.Errorf("图片验证失败: %v", err)
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 处理图片
	processedImage, err := images.ImageSvc.ProcessImage(file, fileHeader, *setting)
	if err != nil {
		return nil, fmt.Errorf("图片处理失败: %v", err)
	}

	// 生成唯一文件名
	uniqueFileName := processedImage.UniqueFileName

	// 创建年/月子目录
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	subDir := filepath.Join("uploads", year, month)
	objectPath := filepath.Join("/", subDir, uniqueFileName)

	// 初始化WebDAV客户端
	client := webdav.Client(webdav.Config{
		BaseURL:  setting.WebdavURL,
		Username: setting.WebdavUser,
		Password: setting.WebdavPass,
		Timeout:  30 * time.Second,
	})

	// 上传文件到WebDAV服务器
	err = client.WebDAVUpload(context.TODO(), objectPath, bytes.NewReader(processedImage.CompressedBytes))
	if err != nil {
		return nil, fmt.Errorf("WebDAV上传失败：%v", err)
	}

	// 检查是否上传缩略图
	thumbnailURL := ""
	if setting.Thumbnail {
		err = client.WebDAVUpload(context.TODO(), filepath.Join("/", subDir, "thumbnails", uniqueFileName), bytes.NewReader(processedImage.ThumbnailBytes))
		if err == nil {
			thumbnailURL = "/uploads/" + year + "/" + month + "/thumbnails/" + uniqueFileName
		}
	}

	// 构建访问URL
	url := "/uploads/" + year + "/" + month + "/" + uniqueFileName

	return &interfaces.ImageUploadResult{
		Success:      true,
		Message:      "上传成功",
		FileName:     uniqueFileName,
		FileSize:     int64(len(processedImage.CompressedBytes)),
		MimeType:     processedImage.MimeType,
		URL:          url,
		ThumbnailURL: thumbnailURL,
		Storage:      setting.StorageType,
		Width:        processedImage.Width,
		Height:       processedImage.Height,
		CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// FTP上传实现
func (u *FTPUploader) Upload(c *gin.Context, cfg *config.Config, setting *models.Settings, fileHeader *multipart.FileHeader) (*interfaces.ImageUploadResult, error) {
	// 1. 验证图片
	if err := images.ValidateImageFile(fileHeader, cfg); err != nil {
		return nil, fmt.Errorf("图片验证失败: %v", err)
	}

	// 2. 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 3. 处理图片（压缩、生成缩略图等）
	processedImage, err := images.ImageSvc.ProcessImage(file, fileHeader, *setting)
	if err != nil {
		return nil, fmt.Errorf("图片处理失败: %v", err)
	}

	// 4. 生成唯一文件名
	uniqueFileName := processedImage.UniqueFileName

	// 5. 构建FTP目录结构（年/月）
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	subDir := PathJoin("uploads", year, month)
	objectPath := PathJoin(subDir, uniqueFileName)

	// 初始化FTP客户端
	ftpUtil := ftp.NewFTPUtil(ftp.FTPConfig{
		Host:     setting.FTPHost,
		Port:     setting.FTPPort,
		User:     setting.FTPUser,
		Password: setting.FTPPass,
		Timeout:  10,
	})
	defer ftpUtil.Close()
	err = ftpUtil.UploadImage(
		objectPath,
		processedImage.CompressedBytes,
		processedImage.MimeType,
	)
	if err != nil {
		return nil, fmt.Errorf("FTP上传失败：%v", err)
	}

	// 检查是否上传缩略图
	thumbnailURL := ""
	if setting.Thumbnail {
		err := ftpUtil.UploadImage(
			PathJoin(subDir, "thumbnails", uniqueFileName),
			processedImage.ThumbnailBytes,
			"image/webp",
		)
		if err == nil {
			thumbnailURL = "/uploads/" + year + "/" + month + "/thumbnails/" + uniqueFileName
		}
	}

	url := "/uploads/" + year + "/" + month + "/" + uniqueFileName

	return &interfaces.ImageUploadResult{
		Success:      true,
		Message:      "上传成功",
		FileName:     uniqueFileName,
		FileSize:     int64(len(processedImage.CompressedBytes)),
		MimeType:     processedImage.MimeType,
		URL:          url,
		ThumbnailURL: thumbnailURL,
		Storage:      setting.StorageType,
		Width:        processedImage.Width,
		Height:       processedImage.Height,
		CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// 本地默认上传实现
func (u *DefaultUploader) Upload(c *gin.Context, cfg *config.Config, setting *models.Settings, fileHeader *multipart.FileHeader) (*interfaces.ImageUploadResult, error) {
	// 验证图片
	if err := images.ValidateImageFile(fileHeader, cfg); err != nil {
		return nil, fmt.Errorf("图片验证失败: %v", err)
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 处理图片
	processedImage, err := images.ImageSvc.ProcessImage(file, fileHeader, *setting)
	if err != nil {
		return nil, fmt.Errorf("图片处理失败: %v", err)
	}

	uniqueFileName := processedImage.UniqueFileName

	// 创建年/月子目录
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	subDir := filepath.Join("uploads", year, month)

	// 确保年月子目录存在
	fullSubDir := filepath.Join(".", subDir)
	if err := ensureUploadDir(fullSubDir); err != nil {
		return nil, fmt.Errorf("创建子目录失败：%v", err)
	}

	// 构建文件路径
	filePath := filepath.Join(fullSubDir, uniqueFileName)

	// 保存处理后的图片文件
	if err := saveFile(filePath, processedImage.CompressedBytes); err != nil {
		return nil, fmt.Errorf("保存文件失败：%v", err)
	}
	thumbnailURL := ""
	// 检查是否上传缩略图
	if setting.Thumbnail {
		if err := ensureUploadDir(filepath.Join(fullSubDir, "thumbnails")); err == nil {
			// 构建缩略图文件路径
			thumbFilePath := filepath.Join(fullSubDir, "thumbnails", uniqueFileName)
			// 保存缩略图文件
			if err := saveFile(thumbFilePath, processedImage.ThumbnailBytes); err != nil {
				log.Println(err)
				// 忽略错误
			}
			thumbnailURL = "/uploads/" + year + "/" + month + "/thumbnails/" + uniqueFileName
		}
	}

	// 构建访问URL (包含年/月子目录)
	fileURL := "/uploads/" + year + "/" + month + "/" + uniqueFileName

	return &interfaces.ImageUploadResult{
		Success:      true,
		Message:      "上传成功",
		URL:          fileURL,
		ThumbnailURL: thumbnailURL,
		Storage:      setting.StorageType,
		FileName:     uniqueFileName,
		FileSize:     int64(len(processedImage.CompressedBytes)),
		MimeType:     processedImage.MimeType,
		Width:        processedImage.Width,
		Height:       processedImage.Height,
		CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// Telegram上传实现
func (u *TelegramUploader) Upload(c *gin.Context, cfg *config.Config, setting *models.Settings, fileHeader *multipart.FileHeader) (*interfaces.ImageUploadResult, error) {
	// 1. 验证图片
	if err := images.ValidateImageFile(fileHeader, cfg); err != nil {
		return nil, fmt.Errorf("图片验证失败: %v", err)
	}

	// 2. 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 3. 处理图片
	processedImage, err := images.ImageSvc.ProcessImage(file, fileHeader, *setting)
	if err != nil {
		return nil, fmt.Errorf("图片处理失败: %v", err)
	}

	// 4. 基础参数校验
	if setting.TGBotToken == "" {
		return nil, fmt.Errorf("telegram bot token 不能为空")
	}
	if setting.TGReceivers == "" {
		return nil, fmt.Errorf("telegram receivers 不能为空")
	}

	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	// 5. 初始化TG客户端
	tgClient := telegram.NewClient(setting.TGBotToken)
	tgClient.Timeout = 20 * time.Second
	tgClient.Retry = 3

	uniqueFileName := processedImage.UniqueFileName

	// 上传主图片
	fileID, messageID, err := tgClient.UploadPhotoByBytes(
		setting.TGReceivers,
		processedImage.CompressedBytes,
		processedImage.UniqueFileName,
		fmt.Sprintf("上传图片: %s", processedImage.UniqueFileName),
	)
	if err != nil {
		return nil, fmt.Errorf("Telegram上传图片失败")
	}

	// 7. 上传缩略图（如果开启）
	thumbnailURL := ""
	thumbFileIDURL := ""
	thumbFileMessageID := 0
	if setting.Thumbnail && len(processedImage.ThumbnailBytes) > 0 {
		thumbFileID, thumbMessageID, err := tgClient.UploadPhotoByBytes(
			setting.TGReceivers,
			processedImage.ThumbnailBytes,
			fmt.Sprintf("thumbnail_%s", uniqueFileName),
			fmt.Sprintf("缩略图: %s", processedImage.UniqueFileName),
		)
		if err == nil {
			// Telegram没有直接的URL，这里存储fileID作为标识
			thumbFileIDURL = thumbFileID
			thumbFileMessageID = thumbMessageID
			thumbnailURL = fmt.Sprintf("/uploads/%s/%s/thumbnails/%s", year, month, uniqueFileName)
		} else {
			log.Printf("Telegram上传缩略图失败: %v", err)
		}
	}

	url := "/uploads/" + year + "/" + month + "/" + uniqueFileName

	telegramModel := models.ImageTeleGram{
		TGFileId:             fileID,
		TGThumbnailFileId:    thumbFileIDURL,
		TGMessageId:          messageID,
		TGThumbnailMessageId: thumbFileMessageID,
		FileName:             uniqueFileName,
	}
	db := database.GetDB()
	if db != nil {
		if err := db.DB.Create(&telegramModel).Error; err != nil {
			return nil, fmt.Errorf("保存telegram图片信息到数据库失败")
		}
	}

	return &interfaces.ImageUploadResult{
		Success:      true,
		Message:      "Telegram上传成功",
		FileName:     processedImage.UniqueFileName,
		FileSize:     int64(len(processedImage.CompressedBytes)),
		MimeType:     processedImage.MimeType,
		URL:          url,
		ThumbnailURL: thumbnailURL,
		Storage:      setting.StorageType, // 存储类型标识
		CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
		Width:        processedImage.Width,
		Height:       processedImage.Height,
	}, nil
}

// CustomAPI上传实现
func (u *CustomApiUploader) Upload(c *gin.Context, cfg *config.Config, setting *models.Settings, fileHeader *multipart.FileHeader) (*interfaces.ImageUploadResult, error) {
	// 1. 验证图片
	if err := images.ValidateImageFile(fileHeader, cfg); err != nil {
		return nil, fmt.Errorf("图片验证失败: %v", err)
	}

	// 2. 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 3. 处理图片（Custom API通常自处理压缩，但也可能需要我们预处理。这里先做标准预处理）
	processedImage, err := images.ImageSvc.ProcessImage(file, fileHeader, *setting)
	if err != nil {
		return nil, fmt.Errorf("图片处理失败: %v", err)
	}

	// 4. 调用Custom API上传
	apiClient := customapi.NewCustomApiUploader(setting.CustomApiUrl, setting.CustomApiKey)
	
	// 使用处理后的 compressed bytes 上传
	resp, err := apiClient.Upload(processedImage.CompressedBytes, processedImage.UniqueFileName)
	if err != nil {
		return nil, fmt.Errorf("Custom API上传失败: %v", err)
	}

	// 5. 组装结果
    // 注意：NodeSeek API 返回的 URL 通常是完整 URL，但我们的系统可能期望相对路径或者能够直接访问的路径。
    // 如果返回的是绝对路径，直接使用。
    // filename 最好使用 API 返回的 hash 或者是我们自己生成的 unique name，视 API 行为而定。
    // NodeSeek API 本身不返回 ID，只返回 URL。删除需要 ID。
    // 观察 UploadResponse，没有显式的 ID 字段。
    // 假设：URL 的最后一部分是文件名/ID，或者 hash 是 ID。
    // 如果 Delete 需要 ID，我们需要从 URL 解析或者使用 API 返回的其他字段。
    // 根据提供的文档：DELETE /api/image/{image_id}
    // 再次检查 UploadResponse：Data 中有 Hash，Url。
    // 通常 NodeImage 的 image_id 可能是 Hash 或者 URL 的文件名部分。
    // 暂时假设 Hash 是 ID，或者我们需要解析 URL。
    // 我们将 URL 存入数据库。
    
    // 为了支持删除，我们需要存储足够的信息。
    // 这里的 FileName 字段通常存文件名。
    // 我们可能需要将 Hash 存入 Image 表的某个字段，或者 Metadata。
    // 目前 Image 表没有 Extra 字段。利用 FileName 存储 "Hash|UniqueName"？ 不太好。
    // 观察 Services/images.go，UploadResult 要返回 FileName。
    
    // 如果 API 返回的 URL 是 http://host/image/12345.jpg
    // 那么 delete id 可能是 12345。
    
	return &interfaces.ImageUploadResult{
		Success:      true,
		Message:      "上传成功",
		FileName:     resp.Data.Hash, // 暂存 Hash 作为文件名标识，方便删除时使用？ 或者存原始文件名
        // 修正：FileName 还是存我们生成的 UniqueFileName 以保持一致性，
        // 但为了删除，我们需要知道 API 的 ID。
        // 如果 Delete 接口需要 Hash，那我们就得想办法存 Hash。
        // 现有的 Image 模型有 URL, FileName, Storage。
        // 如果 URL 是完整的，DeleteLogic 需要解析它。
        // 让我们看看 DeleteImage.go 怎么做的。
		FileSize:     resp.Data.Size,
		MimeType:     resp.Data.Type,
		URL:          resp.Data.Url, // 外部链接
		ThumbnailURL: "", // Custom API 可能不直接返回缩略图 URL，除非我们在前端解析
		Storage:      "custom",
		Width:        resp.Data.Width,
		Height:       resp.Data.Height,
		CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// ensureUploadDir 确保上传目录存在
func ensureUploadDir(uploadPath string) error {
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		return os.MkdirAll(uploadPath, 0755)
	}
	return nil
}

// saveFile 保存文件到磁盘
func saveFile(filePath string, data []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// 辅助函数
func PathJoin(parts ...string) string {
	return strings.Join(parts, "/")
}
