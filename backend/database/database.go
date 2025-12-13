package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"oneimg/backend/config"
	"oneimg/backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 数据库操作类，支持 PostgreSQL + MySQL + SQLite3
type Database struct {
	DB     *gorm.DB
	DBType string // 当前数据库类型: "postgresql", "mysql", "sqlite"
}

var db *Database

// NewDB 创建新的数据库连接
func NewDB(dialector gorm.Dialector) (*Database, error) {
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Database{DB: gormDB}, nil
}

// GetDB 获取数据库实例
func GetDB() *Database {
	return db
}

// InitDB 初始化数据库连接
// 优先级：PostgreSQL > MySQL > SQLite
func InitDB(cfg *config.Config) {
	var err error
	var dialector gorm.Dialector
	var dbType string

	// 根据配置选择数据库类型（优先级：PostgreSQL > MySQL > SQLite）
	if cfg.IsPostgres {
		// 使用 PostgreSQL
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
			cfg.PostgresHost,
			cfg.PostgresPort,
			cfg.PostgresUser,
			cfg.PostgresPassword,
			cfg.PostgresDB)
		dialector = postgres.Open(dsn)
		dbType = "postgresql"
		log.Println("使用 PostgreSQL 数据库")
	} else if cfg.IsMysql {
		// 使用 MySQL
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DbUser,
			cfg.DbPassword,
			cfg.DbHost,
			cfg.DbPort,
			cfg.DbName)
		dialector = mysql.Open(dsn)
		dbType = "mysql"
		log.Println("使用 MySQL 数据库")
	} else {
		// 使用 SQLite
		ensureDirExists(cfg.SqlitePath)
		dialector = sqlite.Open(cfg.SqlitePath)
		dbType = "sqlite"
		log.Printf("使用 SQLite 数据库: %s", cfg.SqlitePath)
	}

	gormDB, gormErr := gorm.Open(dialector, &gorm.Config{})
	if gormErr != nil {
		log.Fatal("数据库连接失败:", gormErr)
	}

	db = &Database{DB: gormDB, DBType: dbType}

	log.Println("数据库连接成功")

	// 自动迁移数据表
	err = db.DB.AutoMigrate(&models.User{}, &models.Image{}, &models.Settings{}, &models.ImageTeleGram{})
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	log.Println("数据库表迁移完成")
}

// 辅助函数，如果数据库目录不存在则创建
func ensureDirExists(path string) {
	dir := filepath.Dir(path)
	// 检查目录状态
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// 创建目录（0755：生产环境安全权限）
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("创建数据库目录失败 [路径: %s]: %v", dir, err)
		}
		return
	}

	// 处理其他错误（如权限不足）
	if err != nil {
		log.Fatalf("检查数据库目录失败 [路径: %s]: %v", dir, err)
	}
}
