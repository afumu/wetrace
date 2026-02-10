package api

import (
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// GetSyncConfig returns the current auto-sync configuration.
func (a *API) GetSyncConfig(c *gin.Context) {
	if a.SyncScheduler == nil {
		transport.InternalServerError(c, "同步调度器未初始化")
		return
	}
	transport.SendSuccess(c, a.SyncScheduler.GetStatus())
}

// UpdateSyncConfig updates the auto-sync configuration.
func (a *API) UpdateSyncConfig(c *gin.Context) {
	if a.SyncScheduler == nil {
		transport.InternalServerError(c, "同步调度器未初始化")
		return
	}

	var req struct {
		Enabled     bool `json:"enabled"`
		IntervalMin int  `json:"interval_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "参数错误")
		return
	}

	if req.IntervalMin < 5 || req.IntervalMin > 1440 {
		transport.BadRequest(c, "同步间隔必须在 5-1440 分钟之间")
		return
	}

	a.SyncScheduler.Configure(req.Enabled, req.IntervalMin)

	// Persist to viper config
	viper.Set("SYNC_ENABLED", req.Enabled)
	viper.Set("SYNC_INTERVAL_MINUTES", req.IntervalMin)
	_ = viper.WriteConfig()

	transport.SendSuccess(c, gin.H{"status": "ok"})
}

// TriggerSync manually triggers a sync operation.
func (a *API) TriggerSync(c *gin.Context) {
	if a.SyncScheduler == nil {
		transport.InternalServerError(c, "同步调度器未初始化")
		return
	}

	go a.SyncScheduler.RunSync()
	transport.SendSuccess(c, gin.H{"status": "syncing"})
}
