package web

import (
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// setupMiddleware 配置 Gin 引擎所需的中间件。
func (s *Service) setupMiddleware() {
	s.router.Use(
		gin.LoggerWithWriter(log.Logger, "/health"),
		recoveryMiddleware(),
		corsMiddleware(),
	)
}

// corsMiddleware 提供一个宽松的 CORS 策略。
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// recoveryMiddleware 从任何 panic 中恢复并写入一个 500 错误。
func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Interface("error", err).Msg("Panic recovered")
				transport.InternalServerError(c, "服务器内部发生错误。")
			}
		}()
		c.Next()
	}
}
