package controllers

import (
	"net/http"
	"time"

	"oneimg/backend/database"
	"oneimg/backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StatsResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

type DashboardStats struct {
	TotalImages      int64                  `json:"total_images"`
	TotalSize        int64                  `json:"total_size"`
	AverageSize      int64                  `json:"average_size"`
	LargestSize      int64                  `json:"largest_size"`
	TodayUploads     int64                  `json:"today_uploads"`
	MonthUploads     int64                  `json:"month_uploads"`
	PermanentImages  int64                  `json:"permanent_images"`
	ExpiringSoon     int64                  `json:"expiring_soon"`
	ActiveOwners     int64                  `json:"active_owners"`
	RecentImages     []models.Image         `json:"recent_images"`
	UploadTrend      []UploadTrendItem      `json:"upload_trend"`
	FormatStats      []FormatStatsItem      `json:"format_stats"`
	StorageStats     []StorageStatsItem     `json:"storage_stats"`
	SizeDistribution []SizeDistributionItem `json:"size_distribution"`
}

type UploadTrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
	Size  int64  `json:"size"`
}

type FormatStatsItem struct {
	Format string `json:"format"`
	Count  int64  `json:"count"`
	Size   int64  `json:"size"`
}

type StorageStatsItem struct {
	Storage string `json:"storage"`
	Count   int64  `json:"count"`
	Size    int64  `json:"size"`
}

type SizeDistributionItem struct {
	Range string `json:"range"`
	Count int64  `json:"count"`
}

func GetDashboardStats(c *gin.Context) {
	dbInstance := database.GetDB()
	if dbInstance == nil || dbInstance.DB == nil {
		c.JSON(http.StatusInternalServerError, StatsResponse{Code: 500, Message: "数据库连接失败"})
		return
	}

	stats, err := buildDashboardStats(c, dbInstance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StatsResponse{Code: 500, Message: "获取统计数据失败"})
		return
	}
	c.JSON(http.StatusOK, StatsResponse{Code: 200, Message: "获取统计数据成功", Success: true, Data: stats})
}

func buildDashboardStats(c *gin.Context, dbInstance *database.Database) (DashboardStats, error) {
	db := dbInstance.DB
	var stats DashboardStats
	var totals struct {
		TotalImages int64
		TotalSize   int64
		AverageSize float64
		LargestSize int64
	}
	if err := statsQuery(c, db).
		Select("COUNT(*) AS total_images, COALESCE(SUM(file_size), 0) AS total_size, COALESCE(AVG(file_size), 0) AS average_size, COALESCE(MAX(file_size), 0) AS largest_size").
		Scan(&totals).Error; err != nil {
		return stats, err
	}
	stats.TotalImages = totals.TotalImages
	stats.TotalSize = totals.TotalSize
	stats.AverageSize = int64(totals.AverageSize)
	stats.LargestSize = totals.LargestSize

	today := time.Now().Format("2006-01-02")
	if err := statsQuery(c, db).Where("DATE(created_at) = ?", today).Count(&stats.TodayUploads).Error; err != nil {
		return stats, err
	}
	month := time.Now().Format("2006-01")
	monthQuery := getMonthQuery(dbInstance.DBType, month)
	if err := statsQuery(c, db).Where(monthQuery.condition, monthQuery.args...).Count(&stats.MonthUploads).Error; err != nil {
		return stats, err
	}
	if err := statsQuery(c, db).Where("expires_at IS NULL").Count(&stats.PermanentImages).Error; err != nil {
		return stats, err
	}
	if err := statsQuery(c, db).
		Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now().Add(7*24*time.Hour)).
		Count(&stats.ExpiringSoon).Error; err != nil {
		return stats, err
	}
	if err := statsQuery(c, db).Distinct("user_id").Count(&stats.ActiveOwners).Error; err != nil {
		return stats, err
	}
	if err := statsQuery(c, db).Order("created_at DESC").Limit(8).Find(&stats.RecentImages).Error; err != nil {
		return stats, err
	}

	stats.UploadTrend = getUploadTrend(c, db, 14)
	stats.FormatStats = getFormatStats(c, db)
	stats.StorageStats = getStorageStats(c, db)
	stats.SizeDistribution = getSizeDistribution(c, db)
	return stats, nil
}

func statsQuery(c *gin.Context, db *gorm.DB) *gorm.DB {
	return scopedImagesQuery(c, db).Where("hidden = ?", false)
}

type dateQuery struct {
	condition string
	args      []interface{}
}

func getMonthQuery(dbType, month string) dateQuery {
	switch dbType {
	case "postgresql":
		return dateQuery{"TO_CHAR(created_at, 'YYYY-MM') = ?", []interface{}{month}}
	case "mysql":
		return dateQuery{"DATE_FORMAT(created_at, '%Y-%m') = ?", []interface{}{month}}
	default:
		return dateQuery{"strftime('%Y-%m', created_at) = ?", []interface{}{month}}
	}
}

