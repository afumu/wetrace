package export

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/types"
	"github.com/rs/zerolog/log"
)

// ForensicMetadata holds the metadata for a forensic export package.
type ForensicMetadata struct {
	ExportTime      string            `json:"export_time"`
	SoftwareVersion string            `json:"software_version"`
	Talker          string            `json:"talker"`
	TalkerName      string            `json:"talker_name"`
	TimeRange       map[string]string `json:"time_range"`
	MessageCount    int               `json:"message_count"`
	Files           map[string]string `json:"files"`
}

// ExportForensic generates a forensic export ZIP containing:
// - report.html (evidence-grade HTML report with watermark)
// - chat_data.csv (raw chat data)
// - checksums.sha256 (SHA-256 hashes of all files)
// - metadata.json (export context metadata)
func (s *Service) ExportForensic(ctx context.Context, talker string, talkerName string, startTime, endTime time.Time) ([]byte, error) {
	query := types.MessageQuery{
		Talker:    talker,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     200000,
		Offset:    0,
	}
	messages, err := s.Store.GetMessages(ctx, query)
	if err != nil {
		return nil, err
	}

	log.Info().Int("count", len(messages)).Str("talker", talkerName).Msg("ExportForensic processing")

	exportTime := time.Now()

	// 1. Generate CSV data
	csvData, err := s.ExportChatCSV(ctx, talker, talkerName, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("生成CSV数据失败: %w", err)
	}

	// 2. Generate forensic HTML report (replaces PDF)
	htmlData, err := s.buildForensicHTML(talkerName, talker, exportTime, startTime, endTime, messages)
	if err != nil {
		return nil, fmt.Errorf("生成取证HTML报告失败: %w", err)
	}

	// 3. Calculate SHA-256 checksums
	fileChecksums := map[string]string{
		"report.html":   fmt.Sprintf("%x", sha256.Sum256(htmlData)),
		"chat_data.csv": fmt.Sprintf("%x", sha256.Sum256(csvData)),
	}

	// 4. Build metadata.json
	metadata := ForensicMetadata{
		ExportTime:      exportTime.Format(time.RFC3339),
		SoftwareVersion: "1.0.0",
		Talker:          talker,
		TalkerName:      talkerName,
		TimeRange: map[string]string{
			"start": startTime.Format(time.RFC3339),
			"end":   endTime.Format(time.RFC3339),
		},
		MessageCount: len(messages),
		Files: map[string]string{
			"report.html":   "sha256:" + fileChecksums["report.html"],
			"chat_data.csv": "sha256:" + fileChecksums["chat_data.csv"],
		},
	}
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("生成metadata失败: %w", err)
	}

	// 5. Build checksums.sha256 file
	var checksumBuf bytes.Buffer
	for name, hash := range fileChecksums {
		fmt.Fprintf(&checksumBuf, "%s  %s\n", hash, name)
	}
	metaHash := fmt.Sprintf("%x", sha256.Sum256(metadataJSON))
	fmt.Fprintf(&checksumBuf, "%s  %s\n", metaHash, "metadata.json")

	// 6. Pack everything into a ZIP
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)

	zipFiles := map[string][]byte{
		"report.html":     htmlData,
		"chat_data.csv":   csvData,
		"metadata.json":   metadataJSON,
		"checksums.sha256": checksumBuf.Bytes(),
	}
	for name, data := range zipFiles {
		f, err := zw.Create(name)
		if err != nil {
			return nil, fmt.Errorf("创建ZIP条目失败: %w", err)
		}
		if _, err := f.Write(data); err != nil {
			return nil, fmt.Errorf("写入ZIP条目失败: %w", err)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("关闭ZIP失败: %w", err)
	}

	return zipBuf.Bytes(), nil
}

// forensicDateGroup groups messages by date for the HTML template.
type forensicDateGroup struct {
	Date     string
	Messages []forensicMsg
}

// forensicMsg holds a single message for the HTML template.
type forensicMsg struct {
	Time    string
	Sender  string
	Content string
	IsSelf  bool
}

// forensicTemplateData holds all data passed to the HTML template.
type forensicTemplateData struct {
	Title        string
	ReportID     string
	ExportTime   string
	TalkerName   string
	Talker       string
	StartTime    string
	EndTime      string
	MessageCount int
	DateGroups   []forensicDateGroup
}

