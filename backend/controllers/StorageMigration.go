package controllers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"oneimg/backend/database"
	"oneimg/backend/models"
	"oneimg/backend/utils/result"
	s3util "oneimg/backend/utils/s3"
	settingsutil "oneimg/backend/utils/settings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	storageMigrationPollInterval  = 3 * time.Second
	storageMigrationObjectTimeout = 2 * time.Minute
	storageMigrationMaxAttempts   = 3
)

var (
	storageMigrationWorkerOnce sync.Once
	storageMigrationWake       = make(chan struct{}, 1)
	storageMutationMu          sync.RWMutex
)

type CreateStorageMigrationRequest struct {
	TargetType     string `json:"target_type" binding:"required"`
	Endpoint       string `json:"endpoint" binding:"required"`
	Region         string `json:"region"`
	Bucket         string `json:"bucket" binding:"required"`
	AccessKey      string `json:"access_key" binding:"required"`
	SecretKey      string `json:"secret_key" binding:"required"`
	ForcePathStyle bool   `json:"force_path_style"`
}

type storageMigrationResponse struct {
	models.StorageMigration
	ProgressPercent int                           `json:"progress_percent"`
	FailedItems     []models.StorageMigrationItem `json:"failed_items,omitempty"`
}

// StorageMutationGuard prevents a write that could be missed by the migration
// snapshot. Image reads remain available for the full migration.
func StorageMutationGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		storageMutationMu.RLock()
		defer storageMutationMu.RUnlock()

		active, err := hasActiveStorageMigration(database.GetDB())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, result.Error(500, "检查存储迁移状态失败"))
			return
		}
		if active {
			c.AbortWithStatusJSON(http.StatusConflict, result.Error(409, "对象存储正在迁移，上传、删除和设置修改已暂时暂停"))
			return
		}
		c.Next()
	}
}

func hasActiveStorageMigration(db *database.Database) (bool, error) {
	if db == nil || db.DB == nil {
		return false, errors.New("database is unavailable")
	}
	var count int64
	err := db.DB.Model(&models.StorageMigration{}).
		Where("status IN ?", []string{models.StorageMigrationPending, models.StorageMigrationRunning}).
		Count(&count).Error
	return count > 0, err
}

func StartStorageMigrationWorker() {
	storageMigrationWorkerOnce.Do(func() {
		go func() {
			db := database.GetDB()
			if db != nil && db.DB != nil {
				_ = db.DB.Model(&models.StorageMigrationItem{}).
					Where("status = ? AND migration_id IN (?)",
						models.StorageMigrationItemCopying,
						db.DB.Model(&models.StorageMigration{}).
							Select("id").Where("status IN ?", []string{models.StorageMigrationPending, models.StorageMigrationRunning})).
					Updates(map[string]any{"status": models.StorageMigrationItemPending, "error": ""}).Error
			}

			ticker := time.NewTicker(storageMigrationPollInterval)
			defer ticker.Stop()
			for {
				runNextStorageMigration()
				select {
				case <-storageMigrationWake:
				case <-ticker.C:
				}
			}
		}()
	})
}

func wakeStorageMigrationWorker() {
	select {
	case storageMigrationWake <- struct{}{}:
	default:
	}
}

