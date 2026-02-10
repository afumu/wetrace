package api

import (
	"os"
	"strconv"

	"github.com/afumu/wetrace/internal/monitor"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
)

// GetMonitorConfigs 获取所有监控配置
func (a *API) GetMonitorConfigs(c *gin.Context) {
	if a.Monitor == nil {
		transport.BadRequest(c, "监控功能未初始化")
		return
	}
	configs := a.Monitor.ListConfigs()
	transport.SendSuccess(c, configs)
}

// CreateMonitorConfig 创建监控配置
func (a *API) CreateMonitorConfig(c *gin.Context) {
	if a.Monitor == nil {
		transport.BadRequest(c, "监控功能未初始化")
		return
	}

	var req monitor.MonitorConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if req.Name == "" {
		transport.BadRequest(c, "配置名称不能为空")
		return
	}
	if req.Type != "keyword" && req.Type != "ai" {
		transport.BadRequest(c, "type 必须为 keyword 或 ai")
		return
	}

	id, err := a.Monitor.CreateConfig(req)
	if err != nil {
		transport.InternalServerError(c, "创建配置失败: "+err.Error())
		return
	}
	transport.SendSuccess(c, gin.H{"id": id})
}

// UpdateMonitorConfig 更新监控配置
func (a *API) UpdateMonitorConfig(c *gin.Context) {
	if a.Monitor == nil {
		transport.BadRequest(c, "监控功能未初始化")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		transport.BadRequest(c, "无效的配置ID")
		return
	}

	var req monitor.MonitorConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := a.Monitor.UpdateConfig(id, req); err != nil {
		if os.IsNotExist(err) {
			transport.NotFound(c, "配置不存在")
			return
		}
		transport.InternalServerError(c, "更新配置失败: "+err.Error())
		return
	}
	transport.SendSuccess(c, gin.H{"status": "ok"})
}

// DeleteMonitorConfig 删除监控配置
func (a *API) DeleteMonitorConfig(c *gin.Context) {
	if a.Monitor == nil {
		transport.BadRequest(c, "监控功能未初始化")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		transport.BadRequest(c, "无效的配置ID")
		return
	}

	if err := a.Monitor.DeleteConfig(id); err != nil {
		if os.IsNotExist(err) {
			transport.NotFound(c, "配置不存在")
			return
		}
		transport.InternalServerError(c, "删除配置失败: "+err.Error())
		return
	}
	transport.SendSuccess(c, gin.H{"status": "deleted"})
}

// TestMonitorPush 测试推送连通性
func (a *API) TestMonitorPush(c *gin.Context) {
	if a.Monitor == nil {
		transport.BadRequest(c, "监控功能未初始化")
		return
	}

	var req struct {
		URL      string `json:"url"`
		Secret   string `json:"secret"`
		Platform string `json:"platform"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if req.URL == "" {
		transport.BadRequest(c, "URL不能为空")
		return
	}

	switch req.Platform {
	case "feishu":
		if err := monitor.TestFeishuBot(req.URL, req.Secret); err != nil {
			transport.InternalServerError(c, "飞书测试失败: "+err.Error())
			return
		}
		transport.SendSuccess(c, gin.H{"status": "ok", "message": "测试消息已发送"})
	default:
		code, err := monitor.TestWebhookURL(req.URL)
		if err != nil {
			transport.InternalServerError(c, "Webhook测试失败: "+err.Error())
			return
		}
		transport.SendSuccess(c, gin.H{"status": "ok", "response_code": code})
	}
}
