package api

import (
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// GetBackupConfig returns the current auto-backup configuration.
func (a *API) GetBackupConfig(c *gin.Context) {
	if a.BackupScheduler == nil {
		transport.InternalServerError(c, "备份调度器未初始化")
		return
	}
	transport.SendSuccess(c, a.BackupScheduler.GetStatus())
}

// UpdateBackupConfig updates the auto-backup configuration.
func (a *API) UpdateBackupConfig(c *gin.Context) {
	if a.BackupScheduler == nil {
		transport.InternalServerError(c, "备份调度器未初始化")
		return
	}

	var req struct {
		Enabled       bool   `json:"enabled"`
		IntervalHours int    `json:"interval_hours"`
		BackupPath    string `json:"backup_path"`
		Format        string `json:"format"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "参数错误")
		return
	}

	if req.IntervalHours < 1 {
		transport.BadRequest(c, "备份间隔必须大于等于 1 小时")
		return
	}

	if req.Enabled && req.BackupPath == "" {
		transport.BadRequest(c, "启用备份时必须指定备份路径")
		return
	}

	a.BackupScheduler.Configure(req.Enabled, req.IntervalHours, req.BackupPath, req.Format)

	// Persist to viper config
	viper.Set("BACKUP_ENABLED", req.Enabled)
	viper.Set("BACKUP_INTERVAL_HOURS", req.IntervalHours)
	viper.Set("BACKUP_PATH", req.BackupPath)
	viper.Set("BACKUP_FORMAT", req.Format)
	_ = viper.WriteConfig()

	transport.SendSuccess(c, gin.H{"status": "configured"})
}

// RunBackup manually triggers a backup operation.
func (a *API) RunBackup(c *gin.Context) {
	if a.BackupScheduler == nil {
		transport.InternalServerError(c, "备份调度器未初始化")
		return
	}

	go a.BackupScheduler.RunBackup()
	transport.SendSuccess(c, gin.H{"status": "backup_started"})
}

// GetBackupHistory returns backup history records.
func (a *API) GetBackupHistory(c *gin.Context) {
	if a.BackupScheduler == nil {
		transport.InternalServerError(c, "备份调度器未初始化")
		return
	}

	var query struct {
		Limit  int `form:"limit"`
		Offset int `form:"offset"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		transport.BadRequest(c, "参数错误")
		return
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}

	records := a.BackupScheduler.GetHistory(query.Limit, query.Offset)
	transport.SendSuccess(c, records)
}
