package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/afumu/wetrace/pkg/util"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
)

// ExportChat 处理导出聊天记录的请求
func (a *API) ExportChat(c *gin.Context) {
	talker := c.Query("talker")
	talkerName := c.Query("name")
	timeRange := c.Query("time_range")

	if talker == "" {
		transport.BadRequest(c, "talker 参数是必需的")
		return
	}

	if talkerName == "" {
		talkerName = talker
	}

	start, end, ok := util.TimeRangeOf(timeRange)
	if !ok {
		// 默认导出所有 (2000-01-01 ~ Now+24h 类似之前的逻辑，或者直接用 All)
		start = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		end = time.Now().Add(24 * time.Hour)
	}

	format := c.Query("format")
	if format == "txt" {
		txtData, err := a.Export.ExportChatTxt(c.Request.Context(), talker, talkerName, start, end)
		if err != nil {
			transport.InternalServerError(c, fmt.Sprintf("导出失败: %v", err))
			return
		}
		fileName := fmt.Sprintf("chat_export_%s_%s.txt", talkerName, talker)
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Data(http.StatusOK, "text/plain; charset=utf-8", txtData)
		return
	}

	// 执行导出
	zipData, err := a.Export.ExportChat(c.Request.Context(), talker, talkerName, start, end)
	if err != nil {
		transport.InternalServerError(c, fmt.Sprintf("导出失败: %v", err))
		return
	}

	// 设置响应头并发送 ZIP 文件
	fileName := fmt.Sprintf("chat_export_%s_%s.zip", talkerName, talker)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", zipData)
}