func CreateStorageMigration(c *gin.Context) {
	var req CreateStorageMigrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, "目标存储配置不完整"))
		return
	}

	targetConfig, err := migrationTargetConfig(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, err.Error()))
		return
	}

	probeCtx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()
	if err := probeStorage(probeCtx, targetConfig); err != nil {
		c.JSON(http.StatusBadGateway, result.Error(502, "目标存储连接测试失败："+err.Error()))
		return
	}

	storageMutationMu.Lock()
	defer storageMutationMu.Unlock()

	db := database.GetDB()
	if db == nil || db.DB == nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "数据库连接未初始化"))
		return
	}
	active, err := hasActiveStorageMigration(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "检查迁移任务失败"))
		return
	}
	if active {
		c.JSON(http.StatusConflict, result.Error(409, "已有对象存储迁移正在进行"))
		return
	}

	setting, err := settingsutil.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "读取当前存储配置失败"))
		return
	}
	sourceType := setting.GetEffectiveStorageType()
	if sourceType != "s3" && sourceType != "r2" {
		c.JSON(http.StatusBadRequest, result.Error(400, "当前存储不是 S3 或 R2"))
		return
	}
	sourceConfig := s3util.ConfigFromSettings(setting, sourceType)
	if err := sourceConfig.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, "当前存储配置无效："+err.Error()))
		return
	}
	if sameStorageLocation(sourceConfig, targetConfig) {
		c.JSON(http.StatusBadRequest, result.Error(400, "目标 Endpoint 和 Bucket 与当前存储相同"))
		return
	}

	migration := models.StorageMigration{
		Status:          models.StorageMigrationPending,
		SourceType:      sourceConfig.Type,
		SourceEndpoint:  sourceConfig.Endpoint,
		SourceRegion:    sourceConfig.Region,
		SourceAccessKey: sourceConfig.AccessKey,
		SourceSecretKey: sourceConfig.SecretKey,
		SourceBucket:    sourceConfig.Bucket,
		SourcePathStyle: sourceConfig.ForcePathStyle,
		TargetType:      targetConfig.Type,
		TargetEndpoint:  targetConfig.Endpoint,
		TargetRegion:    targetConfig.Region,
		TargetAccessKey: targetConfig.AccessKey,
		TargetSecretKey: targetConfig.SecretKey,
		TargetBucket:    targetConfig.Bucket,
		TargetPathStyle: targetConfig.ForcePathStyle,
	}

	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&migration).Error; err != nil {
			return err
		}
		return enqueueMigrationItems(tx, &migration)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "创建迁移任务失败"))
		return
	}

	wakeStorageMigrationWorker()
	response, _ := loadStorageMigrationResponse(db.DB, migration.ID)
	c.JSON(http.StatusAccepted, result.Success("迁移任务已开始", response))
}

func GetLatestStorageMigration(c *gin.Context) {
	db := database.GetDB()
	if db == nil || db.DB == nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "数据库连接未初始化"))
		return
	}
	var migration models.StorageMigration
	query := db.DB.Order("id DESC").Limit(1).Find(&migration)
	if query.Error != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "获取迁移任务失败"))
		return
	}
	if query.RowsAffected == 0 {
		c.JSON(http.StatusOK, result.Success("ok", nil))
		return
	}
	response, err := loadStorageMigrationResponse(db.DB, migration.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "获取迁移进度失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success("ok", response))
}

func RetryStorageMigration(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, result.Error(400, "迁移任务 ID 无效"))
		return
	}

	storageMutationMu.Lock()
	defer storageMutationMu.Unlock()

	db := database.GetDB()
	if db == nil || db.DB == nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "数据库连接未初始化"))
		return
	}
	active, activeErr := hasActiveStorageMigration(db)
	if activeErr != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "检查迁移任务失败"))
		return
	}
	if active {
		c.JSON(http.StatusConflict, result.Error(409, "已有对象存储迁移正在进行"))
		return
	}

	var migration models.StorageMigration
	if err := db.DB.First(&migration, id).Error; err != nil {
		c.JSON(http.StatusNotFound, result.Error(404, "迁移任务不存在"))
		return
	}
	if migration.Status != models.StorageMigrationFailed {
		c.JSON(http.StatusConflict, result.Error(409, "只有失败的迁移任务可以重试"))
		return
	}

	setting, err := settingsutil.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "读取当前存储配置失败"))
		return
	}
	currentConfig := s3util.ConfigFromSettings(setting, setting.GetEffectiveStorageType())
	if setting.GetEffectiveStorageType() != migration.SourceType || !sameStorageLocation(currentConfig, sourceConfigFromMigration(migration)) {
		c.JSON(http.StatusConflict, result.Error(409, "当前存储已变更，不能重试这项迁移"))
		return
	}

	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := enqueueMigrationItems(tx, &migration); err != nil {
			return err
		}
		if err := tx.Model(&models.StorageMigrationItem{}).
			Where("migration_id = ? AND status IN ?", migration.ID, []string{
				models.StorageMigrationItemFailed,
				models.StorageMigrationItemCopying,
			}).
			Updates(map[string]any{"status": models.StorageMigrationItemPending, "error": ""}).Error; err != nil {
			return err
		}
		return tx.Model(&models.StorageMigration{}).Where("id = ?", migration.ID).
			Updates(map[string]any{
				"status":         models.StorageMigrationPending,
				"failed_objects": 0,
				"error":          "",
				"completed_at":   nil,
			}).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "重试迁移任务失败"))
		return
	}

	_ = refreshMigrationCounters(db.DB, migration.ID)
	wakeStorageMigrationWorker()
	response, _ := loadStorageMigrationResponse(db.DB, migration.ID)
	c.JSON(http.StatusAccepted, result.Success("迁移任务已重新开始", response))
}

