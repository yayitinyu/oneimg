package controllers

import (
	"net/http"
	"testing"
	"time"

	"oneimg/backend/models"
	md5util "oneimg/backend/utils/md5"

	"github.com/gin-gonic/gin"
)

func TestParseImageExpiration(t *testing.T) {
	now := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		value        string
		wantDuration time.Duration
		wantNil      bool
		wantError    bool
	}{
		{value: "never", wantNil: true},
		{value: "", wantNil: true},
		{value: "1h", wantDuration: time.Hour},
		{value: "7D", wantDuration: 7 * 24 * time.Hour},
		{value: "forever-ish", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			expiresAt, err := parseImageExpiration(test.value, now)
			if (err != nil) != test.wantError {
				t.Fatalf("parseImageExpiration() error=%v, wantError=%v", err, test.wantError)
			}
			if test.wantError {
				return
			}
			if test.wantNil {
				if expiresAt != nil {
					t.Fatalf("parseImageExpiration()=%v, want nil", expiresAt)
				}
				return
			}
			if expiresAt == nil || !expiresAt.Equal(now.Add(test.wantDuration)) {
				t.Fatalf("parseImageExpiration()=%v, want %v", expiresAt, now.Add(test.wantDuration))
			}
		})
	}
}

func TestValidateRegistrationCredentials(t *testing.T) {
	if username, err := validateRegistrationCredentials("  Sakura_23  ", "secret12"); err != nil || username != "Sakura_23" {
		t.Fatalf("valid credentials rejected: username=%q error=%v", username, err)
	}
	if _, err := validateRegistrationCredentials("樱花用户", "secret12"); err != nil {
		t.Fatalf("Unicode username rejected: %v", err)
	}
	if _, err := validateRegistrationCredentials("樱花用户", "樱花密码安全"); err != nil {
		t.Fatalf("valid Unicode password rejected: %v", err)
	}

	invalid := []struct {
		username string
		password string
	}{
		{"ab", "secret12"},
		{"guest", "secret12"},
		{"GUEST_archive", "secret12"},
		{"name with space", "secret12"},
		{"valid-name", "short"},
		{"valid-name", "五个字符短"},
		{"valid-name", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}
	for _, test := range invalid {
		if _, err := validateRegistrationCredentials(test.username, test.password); err == nil {
			t.Fatalf("invalid credentials accepted: username=%q", test.username)
		}
	}
}

func TestImageCacheHeadersRespectExpiration(t *testing.T) {
	now := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)

	cacheControl, expiresHeader := imageCacheHeaders(nil, now)
	if cacheControl != "public, max-age=31536000, immutable" || expiresHeader != "" {
		t.Fatalf("permanent cache headers = %q, %q", cacheControl, expiresHeader)
	}

	expiresAt := now.Add(90 * time.Second)
	cacheControl, expiresHeader = imageCacheHeaders(&expiresAt, now)
	if cacheControl != "public, max-age=90, must-revalidate" {
		t.Fatalf("temporary cache control = %q", cacheControl)
	}
	if expiresHeader != expiresAt.Format(http.TimeFormat) {
		t.Fatalf("expires header = %q, want %q", expiresHeader, expiresAt.Format(http.TimeFormat))
	}

	expiredAt := now.Add(-time.Second)
	cacheControl, _ = imageCacheHeaders(&expiredAt, now)
	if cacheControl != "no-store" {
		t.Fatalf("expired cache control = %q", cacheControl)
	}
}

func TestCheckImageAccessPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fileName := "image.webp"

	t.Run("administrator", func(t *testing.T) {
		context, _ := gin.CreateTestContext(nil)
		context.Set("user_role", models.RoleAdmin)
		if !CheckImageAccessPermission(context, models.Image{UserId: 999}) {
			t.Fatal("administrator should be allowed")
		}
	})

	t.Run("registered user owns only matching user id", func(t *testing.T) {
		context, _ := gin.CreateTestContext(nil)
		context.Set("user_role", models.RoleUser)
		context.Set("user_id", 42)
		if !CheckImageAccessPermission(context, models.Image{UserId: 42}) {
			t.Fatal("owner should be allowed")
		}
		if CheckImageAccessPermission(context, models.Image{UserId: 43}) {
			t.Fatal("different registered user should be denied")
		}
	})

	t.Run("guest requires uuid and signature", func(t *testing.T) {
		context, _ := gin.CreateTestContext(nil)
		context.Set("user_role", models.RoleGuest)
		context.Set("is_guest", true)
		context.Set("username", "guest-uuid")
		owned := models.Image{UUID: "guest-uuid", FileName: fileName, MD5: md5util.Md5("guest-uuid" + fileName)}
		if !CheckImageAccessPermission(context, owned) {
			t.Fatal("matching guest image should be allowed")
		}
		owned.MD5 = "invalid"
		if CheckImageAccessPermission(context, owned) {
			t.Fatal("guest image with invalid signature should be denied")
		}
	})
}

func TestStorageObjectKey(t *testing.T) {
	tests := map[string]string{
		"/uploads/2026/07/image.webp":                    "uploads/2026/07/image.webp",
		"https://cdn.example.com/uploads/old/image.webp": "uploads/old/image.webp",
	}
	for input, want := range tests {
		if got := storageObjectKey(input); got != want {
			t.Fatalf("storageObjectKey(%q)=%q, want %q", input, got, want)
		}
	}
}
