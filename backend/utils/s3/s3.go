package s3

import (
	"context"
	"fmt"
	"strings"

	"oneimg/backend/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type ClientConfig struct {
	Type           string
	Endpoint       string
	Region         string
	AccessKey      string
	SecretKey      string
	Bucket         string
	ForcePathStyle bool
}

func ConfigFromSettings(setting models.Settings, storageType string) ClientConfig {
	storageType = strings.ToLower(strings.TrimSpace(storageType))
	if storageType == "r2" {
		return ClientConfig{
			Type:      "r2",
			Endpoint:  setting.R2Endpoint,
			Region:    "auto",
			AccessKey: setting.R2AccessKey,
			SecretKey: setting.R2SecretKey,
			Bucket:    setting.R2Bucket,
		}
	}
	region := strings.TrimSpace(setting.S3Region)
	if region == "" {
		region = "us-east-1"
	}
	return ClientConfig{
		Type:           "s3",
		Endpoint:       setting.S3Endpoint,
		Region:         region,
		AccessKey:      setting.S3AccessKey,
		SecretKey:      setting.S3SecretKey,
		Bucket:         setting.S3Bucket,
		ForcePathStyle: setting.S3PathStyle,
	}
}

func (cfg ClientConfig) Validate() error {
	if cfg.Type != "s3" && cfg.Type != "r2" {
		return fmt.Errorf("不支持的对象存储类型：%s", cfg.Type)
	}
	if strings.TrimSpace(cfg.AccessKey) == "" || strings.TrimSpace(cfg.SecretKey) == "" {
		return fmt.Errorf("S3/R2密钥为空")
	}
	if strings.TrimSpace(cfg.Bucket) == "" || strings.TrimSpace(cfg.Endpoint) == "" {
		return fmt.Errorf("S3/R2配置缺失 [bucket:%s, endpoint:%s]", cfg.Bucket, cfg.Endpoint)
	}
	return nil
}

func NewClient(ctx context.Context, cfg ClientConfig) (*awss3.Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	region := strings.TrimSpace(cfg.Region)
	if cfg.Type == "r2" {
		region = "auto"
	} else if region == "" {
		region = "us-east-1"
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			strings.TrimSpace(cfg.AccessKey),
			strings.TrimSpace(cfg.SecretKey),
			"", // Token
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("加载 AWS 配置失败: %w", err)
	}

	client := awss3.NewFromConfig(awsCfg, func(options *awss3.Options) {
		options.BaseEndpoint = aws.String(strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/"))
		options.UsePathStyle = cfg.ForcePathStyle
	})
	return client, nil
}

// NewS3Client keeps existing upload/read call sites compatible while all new
// migration code can construct independent source and target clients.
func NewS3Client(setting models.Settings) (*awss3.Client, error) {
	return NewClient(context.Background(), ConfigFromSettings(setting, setting.GetEffectiveStorageType()))
}

func GetObject(client awss3.Client, ctx context.Context, input *awss3.GetObjectInput) (*awss3.GetObjectOutput, error) {
	return client.GetObject(ctx, input)
}

func DeleteObject(client awss3.Client, ctx context.Context, input *awss3.DeleteObjectInput) (*awss3.DeleteObjectOutput, error) {
	return client.DeleteObject(ctx, input)
}