func migrationTargetConfig(req CreateStorageMigrationRequest) (s3util.ClientConfig, error) {
	storageType := strings.ToLower(strings.TrimSpace(req.TargetType))
	if storageType != "s3" && storageType != "r2" {
		return s3util.ClientConfig{}, errors.New("目标存储只支持 S3 或 R2")
	}
	endpoint, err := normalizeStorageEndpoint(req.Endpoint)
	if err != nil {
		return s3util.ClientConfig{}, err
	}
	region := strings.TrimSpace(req.Region)
	forcePathStyle := req.ForcePathStyle
	if storageType == "r2" {
		region = "auto"
		forcePathStyle = false
	} else if region == "" {
		region = "us-east-1"
	}
	cfg := s3util.ClientConfig{
		Type:           storageType,
		Endpoint:       endpoint,
		Region:         region,
		AccessKey:      strings.TrimSpace(req.AccessKey),
		SecretKey:      strings.TrimSpace(req.SecretKey),
		Bucket:         strings.TrimSpace(req.Bucket),
		ForcePathStyle: forcePathStyle,
	}
	return cfg, cfg.Validate()
}

func normalizeStorageEndpoint(raw string) (string, error) {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", errors.New("Endpoint 必须是有效的 HTTP 或 HTTPS 地址")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("Endpoint 不能包含查询参数或片段")
	}
	return raw, nil
}

func sameStorageLocation(left, right s3util.ClientConfig) bool {
	return strings.EqualFold(strings.TrimRight(strings.TrimSpace(left.Endpoint), "/"), strings.TrimRight(strings.TrimSpace(right.Endpoint), "/")) &&
		strings.EqualFold(strings.TrimSpace(left.Bucket), strings.TrimSpace(right.Bucket))
}

func probeStorage(ctx context.Context, cfg s3util.ClientConfig) error {
	client, err := s3util.NewClient(ctx, cfg)
	if err != nil {
		return err
	}
	key := ".oneimg-migration-probe/" + uuid.NewString()
	payload := []byte("oneimg-storage-migration-probe")
	_, err = client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:        aws.String(cfg.Bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(payload),
		ContentLength: aws.Int64(int64(len(payload))),
		ContentType:   aws.String("text/plain"),
	})
	if err != nil {
		return fmt.Errorf("写入测试对象失败: %w", err)
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, _ = client.DeleteObject(cleanupCtx, &awss3.DeleteObjectInput{Bucket: aws.String(cfg.Bucket), Key: aws.String(key)})
	}()

	object, err := client.GetObject(ctx, &awss3.GetObjectInput{Bucket: aws.String(cfg.Bucket), Key: aws.String(key)})
	if err != nil {
		return fmt.Errorf("读取测试对象失败: %w", err)
	}
	readBack, readErr := io.ReadAll(io.LimitReader(object.Body, int64(len(payload))+1))
	closeErr := object.Body.Close()
	if readErr != nil {
		return fmt.Errorf("读取测试对象内容失败: %w", readErr)
	}
	if closeErr != nil {
		return fmt.Errorf("关闭测试对象失败: %w", closeErr)
	}
	if !bytes.Equal(readBack, payload) {
		return errors.New("测试对象内容校验失败")
	}
	if _, err := client.DeleteObject(ctx, &awss3.DeleteObjectInput{Bucket: aws.String(cfg.Bucket), Key: aws.String(key)}); err != nil {
		return fmt.Errorf("删除测试对象失败: %w", err)
	}
	return nil
}

