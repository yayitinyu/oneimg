package controllers

import (
	"time"

	"oneimg/backend/models"
	"oneimg/backend/utils/md5"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func unexpiredImagesQuery(db *gorm.DB) *gorm.DB {
	return db.Model(&models.Image{}).
		Where("expires_at IS NULL OR expires_at > ?", time.Now())
}

func activeImagesQuery(db *gorm.DB) *gorm.DB {
	return unexpiredImagesQuery(db).Where("hidden = ?", false)
}

func scopedImagesQuery(c *gin.Context, db *gorm.DB) *gorm.DB {
	query := unexpiredImagesQuery(db)
	role := c.GetInt("user_role")

	if role == models.RoleAdmin {
		switch c.Query("owner") {
		case "admins":
			query = query.Where("user_id IN (?)", db.Model(&models.User{}).Select("id").Where("role = ?", models.RoleAdmin))
		case "users":
			query = query.Where("user_id IN (?)", db.Model(&models.User{}).Select("id").Where("role = ?", models.RoleUser))
		case "guests":
			query = query.Where("user_id NOT IN (?)", db.Model(&models.User{}).Select("id"))
		case "mine":
			query = query.Where("user_id = ?", c.GetInt("user_id"))
		}
		return query
	}

	if c.GetBool("is_guest") || role == models.RoleGuest {
		return query.Where("uuid = ?", GetUUID(c))
	}
	return query.Where("user_id = ?", c.GetInt("user_id"))
}

func CheckImageAccessPermission(c *gin.Context, image models.Image) bool {
	role := c.GetInt("user_role")
	if role == models.RoleAdmin {
		return true
	}
	if !c.GetBool("is_guest") && role != models.RoleGuest {
		return image.UserId == c.GetInt("user_id")
	}
	currentUUID := GetUUID(c)
	username := c.GetString("username")
	return image.UUID == currentUUID && md5.Md5(username+image.FileName) == image.MD5
}
