package web

import (
	"io"
	"net/http"
	"strings"

	"github.com/afumu/wetrace/web/middleware"
	"github.com/gin-gonic/gin"
)

// setupRoutes 初始化所有应用程序路由。
func (s *Service) setupRoutes() {
	// 密码保护中间件
	s.router.Use(middleware.AuthMiddleware(s.api))

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

			// AI 配置路由 (需求3)
			system.GET("/ai_config", s.api.GetAIConfig)
			system.POST("/ai_config", s.api.UpdateAIConfig)
			system.POST("/ai_config/test", s.api.TestAIConfig)

			// AI 提示词配置路由
			system.GET("/ai_prompts", s.api.GetAIPrompts)
			system.POST("/ai_prompts", s.api.UpdateAIPrompts)

			// 密码保护路由 (需求6)
			system.GET("/password/status", s.api.GetPasswordStatus)
			system.POST("/password/set", s.api.SetPassword)
			system.POST("/password/verify", s.api.VerifyPassword)
			system.POST("/password/disable", s.api.DisablePassword)

			// 合规提示路由 (需求9)
			system.GET("/compliance", s.api.GetCompliance)
			system.POST("/compliance/agree", s.api.AgreeCompliance)

			// 自动同步路由 (需求5)
			system.GET("/sync_config", s.api.GetSyncConfig)
			system.POST("/sync_config", s.api.UpdateSyncConfig)
			system.POST("/sync", s.api.TriggerSync)
			system.GET("/sync_status", s.api.GetSyncStatus)

			// 自动备份路由 (需求8)
			system.GET("/backup_config", s.api.GetBackupConfig)
			system.POST("/backup_config", s.api.UpdateBackupConfig)
			system.POST("/backup/run", s.api.RunBackup)
			system.GET("/backup/history", s.api.GetBackupHistory)

			// TTS 语音转文字配置路由
			system.GET("/tts_config", s.api.GetTTSConfig)
			system.POST("/tts_config", s.api.UpdateTTSConfig)
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
		v1.GET("/contacts/need-contact", s.api.GetNeedContactList)
		v1.GET("/contacts/:id", s.api.GetContactByID)
		v1.GET("/contacts/export", s.api.ExportContacts)

		// 群聊路由
		v1.GET("/chatrooms", s.api.GetChatRooms)
		v1.GET("/chatrooms/:id", s.api.GetChatRoomByID)

		// 媒体路由
		v1.GET("/media/images", s.api.GetImageList)
		v1.GET("/media/:type/:key", s.api.GetMedia)
		v1.GET("/media/emoji", s.api.GetEmoji)
		v1.POST("/media/cache/start", s.api.HandleStartCache)
		v1.GET("/media/cache/status", s.api.GetCacheStatus)
		v1.POST("/media/voice/transcribe", s.api.TranscribeVoice)

		// 导出路由
		v1.GET("/export/chat", s.api.ExportChat)
		v1.GET("/export/forensic", s.api.ExportForensic)
		v1.GET("/export/voices", s.api.ExportVoices)

		// 搜索路由
		searchGroup := v1.Group("/search")
		{
			searchGroup.GET("", s.api.Search)
			searchGroup.GET("/context", s.api.SearchContext)
		}

		// 年度报告路由
		reportGroup := v1.Group("/report")
		{
			reportGroup.GET("/annual", s.api.GetAnnualReport)
		}

		// AI 路由
		aiGroup := v1.Group("/ai")
		{
			aiGroup.POST("/summarize", s.api.AISummarize)
			aiGroup.POST("/simulate", s.api.AISimulate)
			aiGroup.POST("/sentiment", s.api.AISentiment)
			aiGroup.POST("/summary", s.api.AISummary)
			aiGroup.POST("/todos", s.api.AIExtractTodos)
			aiGroup.POST("/extract", s.api.AIExtractInfo)
			aiGroup.POST("/voice2text", s.api.AIVoice2Text)
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
			analysisGroup.GET("/wordcloud/global", s.api.GetWordCloudGlobal)
			analysisGroup.GET("/wordcloud/:id", s.api.GetWordCloud)
		}

		// 回放路由 (需求16)
		v1.GET("/messages/replay", s.api.GetReplayMessages)

		// 回放导出路由 (需求17)
		replayExport := v1.Group("/export/replay")
		{
			replayExport.POST("", s.api.CreateReplayExport)
			replayExport.GET("/status/:task_id", s.api.GetReplayExportStatus)
			replayExport.GET("/download/:task_id", s.api.DownloadReplayExport)
		}

		// 监控配置路由 (需求14+15)
		monitorGroup := v1.Group("/monitor")
		{
			monitorGroup.GET("/configs", s.api.GetMonitorConfigs)
			monitorGroup.POST("/configs", s.api.CreateMonitorConfig)
			monitorGroup.PUT("/configs/:id", s.api.UpdateMonitorConfig)
			monitorGroup.DELETE("/configs/:id", s.api.DeleteMonitorConfig)
			monitorGroup.POST("/test", s.api.TestMonitorPush)
		}

		// 飞书配置路由 (需求15)
		feishuGroup := v1.Group("/feishu")
		{
			feishuGroup.GET("/config", s.api.GetFeishuConfig)
			feishuGroup.PUT("/config", s.api.UpdateFeishuConfig)
			feishuGroup.POST("/test", s.api.TestFeishuBot)
			feishuGroup.POST("/test_bitable", s.api.TestFeishuBitable)
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