func enqueueMigrationItems(tx *gorm.DB, migration *models.StorageMigration) error {
	var images []models.Image
	if err := tx.Where("storage = ?", migration.SourceType).Order("id ASC").Find(&images).Error; err != nil {
		return err
	}
	for _, image := range images {
		objects := []struct {
			kind string
			path string
		}{
			{kind: "original", path: image.Url},
			{kind: "thumbnail", path: image.Thumbnail},
		}
		for _, object := range objects {
			key := migrationObjectKey(object.path)
			if key == "" {
				continue
			}
			item := models.StorageMigrationItem{
				MigrationID: migration.ID,
				ImageID:     image.Id,
				Kind:        object.kind,
				ObjectKey:   key,
				Status:      models.StorageMigrationItemPending,
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&item).Error; err != nil {
				return err
			}
		}
	}
	var total int64
	if err := tx.Model(&models.StorageMigrationItem{}).Where("migration_id = ?", migration.ID).Count(&total).Error; err != nil {
		return err
	}
	migration.TotalObjects = int(total)
	return tx.Model(&models.StorageMigration{}).Where("id = ?", migration.ID).Update("total_objects", migration.TotalObjects).Error
}

func runNextStorageMigration() {
	db := database.GetDB()
	if db == nil || db.DB == nil {
		return
	}
	var migration models.StorageMigration
	query := db.DB.Where("status IN ?", []string{models.StorageMigrationPending, models.StorageMigrationRunning}).Order("id ASC").Limit(1).Find(&migration)
	if query.Error != nil || query.RowsAffected == 0 {
		return
	}

	now := time.Now()
	updates := map[string]any{"status": models.StorageMigrationRunning, "error": ""}
	if migration.StartedAt == nil {
		updates["started_at"] = now
	}
	if err := db.DB.Model(&models.StorageMigration{}).Where("id = ?", migration.ID).Updates(updates).Error; err != nil {
		return
	}
	migration.Status = models.StorageMigrationRunning

	sourceClient, err := s3util.NewClient(context.Background(), sourceConfigFromMigration(migration))
	if err != nil {
		failStorageMigration(db.DB, migration.ID, "创建源存储客户端失败："+err.Error())
		return
	}
	targetClient, err := s3util.NewClient(context.Background(), targetConfigFromMigration(migration))
	if err != nil {
		failStorageMigration(db.DB, migration.ID, "创建目标存储客户端失败："+err.Error())
		return
	}

	var items []models.StorageMigrationItem
	if err := db.DB.Where("migration_id = ? AND status = ?", migration.ID, models.StorageMigrationItemPending).Order("id ASC").Find(&items).Error; err != nil {
		failStorageMigration(db.DB, migration.ID, "读取迁移清单失败："+err.Error())
		return
	}
	for i := range items {
		copyStorageMigrationItem(db.DB, migration, &items[i], sourceClient, targetClient)
	}
	_ = refreshMigrationCounters(db.DB, migration.ID)

	var failed int64
	if err := db.DB.Model(&models.StorageMigrationItem{}).
		Where("migration_id = ? AND status = ?", migration.ID, models.StorageMigrationItemFailed).
		Count(&failed).Error; err != nil {
		failStorageMigration(db.DB, migration.ID, "统计迁移结果失败："+err.Error())
		return
	}
	if failed > 0 {
		failStorageMigration(db.DB, migration.ID, fmt.Sprintf("%d 个对象迁移失败，可重试失败项", failed))
		return
	}

	storageMutationMu.Lock()
	defer storageMutationMu.Unlock()
	if err := completeStorageMigration(db.DB, &migration); err != nil {
		failStorageMigration(db.DB, migration.ID, "切换目标存储失败："+err.Error())
		return
	}
}

