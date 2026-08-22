package controllers

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"oneimg/backend/models"
	s3util "oneimg/backend/utils/s3"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCopyAndVerifyObjectAcrossS3CompatibleStores(t *testing.T) {
	sourceStore := newFakeS3Store(t, false)
	defer sourceStore.Close()
	targetStore := newFakeS3Store(t, false)
	defer targetStore.Close()

	payload := []byte("verified migration payload")
	sourceStore.put("source-bucket", "uploads/2026/08/image.webp", payload)
	sourceClient := newFakeS3Client(t, sourceStore.URL)
	targetClient := newFakeS3Client(t, targetStore.URL)
	migration := models.StorageMigration{SourceBucket: "source-bucket", TargetBucket: "target-bucket"}

	size, digest, err := copyAndVerifyObject(context.Background(), migration, "uploads/2026/08/image.webp", sourceClient, targetClient)
	if err != nil {
		t.Fatalf("copyAndVerifyObject returned error: %v", err)
	}
	if size != int64(len(payload)) || len(digest) != 64 {
		t.Fatalf("unexpected verification result: size=%d digest=%q", size, digest)
	}
	if got := targetStore.get("target-bucket", "uploads/2026/08/image.webp"); !bytes.Equal(got, payload) {
		t.Fatalf("target payload=%q, want %q", got, payload)
	}
}

func TestProbeStorageChecksReadWriteAndDelete(t *testing.T) {
	store := newFakeS3Store(t, false)
	defer store.Close()
	cfg := s3util.ClientConfig{
		Type: "s3", Endpoint: store.URL, Region: "us-east-1", Bucket: "target-bucket",
		AccessKey: "access", SecretKey: "secret", ForcePathStyle: true,
	}
	if err := probeStorage(context.Background(), cfg); err != nil {
		t.Fatalf("probeStorage returned error: %v", err)
	}
	store.mu.Lock()
	remaining := len(store.objects)
	store.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("probe object was not removed; remaining=%d", remaining)
	}
}

func TestCopyAndVerifyObjectRemovesCorruptTarget(t *testing.T) {
	sourceStore := newFakeS3Store(t, false)
	defer sourceStore.Close()
	targetStore := newFakeS3Store(t, true)
	defer targetStore.Close()
	sourceStore.put("source-bucket", "uploads/image.webp", []byte("source payload"))

	_, _, err := copyAndVerifyObject(
		context.Background(),
		models.StorageMigration{SourceBucket: "source-bucket", TargetBucket: "target-bucket"},
		"uploads/image.webp",
		newFakeS3Client(t, sourceStore.URL),
		newFakeS3Client(t, targetStore.URL),
	)
	if err == nil || !strings.Contains(err.Error(), "校验不一致") {
		t.Fatalf("expected verification mismatch, got %v", err)
	}
	if got := targetStore.get("target-bucket", "uploads/image.webp"); got != nil {
		t.Fatalf("corrupt target object was not removed: %q", got)
	}
}

func TestCopyAndVerifyObjectDoesNotOverwriteDifferentTarget(t *testing.T) {
	sourceStore := newFakeS3Store(t, false)
	defer sourceStore.Close()
	targetStore := newFakeS3Store(t, false)
	defer targetStore.Close()
	sourceStore.put("source-bucket", "uploads/image.webp", []byte("source payload"))
	targetStore.put("target-bucket", "uploads/image.webp", []byte("existing target payload"))

	_, _, err := copyAndVerifyObject(
		context.Background(),
		models.StorageMigration{SourceBucket: "source-bucket", TargetBucket: "target-bucket"},
		"uploads/image.webp",
		newFakeS3Client(t, sourceStore.URL),
		newFakeS3Client(t, targetStore.URL),
	)
	if err == nil || !strings.Contains(err.Error(), "未覆盖") {
		t.Fatalf("expected a no-overwrite conflict, got %v", err)
	}
	if got := targetStore.get("target-bucket", "uploads/image.webp"); !bytes.Equal(got, []byte("existing target payload")) {
		t.Fatalf("existing target object was changed: %q", got)
	}
}

