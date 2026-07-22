package controllers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"oneimg/backend/config"
	"oneimg/backend/database"
	"oneimg/backend/models"
	"oneimg/backend/utils/result"
	"oneimg/backend/utils/settings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// 登录请求结构
type LoginRequest struct {
	Username       string         `json:"username" binding:"required"`
	Password       string         `json:"password" binding:"required"`
	TurnstileToken string         `json:"turnstileToken"`
	FusionHash     string         `json:"fusionHash"`
	StableFeatures map[string]any `json:"stableFeatures"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Token string       `json:"token,omitempty"`
	User  *models.User `json:"user,omitempty"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, result.Error(
			400,
			"请求参数错误",
		))
		return
	}

	// 获取数据库实例
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "数据库连接失败"))
		return
	}

	// 获取系统设置 (使用统一的辅助函数，支持环境变量覆盖)
	sysSettings, err := settings.GetSettings()
	if err != nil {
		// 区分记录不存在和数据库错误
		if strings.Contains(err.Error(), "record not found") {
			c.JSON(http.StatusInternalServerError, result.Error(500, "系统配置未初始化"))
		} else {
			c.JSON(http.StatusInternalServerError, result.Error(500, "配置信息查询失败"))
		}
		return
	}

	// 检查是否开启了 Turnstile 验证
	if sysSettings.Turnstile {
		if req.TurnstileToken == "" {
			c.JSON(http.StatusBadRequest, result.Error(400, "请完成人机验证"))
			return
		}
		if !ValidateTurnstileToken(req.TurnstileToken, c.ClientIP()) {
			c.JSON(http.StatusBadRequest, result.Error(400, "人机验证失败，请重试"))
			return
		}
	}

	// 普通用户登录逻辑
	var user models.User
	userInfo := db.DB.Where("LOWER(username) = LOWER(?)", strings.TrimSpace(req.Username)).First(&user)

	// 用户不存在
	if userInfo.Error != nil {
		c.JSON(http.StatusBadRequest, result.Error(401, "用户名或密码错误"))
		return
	}
	if user.Role != models.RoleAdmin && user.Role != models.RoleUser {
		c.JSON(http.StatusForbidden, result.Error(403, "该账户类型已停用"))
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusBadRequest, result.Error(401, "用户名或密码错误"))
		return
	}

	// 设置session
	session, err := SetSession(c, &user)
	if err != nil {
		return
	}

	// 返回结果去除密码
	user.Password = ""
	// 返回结果
	c.JSON(http.StatusOK, result.Success("登录成功", map[string]any{
		"token": session.ID(),
		"user":  user,
	}))
}

// 设置Session
func SetSession(c *gin.Context, user *models.User) (sessions.Session, error) {
	// 获取session
	session := sessions.Default(c)

	// 设置session数据
	session.Set("user_id", user.Id)
	session.Set("user_role", user.Role)
	session.Set("username", user.Username)
	session.Set("is_guest", false)
	session.Set("logged_in", true)

	// 设置session选项
	session.Options(sessions.Options{
		MaxAge:   24 * 60 * 60, // 24小时，单位秒
		HttpOnly: true,         // 防止XSS攻击
		Secure:   config.App != nil && config.App.SessionSecure,
		SameSite: http.SameSiteStrictMode, // 防止CSRF攻击
		Path:     "/",                     // cookie路径
	})

	// 保存session
	if err := session.Save(); err != nil {
		errMsg := "session保存失败：" + err.Error()
		c.JSON(http.StatusInternalServerError, result.Error(500, errMsg))
		return nil, err
	}

	return session, nil
}

// ValidateTurnstileToken 验证 Cloudflare Turnstile token
func ValidateTurnstileToken(token string, clientIP string) bool {
	if token == "" {
		log.Println("[Turnstile] Token is empty")
		return false
	}

	// 从系统设置获取密钥
	sysSettings, err := settings.GetSettings()
	if err != nil {
		log.Printf("[Turnstile] Error getting settings: %v\n", err)
		return false
	}
	// 1. Trim whitespace (defensive, in case settings weren't updated yet)
	secretKey := strings.TrimSpace(sysSettings.TurnstileSecretKey)

	if secretKey == "" {
		log.Println("[Turnstile] Secret key is empty in settings")
		return false
	}

	// Log masked key for debugging
	maskedKey := secretKey
	if len(secretKey) > 8 {
		maskedKey = secretKey[:4] + "..." + secretKey[len(secretKey)-4:]
	}
	log.Printf("[Turnstile] Using Secret Key: %s (len: %d)", maskedKey, len(secretKey))

	// 构建请求
	formData := url.Values{}
	formData.Set("secret", secretKey)
	formData.Set("response", token)
	if clientIP != "" {
		formData.Set("remoteip", clientIP)
	}

	// 发送验证请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", formData)
	if err != nil {
		log.Printf("[Turnstile] HTTP request failed: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Turnstile] Error reading response body: %v\n", err)
		return false
	}

	var verifyResp struct {
		Success     bool     `json:"success"`
		ErrorCodes  []string `json:"error-codes"`
		ChallengeTS string   `json:"challenge_ts"`
		Hostname    string   `json:"hostname"`
	}
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		log.Printf("[Turnstile] Error unmarshaling response: %v\n", err)
		return false
	}

	if !verifyResp.Success {
		log.Printf("[Turnstile] Verification failed. Response: %+v\n", verifyResp)
		prefixLength := min(len(token), 10)
		log.Printf("[Turnstile] Token used: %s...\n", token[:prefixLength])
	} else {
		log.Println("[Turnstile] Verification successful")
	}

	return verifyResp.Success
}

// 退出登录
func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   config.App != nil && config.App.SessionSecure,
		SameSite: http.SameSiteStrictMode,
	})
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "退出登录失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success("退出登录成功", nil))
}