func copyStorageMigrationItem(db *gorm.DB, migration models.StorageMigration, item *models.StorageMigrationItem, sourceClient, targetClient *awss3.Client) {
	var lastErr error
	for attempt := 1; attempt <= storageMigrationMaxAttempts; attempt++ {
		_ = db.Model(&models.StorageMigrationItem{}).Where("id = ?", item.ID).Updates(map[string]any{
			"status":   models.StorageMigrationItemCopying,
			"attempts": gorm.Expr("attempts + 1"),
			"error":    "",
		}).Error

		ctx, cancel := context.WithTimeout(context.Background(), storageMigrationObjectTimeout)
		size, digest, err := copyAndVerifyObject(ctx, migration, item.ObjectKey, sourceClient, targetClient)
		cancel()
		if err == nil {
			_ = db.Model(&models.StorageMigrationItem{}).Where("id = ?", item.ID).Updates(map[string]any{
				"status": models.StorageMigrationItemCompleted,
				"size":   size,
				"sha256": digest,
				"error":  "",
			}).Error
			_ = refreshMigrationCounters(db, migration.ID)
			return
		}
		lastErr = err
		if attempt < storageMigrationMaxAttempts {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	log.Printf("对象迁移失败 [job:%d key:%s]: %v", migration.ID, item.ObjectKey, lastErr)
	_ = db.Model(&models.StorageMigrationItem{}).Where("id = ?", item.ID).Updates(map[string]any{
		"status": models.StorageMigrationItemFailed,
		"error":  lastErr.Error(),
	}).Error
	_ = refreshMigrationCounters(db, migration.ID)
}

func copyAndVerifyObject(ctx context.Context, migration models.StorageMigration, key string, sourceClient, targetClient *awss3.Client) (int64, string, error) {
	source, err := sourceClient.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(migration.SourceBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, "", fmt.Errorf("读取源对象失败: %w", err)
	}
	defer source.Body.Close()

	tempFile, err := os.CreateTemp("", "oneimg-storage-migration-*")
	if err != nil {
		return 0, "", fmt.Errorf("创建迁移临时文件失败: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	sourceHash := sha256.New()
	size, err := io.Copy(io.MultiWriter(tempFile, sourceHash), source.Body)
	if err != nil {
		return 0, "", fmt.Errorf("读取源对象内容失败: %w", err)
	}
	if source.ContentLength != nil && *source.ContentLength >= 0 && size != *source.ContentLength {
		return 0, "", fmt.Errorf("源对象长度不一致: expected=%d actual=%d", *source.ContentLength, size)
	}
	sourceDigest := sourceHash.Sum(nil)
	targetSize, targetDigest, targetExists, err := readObjectDigest(ctx, targetClient, migration.TargetBucket, key)
	if err != nil {
		return 0, "", fmt.Errorf("检查目标对象失败: %w", err)
	}
	if targetExists {
		if targetSize == size && bytes.Equal(targetDigest, sourceDigest) {
			return size, hex.EncodeToString(sourceDigest), nil
		}
		return 0, "", fmt.Errorf("目标对象已存在不同内容，未覆盖: source=%d target=%d", size, targetSize)
	}
	if _, err := tempFile.Seek(0, io.SeekStart); err != nil {
		return 0, "", fmt.Errorf("重置迁移临时文件失败: %w", err)
	}

	_, err = targetClient.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:             aws.String(migration.TargetBucket),
		Key:                aws.String(key),
		Body:               tempFile,
		ContentLength:      aws.Int64(size),
		ContentType:        source.ContentType,
		CacheControl:       source.CacheControl,
		ContentDisposition: source.ContentDisposition,
		ContentEncoding:    source.ContentEncoding,
		ContentLanguage:    source.ContentLanguage,
		Expires:            source.Expires,
		Metadata:           source.Metadata,
	})
	if err != nil {
		return 0, "", fmt.Errorf("写入目标对象失败: %w", err)
	}

	verifiedSize, verifiedDigest, found, verifyErr := readObjectDigest(ctx, targetClient, migration.TargetBucket, key)
	if verifyErr != nil {
		cleanupTargetObject(targetClient, migration.TargetBucket, key)
		return 0, "", fmt.Errorf("读取目标对象校验失败: %w", verifyErr)
	}
	if !found || verifiedSize != size || !bytes.Equal(verifiedDigest, sourceDigest) {
		cleanupTargetObject(targetClient, migration.TargetBucket, key)
		return 0, "", fmt.Errorf("目标对象校验不一致: source=%d target=%d", size, verifiedSize)
	}
	return size, hex.EncodeToString(sourceDigest), nil
}

func readObjectDigest(ctx context.Context, client *awss3.Client, bucket, key string) (int64, []byte, bool, error) {
	object, err := client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		var responseErr *smithyhttp.ResponseError
		if errors.As(err, &noSuchKey) || (errors.As(err, &responseErr) && responseErr.HTTPStatusCode() == http.StatusNotFound) {
			return 0, nil, false, nil
		}
		return 0, nil, false, err
	}

	digest := sha256.New()
	size, readErr := io.Copy(digest, object.Body)
	closeErr := object.Body.Close()
	if readErr != nil {
		return 0, nil, true, readErr
	}
	if closeErr != nil {
		return 0, nil, true, closeErr
	}
	return size, digest.Sum(nil), true, nil
}

func cleanupTargetObject(client *awss3.Client, bucket, key string) {
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cleanupCancel()
	_, _ = client.DeleteObject(cleanupCtx, &awss3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
}

func refreshMigrationCounters(db *gorm.DB, migrationID int) error {
	type counters struct {
		Total  int64
		Copied int64
		Failed int64
		Bytes  int64
	}
	var counts counters
	if err := db.Model(&models.StorageMigrationItem{}).
		Where("migration_id = ?", migrationID).
		Select(`COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS copied,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS failed,
			COALESCE(SUM(CASE WHEN status = ? THEN size ELSE 0 END), 0) AS bytes`,
			models.StorageMigrationItemCompleted,
			models.StorageMigrationItemFailed,
			models.StorageMigrationItemCompleted).
		Scan(&counts).Error; err != nil {
		return err
	}
	return db.Model(&models.StorageMigration{}).Where("id = ?", migrationID).Updates(map[string]any{
		"total_objects":  int(counts.Total),
		"copied_objects": int(counts.Copied),
		"failed_objects": int(counts.Failed),
		"copied_bytes":   counts.Bytes,
	}).Error
}

func completeStorageMigration(db *gorm.DB, migration *models.StorageMigration) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var current models.StorageMigration
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, migration.ID).Error; err != nil {
			return err
		}
		if current.Status != models.StorageMigrationRunning {
			return fmt.Errorf("迁移任务状态已变更为 %s", current.Status)
		}
		if err := ensureMigrationCoverage(tx, current); err != nil {
			return err
		}
		var setting models.Settings
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&setting, 1).Error; err != nil {
			return fmt.Errorf("读取当前存储配置失败: %w", err)
		}
		if setting.GetEffectiveStorageType() != current.SourceType ||
			!sameStorageLocation(s3util.ConfigFromSettings(setting, current.SourceType), sourceConfigFromMigration(current)) {
			return errors.New("当前存储配置已变更，拒绝切换")
		}

		settingUpdates := map[string]any{"storage_type": current.TargetType}
		if current.TargetType == "r2" {
			settingUpdates["r2_endpoint"] = current.TargetEndpoint
			settingUpdates["r2_access_key"] = current.TargetAccessKey
			settingUpdates["r2_secret_key"] = current.TargetSecretKey
			settingUpdates["r2_bucket"] = current.TargetBucket
		} else {
			settingUpdates["s3_endpoint"] = current.TargetEndpoint
			settingUpdates["s3_region"] = current.TargetRegion
			settingUpdates["s3_access_key"] = current.TargetAccessKey
			settingUpdates["s3_secret_key"] = current.TargetSecretKey
			settingUpdates["s3_bucket"] = current.TargetBucket
			settingUpdates["s3_path_style"] = current.TargetPathStyle
		}
		settingResult := tx.Model(&setting).Updates(settingUpdates)
		if settingResult.Error != nil {
			return settingResult.Error
		}
		if settingResult.RowsAffected != 1 {
			return errors.New("当前存储配置未更新")
		}
		if err := tx.Model(&models.Image{}).Where("storage = ?", current.SourceType).Update("storage", current.TargetType).Error; err != nil {
			return err
		}

		now := time.Now()
		return tx.Model(&models.StorageMigration{}).Where("id = ?", current.ID).Updates(map[string]any{
			"status":            models.StorageMigrationCompleted,
			"completed_at":      now,
			"error":             "",
			"source_access_key": "",
			"source_secret_key": "",
			"target_access_key": "",
			"target_secret_key": "",
		}).Error
	})
}

