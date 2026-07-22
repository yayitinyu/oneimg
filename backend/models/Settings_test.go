package models

import (
	"testing"
	"time"
)

func TestGetTGStorageTarget(t *testing.T) {
	tests := []struct {
		name     string
		settings Settings
		want     string
	}{
		{"channel takes priority", Settings{TGChannelID: " -10099 ", TGReceivers: "111,222"}, "-10099"},
		{"first receiver fallback", Settings{TGReceivers: " 111, 222 "}, "111"},
		{"empty configuration", Settings{}, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.settings.GetTGStorageTarget(); got != test.want {
				t.Fatalf("GetTGStorageTarget()=%q, want %q", got, test.want)
			}
		})
	}
}

func TestImageIsExpired(t *testing.T) {
	now := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Second)
	future := now.Add(time.Second)

	tests := []struct {
		name  string
		image Image
		want  bool
	}{
		{"permanent image", Image{}, false},
		{"past expiration", Image{ExpiresAt: &past}, true},
		{"exact expiration", Image{ExpiresAt: &now}, true},
		{"future expiration", Image{ExpiresAt: &future}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.image.IsExpired(now); got != test.want {
				t.Fatalf("IsExpired()=%v, want %v", got, test.want)
			}
		})
	}
}

func TestIsValidTelegramStorageConfig(t *testing.T) {
	settings := Settings{StorageType: "telegram", TGBotToken: "token", TGChannelID: "-100123"}
	if !settings.IsValidStorageConfig() {
		t.Fatal("channel-only Telegram storage configuration should be valid")
	}
}
