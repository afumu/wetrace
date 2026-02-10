package middleware

import (
	"net/http"
	"strings"

	"github.com/afumu/wetrace/web/api"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// AuthMiddleware 密码保护中间件
func AuthMiddleware(a *api.API) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用了密码保护
		hash := viper.GetString("PASSWORD_HASH")
		if hash == "" {
			c.Next()
			return
		}

		// 白名单路径不需要验证
		path := c.Request.URL.Path
		whitelist := []string{
			"/api/v1/system/password/status",
			"/api/v1/system/password/verify",
			"/api/v1/system/compliance",
			"/health",
		}
		for _, w := range whitelist {
			if strings.HasPrefix(path, w) {
				c.Next()
				return
			}
		}

		// 非 API 路径不需要验证（静态文件等）
		if !strings.HasPrefix(path, "/api/") {
			c.Next()
			return
		}

		// 从 header 或 cookie 获取 token
		token := c.GetHeader("X-Auth-Token")
		if token == "" {
			token, _ = c.Cookie("auth_token")
		}

		if token == "" || !a.Password.IsValidSession(token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    401,
					"message": "请先验证密码",
				},
			})
			return
		}

		c.Next()
	}
}
