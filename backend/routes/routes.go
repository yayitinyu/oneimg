package routes

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"time"

	"oneimg/backend/config"
	"oneimg/backend/controllers"
	"oneimg/backend/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// 设置路由
func SetupRoutes(frontendFS embed.FS) *gin.Engine {
	cfg := config.App
	controllers.StartImageLifecycleWorker()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 基础中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middlewares.ConfigMiddleware(cfg))
	r.Use(middlewares.SessionMiddleware(cfg))

	// 默认仅允许同源访问。显式配置来源时才启用 CORS，避免“* + Cookie”无效且不安全的组合。
	if len(cfg.CORSAllowedOrigins) > 0 {
		allowCredentials := true
		for _, origin := range cfg.CORSAllowedOrigins {
			if origin == "*" {
				allowCredentials = false
				break
			}
		}
		r.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.CORSAllowedOrigins,
			AllowMethods:     []string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Authorization", "Content-Type", "Accept", "Origin"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: allowCredentials,
			MaxAge:           12 * time.Hour,
		}))
	}

	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		panic("加载前端文件失败：" + err.Error())
	}
	assetsFS, _ := fs.Sub(distFS, "assets")
	r.StaticFS("/assets", http.FS(assetsFS))

	// 静态资源
	r.GET("/uploads/*path", controllers.ImageProxy)
	r.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")

	// API路由分组
	api := r.Group("/api")
	{
		// 公开接口（无需认证）
		api.POST("/login", controllers.Login)
		api.POST("/register", controllers.Register)
		api.POST("/logout", controllers.Logout)
		api.GET("/logout", controllers.Logout)
		// 返回登录设置
		api.GET("/settings/login", controllers.GetLoginSettings)
		// 健康检查
		api.Match([]string{"GET", "HEAD"}, "/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "time": time.Now().Unix()})
		})
		// Telegram Bot Webhook（公开端点，无需认证）
		api.POST("/telegram/webhook", controllers.TelegramWebhook)

		// 需要认证的接口分组（应用AuthMiddleware）
		auth := api.Group("")
		auth.Use(middlewares.AuthMiddleware())
		{
			// 用户信息接口
			auth.GET("/user/status", controllers.CheckLoginStatus)
			auth.GET("/user/profile", controllers.GetUserProfile)
			auth.PUT("/user/profile", controllers.UpdateUserProfile)
			auth.POST("/account/change", controllers.ChangeAccountInfo)

			// 图片相关接口
			auth.POST("/upload", controllers.UploadImage)
			auth.POST("/upload/images", controllers.UploadImages)
			auth.POST("/upload/url", controllers.UploadImageByURL)
			auth.DELETE("/images/:id", controllers.DeleteImage)
			auth.DELETE("/images/:id/record", controllers.DeleteImageRecord) // Old endpoint for deletion
			auth.DELETE("/images/:id/recent", controllers.DismissImage)      // New endpoint for dismissing from recent
			auth.GET("/images", controllers.GetImageList)
			auth.GET("/images/:id", controllers.GetImageDetail)

			// 管理员查看全站统计，普通用户只查看自己的图片统计。
			auth.GET("/stats/dashboard", controllers.GetDashboardStats)
			auth.GET("/stats/images", controllers.GetImageStats)

			// 需要管理员权限
			auth.Use(middlewares.AdminOnlyMiddleware())
			{
				// 账户与邀请码管理接口
				auth.POST("/sessions/clear", controllers.ClearAllSessions)
				auth.GET("/invitations", controllers.ListInvitations)
				auth.POST("/invitations", controllers.CreateInvitations)
				auth.DELETE("/invitations/:id", controllers.DeleteInvitation)

				// 系统设置接口
				auth.Any("/settings/get", controllers.GetSettings)
				auth.POST("/settings/update", controllers.UpdateSettings)

				// 数据库状态接口
				auth.GET("/database/status", controllers.GetDatabaseStatus)
			}
		}
	}

	// 前端SPA路由支持
	r.NoRoute(func(c *gin.Context) {
		// API路径返回404
		if len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "API Not Found"})
			return
		}

		if files, err := fs.ReadDir(distFS, "."); err == nil {
			var fileNames []string
			for _, f := range files {
				fileNames = append(fileNames, f.Name())
			}
		} else {
			log.Printf("读取distFS文件列表失败：%s", err)
		}
		indexContent, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "加载前端页面失败：%s", err)
			return
		}

		// 返回HTML内容
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, string(indexContent))
	})

	return r
}
