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

	var (
		data        []byte
		err         error
		fileName    string
		contentType string
	)

	ctx := c.Request.Context()

	switch format {
	case "txt":
		data, err = a.Export.ExportChatTxt(ctx, talker, talkerName, start, end)
		fileName = fmt.Sprintf("chat_export_%s_%s.txt", talkerName, talker)
		contentType = "text/plain; charset=utf-8"
	case "csv":
		data, err = a.Export.ExportChatCSV(ctx, talker, talkerName, start, end)
		fileName = fmt.Sprintf("chat_export_%s_%s.csv", talkerName, talker)
		contentType = "text/csv; charset=utf-8"
	case "xlsx":
		data, err = a.Export.ExportChatXLSX(ctx, talker, talkerName, start, end)
		fileName = fmt.Sprintf("chat_export_%s_%s.xlsx", talkerName, talker)
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "docx":
		data, err = a.Export.ExportChatDOCX(ctx, talker, talkerName, start, end)
		fileName = fmt.Sprintf("chat_export_%s_%s.docx", talkerName, talker)
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "pdf":
		data, err = a.Export.ExportChatPDF(ctx, talker, talkerName, start, end)
		fileName = fmt.Sprintf("chat_export_%s_%s.pdf", talkerName, talker)
		contentType = "application/pdf"
	default:
		// 默认导出 HTML ZIP
		data, err = a.Export.ExportChat(ctx, talker, talkerName, start, end)
		fileName = fmt.Sprintf("chat_export_%s_%s.zip", talkerName, talker)
		contentType = "application/octet-stream"
	}

	if err != nil {
		transport.InternalServerError(c, fmt.Sprintf("导出失败: %v", err))
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, data)
}

// ExportForensic 处理法律取证导出请求，返回包含 report.html + chat_data.csv + checksums.sha256 + metadata.json 的 ZIP 包
func (a *API) ExportForensic(c *gin.Context) {
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
		start = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		end = time.Now().Add(24 * time.Hour)
	}

	ctx := c.Request.Context()
	data, err := a.Export.ExportForensic(ctx, talker, talkerName, start, end)
	if err != nil {
		transport.InternalServerError(c, fmt.Sprintf("取证导出失败: %v", err))
		return
	}

	fileName := fmt.Sprintf("forensic_export_%s_%s.zip", talkerName, talker)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/zip")
	c.Data(http.StatusOK, "application/zip", data)
}
