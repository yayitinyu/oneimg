package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"oneimg/backend/database"
	"oneimg/backend/models"
	"oneimg/backend/utils/result"
	"oneimg/backend/utils/settings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	errUsernameExists     = errors.New("username already exists")
	errInvitationRequired = errors.New("invitation code required")
	errInvitationInvalid  = errors.New("invitation code invalid")
)

type validationError string

func (e validationError) Error() string { return string(e) }

func newValidationError(message string) error { return validationError(message) }

func isLetterOrNumber(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) }

func nowUTC() time.Time { return time.Now().UTC() }

type RegisterRequest struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	InvitationCode string `json:"invitation_code"`
	TurnstileToken string `json:"turnstileToken"`
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, "请完整填写注册信息"))
		return
	}

	sysSettings, err := settings.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "获取注册配置失败"))
		return
	}

	mode := strings.ToLower(strings.TrimSpace(sysSettings.RegistrationMode))
	if mode == "" {
		mode = models.RegistrationOpen
	}
	if mode == models.RegistrationClosed {
		c.JSON(http.StatusForbidden, result.Error(403, "当前未开放注册"))
		return
	}
	if mode != models.RegistrationOpen && mode != models.RegistrationInvite {
		c.JSON(http.StatusInternalServerError, result.Error(500, "注册配置无效"))
		return
	}

	if sysSettings.Turnstile {
		if req.TurnstileToken == "" || !ValidateTurnstileToken(req.TurnstileToken, c.ClientIP()) {
			c.JSON(http.StatusBadRequest, result.Error(400, "人机验证失败，请重试"))
			return
		}
	}

	username, err := validateRegistrationCredentials(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, result.Error(400, err.Error()))
		return
	}

	db := database.GetDB()
	if db == nil || db.DB == nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "数据库连接失败"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Error(500, "密码处理失败"))
		return
	}

	user := models.User{
		Role:     models.RoleUser,
		Username: username,
		Password: string(hashedPassword),
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var existing int64
		if err := tx.Model(&models.User{}).
			Where("LOWER(username) = LOWER(?)", username).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return errUsernameExists
		}

		var invitation models.InvitationCode
		if mode == models.RegistrationInvite {
			code := normalizeInvitationCode(req.InvitationCode)
			if code == "" {
				return errInvitationRequired
			}
			if err := tx.Where("code_hash = ? AND used_at IS NULL", hashInvitationCode(code)).First(&invitation).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return errInvitationInvalid
				}
				return err
			}
		}

		if err := tx.Create(&user).Error; err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unique") ||
				strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				return errUsernameExists
			}
			return err
		}

		if mode == models.RegistrationInvite {
			now := nowUTC()
			update := tx.Model(&models.InvitationCode{}).
				Where("id = ? AND used_at IS NULL", invitation.Id).
				Updates(map[string]any{"used_at": now, "used_by": user.Id})
			if update.Error != nil {
				return update.Error
			}
			if update.RowsAffected != 1 {
				return errInvitationInvalid
			}
		}
		return nil
	})
	if err != nil {
		switch err {
		case errUsernameExists:
			c.JSON(http.StatusConflict, result.Error(409, "用户名已存在"))
		case errInvitationRequired:
			c.JSON(http.StatusBadRequest, result.Error(400, "请输入邀请码"))
		case errInvitationInvalid:
			c.JSON(http.StatusBadRequest, result.Error(400, "邀请码无效或已使用"))
		default:
			c.JSON(http.StatusInternalServerError, result.Error(500, "注册失败，请稍后重试"))
		}
		return
	}

	session, err := SetSession(c, &user)
	if err != nil {
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, result.Success("注册成功", map[string]any{
		"token": session.ID(),
		"user":  user,
	}))
}

func validateRegistrationCredentials(rawUsername, password string) (string, error) {
	username, err := validateAccountUsername(rawUsername)
	if err != nil {
		return "", err
	}
	if err := validateAccountPassword(password); err != nil {
		return "", err
	}
	return username, nil
}

func validateAccountPassword(password string) error {
	// bcrypt accepts at most 72 bytes. Count runes for the minimum so a short
	// multibyte password cannot accidentally satisfy the six-character rule.
	if utf8.RuneCountInString(password) < 6 || len([]byte(password)) > 72 {
		return newValidationError("密码至少需要 6 个字符，且不能超过 72 字节")
	}
	return nil
}

func validateAccountUsername(rawUsername string) (string, error) {
	username := strings.TrimSpace(rawUsername)
	usernameLength := utf8.RuneCountInString(username)
	if usernameLength < 3 || usernameLength > 32 {
		return "", newValidationError("用户名长度需为 3–32 个字符")
	}
	if isTouristUsername(username) {
		return "", newValidationError("该用户名为系统保留名称")
	}
	for _, r := range username {
		if !(r == '_' || r == '-' || isLetterOrNumber(r)) {
			return "", newValidationError("用户名只能包含文字、数字、下划线和连字符")
		}
	}
	return username, nil
}

func hashInvitationCode(code string) string {
	sum := sha256.Sum256([]byte(normalizeInvitationCode(code)))
	return hex.EncodeToString(sum[:])
}

func normalizeInvitationCode(code string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), "-", ""))
}
