package api

import (
	"github.com/afumu/wetrace/internal/monitor"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
)

// GetFeishuConfig 获取飞书平台配置
func (a *API) GetFeishuConfig(c *gin.Context) {
	if a.Monitor == nil {
		transport.BadRequest(c, "监控功能未初始化")
		return
	}
	cfg := a.Monitor.GetFeishuConfig()
	transport.SendSuccess(c, cfg)
}

// UpdateFeishuConfig 更新飞书平台配置
func (a *API) UpdateFeishuConfig(c *gin.Context) {
	if a.Monitor == nil {
		transport.BadRequest(c, "监控功能未初始化")
		return
	}

	var req monitor.FeishuConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := a.Monitor.UpdateFeishuConfig(req); err != nil {
		transport.InternalServerError(c, "更新飞书配置失败: "+err.Error())
		return
	}
	transport.SendSuccess(c, gin.H{"status": "ok"})
}

// TestFeishuBot 测试飞书机器人连通性
func (a *API) TestFeishuBot(c *gin.Context) {
	if a.Monitor == nil {
		transport.BadRequest(c, "监控功能未初始化")
		return
	}

	cfg := a.Monitor.GetFeishuConfig()
	if cfg.BotWebhook == "" {
		transport.BadRequest(c, "飞书机器人Webhook URL未配置")
		return
	}

	if err := monitor.TestFeishuBot(cfg.BotWebhook, cfg.SignSecret); err != nil {
		transport.InternalServerError(c, "飞书测试失败: "+err.Error())
		return
	}
	transport.SendSuccess(c, gin.H{"status": "ok", "message": "测试消息已发送"})
}

// TestFeishuBitable 测试飞书多维表格连通性
func (a *API) TestFeishuBitable(c *gin.Context) {
	if a.Monitor == nil {
		transport.BadRequest(c, "监控功能未初始化")
		return
	}

	cfg := a.Monitor.GetFeishuConfig()
	if cfg.AppID == "" || cfg.AppSecret == "" {
		transport.BadRequest(c, "飞书应用 AppID/AppSecret 未配置")
		return
	}
	if cfg.AppToken == "" || cfg.TableID == "" {
		transport.BadRequest(c, "多维表格 AppToken/TableID 未配置")
		return
	}

	if err := monitor.TestBitableConnection(cfg.AppID, cfg.AppSecret, cfg.AppToken, cfg.TableID); err != nil {
		transport.InternalServerError(c, "多维表格测试失败: "+err.Error())
		return
	}
	transport.SendSuccess(c, gin.H{"status": "ok", "message": "多维表格连通性测试成功"})
}