func ensureMigrationCoverage(tx *gorm.DB, migration models.StorageMigration) error {
	var images []models.Image
	if err := tx.Where("storage = ?", migration.SourceType).Find(&images).Error; err != nil {
		return err
	}
	for _, image := range images {
		objects := []struct {
			kind string
			path string
		}{{"original", image.Url}, {"thumbnail", image.Thumbnail}}
		for _, object := range objects {
			if migrationObjectKey(object.path) == "" {
				continue
			}
			var count int64
			if err := tx.Model(&models.StorageMigrationItem{}).
				Where("migration_id = ? AND image_id = ? AND kind = ? AND status = ?",
					migration.ID, image.Id, object.kind, models.StorageMigrationItemCompleted).
				Count(&count).Error; err != nil {
				return err
			}
			if count != 1 {
				return fmt.Errorf("图片 %d 的%s尚未完成迁移", image.Id, object.kind)
			}
		}
	}
	return nil
}

func failStorageMigration(db *gorm.DB, migrationID int, message string) {
	_ = refreshMigrationCounters(db, migrationID)
	_ = db.Model(&models.StorageMigration{}).Where("id = ?", migrationID).Updates(map[string]any{
		"status": models.StorageMigrationFailed,
		"error":  message,
	}).Error
	log.Printf("存储迁移失败 [job:%d]: %s", migrationID, message)
}

