package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"oneimg/backend/database"
	"oneimg/backend/models"
	"oneimg/backend/utils/result"

	"github.com/gin-gonic/gin"
)

type imageListItem struct {
	models.Image
	OwnerType string `json:"owner_type"`
	OwnerName string `json:"owner_name"`
}

// GetImageList 获取图片列表
func GetImageList(c *gin.Context) {
	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// 获取排序参数
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	// 获取搜索参数
	search := c.Query("search")

	// 计算偏移量
	offset := (page - 1) * limit

	db := database.GetDB().DB

	var images []models.Image
	var total int64

	// 管理员可使用 owner 参数筛选；普通用户和游客始终只能看到自己的图片。
	query := scopedImagesQuery(c, db)

	// 过滤最近上传
	if c.Query("recent") == "true" {
		query = query.Where("show_in_recent = ?", true)
	}

	// 获取可见性参数
	visibility := c.Query("visibility")
	switch visibility {
	case "visible":
		query = query.Where("hidden = ?", false)
	case "hidden":
		query = query.Where("hidden = ?", true)
	case "all":
		// do nothing, return all
	default:
		// 默认只显示非隐藏图片
		query = query.Where("hidden = ?", false)
	}

	// 添加搜索条件
	if search != "" {
		query = query.Where("file_name LIKE ?", "%"+search+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取图片总数失败",
		})
		return
	}

	// 验证排序字段
	validSortFields := map[string]bool{
		"created_at": true,
		"file_size":  true,
		"filename":   true,
	}

	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// 映射前端字段名到数据库字段名
	fieldMapping := map[string]string{
		"filename":   "file_name",
		"created_at": "created_at",
		"file_size":  "file_size",
	}

	dbField := fieldMapping[sortBy]
	if dbField == "" {
		dbField = "created_at"
	}

	// 获取图片列表
	orderClause := dbField + " " + sortOrder
	if err := query.Order(orderClause).Offset(offset).Limit(limit).Find(&images).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取图片列表失败",
		})
		return
	}

	items := make([]imageListItem, 0, len(images))
	userIDs := make([]int, 0, len(images))
	seenUserIDs := make(map[int]struct{}, len(images))
	for _, image := range images {
		if image.UserId <= 0 {
			continue
		}
		if _, exists := seenUserIDs[image.UserId]; exists {
			continue
		}
		seenUserIDs[image.UserId] = struct{}{}
		userIDs = append(userIDs, image.UserId)
	}

	owners := make(map[int]models.User, len(userIDs))
	if len(userIDs) > 0 {
		var users []models.User
		if err := db.Select("id", "role", "username", "nickname").Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "获取图片归属失败",
			})
			return
		}
		for _, user := range users {
			owners[user.Id] = user
		}
	}

	for _, image := range images {
		item := imageListItem{Image: image, OwnerType: "guest", OwnerName: "游客"}
		if image.UserId == 0 && image.Storage == "telegram" {
			item.OwnerType = "external"
			item.OwnerName = "Telegram"
		} else if owner, exists := owners[image.UserId]; exists {
			item.OwnerType = "user"
			if owner.Role == models.RoleAdmin {
				item.OwnerType = "admin"
			}
			item.OwnerName = strings.TrimSpace(owner.Nickname)
			if item.OwnerName == "" {
				item.OwnerName = owner.Username
			}
		}
		items = append(items, item)
	}

	// 计算总页数
	totalPages := (total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, result.Success("获取图片列表成功", gin.H{
		"images":      items,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	}))
}
