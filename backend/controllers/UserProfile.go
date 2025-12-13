package controllers

import (
	"net/http"

	"oneimg/backend/database"
	"oneimg/backend/models"
	"oneimg/backend/utils/result"

	"github.com/gin-gonic/gin"
)

// ProfileUpdateRequest 用户资料更新请求
type ProfileUpdateRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// GetUserProfile 获取用户资料
func GetUserProfile(c *gin.Context) {
	userId := c.GetInt("user_id")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, result.Error("未登录"))
		return
	}

	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, result.Error("数据库连接失败"))
		return
	}

	var user models.User
	if err := db.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, result.Error("用户不存在"))
		return
	}

	c.JSON(http.StatusOK, result.Success(map[string]any{
		"id":       user.Id,
		"username": user.Username,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"role":     user.Role,
	}))
}

// UpdateUserProfile 更新用户资料
func UpdateUserProfile(c *gin.Context) {
	userId := c.GetInt("user_id")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, result.Error("未登录"))
		return
	}

	var req ProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, result.Error("请求参数错误"))
		return
	}

	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, result.Error("数据库连接失败"))
		return
	}

	// 更新用户资料
	updates := map[string]interface{}{}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, result.Error("没有要更新的内容"))
		return
	}

	if err := db.DB.Model(&models.User{}).Where("id = ?", userId).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, result.Error("更新失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success("更新成功"))
}
