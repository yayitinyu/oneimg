package controllers

import (
	"net/http"

	"oneimg/backend/database"
	"oneimg/backend/utils/result"

	"github.com/gin-gonic/gin"
)

// DatabaseStatusResponse 数据库状态响应
type DatabaseStatusResponse struct {
	Type      string `json:"type"`       // 数据库类型: postgresql, mysql, sqlite
	Connected bool   `json:"connected"`  // 是否已连接
}

// GetDatabaseStatus 获取数据库连接状态
func GetDatabaseStatus(c *gin.Context) {
	db := database.GetDB()
	
	if db == nil {
		c.JSON(http.StatusOK, result.Error("数据库未初始化"))
		return
	}

	// 检查数据库连接
	sqlDB, err := db.DB.DB()
	connected := err == nil && sqlDB.Ping() == nil

	c.JSON(http.StatusOK, result.Success(DatabaseStatusResponse{
		Type:      db.DBType,
		Connected: connected,
	}))
}