func TestCopyAndVerifyObjectAcceptsMatchingTarget(t *testing.T) {
	sourceStore := newFakeS3Store(t, false)
	defer sourceStore.Close()
	targetStore := newFakeS3Store(t, false)
	defer targetStore.Close()
	payload := []byte("already copied payload")
	sourceStore.put("source-bucket", "uploads/image.webp", payload)
	targetStore.put("target-bucket", "uploads/image.webp", payload)

	size, digest, err := copyAndVerifyObject(
		context.Background(),
		models.StorageMigration{SourceBucket: "source-bucket", TargetBucket: "target-bucket"},
		"uploads/image.webp",
		newFakeS3Client(t, sourceStore.URL),
		newFakeS3Client(t, targetStore.URL),
	)
	if err != nil {
		t.Fatalf("matching target was rejected: %v", err)
	}
	if size != int64(len(payload)) || len(digest) != 64 {
		t.Fatalf("unexpected verification result: size=%d digest=%q", size, digest)
	}
}

func TestMigrationTargetConfigNormalizesR2(t *testing.T) {
	cfg, err := migrationTargetConfig(CreateStorageMigrationRequest{
		TargetType:     " R2 ",
		Endpoint:       "https://example.r2.cloudflarestorage.com/",
		Region:         "wrong-region",
		Bucket:         "images",
		AccessKey:      "access",
		SecretKey:      "secret",
		ForcePathStyle: true,
	})
	if err != nil {
		t.Fatalf("migrationTargetConfig returned error: %v", err)
	}
	if cfg.Type != "r2" || cfg.Region != "auto" || cfg.ForcePathStyle {
		t.Fatalf("unexpected R2 config: %+v", cfg)
	}
	if cfg.Endpoint != "https://example.r2.cloudflarestorage.com" {
		t.Fatalf("endpoint was not normalized: %q", cfg.Endpoint)
	}
}

func TestMigrationTargetConfigDefaultsS3Region(t *testing.T) {
	cfg, err := migrationTargetConfig(CreateStorageMigrationRequest{
		TargetType: "s3",
		Endpoint:   "https://s3.example.com",
		Bucket:     "images",
		AccessKey:  "access",
		SecretKey:  "secret",
	})
	if err != nil {
		t.Fatalf("migrationTargetConfig returned error: %v", err)
	}
	if cfg.Region != "us-east-1" {
		t.Fatalf("region=%q, want us-east-1", cfg.Region)
	}
}

func TestCompleteStorageMigrationSwitchesOnlyAfterCoverage(t *testing.T) {
	db := newStorageMigrationTestDB(t)
	setting := models.Settings{
		ID:          1,
		StorageType: "s3",
		S3Endpoint:  "https://old.example.com",
		S3Region:    "us-east-1",
		S3AccessKey: "old-access",
		S3SecretKey: "old-secret",
		S3Bucket:    "old-images",
	}
	if err := db.Create(&setting).Error; err != nil {
		t.Fatal(err)
	}
	image := models.Image{
		Url:       "/uploads/2026/08/image.webp",
		Thumbnail: "/uploads/2026/08/thumbnails/image.webp",
		FileName:  "image.webp",
		FileSize:  128,
		Storage:   "s3",
		UserId:    1,
	}
	if err := db.Create(&image).Error; err != nil {
		t.Fatal(err)
	}
	migration := models.StorageMigration{
		Status:          models.StorageMigrationRunning,
		SourceType:      "s3",
		SourceEndpoint:  setting.S3Endpoint,
		SourceRegion:    setting.S3Region,
		SourceAccessKey: setting.S3AccessKey,
		SourceSecretKey: setting.S3SecretKey,
		SourceBucket:    setting.S3Bucket,
		TargetType:      "r2",
		TargetEndpoint:  "https://account.r2.cloudflarestorage.com",
		TargetRegion:    "auto",
		TargetAccessKey: "new-access",
		TargetSecretKey: "new-secret",
		TargetBucket:    "new-images",
	}
	if err := db.Create(&migration).Error; err != nil {
		t.Fatal(err)
	}
	if err := enqueueMigrationItems(db, &migration); err != nil {
		t.Fatal(err)
	}

	if err := completeStorageMigration(db, &migration); err == nil {
		t.Fatal("cutover succeeded before migration items completed")
	}
	var unchanged models.Settings
	if err := db.First(&unchanged, 1).Error; err != nil {
		t.Fatal(err)
	}
	if unchanged.StorageType != "s3" {
		t.Fatalf("storage switched early to %q", unchanged.StorageType)
	}

	if err := db.Model(&models.StorageMigrationItem{}).
		Where("migration_id = ?", migration.ID).
		Updates(map[string]any{"status": models.StorageMigrationItemCompleted, "size": 64}).Error; err != nil {
		t.Fatal(err)
	}
	if err := refreshMigrationCounters(db, migration.ID); err != nil {
		t.Fatal(err)
	}
	if err := completeStorageMigration(db, &migration); err != nil {
		t.Fatalf("cutover failed after all items completed: %v", err)
	}

	var updatedSetting models.Settings
	if err := db.First(&updatedSetting, 1).Error; err != nil {
		t.Fatal(err)
	}
	if updatedSetting.StorageType != "r2" || updatedSetting.R2Bucket != "new-images" || updatedSetting.R2SecretKey != "new-secret" {
		t.Fatalf("target settings were not applied: %+v", updatedSetting)
	}
	var updatedImage models.Image
	if err := db.First(&updatedImage, image.Id).Error; err != nil {
		t.Fatal(err)
	}
	if updatedImage.Storage != "r2" {
		t.Fatalf("image storage=%q, want r2", updatedImage.Storage)
	}
	var completed models.StorageMigration
	if err := db.First(&completed, migration.ID).Error; err != nil {
		t.Fatal(err)
	}
	if completed.Status != models.StorageMigrationCompleted || completed.TargetSecretKey != "" || completed.SourceSecretKey != "" {
		t.Fatalf("migration was not completed and scrubbed: %+v", completed)
	}
	if completed.CopiedObjects != 2 || completed.CopiedBytes != 128 {
		t.Fatalf("unexpected counters: copied=%d bytes=%d", completed.CopiedObjects, completed.CopiedBytes)
	}
}

