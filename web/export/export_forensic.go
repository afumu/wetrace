package export

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/afumu/wetrace/store/types"
	"github.com/rs/zerolog/log"
)

// ExportForensic 重新实现：复用精美版 HTML，增加专业取证特性
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

	log.Info().Int("count", len(messages)).Str("talker", talkerName).Msg("ExportForensic processing (Beautiful HTML Mode)")

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// 1. 处理媒体文件并注入 _url 到消息内容中 (复用核心逻辑)
	for _, msg := range messages {
		s.processMedia(ctx, zw, msg)
	}

	// 2. 生成 data.js
	msgJson, _ := json.Marshal(messages)
	dataJs := fmt.Sprintf("window.CHAT_DATA = %s;", string(msgJson))
	fData, _ := zw.Create("data.js")
	fData.Write([]byte(dataJs))

	// 3. 生成增强取证特性的 index.html
	exportTime := time.Now()
	reportID := fmt.Sprintf("FORENSIC-%s-%s", exportTime.Format("20060102150405"), fmt.Sprintf("%x", sha256.Sum256([]byte(talker+exportTime.String())))[:6])

	// 计算数据指纹
	h := sha256.New()
	h.Write(msgJson)
	dataHash := fmt.Sprintf("%x", h.Sum(nil))

	html, err := s.buildForensicBeautifulHTML(talkerName, talker, reportID, exportTime, startTime, endTime, len(messages), dataHash)
	if err != nil {
		return nil, err
	}
	fHtml, _ := zw.Create("index.html")
	fHtml.Write([]byte(html))

	// 4. 生成 metadata.json
	metadata := map[string]interface{}{
		"type":             "forensic_report",
		"report_id":        reportID,
		"export_time":      exportTime.Format(time.RFC3339),
		"talker":           talker,
		"talker_name":      talkerName,
		"message_count":    len(messages),
		"data_fingerprint": dataHash,
	}
	metaJson, _ := json.MarshalIndent(metadata, "", "  ")
	fMeta, _ := zw.Create("metadata.json")
	fMeta.Write(metaJson)

	// 5. 复制静态资源
	s.copyAssets(zw)

	zw.Close()
	return buf.Bytes(), nil
}

// buildForensicBeautifulHTML 在基础 HTML 模板上注入取证样式和内容
func (s *Service) buildForensicBeautifulHTML(talkerName, talker, reportID string, exportTime, startTime, endTime time.Time, count int, dataHash string) (string, error) {
	html := exportTemplate

	// 1. 注入 CSS 样式 (水印 + 取证卡片)
	forensicStyles := `
        /* 专业取证水印 */
        body::before {
            content: "";
            position: fixed;
            top: 0; left: 0; width: 100%; height: 100%;
            z-index: 9999;
            pointer-events: none;
            background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='300' viewBox='0 0 400 300'%3E%3Ctext x='50%25' y='50%25' font-family='sans-serif' font-weight='bold' font-size='12' fill='rgba(150, 0, 0, 0.04)' text-anchor='middle' transform='rotate(-30 200 150)'%3E取证证据 FORENSIC - ` + reportID + ` - ` + exportTime.Format("2006/01/02") + `%3C/text%3E%3C/svg%3E");
            background-repeat: repeat;
        }

        /* 取证摘要卡片 */
        .forensic-summary {
            background: #fff;
            margin: 20px 30px 10px;
            padding: 24px;
            border-radius: 8px;
            border: 1px solid #e0e0e0;
            border-left: 5px solid #1a1a2e;
            box-shadow: 0 2px 8px rgba(0,0,0,0.05);
            flex-shrink: 0;
        }
        .forensic-summary h2 { 
            font-size: 18px; color: #1a1a2e; margin-bottom: 15px; 
            display: flex; align-items: center; gap: 10px;
        }
        .forensic-summary .grid {
            display: grid; grid-template-columns: 1fr 1fr; gap: 12px;
        }
        .forensic-summary .item { font-size: 13px; color: #666; }
        .forensic-summary .label { font-weight: bold; color: #333; margin-right: 8px; }
        .forensic-summary .hash { 
            grid-column: span 2; margin-top: 8px; padding: 8px; 
            background: #f8f9fa; border-radius: 4px; font-family: monospace; 
            font-size: 11px; color: #888; border: 1px dashed #ddd;
            word-break: break-all;
        }

        .forensic-footer {
            margin: 40px 30px; padding: 20px;
            border-top: 1px solid #ddd;
            text-align: center; font-size: 12px; color: #999;
            flex-shrink: 0;
        }
    `
	html = strings.Replace(html, "<!-- CSS_PLACEHOLDER -->", "<style>"+forensicStyles+"</style>", 1)

	// 2. 构造取证摘要 HTML
	summaryHTML := `
        <div class="forensic-summary">
            <h2><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg> 电子数据取证摘要 (Forensic Summary)</h2>
            <div class="grid">
                <div class="item"><span class="label">报告编号:</span>` + reportID + `</div>
                <div class="item"><span class="label">导出时间:</span>` + exportTime.Format("2006-01-02 15:04:05") + `</div>
                <div class="item"><span class="label">聊天对象:</span>` + talkerName + ` (` + talker + `)</div>
                <div class="item"><span class="label">消息总数:</span>` + fmt.Sprintf("%d", count) + ` 条</div>
                <div class="item"><span class="label">起始时间:</span>` + startTime.Format("2006-01-02 15:04:05") + `</div>
                <div class="item"><span class="label">截止时间:</span>` + endTime.Format("2006-01-02 15:04:05") + `</div>
                <div class="hash">
                    <div style="margin-bottom:4px; font-weight:bold; color:#555;">数据校验指纹 (SHA-256 Fingerprint):</div>
                    ` + dataHash + `
                </div>
            </div>
        </div>
    `
	// 注入到渲染列表之前
	html = strings.Replace(html, `<div id="msg-list" class="msg-list"></div>`, summaryHTML+`<div id="msg-list" class="msg-list"></div>`, 1)

	// 3. 注入脚注
	footerHTML := `
        <div class="forensic-footer">
            <p>WeTrace 取证报告 | 声明：本数据由 WeTrace 取证工具自动采集，数据完整性由上述 SHA-256 指纹校验。任何篡改均会导致指纹失效。</p>
            <p style="margin-top:8px;">导出人员签名：____________________  日期：____________________</p>
        </div>
    `
	html = strings.Replace(html, `</div><!-- #app -->`, footerHTML+`</div><!-- #app -->`, 1)

	return html, nil
}
