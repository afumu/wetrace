package web

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// setupRoutes 初始化所有应用程序路由。
func (s *Service) setupRoutes() {
	// API v1 路由组, 使用在 service 中初始化的处理器
	v1 := s.router.Group("/api/v1")
	{
		// 系统路由
		system := v1.Group("/system")
		{
			system.GET("/status", s.api.GetSystemStatus)
			system.POST("/decrypt", s.api.HandleEnvDecrypt)
			system.GET("/wxkey/db", s.api.GetWeChatDbKey)
			system.GET("/wxkey/image", s.api.GetWeChatImageKey)
			system.GET("/detect/wechat_path", s.api.DetectWeChatInstallPath)
			system.GET("/detect/db_path", s.api.DetectWeChatDataPath)
			system.POST("/select_path", s.api.SelectPath)
			system.POST("/config", s.api.UpdateConfig)
		}

		// 会话路由
		v1.GET("/sessions", s.api.GetSessions)
		v1.DELETE("/sessions/:id", s.api.DeleteSession)

		// 总览路由
		v1.GET("/dashboard", s.api.GetDashboard)

		// 消息路由
		v1.GET("/messages", s.api.GetMessages)

		// 联系人路由
		v1.GET("/contacts", s.api.GetContacts)
		v1.GET("/contacts/:id", s.api.GetContactByID)

		// 群聊路由
		v1.GET("/chatrooms", s.api.GetChatRooms)
		v1.GET("/chatrooms/:id", s.api.GetChatRoomByID)

		// 媒体路由
		v1.GET("/media/:type/:key", s.api.GetMedia)
		v1.GET("/media/emoji", s.api.GetEmoji)
		v1.POST("/media/cache/start", s.api.HandleStartCache)
		v1.GET("/media/cache/status", s.api.GetCacheStatus)

		// 导出路由
		v1.GET("/export/chat", s.api.ExportChat)

		// AI 路由
		aiGroup := v1.Group("/ai")
		{
			aiGroup.POST("/summarize", s.api.AISummarize)
			aiGroup.POST("/simulate", s.api.AISimulate)
		}

		// 分析路由
		analysisGroup := v1.Group("/analysis")
		{
			analysisGroup.GET("/personal/top_contacts", s.api.GetPersonalTopContacts)
			analysisGroup.GET("/hourly/:id", s.api.GetHourlyActivity)
			analysisGroup.GET("/daily/:id", s.api.GetDailyActivity)
			analysisGroup.GET("/weekday/:id", s.api.GetWeekdayActivity)
			analysisGroup.GET("/monthly/:id", s.api.GetMonthlyActivity)
			analysisGroup.GET("/type_distribution/:id", s.api.GetMessageTypeDistribution)
			analysisGroup.GET("/member_activity/:id", s.api.GetMemberActivity)
			analysisGroup.GET("/repeat/:id", s.api.GetRepeatAnalysis)
		}
	}

	// 健康检查
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 静态文件服务 (UI)
	if s.staticFS != nil {
		s.router.StaticFS("/assets", http.FS(s.staticFS))
		// 处理 SPA 的 fallback，除了 /api 开头的路径外，都返回 index.html
		s.router.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api") {
				c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
				return
			}

			// Try to serve file directly first (e.g. favicon.ico)
			file, err := s.staticFS.Open(strings.TrimPrefix(c.Request.URL.Path, "/"))
			if err == nil {
				defer file.Close()
				stat, err := file.Stat()
				if err == nil && !stat.IsDir() {
					http.FileServer(http.FS(s.staticFS)).ServeHTTP(c.Writer, c.Request)
					return
				}
			}

			// Serve index.html
			f, err := s.staticFS.Open("index.html")
			if err != nil {
				c.String(http.StatusNotFound, "UI not found")
				return
			}
			defer f.Close()
			c.Status(http.StatusOK)
			c.Header("Content-Type", "text/html")
			_, _ = io.Copy(c.Writer, f)
		})
	}
}
