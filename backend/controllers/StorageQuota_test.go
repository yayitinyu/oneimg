package controllers

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"oneimg/backend/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestStorageQuotaExceeded(t *testing.T) {
	tests := []struct {
		name                  string
		used, incoming, quota int64
		want                  bool
	}{
		{name: "unlimited", used: 500, incoming: 500, quota: 0, want: false},
		{name: "under limit", used: 400, incoming: 500, quota: 1000, want: false},
		{name: "exact fit", used: 400, incoming: 600, quota: 1000, want: false},
		{name: "over limit", used: 400, incoming: 601, quota: 1000, want: true},
		{name: "already full", used: 1000, incoming: 1, quota: 1000, want: true},
		{name: "invalid negative size", used: 0, incoming: -1, quota: 1000, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := storageQuotaExceeded(test.used, test.incoming, test.quota); got != test.want {
				t.Fatalf("storageQuotaExceeded(%d, %d, %d)=%v, want %v", test.used, test.incoming, test.quota, got, test.want)
			}
		})
	}
}

func TestValidateUserStorageQuota(t *testing.T) {
	if err := validateSettingData("user_storage_quota", int64(0)); err != nil {
		t.Fatalf("zero should mean unlimited: %v", err)
	}
	if err := validateSettingData("user_storage_quota", float64(5*1024*1024*1024)); err != nil {
		t.Fatalf("valid quota rejected: %v", err)
	}
	if err := validateSettingData("user_storage_quota", -1); err == nil {
		t.Fatal("negative quota should be rejected")
	}
}

func TestPersistUploadedImagesEnforcesQuotaAtomically(t *testing.T) {
	dsn := fmt.Sprintf("file:quota-%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "requires cgo") {
			t.Skip("SQLite integration test requires CGO")
		}
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Image{}); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	user := models.User{Username: "quota-user", Password: "not-used", Role: models.RoleUser}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	existing := models.Image{Url: "/existing", FileName: "existing.png", FileSize: 400, UserId: user.Id}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("create existing image: %v", err)
	}

	batch := []*models.Image{
		{Url: "/first", FileName: "first.png", FileSize: 350, UserId: user.Id},
		{Url: "/second", FileName: "second.png", FileSize: 251, UserId: user.Id},
	}
	err = persistUploadedImages(db, user.Id, models.RoleUser, 1000, batch)
	if !errors.Is(err, errUserStorageQuotaExceeded) {
		t.Fatalf("persist over quota error=%v, want quota error", err)
	}

	var count int64
	if err := db.Model(&models.Image{}).Where("user_id = ?", user.Id).Count(&count).Error; err != nil {
		t.Fatalf("count images: %v", err)
	}
	if count != 1 {
		t.Fatalf("over-quota batch was partially persisted: count=%d", count)
	}

	exactFit := []*models.Image{{Url: "/exact", FileName: "exact.png", FileSize: 600, UserId: user.Id}}
	if err := persistUploadedImages(db, user.Id, models.RoleUser, 1000, exactFit); err != nil {
		t.Fatalf("exact-fit upload rejected: %v", err)
	}
	used, err := userStorageUsage(db, user.Id)
	if err != nil {
		t.Fatalf("read usage: %v", err)
	}
	if used != 1000 {
		t.Fatalf("usage=%d, want 1000", used)
	}
}
