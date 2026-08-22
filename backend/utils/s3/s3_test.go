package s3

import (
	"testing"

	"oneimg/backend/models"
)

func TestConfigFromSettingsSelectsIndependentBackends(t *testing.T) {
	setting := models.Settings{
		StorageType: "s3",
		S3Endpoint:  "https://s3.example.com",
		S3Region:    "ap-east-1",
		S3AccessKey: "s3-access",
		S3SecretKey: "s3-secret",
		S3Bucket:    "s3-bucket",
		S3PathStyle: true,
		R2Endpoint:  "https://account.r2.cloudflarestorage.com",
		R2AccessKey: "r2-access",
		R2SecretKey: "r2-secret",
		R2Bucket:    "r2-bucket",
	}

	s3Config := ConfigFromSettings(setting, "s3")
	if s3Config.Region != "ap-east-1" || !s3Config.ForcePathStyle || s3Config.Bucket != "s3-bucket" {
		t.Fatalf("unexpected S3 config: %+v", s3Config)
	}
	r2Config := ConfigFromSettings(setting, "r2")
	if r2Config.Region != "auto" || r2Config.ForcePathStyle || r2Config.Bucket != "r2-bucket" {
		t.Fatalf("unexpected R2 config: %+v", r2Config)
	}
}

func TestClientConfigValidate(t *testing.T) {
	valid := ClientConfig{
		Type: "s3", Endpoint: "https://s3.example.com", Region: "us-east-1",
		AccessKey: "access", SecretKey: "secret", Bucket: "images",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	valid.SecretKey = ""
	if err := valid.Validate(); err == nil {
		t.Fatal("config without a secret key was accepted")
	}
}
