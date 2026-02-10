package api

import (
	"time"

	"github.com/afumu/wetrace/internal/ai"
	"github.com/afumu/wetrace/pkg/util"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// SelectPath 让用户在服务端（本地）选择路径
func (a *API) SelectPath(c *gin.Context) {
	type SelectReq struct {
		Type string `json:"type"` // "file" or "folder"
	}
	var req SelectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "参数错误")
		return
	}

	var path string
	var err error

	if req.Type == "file" {
		path, err = util.OpenFileDialog("选择微信可执行文件", "WeChat Executable|WeChat.exe;Weixin.exe|All Files|*.*")
	} else {
		path, err = util.OpenFolderDialog("选择微信数据目录 (xwechat_files)")
	}

	if err != nil {
		// 用户取消
		if err.Error() == "cancelled" {
			transport.SendSuccess(c, gin.H{"path": ""})
			return
		}
		// 其他错误
		transport.InternalServerError(c, "打开对话框失败: "+err.Error())
		return
	}

	transport.SendSuccess(c, gin.H{"path": path})
}

// GetSystemStatus 返回应用程序的当前状态。
// 目前，它只确认服务正在运行。
func (a *API) GetSystemStatus(c *gin.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 获取当前配置中的密钥，用于前端判断是否存在
	status := gin.H{
		"store_initialized": true,
		"config": gin.H{
			"wechat_db_key":      a.Conf.WechatDbKey,
			"image_key":          a.Media.ImageKey,
			"xor_key":            a.Media.XorKey,
			"wechat_path":        a.Conf.WechatPath,
			"wechat_db_src_path": a.Conf.WechatDbSrcPath,
		},
	}
	transport.SendSuccess(c, status)
}

// DetectWeChatInstallPath 检测微信安装路径
func (a *API) DetectWeChatInstallPath(c *gin.Context) {
	paths := util.FindWeChatInstallPaths()
	transport.SendSuccess(c, paths)
}

// DetectWeChatDataPath 检测微信数据路径
func (a *API) DetectWeChatDataPath(c *gin.Context) {
	paths := util.FindWeChatDataPaths()
	transport.SendSuccess(c, paths)
}

// UpdateConfig 更新系统配置
func (a *API) UpdateConfig(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "参数错误")
		return
	}

	// 允许更新的键白名单
	allowedKeys := map[string]bool{
		"WXKEY_WECHAT_PATH":  true,
		"WECHAT_DB_SRC_PATH": true,
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	changed := false
	for k, v := range req {
		if allowedKeys[k] {
			viper.Set(k, v)
			// 同步更新内存中的配置对象
			if k == "WXKEY_WECHAT_PATH" {
				a.Conf.WechatPath = v
			} else if k == "WECHAT_DB_SRC_PATH" {
				a.Conf.WechatDbSrcPath = v
			}
			changed = true
		}
	}

	if changed {
		if err := viper.WriteConfig(); err != nil {
			// 如果文件不存在，尝试创建
			if err := viper.WriteConfigAs(".env"); err != nil {
				transport.InternalServerError(c, "保存配置文件失败: "+err.Error())
				return
			}
		}
	}

	transport.SendSuccess(c, gin.H{"status": "ok"})
}

// maskAPIKey 对 API Key 做脱敏处理，仅显示前4位和后4位
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// GetAIConfig 获取 AI 配置
func (a *API) GetAIConfig(c *gin.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()

	masked := ""
	if a.Conf.AIAPIKey != "" {
		masked = maskAPIKey(a.Conf.AIAPIKey)
	}

	transport.SendSuccess(c, gin.H{
		"enabled":        a.Conf.AIEnabled,
		"provider":       a.Conf.AIProvider,
		"model":          a.Conf.AIModel,
		"base_url":       a.Conf.AIBaseURL,
		"api_key_masked": masked,
	})
}

// UpdateAIConfig 更新 AI 配置
func (a *API) UpdateAIConfig(c *gin.Context) {
	var req struct {
		Enabled  bool   `json:"enabled"`
		Provider string `json:"provider"`
		Model    string `json:"model"`
		BaseURL  string `json:"base_url"`
		APIKey   string `json:"api_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "参数错误")
		return
	}

	if req.Enabled {
		if req.Model == "" || req.BaseURL == "" || req.APIKey == "" {
			transport.BadRequest(c, "启用 AI 时必须提供 model、base_url 和 api_key")
			return
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// 持久化到 viper
	viper.Set("AI_ENABLED", req.Enabled)
	viper.Set("AI_PROVIDER", req.Provider)
	viper.Set("AI_MODEL", req.Model)
	viper.Set("AI_BASE_URL", req.BaseURL)
	viper.Set("AI_API_KEY", req.APIKey)

	if err := viper.WriteConfig(); err != nil {
		transport.InternalServerError(c, "保存配置失败: "+err.Error())
		return
	}

	// 同步更新内存配置
	a.Conf.AIEnabled = req.Enabled
	a.Conf.AIProvider = req.Provider
	a.Conf.AIModel = req.Model
	a.Conf.AIBaseURL = req.BaseURL
	a.Conf.AIAPIKey = req.APIKey

	// 重建 AI 客户端
	if req.Enabled {
		a.AI = ai.NewClient(req.APIKey, req.BaseURL, req.Model)
	} else {
		a.AI = nil
	}

	transport.SendSuccess(c, gin.H{"status": "ok"})
}

// TestAIConfig 测试 AI 连接
func (a *API) TestAIConfig(c *gin.Context) {
	a.mu.Lock()
	client := a.AI
	a.mu.Unlock()

	if client == nil {
		transport.BadRequest(c, "AI 未启用或未配置")
		return
	}

	start := time.Now()
	_, err := client.Chat([]ai.Message{
		{Role: "user", Content: "ping"},
	})
	latency := time.Since(start).Milliseconds()

	if err != nil {
		transport.InternalServerError(c, "AI 连接测试失败: "+err.Error())
		return
	}

	a.mu.Lock()
	model := a.Conf.AIModel
	a.mu.Unlock()

	transport.SendSuccess(c, gin.H{
		"status":     "connected",
		"model":      model,
		"latency_ms": latency,
	})
}

// GetCompliance 获取合规同意状态
func (a *API) GetCompliance(c *gin.Context) {
	agreed := viper.GetBool("COMPLIANCE_AGREED")
	agreedAt := viper.GetString("COMPLIANCE_AGREED_AT")
	version := viper.GetString("COMPLIANCE_VERSION")
	if version == "" {
		version = "1.0"
	}

	transport.SendSuccess(c, gin.H{
		"agreed":    agreed,
		"agreed_at": agreedAt,
		"version":   version,
	})
}

// AgreeCompliance 提交合规同意
func (a *API) AgreeCompliance(c *gin.Context) {
	var req struct {
		Version string `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "参数错误")
		return
	}

	if req.Version == "" {
		req.Version = "1.0"
	}

	now := time.Now().Format(time.RFC3339)

	viper.Set("COMPLIANCE_AGREED", true)
	viper.Set("COMPLIANCE_AGREED_AT", now)
	viper.Set("COMPLIANCE_VERSION", req.Version)

	if err := viper.WriteConfig(); err != nil {
		transport.InternalServerError(c, "保存配置失败: "+err.Error())
		return
	}

	transport.SendSuccess(c, gin.H{"status": "agreed"})
}
