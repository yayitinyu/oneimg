package controllers

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"oneimg/backend/database"
	"oneimg/backend/models"
	"oneimg/backend/utils/result"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateInvitationsRequest struct {
	Count int `json:"count"`
}

func ListInvitations(c *gin.Context) {
	var codes []models.InvitationCode
	if err := database.GetDB().DB.Order("created_at DESC").Find(&codes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "获取邀请码失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success("获取成功", codes))
}

func CreateInvitations(c *gin.Context) {
	var req CreateInvitationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, "请求参数错误"))
		return
	}
	if req.Count == 0 {
		req.Count = 1
	}
	if req.Count < 1 || req.Count > 20 {
		c.JSON(http.StatusBadRequest, result.Error(400, "一次可生成 1–20 个邀请码"))
		return
	}

	db := database.GetDB()
	if db == nil || db.DB == nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "数据库连接失败"))
		return
	}

	plainCodes := make([]string, 0, req.Count)
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Collisions are cryptographically unlikely; the bound prevents an
		// unexpected database error from turning into an infinite request.
		maxAttempts := req.Count * 3
		for attempts := 0; attempts < maxAttempts && len(plainCodes) < req.Count; attempts++ {
			code, err := generateInvitationCode()
			if err != nil {
				return err
			}
			record := models.InvitationCode{
				CodeHash: hashInvitationCode(code),
				Hint:     code[len(code)-4:],
			}
			if err := tx.Create(&record).Error; err != nil {
				message := strings.ToLower(err.Error())
				if strings.Contains(message, "unique") || strings.Contains(message, "duplicate") {
					continue
				}
				return err
			}
			plainCodes = append(plainCodes, code)
		}
		if len(plainCodes) != req.Count {
			return fmt.Errorf("无法生成足够的不重复邀请码")
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "生成邀请码失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success("邀请码已生成，请立即保存", map[string]any{
		"codes": plainCodes,
	}))
}

func DeleteInvitation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, result.Error(400, "邀请码 ID 无效"))
		return
	}
	deleteResult := database.GetDB().DB.
		Where("id = ? AND used_at IS NULL", id).
		Delete(&models.InvitationCode{})
	if deleteResult.Error != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "删除邀请码失败"))
		return
	}
	if deleteResult.RowsAffected == 0 {
		c.JSON(http.StatusConflict, result.Error(409, "邀请码不存在或已被使用"))
		return
	}
	c.JSON(http.StatusOK, result.Success("邀请码已删除", nil))
}

func generateInvitationCode() (string, error) {
	bytes := make([]byte, 10)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	raw := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)
	raw = strings.ToUpper(raw)
	return raw[:4] + "-" + raw[4:8] + "-" + raw[8:12] + "-" + raw[12:16], nil
}