// buildForensicHTML generates an evidence-grade HTML report with watermark,
// header, metadata, chat records grouped by date, integrity verification,
// and signature area.
func (s *Service) buildForensicHTML(talkerName, talker string, exportTime, startTime, endTime time.Time, messages []*model.Message) ([]byte, error) {
	reportID := fmt.Sprintf("WT-%s-%s", exportTime.Format("20060102150405"), fmt.Sprintf("%x", sha256.Sum256([]byte(talker+exportTime.String())))[:8])

	// Group messages by date
	dateGroups := s.groupMessagesByDate(messages)

	data := forensicTemplateData{
		Title:        "WeTrace 取证报告",
		ReportID:     reportID,
		ExportTime:   exportTime.Format("2006-01-02 15:04:05"),
		TalkerName:   talkerName,
		Talker:       talker,
		StartTime:    startTime.Format("2006-01-02 15:04:05"),
		EndTime:      endTime.Format("2006-01-02 15:04:05"),
		MessageCount: len(messages),
		DateGroups:   dateGroups,
	}

	tmpl, err := template.New("forensic").Parse(forensicHTMLTemplate)
	if err != nil {
		return nil, fmt.Errorf("解析HTML模板失败: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("渲染HTML模板失败: %w", err)
	}

	return buf.Bytes(), nil
}

// groupMessagesByDate groups messages into date-based groups for display.
func (s *Service) groupMessagesByDate(messages []*model.Message) []forensicDateGroup {
	var groups []forensicDateGroup
	currentDate := ""

	for _, msg := range messages {
		dateStr := msg.Time.Format("2006-01-02")

		sender := msg.SenderName
		if sender == "" {
			sender = msg.Sender
		}
		msg.SetContent("host", "127.0.0.1:5200/api/v1/media")
		content := msg.PlainTextContent()

		fMsg := forensicMsg{
			Time:    msg.Time.Format("15:04:05"),
			Sender:  sender,
			Content: content,
			IsSelf:  msg.IsSelf,
		}

		if dateStr != currentDate {
			currentDate = dateStr
			groups = append(groups, forensicDateGroup{
				Date:     dateStr,
				Messages: []forensicMsg{fMsg},
			})
		} else if len(groups) > 0 {
			groups[len(groups)-1].Messages = append(groups[len(groups)-1].Messages, fMsg)
		}
	}

	return groups
}

// forensicHTMLTemplate is the embedded HTML template for forensic reports.
const forensicHTMLTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.Title}} - {{.ReportID}}</title>
<style>
/* === Base Reset === */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

body {
    font-family: "Microsoft YaHei", "PingFang SC", "Helvetica Neue", Arial, sans-serif;
    background: #f0f2f5; color: #1a1a2e; line-height: 1.6; position: relative;
}

/* === Watermark (pure CSS, no JS needed) === */
body::after {
    content: "";
    position: fixed; top: -50%; left: -50%; width: 200%; height: 200%;
    z-index: 9999; pointer-events: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='480' height='320'%3E%3Ctext x='50%25' y='50%25' dominant-baseline='middle' text-anchor='middle' font-family='Microsoft YaHei,PingFang SC,sans-serif' font-size='22' font-weight='700' fill='rgba(180,40,40,0.055)' transform='rotate(-35,240,160)'%3E%E5%8F%96%E8%AF%81%E8%AF%81%E6%8D%AE FORENSIC EVIDENCE%3C/text%3E%3C/svg%3E");
    background-repeat: repeat;
}

/* === Page Container === */
.page { max-width: 900px; margin: 40px auto; padding: 0; }

/* === Report Header === */
.report-header {
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
    color: #fff; padding: 50px 60px; border-radius: 8px 8px 0 0;
    position: relative; overflow: hidden;
}
.report-header::after {
    content: "FORENSIC"; position: absolute; right: -20px; bottom: -10px;
    font-size: 120px; font-weight: 900; opacity: 0.04; letter-spacing: 8px;
}
.report-header h1 { font-size: 28px; font-weight: 700; margin-bottom: 8px; }
.report-header .subtitle { font-size: 14px; opacity: 0.7; }
.report-header .report-id {
    display: inline-block; margin-top: 16px; padding: 6px 16px;
    background: rgba(255,255,255,0.12); border-radius: 4px;
    font-family: "Courier New", monospace; font-size: 13px; letter-spacing: 1px;
}

/* === Content Body === */
.report-body {
    background: #fff; padding: 50px 60px; border-radius: 0 0 8px 8px;
    box-shadow: 0 2px 20px rgba(0,0,0,0.06);
}

/* === Section === */
.section { margin-bottom: 40px; }
.section-title {
    font-size: 18px; font-weight: 700; color: #1a1a2e;
    padding-bottom: 12px; margin-bottom: 20px;
    border-bottom: 2px solid #e8e8e8; position: relative;
}
.section-title::before {
    content: ""; position: absolute; bottom: -2px; left: 0;
    width: 60px; height: 2px; background: #0f3460;
}

/* === Metadata Table === */
.meta-table { width: 100%; border-collapse: collapse; }
.meta-table td {
    padding: 10px 16px; border: 1px solid #e8e8e8; font-size: 14px;
}
.meta-table td:first-child {
    width: 160px; background: #f8f9fa; font-weight: 600; color: #555;
    white-space: nowrap;
}

