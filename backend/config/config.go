package config

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// 服务器配置
	Port string

	// Sqlite3数据库
	SqlitePath string

	// Mysql数据库
	IsMysql    bool
	DbHost     string
	DbPort     int
	DbUser     string
	DbPassword string
	DbName     string

	// PostgreSQL数据库
	IsPostgres       bool
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	// 上传文件配置
	MaxFileSize  int64
	AllowedTypes []string

	// 默认用户
	DefaultUser string
	DefaultPass string

	// JWT配置
	JWTSecret string

	// Session配置
	SessionSecret string
	SessionSecure bool

	// 跨域配置。留空时仅允许浏览器同源访问。
	CORSAllowedOrigins []string
}

// 全局配置实例
var App *Config

// 检查.env文件是否存在
func EnvExists() bool {
	_, err := os.Stat(getEnvFilePath())
	return !os.IsNotExist(err)
}

func getEnvFilePath() string {
	if path := strings.TrimSpace(os.Getenv("ONEIMG_ENV_FILE")); path != "" {
		return path
	}
	return ".env"
}

// 创建默认.env文件
func CreateDefaultEnv() {
	// 1. 创建data目录和配置文件目录
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("创建data目录失败：%v", err)
	}
	envPath := getEnvFilePath()
	if envDir := filepath.Dir(envPath); envDir != "." {
		if err := os.MkdirAll(envDir, 0700); err != nil {
			log.Fatalf("创建配置目录失败：%v", err)
		}
	}

	// 2. 生成随机的SESSION_SECRET（32位base64编码）
	sessionSecret := generateRandomSecret(32)

	// 3. 直接定义.env模板内容
	envTemplate := `# 服务器配置
SERVER_PORT=8080

# 数据库配置（优先级：PostgreSQL > MySQL > SQLite）
SQLITE_PATH=./data/data.db

# MySQL配置
IS_MYSQL=false
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=oneimg

# PostgreSQL配置
IS_POSTGRES=false
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=
POSTGRES_DB=oneimg

# 文件上传配置
MAX_FILE_SIZE=10485760
ALLOWED_TYPES=image/jpeg,image/png,image/gif,image/webp

# 默认用户配置
DEFAULT_USER=admin
DEFAULT_PASS=123456

# Session配置
SESSION_SECRET=
SESSION_SECURE=false

# 跨域来源（逗号分隔；留空表示仅同源访问）
CORS_ALLOWED_ORIGINS=
`

	// 4. 替换模板中的SESSION_SECRET占位符
	envContent := strings.Replace(envTemplate, "SESSION_SECRET=", "SESSION_SECRET="+sessionSecret, 1)

	// 5. 写入.env文件
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		log.Fatalf("生成默认.env文件失败：%v", err)
	}

	log.Printf("✅ 首次启动：自动生成.env文件（路径：%s）", envPath)
}

// 生成指定长度的随机密钥（base64编码）
func generateRandomSecret(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Fatalf("生成随机密钥失败：%v", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// 初始化配置（优先加载外部.env，无则生成默认）
func NewConfig() {
	// 1. 检查.env文件，不存在则生成
	if !EnvExists() {
		CreateDefaultEnv()
	}

	// 2. 加载.env文件（此时必存在）
	err := godotenv.Load(getEnvFilePath())
	if err != nil {
		log.Fatalf("加载.env文件失败：%v", err)
	}

	// 3. 解析配置项
	maxFileSize := getEnvInt64("MAX_FILE_SIZE", 10485760)
	allowedTypes := splitAndTrim(getEnv("ALLOWED_TYPES", "image/jpeg,image/png,image/gif,image/webp"))
	port := getEnv("SERVER_PORT", getEnv("PORT", "8080"))

	// Sqlite3配置
	sqlitePath := getEnv("SQLITE_PATH", "./data/data.db")
	// 确保SQLite目录存在
	if err := os.MkdirAll(filepath.Dir(sqlitePath), 0755); err != nil {
		log.Printf("警告：创建SQLite目录失败：%v", err)
	}

	// Mysql配置
	isMysql := getEnv("IS_MYSQL", "false") == "true"
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnvInt("DB_PORT", 3306)
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "oneimg")

	// PostgreSQL配置
	isPostgres := getEnv("IS_POSTGRES", "false") == "true"
	postgresHost := getEnv("POSTGRES_HOST", "localhost")
	postgresPort := getEnvInt("POSTGRES_PORT", 5432)
	postgresUser := getEnv("POSTGRES_USER", "postgres")
	postgresPassword := getEnv("POSTGRES_PASSWORD", "")
	postgresDB := getEnv("POSTGRES_DB", "oneimg")

	// 默认用户配置
	defaultUser := getEnv("DEFAULT_USER", "admin")
	defaultPass := getEnv("DEFAULT_PASS", "123456")

	// JWT配置（默认生成随机密钥，避免硬编码）
	jwtSecret := getEnv("JWT_SECRET", generateRandomSecret(32))

	// Session配置（读取.env中的值，无则生成）
	sessionSecret := getEnv("SESSION_SECRET", generateRandomSecret(32))
	sessionSecure := getEnvBool("SESSION_SECURE", false)
	corsAllowedOrigins := splitAndTrim(getEnv("CORS_ALLOWED_ORIGINS", ""))

	// 初始化全局配置
	App = &Config{
		Port:               port,
		SqlitePath:         sqlitePath,
		IsMysql:            isMysql,
		DbHost:             dbHost,
		DbPort:             dbPort,
		DbUser:             dbUser,
		DbPassword:         dbPassword,
		DbName:             dbName,
		IsPostgres:         isPostgres,
		PostgresHost:       postgresHost,
		PostgresPort:       postgresPort,
		PostgresUser:       postgresUser,
		PostgresPassword:   postgresPassword,
		PostgresDB:         postgresDB,
		MaxFileSize:        maxFileSize,
		AllowedTypes:       allowedTypes,
		DefaultUser:        defaultUser,
		DefaultPass:        defaultPass,
		JWTSecret:          jwtSecret,
		SessionSecret:      sessionSecret,
		SessionSecure:      sessionSecure,
		CORSAllowedOrigins: corsAllowedOrigins,
	}

	log.Println("✅ 配置初始化完成")
}

func getEnvInt(key string, defaultValue int) int {
	value := getEnv(key, strconv.Itoa(defaultValue))
	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("警告：%s=%q 不是有效整数，使用默认值 %d", key, value, defaultValue)
		return defaultValue
	}
	return parsed
}

func getEnvInt64(key string, defaultValue int64) int64 {
	value := getEnv(key, strconv.FormatInt(defaultValue, 10))
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		log.Printf("警告：%s=%q 不是有效正整数，使用默认值 %d", key, value, defaultValue)
		return defaultValue
	}
	return parsed
}

func getEnvBool(key string, defaultValue bool) bool {
	value := getEnv(key, strconv.FormatBool(defaultValue))
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("警告：%s=%q 不是有效布尔值，使用默认值 %t", key, value, defaultValue)
		return defaultValue
	}
	return parsed
}

func splitAndTrim(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// 获取环境变量（带默认值）
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