func migrationObjectKey(rawURL string) string {
	if parsed, err := url.Parse(rawURL); err == nil && parsed.Path != "" {
		rawURL = parsed.Path
	}
	return strings.TrimPrefix(strings.TrimSpace(rawURL), "/")
}

func sourceConfigFromMigration(migration models.StorageMigration) s3util.ClientConfig {
	return s3util.ClientConfig{
		Type:           migration.SourceType,
		Endpoint:       migration.SourceEndpoint,
		Region:         migration.SourceRegion,
		AccessKey:      migration.SourceAccessKey,
		SecretKey:      migration.SourceSecretKey,
		Bucket:         migration.SourceBucket,
		ForcePathStyle: migration.SourcePathStyle,
	}
}

func targetConfigFromMigration(migration models.StorageMigration) s3util.ClientConfig {
	return s3util.ClientConfig{
		Type:           migration.TargetType,
		Endpoint:       migration.TargetEndpoint,
		Region:         migration.TargetRegion,
		AccessKey:      migration.TargetAccessKey,
		SecretKey:      migration.TargetSecretKey,
		Bucket:         migration.TargetBucket,
		ForcePathStyle: migration.TargetPathStyle,
	}
}

func loadStorageMigrationResponse(db *gorm.DB, id int) (*storageMigrationResponse, error) {
	var migration models.StorageMigration
	if err := db.First(&migration, id).Error; err != nil {
		return nil, err
	}
	response := &storageMigrationResponse{StorageMigration: migration}
	if migration.TotalObjects > 0 {
		response.ProgressPercent = migration.CopiedObjects * 100 / migration.TotalObjects
	} else if migration.Status == models.StorageMigrationCompleted {
		response.ProgressPercent = 100
	}
	if migration.FailedObjects > 0 {
		if err := db.Where("migration_id = ? AND status = ?", migration.ID, models.StorageMigrationItemFailed).
			Order("id ASC").Limit(20).Find(&response.FailedItems).Error; err != nil {
			return nil, err
		}
	}
	return response, nil
}
