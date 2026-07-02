package images

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"oneimg/backend/config"
	"oneimg/backend/models"
	"oneimg/backend/utils/watermark"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

// 常量定义 - 提取魔法数字和固定值
const (
	DefaultCompressQuality = 85
	ThumbnailMaxWidth      = 300
	ThumbnailMaxHeight     = 300
	ThumbnailQuality       = 80
	CompressSizeThreshold  = 1024 * 1024 // 1MB
	MaxImageDimension      = 20000
	MaxImagePixels         = 50000000
)

// 特殊格式常量
var (
	specialFormats   = []string{"gif"}
	specialMimeTypes = []string{
		"image/gif",
		"image/svg+xml",
	}
	ErrUnsupportedFormat  = errors.New("unsupported image format")
	ErrFileTooLarge       = errors.New("file size exceeds limit")
	ErrMissingContentType = errors.New("missing content type")
)

type ImageService struct{}

var ImageSvc *ImageService

// InitImageService 初始化图片服务（线程安全）
func InitImageService() {
	if ImageSvc == nil {
		ImageSvc = &ImageService{}
	}
}

// ProcessedImage 处理后的图片数据
type ProcessedImage struct {
	CompressedBytes []byte // 处理后的字节
	ThumbnailBytes  []byte // 缩略图字节
	Width           int    // 图片宽度
	Height          int    // 图片高度
	Format          string // 最终格式
	MimeType        string // 最终MIME类型
	OutputExt       string // 输出文件扩展名
	UniqueFileName  string // 唯一文件名
	ThumbnailName   string // 缩略图文件名（固定为WebP扩展名）
}

// ProcessImage 处理图片（压缩、获取尺寸等）
func (s *ImageService) ProcessImage(
	file multipart.File,
	header *multipart.FileHeader,
	setting models.Settings,
) (*ProcessedImage, error) {
	// 1. 读取文件内容（一次性读取，避免多次IO）
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}

	// 验证文件完整性（检查实际读取大小是否与Header声明大小一致）
	if header.Size > 0 && int64(len(fileBytes)) != header.Size {
		return nil, fmt.Errorf("upload truncated: expected %d bytes, got %d bytes", header.Size, len(fileBytes))
	}

	// 2. 解码图片（获取原图信息）
	img, format, err := s.decodeImage(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("decode image failed: %w", err)
	}

	// 3. 获取图片基本信息
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 || width > MaxImageDimension || height > MaxImageDimension || int64(width)*int64(height) > MaxImagePixels {
		return nil, fmt.Errorf("image dimensions exceed limit: %dx%d", width, height)
	}
	mimeType := detectImageMIME(fileBytes)

	// 4. 处理主图片（压缩/格式转换）
	processedBytes, finalFormat, finalMimeType, err := s.processMainImage(
		fileBytes, img, format, mimeType, header.Size, setting,
	)
	if err != nil {
		return nil, fmt.Errorf("process main image failed: %w", err)
	}

	// 5. 处理文件扩展名
	outputExt := map[string]string{
		"image/jpeg":    ".jpg",  // JPEG格式
		"image/png":     ".png",  // PNG格式
		"image/gif":     ".gif",  // GIF格式
		"image/webp":    ".webp", // WebP格式
		"image/svg+xml": ".svg",  // SVG格式
		"image/bmp":     ".bmp",  // BMP格式
		"image/tiff":    ".tiff", // TIFF格式
		"image/heic":    ".heic", // HEIC格式
		"image/heif":    ".heif", // HEIF格式
	}

	var thumbnailBytes []byte
	if setting.Thumbnail {
		// 仅在开启缩略图时进行二次解码，避免无意义的 CPU 和内存消耗。
		reader := bytes.NewReader(processedBytes)
		img, _, err = image.Decode(reader)
		if err != nil {
			return nil, fmt.Errorf("decode image failed: %w", err)
		}
		thumbnailBytes, err = s.generateThumbnail(img)
		if err != nil {
			return nil, fmt.Errorf("generate thumbnail failed: %w", err)
		}
	}

	// 7. 组装返回结果
	uniqueFileName := generateUniqueFileName(outputExt[finalMimeType])
	return &ProcessedImage{
		CompressedBytes: processedBytes,
		ThumbnailBytes:  thumbnailBytes,
		Width:           width,
		Height:          height,
		Format:          finalFormat,
		MimeType:        finalMimeType,
		OutputExt:       outputExt[finalMimeType],
		UniqueFileName:  uniqueFileName,
		ThumbnailName:   strings.TrimSuffix(uniqueFileName, outputExt[finalMimeType]) + ".webp",
	}, nil
}

