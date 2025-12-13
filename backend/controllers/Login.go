package controllers

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"oneimg/backend/database"
	"oneimg/backend/models"
	"oneimg/backend/utils/result"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Cloudflare Turnstile 密钥
const TURNSTILE_SECRET_KEY = "0x4AAAAAACGe2CG9vvdBZFyH7myxpG1E8Lg"

// 登录请求结构
type LoginRequest struct {
	Username           string         `json:"username" binding:"required"`
	Password           string         `json:"password" binding:"required"`
	TurnstileToken     string         `json:"turnstileToken"`
	TouristFingerprint string         `json:"touristFingerprint"`
	FusionHash         string         `json:"fusionHash"`
	StableFeatures     map[string]any `json:"stableFeatures"`
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

	var settings models.Settings
	// 检查记录是否存在
	sqlResult := db.DB.First(&settings)
	if sqlResult.Error != nil {
		// 区分记录不存在和数据库错误
		if strings.Contains(sqlResult.Error.Error(), "record not found") {
			c.JSON(http.StatusInternalServerError, result.Error(500, "系统配置未初始化"))
		} else {
			c.JSON(http.StatusInternalServerError, result.Error(500, "配置信息查询失败"))
		}
		return
	}

	// 先判断是否为游客登录（游客登录跳过验证）
	if settings.Tourist {
		// 判断是否为游客登录（UUID格式/包含guest前缀）
		isTourist := len(req.TouristFingerprint) == 36 ||
			strings.HasPrefix(req.Username, "guest_") ||
			req.Username == "guest" ||
			len(req.Username) == 36 // UUID 格式

		if isTourist {
			// 1. 优先使用传递的游客指纹
			touristUUID := req.TouristFingerprint
			if touristUUID == "" {
				touristUUID = req.Username
				// 兼容旧逻辑，固定guest账号生成随机UUID
				if touristUUID == "guest" {
					touristUUID = generateRandomUUID()
				}
			}

			touristID := int(generateTouristID(touristUUID))
			touristUser := &models.User{
				Id:       touristID,
				Role:     2,
				Username: touristUUID,
			}

			// 设置游客Session
			session, err := SetSession(c, touristUser)
			if err != nil {
				c.JSON(http.StatusInternalServerError, result.Error(500, "游客登录失败："+err.Error()))
				return
			}

			// 返回游客登录结果
			c.JSON(http.StatusOK, result.Success("游客登录成功", map[string]any{
				"token": session.ID(),
				"user": &models.User{
					Id:       touristUser.Id,
					Role:     2,
					Username: touristUser.Username,
				},
			}))
			return
		}
	}

	// 检查是否开启了 Turnstile 验证（仅对管理员生效）
	if settings.Turnstile {
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
	userInfo := db.DB.Where("username = ?", req.Username).First(&user)

	// 用户不存在
	if userInfo.Error != nil {
		c.JSON(http.StatusBadRequest, result.Error(401, "用户名或密码错误"))
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

// 辅助函数：基于UUID生成游客ID（保证唯一性）
func generateTouristID(uuid string) uint {
	var id uint = 2 // 基础ID（避开普通用户ID）
	for _, c := range uuid {
		id = id*31 + uint(c)
	}
	// 保证ID大于2，且不超过uint最大值
	if id <= 2 {
		id += 100000
	}
	return id
}

// 辅助函数：生成随机UUID
func generateRandomUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// 降级方案：使用时间戳+随机数
		return "guest_" + time.Now().Format("20060102150405") + "_" + strings.ReplaceAll(time.Now().String(), ":", "")
	}

	// 设置UUID版本和变体
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // RFC 4122 variant

	// 格式化UUID字符串
	return strings.ToLower(
		string(b[0:4]) + "-" +
			string(b[4:6]) + "-" +
			string(b[6:8]) + "-" +
			string(b[8:10]) + "-" +
			string(b[10:16]),
	)[:36]
}

// 设置Session
func SetSession(c *gin.Context, user *models.User) (sessions.Session, error) {
	// 获取session
	session := sessions.Default(c)

	// 设置session数据
	session.Set("user_id", user.Id)
	session.Set("user_role", user.Role)
	session.Set("username", user.Username)
	session.Set("logged_in", true)

	// 设置session选项
	session.Options(sessions.Options{
		MaxAge:   24 * 60 * 60,            // 24小时，单位秒
		HttpOnly: true,                    // 防止XSS攻击
		Secure:   false,                   // 生产环境应设为true（需要HTTPS）
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
		return false
	}

	// 构建请求
	formData := url.Values{}
	formData.Set("secret", TURNSTILE_SECRET_KEY)
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
		return false
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	var verifyResp struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return false
	}

	return verifyResp.Success
}

// 退出登录
func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "退出登录失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success("退出登录成功", nil))
}