func TestMigrationObjectKey(t *testing.T) {
	tests := map[string]string{
		"/uploads/2026/08/image.webp":                    "uploads/2026/08/image.webp",
		"https://img.example.com/uploads/2026/08/a.webp": "uploads/2026/08/a.webp",
		"": "",
	}
	for input, want := range tests {
		if got := migrationObjectKey(input); got != want {
			t.Errorf("migrationObjectKey(%q)=%q, want %q", input, got, want)
		}
	}
}

func newStorageMigrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&models.Settings{},
		&models.Image{},
		&models.StorageMigration{},
		&models.StorageMigrationItem{},
	); err != nil {
		t.Fatal(err)
	}
	return db
}

type fakeS3Store struct {
	*httptest.Server
	mu         sync.Mutex
	objects    map[string][]byte
	corruptPut bool
}

func newFakeS3Store(t *testing.T, corruptPut bool) *fakeS3Store {
	t.Helper()
	store := &fakeS3Store{objects: make(map[string][]byte), corruptPut: corruptPut}
	store.Server = httptest.NewServer(http.HandlerFunc(store.serveHTTP))
	return store
}

func (store *fakeS3Store) serveHTTP(w http.ResponseWriter, r *http.Request) {
	bucket, key, ok := strings.Cut(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if !ok || bucket == "" || key == "" {
		http.Error(w, "invalid object path", http.StatusBadRequest)
		return
	}
	objectID := bucket + "/" + key
	switch r.Method {
	case http.MethodPut:
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if store.corruptPut && len(data) > 0 {
			data[0] ^= 0xff
		}
		store.mu.Lock()
		store.objects[objectID] = append([]byte(nil), data...)
		store.mu.Unlock()
		w.Header().Set("ETag", `"fake-etag"`)
		w.WriteHeader(http.StatusOK)
	case http.MethodGet:
		store.mu.Lock()
		data, exists := store.objects[objectID]
		data = append([]byte(nil), data...)
		store.mu.Unlock()
		if !exists {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/webp")
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	case http.MethodDelete:
		store.mu.Lock()
		delete(store.objects, objectID)
		store.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "unsupported method", http.StatusMethodNotAllowed)
	}
}

func (store *fakeS3Store) put(bucket, key string, data []byte) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.objects[bucket+"/"+key] = append([]byte(nil), data...)
}

func (store *fakeS3Store) get(bucket, key string) []byte {
	store.mu.Lock()
	defer store.mu.Unlock()
	data, ok := store.objects[bucket+"/"+key]
	if !ok {
		return nil
	}
	return append([]byte(nil), data...)
}

func newFakeS3Client(t *testing.T, endpoint string) *awss3.Client {
	t.Helper()
	client, err := s3util.NewClient(context.Background(), s3util.ClientConfig{
		Type: "s3", Endpoint: endpoint, Region: "us-east-1", Bucket: "unused",
		AccessKey: "access", SecretKey: "secret", ForcePathStyle: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}