// processMainImage 处理主图片（拆分逻辑，提高可读性）
func (s *ImageService) processMainImage(
	fileBytes []byte,
	img image.Image,
	format, mimeType string,
	fileSize int64,
	setting models.Settings,
) ([]byte, string, string, error) {
	webpQuality := setting.WebpQuality
	if webpQuality < 1 || webpQuality > 100 {
		webpQuality = DefaultCompressQuality
	}

	// 特殊格式直接返回原数据
	if s.isSpecialFormat(format, mimeType) {
		return fileBytes, format, mimeType, nil
	}

	// 添加水印
	if setting.WatermarkEnable {
		watermarkCfg := watermark.WatermarkSetting(setting)
		fileReader := bytes.NewReader(fileBytes)
		processedReader, err := watermark.ProcessImageWithWatermark(fileReader, mimeType, watermarkCfg)
		if err != nil {
			return nil, "", "", fmt.Errorf("添加水印失败：%w", err)
		}
		fileBytes, err = io.ReadAll(processedReader)
		if err != nil {
			return nil, "", "", fmt.Errorf("读取水印后图片数据失败：%w", err)
		}
		img, _, err = image.Decode(bytes.NewReader(fileBytes))
		if err != nil {
			return nil, "", "", fmt.Errorf("解码水印后图片失败：%w", err)
		}
	}

	// WebP格式处理
	if strings.ToLower(format) == "webp" {
		if setting.OriginalImage || fileSize <= CompressSizeThreshold {
			return fileBytes, "webp", "image/webp", nil
		}
		compressed, err := s.compressWebP(img, webpQuality)
		if err != nil {
			return nil, "", "", fmt.Errorf("compress webp: %w", err)
		}
		return compressed, "webp", "image/webp", nil
	}

	// 需要转换为WebP
	if setting.SaveWebp {
		webpData, err := s.convertToWebP(img, webpQuality)
		if err != nil {
			return nil, "", "", fmt.Errorf("convert to webp: %w", err)
		}
		log.Println("转换webp")
		return webpData, "webp", "image/webp", nil
	}

	// 保存原图
	if setting.OriginalImage {
		return fileBytes, format, mimeType, nil
	}

	// 默认进行压缩
	compressed, err := s.compressOriginalFormat(img, format, DefaultCompressQuality)
	if err != nil {
		return nil, "", "", fmt.Errorf("compress %s: %w", format, err)
	}
	return compressed, format, mimeForFormat(format), nil
}

func (s *ImageService) compressOriginalFormat(img image.Image, format string, quality int) ([]byte, error) {
	var output bytes.Buffer
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&output, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
	case "png":
		encoder := png.Encoder{CompressionLevel: png.DefaultCompression}
		if err := encoder.Encode(&output, img); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported output format %q", format)
	}
	return output.Bytes(), nil
}

// generateThumbnail 生成缩略图
func (s *ImageService) generateThumbnail(img image.Image) ([]byte, error) {
	// 所有缩略图统一为 WebP；调用方使用独立的 .webp 文件名和 MIME 类型。
	return s.generateWebPThumbnail(img, ThumbnailMaxWidth, ThumbnailMaxHeight, ThumbnailQuality)
}

// isSpecialFormat 检查是否为特殊格式（需要保持原格式）
func (s *ImageService) isSpecialFormat(format, mimeType string) bool {
	// 检查格式
	if slices.Contains(specialFormats, strings.ToLower(format)) {
		return true
	}

	// 检查MIME类型
	if slices.Contains(specialMimeTypes, mimeType) {
		return true
	}

	return false
}