/* === Chat Records === */
.date-group { margin-bottom: 24px; }
.date-header {
    text-align: center; margin: 20px 0 16px; position: relative;
}
.date-header::before {
    content: ""; position: absolute; top: 50%; left: 0; right: 0;
    height: 1px; background: #e0e0e0;
}
.date-header span {
    position: relative; background: #fff; padding: 0 20px;
    font-size: 13px; color: #999; font-weight: 500;
}
.msg-item {
    display: flex; gap: 12px; padding: 8px 0;
    border-bottom: 1px solid #f5f5f5; font-size: 14px;
}
.msg-time {
    flex-shrink: 0; width: 70px; color: #999;
    font-family: "Courier New", monospace; font-size: 13px; padding-top: 2px;
}
.msg-sender {
    flex-shrink: 0; width: 120px; font-weight: 600; color: #333;
    overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.msg-sender.self { color: #0f3460; }
.msg-text { flex: 1; color: #444; word-break: break-all; }

/* === Integrity Section === */
.checksum-box {
    background: #f8f9fa; border: 1px solid #e8e8e8; border-radius: 6px;
    padding: 20px; font-family: "Courier New", monospace; font-size: 13px;
    line-height: 2; word-break: break-all;
}
.checksum-label { color: #999; font-size: 12px; display: block; margin-bottom: 2px; font-family: "Microsoft YaHei", sans-serif; }
.checksum-value { color: #1a1a2e; font-weight: 600; }

/* === Signature Section === */
.signature-area { margin-top: 60px; }
.sig-row {
    display: flex; align-items: flex-end; gap: 16px;
    margin-bottom: 40px;
}
.sig-label { font-size: 14px; color: #555; font-weight: 600; white-space: nowrap; }
.sig-line {
    flex: 1; border-bottom: 1px solid #333;
    min-width: 200px; height: 30px;
}

/* === Footer === */
.report-footer {
    text-align: center; padding: 30px 0; font-size: 12px; color: #bbb;
}

/* === Print Styles === */
@media print {
    body { background: #fff; }
    .page { margin: 0; max-width: 100%; }
    .report-header { border-radius: 0; }
    .report-body { box-shadow: none; border-radius: 0; padding: 30px 40px; }
    body::after { opacity: 0.6; }
    .msg-item { break-inside: avoid; }
    .section { break-inside: avoid; }
    .signature-area { break-before: page; }
}
</style>
</head>
<body>

<div class="page">

<!-- Report Header -->
<div class="report-header">
    <h1>{{.Title}}</h1>
    <div class="subtitle">WeTrace Digital Forensic Evidence Report</div>
    <div class="report-id">{{.ReportID}}</div>
</div>

<div class="report-body">

<!-- Section 1: Metadata -->
<div class="section">
    <div class="section-title">一、报告信息</div>
    <table class="meta-table">
        <tr><td>报告编号</td><td>{{.ReportID}}</td></tr>
        <tr><td>生成时间</td><td>{{.ExportTime}}</td></tr>
        <tr><td>聊天对象</td><td>{{.TalkerName}} ({{.Talker}})</td></tr>
        <tr><td>数据时间范围</td><td>{{.StartTime}} ~ {{.EndTime}}</td></tr>
        <tr><td>消息总数</td><td>{{.MessageCount}} 条</td></tr>
        <tr><td>软件版本</td><td>WeTrace v1.0.0</td></tr>
    </table>
</div>

<!-- Section 2: Chat Records -->
<div class="section">
    <div class="section-title">二、聊天记录</div>
    {{range .DateGroups}}
    <div class="date-group">
        <div class="date-header"><span>{{.Date}}</span></div>
        {{range .Messages}}
        <div class="msg-item">
            <div class="msg-time">{{.Time}}</div>
            <div class="msg-sender{{if .IsSelf}} self{{end}}">{{.Sender}}</div>
            <div class="msg-text">{{.Content}}</div>
        </div>
        {{end}}
    </div>
    {{end}}
</div>

<!-- Section 3: Integrity Verification -->
<div class="section">
    <div class="section-title">三、完整性验证</div>
    <p style="font-size:14px; color:#666; margin-bottom:16px;">
        本导出包中的所有文件均附带 SHA-256 哈希校验值，可通过 <code>checksums.sha256</code> 文件验证数据完整性。
    </p>
    <div class="checksum-box">
        <span class="checksum-label">report.html</span>
        <span class="checksum-value">请参阅 checksums.sha256 文件</span><br>
        <span class="checksum-label">chat_data.csv</span>
        <span class="checksum-value">请参阅 checksums.sha256 文件</span>
    </div>
    <p style="font-size:13px; color:#999; margin-top:12px;">
        验证方法：在命令行中执行 <code>sha256sum -c checksums.sha256</code>，或使用其他 SHA-256 校验工具逐一比对哈希值。
        如哈希值与校验文件中记录的值一致，则表明数据自导出后未被篡改。
    </p>
</div>

<!-- Section 4: Signature -->
<div class="section signature-area">
    <div class="section-title">四、签名确认</div>
    <div class="sig-row">
        <span class="sig-label">导出操作人签名：</span>
        <div class="sig-line"></div>
    </div>
    <div class="sig-row">
        <span class="sig-label">日期：</span>
        <div class="sig-line"></div>
    </div>
    <div class="sig-row">
        <span class="sig-label">备注：</span>
        <div class="sig-line"></div>
    </div>
</div>

</div><!-- .report-body -->

<div class="report-footer">
    本报告由 WeTrace 软件自动生成 | 报告编号: {{.ReportID}} | 生成时间: {{.ExportTime}}
</div>

</div><!-- .page -->
</body>
</html>`
