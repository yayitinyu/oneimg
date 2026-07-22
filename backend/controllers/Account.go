package controllers

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"oneimg/backend/database"
	"oneimg/backend/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ChangeAccountInfoRequest 修改登录信息请求结构
type ChangeAccountInfoRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password"`
	NewUsername     string `json:"new_username"`
}

// AccountResponse 账户响应结构
type AccountResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

var uuidRegex = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

// ChangeAccountInfo 修改密码
func ChangeAccountInfo(c *gin.Context) {
	if c.GetBool("is_guest") {
		c.JSON(http.StatusForbidden, AccountResponse{
			Code: 403, Message: "游客不能修改账户信息", Success: false,
		})
		return
	}
	var req ChangeAccountInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AccountResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
			Success: false,
		})
		return
	}

	// 获取当前用户ID
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, AccountResponse{
			Code:    401,
			Message: "未登录",
			Success: false,
		})
		return
	}

	// 获取数据库实例
	db := database.GetDB().DB

	// 查找用户
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, AccountResponse{
			Code:    404,
			Message: "用户不存在",
			Success: false,
		})
		return
	}

	// 验证当前密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusBadRequest, AccountResponse{
			Code:    400,
			Message: "当前密码错误",
			Success: false,
		})
		return
	}

	if req.NewUsername == "" && req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, AccountResponse{
			Code:    400,
			Message: "请至少修改一项",
			Success: false,
		})
		return
	}

	normalizedUsername := ""
	if req.NewUsername != "" {
		var err error
		normalizedUsername, err = validateAccountUsername(req.NewUsername)
		if err != nil {
			c.JSON(http.StatusBadRequest, AccountResponse{Code: 400, Message: err.Error(), Success: false})
			return
		}
	}
	if req.NewPassword != "" {
		if err := validateAccountPassword(req.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, AccountResponse{Code: 400, Message: err.Error(), Success: false})
			return
		}
	}

	// 开启事务
	tx := db.Begin()
	if err := tx.Error; err != nil {
		c.JSON(http.StatusInternalServerError, AccountResponse{
			Code:    500,
			Message: "数据库操作失败: " + err.Error(),
			Success: false,
		})
		return
	}
	defer tx.Rollback()

	// 如果用户名存在修改用户
	if req.NewUsername != "" {
		var existingUser models.User
		lookupErr := tx.Where("LOWER(username) = LOWER(?) AND id != ?", normalizedUsername, userID).First(&existingUser).Error
		if lookupErr == nil {
			c.JSON(http.StatusBadRequest, AccountResponse{
				Code:    400,
				Message: "用户名已存在",
				Success: false,
			})
			return
		}
		if !errors.Is(lookupErr, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, AccountResponse{Code: 500, Message: "检查用户名失败", Success: false})
			return
		}

		// 更新用户名
		if err := tx.Model(&user).Update("username", normalizedUsername).Error; err != nil {
			c.JSON(http.StatusInternalServerError, AccountResponse{
				Code:    500,
				Message: "用户名更新失败",
				Success: false,
			})
			return
		}
	}

	if req.NewPassword != "" {
		// 加密新密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, AccountResponse{
				Code:    500,
				Message: "密码加密失败",
				Success: false,
			})
			// 回滚事务
			tx.Rollback()
			return
		}

		// 更新密码
		if err := tx.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, AccountResponse{
				Code:    500,
				Message: "密码更新失败",
				Success: false,
			})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, AccountResponse{
			Code:    500,
			Message: "数据库操作失败: " + err.Error(),
			Success: false,
		})
		return
	}

	// 退出登录
	session.Clear()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, AccountResponse{
			Code:    500,
			Message: "会话失效失败: " + err.Error(),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, AccountResponse{
		Code:    200,
		Message: "修改成功",
		Success: true,
	})
}

// isTouristUsername 辅助函数，检查是否为游客账号
func isTouristUsername(username string) bool {
	normalized := strings.ToLower(strings.TrimSpace(username))
	return strings.HasPrefix(normalized, "guest_") || normalized == "guest" || uuidRegex.MatchString(normalized)
}

// ClearAllSessions 清除所有会话
func ClearAllSessions(c *gin.Context) {
	// 获取当前session
	session := sessions.Default(c)

	// 清除当前session
	session.Clear()
	session.Save()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, AccountResponse{
			Code:    500,
			Message: "清除会话失败",
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, AccountResponse{
		Code:    200,
		Message: "所有会话已清除",
		Success: true,
	})
}

// 辅助函数，获取用户UUID
func GetUUID(c *gin.Context) string {
	uuidRegex := regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")
	if uuidRegex.MatchString(c.GetString("username")) {
		return c.GetString("username")
	} else if c.GetString("username") == "00000000-0000-0000-0000-000000000000" {
		return "00000000-0000-0000-0000-000000000000"
	} else {
		return c.GetString("username")
	}
}