// decodeImage 解码图片，支持webp/gif/png/jpeg等格式
// 优化点：减少内存拷贝，按优先级解码
func (s *ImageService) decodeImage(reader io.Reader) (image.Image, string, error) {
	// 读取数据到缓冲区（复用）
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", fmt.Errorf("read image data: %w", err)
	}
	buf := bytes.NewReader(data)

	// 按优先级解码（常用格式优先）
	decodeFuncs := []struct {
		decode func(*bytes.Reader) (image.Image, error)
		format string
	}{
		{func(r *bytes.Reader) (image.Image, error) { return webp.Decode(r) }, "webp"},
		{func(r *bytes.Reader) (image.Image, error) { return gif.Decode(r) }, "gif"},
		{func(r *bytes.Reader) (image.Image, error) { return png.Decode(r) }, "png"},
		{func(r *bytes.Reader) (image.Image, error) { return jpeg.Decode(r) }, "jpeg"},
	}

	for _, df := range decodeFuncs {
		buf.Seek(0, io.SeekStart) // 重置读取指针
		img, err := df.decode(buf)
		if err == nil {
			return img, df.format, nil
		}
	}

	// 最后尝试标准库的自动检测
	buf.Seek(0, io.SeekStart)
	img, format, err := image.Decode(buf)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrUnsupportedFormat, err)
	}

	return img, format, nil
}

// convertToWebP 将图片转换为webp格式
func (s *ImageService) convertToWebP(img image.Image, quality int) ([]byte, error) {
	if quality < 0 || quality > 100 {
		return nil, fmt.Errorf("invalid quality: %d (must be 0-100)", quality)
	}

	data, err := webp.EncodeRGBA(img, float32(quality))
	if err != nil {
		return nil, fmt.Errorf("encode webp: %w", err)
	}

	return data, nil
}

// compressWebP 压缩webp图片（复用转换逻辑）
func (s *ImageService) compressWebP(img image.Image, quality int) ([]byte, error) {
	return s.convertToWebP(img, quality)
}

// ValidateImage 验证图片格式和大小
func (s *ImageService) ValidateImage(
	header *multipart.FileHeader,
	allowedTypes []string,
	maxSize int64,
) error {
	// 检查文件大小
	if header.Size > maxSize {
		return fmt.Errorf("%w: max size %d bytes, got %d bytes",
			ErrFileTooLarge, maxSize, header.Size)
	}

	// 检查Content-Type
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		return ErrMissingContentType
	}

	// 检查是否允许的类型
	if !slices.Contains(allowedTypes, mimeType) {
		return fmt.Errorf("unsupported content type: %s (allowed: %s)",
			mimeType, strings.Join(allowedTypes, ", "))
	}

	file, err := header.Open()
	if err != nil {
		return fmt.Errorf("open image for validation: %w", err)
	}
	defer file.Close()
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("read image signature: %w", err)
	}
	detectedType := detectImageMIME(buffer[:n])
	if !slices.Contains(allowedTypes, detectedType) {
		return fmt.Errorf("unsupported image data type: %s", detectedType)
	}
	if detectedType != mimeType {
		return fmt.Errorf("content type mismatch: declared %s, detected %s", mimeType, detectedType)
	}

	return nil
}

func detectImageMIME(data []byte) string {
	if len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "image/webp"
	}
	return strings.ToLower(strings.TrimSpace(strings.SplitN(http.DetectContentType(data), ";", 2)[0]))
}

func mimeForFormat(format string) string {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// generateWebPThumbnail 生成webp格式缩略图
func (s *ImageService) generateWebPThumbnail(
	img image.Image,
	maxWidth, maxHeight, quality int,
) ([]byte, error) {
	// 调整图片大小
	thumbnail := imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)

	// 转换为WebP
	return s.convertToWebP(thumbnail, quality)
}

// ValidateImageFile 验证图片文件
func ValidateImageFile(header *multipart.FileHeader, cfg *config.Config) error {
	return ImageSvc.ValidateImage(header, cfg.AllowedTypes, cfg.MaxFileSize)
}

// ReadFileContent 读取文件内容
func ReadFileContent(header *multipart.FileHeader) ([]byte, error) {
	file, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败：%v", err)
	}
	defer file.Close()

	return io.ReadAll(file)
}

// GetFileMimeType 获取文件MIME类型
func GetFileMimeType(header *multipart.FileHeader) string {
	return header.Header.Get("Content-Type")
}

// generateUniqueFileName 生成唯一文件名
// 格式: {timestamp}_{random6char}.{ext}
// 示例: 1764076031141_to5nxg.webp
func generateUniqueFileName(ext string) string {
	timestamp := time.Now().UnixMilli()
	randomStr := strings.ReplaceAll(uuid.NewString()[:8], "-", "")
	return fmt.Sprintf("%d_%s%s", timestamp, randomStr, ext)
}