func getYearQuery(dbType, year string) dateQuery {
	switch dbType {
	case "postgresql":
		return dateQuery{"TO_CHAR(created_at, 'YYYY') = ?", []interface{}{year}}
	case "mysql":
		return dateQuery{"DATE_FORMAT(created_at, '%Y') = ?", []interface{}{year}}
	default:
		return dateQuery{"strftime('%Y', created_at) = ?", []interface{}{year}}
	}
}

func getUploadTrend(c *gin.Context, db *gorm.DB, days int) []UploadTrendItem {
	trend := make([]UploadTrendItem, 0, days)
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var item UploadTrendItem
		statsQuery(c, db).
			Where("DATE(created_at) = ?", date).
			Select("COUNT(*) AS count, COALESCE(SUM(file_size), 0) AS size").
			Scan(&item)
		item.Date = date
		trend = append(trend, item)
	}
	return trend
}

func getFormatStats(c *gin.Context, db *gorm.DB) []FormatStatsItem {
	var stats []FormatStatsItem
	statsQuery(c, db).
		Select("mime_type AS format, COUNT(*) AS count, COALESCE(SUM(file_size), 0) AS size").
		Group("mime_type").Order("count DESC").Scan(&stats)
	return stats
}

func getStorageStats(c *gin.Context, db *gorm.DB) []StorageStatsItem {
	var stats []StorageStatsItem
	statsQuery(c, db).
		Select("storage, COUNT(*) AS count, COALESCE(SUM(file_size), 0) AS size").
		Group("storage").Order("count DESC").Scan(&stats)
	return stats
}

func getSizeDistribution(c *gin.Context, db *gorm.DB) []SizeDistributionItem {
	ranges := []struct {
		name string
		min  int64
		max  int64
	}{
		{"< 100 KB", 0, 100 * 1024},
		{"100–500 KB", 100 * 1024, 500 * 1024},
		{"500 KB–1 MB", 500 * 1024, 1024 * 1024},
		{"1–5 MB", 1024 * 1024, 5 * 1024 * 1024},
		{"≥ 5 MB", 5 * 1024 * 1024, 0},
	}
	distribution := make([]SizeDistributionItem, 0, len(ranges))
	for _, itemRange := range ranges {
		var count int64
		query := statsQuery(c, db).Where("file_size >= ?", itemRange.min)
		if itemRange.max > 0 {
			query = query.Where("file_size < ?", itemRange.max)
		}
		query.Count(&count)
		distribution = append(distribution, SizeDistributionItem{Range: itemRange.name, Count: count})
	}
	return distribution
}

func GetImageStats(c *gin.Context) {
	dbInstance := database.GetDB()
	if dbInstance == nil || dbInstance.DB == nil {
		c.JSON(http.StatusInternalServerError, StatsResponse{Code: 500, Message: "数据库连接失败"})
		return
	}
	period := c.DefaultQuery("period", "month")
	var stats []UploadTrendItem
	switch period {
	case "day":
		stats = getPeriodStats(c, dbInstance.DB, 30, func(i int) (string, dateQuery) {
			date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
			return date, dateQuery{"DATE(created_at) = ?", []interface{}{date}}
		})
	case "week":
		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		currentWeekStart := time.Date(now.Year(), now.Month(), now.Day()-(weekday-1), 0, 0, 0, 0, now.Location())
		stats = getPeriodStats(c, dbInstance.DB, 12, func(i int) (string, dateQuery) {
			weekStart := currentWeekStart.AddDate(0, 0, -i*7)
			weekEnd := weekStart.AddDate(0, 0, 7)
			return weekStart.Format("2006-01-02"), dateQuery{"created_at >= ? AND created_at < ?", []interface{}{weekStart, weekEnd}}
		})
	case "year":
		stats = getPeriodStats(c, dbInstance.DB, 5, func(i int) (string, dateQuery) {
			year := time.Now().AddDate(-i, 0, 0).Format("2006")
			return year, getYearQuery(dbInstance.DBType, year)
		})
	default:
		stats = getPeriodStats(c, dbInstance.DB, 12, func(i int) (string, dateQuery) {
			month := time.Now().AddDate(0, -i, 0).Format("2006-01")
			return month, getMonthQuery(dbInstance.DBType, month)
		})
	}
	c.JSON(http.StatusOK, StatsResponse{Code: 200, Message: "获取图片统计成功", Success: true, Data: stats})
}

func getPeriodStats(c *gin.Context, db *gorm.DB, periods int, selector func(int) (string, dateQuery)) []UploadTrendItem {
	stats := make([]UploadTrendItem, 0, periods)
	for i := periods - 1; i >= 0; i-- {
		label, condition := selector(i)
		var item UploadTrendItem
		statsQuery(c, db).Where(condition.condition, condition.args...).
			Select("COUNT(*) AS count, COALESCE(SUM(file_size), 0) AS size").Scan(&item)
		item.Date = label
		stats = append(stats, item)
	}
	return stats
}
