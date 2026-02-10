package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/afumu/wetrace/internal/replay"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
)

// CreateReplayExport 创建回放导出任务
func (a *API) CreateReplayExport(c *gin.Context) {
	if a.ReplayExporter == nil {
		transport.InternalServerError(c, "回放导出服务未初始化")
		return
	}

	var req replay.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "无效的导出参数: "+err.Error())
		return
	}

	// 设置默认值
	if req.Format == "" {
		req.Format = "mp4"
	}
	if req.Speed <= 0 {
		req.Speed = 4
	}
	if req.Resolution == "" {
		req.Resolution = "720p"
	}

	// 校验参数
	if req.Format != "mp4" && req.Format != "gif" {
		transport.BadRequest(c, "format 仅支持 mp4 或 gif")
		return
	}

	task := a.ReplayExporter.CreateTask(req)

	transport.SendSuccess(c, gin.H{
		"task_id": task.TaskID,
		"status":  task.Status,
		"message": "导出任务已创建",
	})
}

// GetReplayExportStatus 查询导出任务状态
func (a *API) GetReplayExportStatus(c *gin.Context) {
	if a.ReplayExporter == nil {
		transport.InternalServerError(c, "回放导出服务未初始化")
		return
	}

	taskID := c.Param("task_id")
	if taskID == "" {
		transport.BadRequest(c, "task_id 参数是必需的")
		return
	}

	task := a.ReplayExporter.TaskManager.GetTask(taskID)
	if task == nil {
		transport.NotFound(c, "导出任务不存在")
		return
	}

	transport.SendSuccess(c, gin.H{
		"task_id":          task.TaskID,
		"status":           task.Status,
		"progress":         task.Progress,
		"total_frames":     task.TotalFrames,
		"processed_frames": task.ProcessedFrames,
		"error":            task.Error,
	})
}

// DownloadReplayExport 下载已完成的导出文件
func (a *API) DownloadReplayExport(c *gin.Context) {
	if a.ReplayExporter == nil {
		transport.InternalServerError(c, "回放导出服务未初始化")
		return
	}

	taskID := c.Param("task_id")
	if taskID == "" {
		transport.BadRequest(c, "task_id 参数是必需的")
		return
	}

	task := a.ReplayExporter.TaskManager.GetTask(taskID)
	if task == nil {
		transport.NotFound(c, "导出任务不存在")
		return
	}

	if task.Status != replay.StatusCompleted {
		transport.BadRequest(c, "导出任务尚未完成")
		return
	}

	if task.FilePath == "" {
		transport.InternalServerError(c, "导出文件路径为空")
		return
	}

	data, err := os.ReadFile(task.FilePath)
	if err != nil {
		transport.InternalServerError(c, fmt.Sprintf("读取导出文件失败: %v", err))
		return
	}

	contentType := "video/mp4"
	fileName := taskID + ".mp4"
	if task.Format == "gif" {
		contentType = "image/gif"
		fileName = taskID + ".gif"
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, data)
}
