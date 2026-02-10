package export

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/afumu/wetrace/store/types"
	"github.com/rs/zerolog/log"
	"github.com/signintech/gopdf"
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
// - report.pdf (evidence-grade PDF report)
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

	// 2. Generate forensic PDF report
	pdfData, err := s.buildForensicPDF(talkerName, talker, exportTime, startTime, endTime, len(messages))
	if err != nil {
		return nil, fmt.Errorf("生成取证PDF报告失败: %w", err)
	}

	// 3. Calculate SHA-256 checksums
	fileChecksums := map[string]string{
		"report.pdf":    fmt.Sprintf("%x", sha256.Sum256(pdfData)),
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
			"report.pdf":    "sha256:" + fileChecksums["report.pdf"],
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
	// Also include metadata.json checksum
	metaHash := fmt.Sprintf("%x", sha256.Sum256(metadataJSON))
	fmt.Fprintf(&checksumBuf, "%s  %s\n", metaHash, "metadata.json")

	// 6. Pack everything into a ZIP
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)

	zipFiles := map[string][]byte{
		"report.pdf":       pdfData,
		"chat_data.csv":    csvData,
		"metadata.json":    metadataJSON,
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

// buildForensicPDF generates an evidence-grade PDF report with cover page,
// summary, and integrity statement.
func (s *Service) buildForensicPDF(talkerName, talker string, exportTime, startTime, endTime time.Time, msgCount int) ([]byte, error) {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	fontPath := chineseFontPath()
	if err := pdf.AddTTFFont("chinese", fontPath); err != nil {
		return nil, fmt.Errorf("加载字体失败: %w", err)
	}

	// === Cover Page ===
	pdf.AddPage()
	curY := 200.0

	if err := pdf.SetFont("chinese", "", 24); err != nil {
		return nil, fmt.Errorf("设置字体失败: %w", err)
	}
	pdf.SetX(pdfMarginLeft)
	pdf.SetY(curY)
	pdf.Cell(nil, "WeTrace 取证报告")
	curY += 50

	if err := pdf.SetFont("chinese", "", 14); err != nil {
		return nil, fmt.Errorf("设置字体失败: %w", err)
	}
	coverLines := []string{
		fmt.Sprintf("聊天对象: %s (%s)", talkerName, talker),
		fmt.Sprintf("导出时间: %s", exportTime.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("数据范围: %s ~ %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")),
		fmt.Sprintf("消息总数: %d", msgCount),
		"软件版本: WeTrace v1.0.0",
	}
	for _, line := range coverLines {
		pdf.SetX(pdfMarginLeft)
		pdf.SetY(curY)
		pdf.Cell(nil, line)
		curY += 28
	}

	// === Summary Page ===
	pdf.AddPage()
	curY = pdfMarginTop

	if err := pdf.SetFont("chinese", "", pdfTitleSize); err != nil {
		return nil, fmt.Errorf("设置字体失败: %w", err)
	}
	pdf.SetX(pdfMarginLeft)
	pdf.SetY(curY)
	pdf.Cell(nil, "一、数据摘要")
	curY += pdfLineHeight * 2

	if err := pdf.SetFont("chinese", "", pdfFontSize); err != nil {
		return nil, fmt.Errorf("设置字体失败: %w", err)
	}
	summaryLines := []string{
		fmt.Sprintf("本报告由 WeTrace 软件自动生成，用于记录微信聊天数据的导出过程。"),
		"",
		fmt.Sprintf("聊天对象标识: %s", talker),
		fmt.Sprintf("聊天对象名称: %s", talkerName),
		fmt.Sprintf("数据时间范围: %s 至 %s", startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("导出消息总数: %d 条", msgCount),
		fmt.Sprintf("报告生成时间: %s", exportTime.Format("2006-01-02 15:04:05")),
		"",
		"本导出包包含以下文件:",
		"  - report.pdf: 本取证报告文档",
		"  - chat_data.csv: 原始聊天记录数据（CSV格式）",
		"  - checksums.sha256: 所有文件的SHA-256哈希校验值",
		"  - metadata.json: 导出元数据（JSON格式）",
	}
	for _, line := range summaryLines {
		if curY+pdfLineHeight > pdfPageHeight-pdfMarginBot {
			pdf.AddPage()
			curY = pdfMarginTop
		}
		pdf.SetX(pdfMarginLeft)
		pdf.SetY(curY)
		pdf.Cell(nil, line)
		curY += pdfLineHeight
	}

	// === Integrity Statement Page ===
	pdf.AddPage()
	curY = pdfMarginTop

	if err := pdf.SetFont("chinese", "", pdfTitleSize); err != nil {
		return nil, fmt.Errorf("设置字体失败: %w", err)
	}
	pdf.SetX(pdfMarginLeft)
	pdf.SetY(curY)
	pdf.Cell(nil, "二、完整性声明")
	curY += pdfLineHeight * 2

	if err := pdf.SetFont("chinese", "", pdfFontSize); err != nil {
		return nil, fmt.Errorf("设置字体失败: %w", err)
	}
	integrityLines := []string{
		"本导出包中的所有文件均附带 SHA-256 哈希校验值，",
		"可通过 checksums.sha256 文件验证数据完整性。",
		"",
		"验证方法:",
		"  在命令行中执行: sha256sum -c checksums.sha256",
		"  或使用其他 SHA-256 校验工具逐一比对哈希值。",
		"",
		"如哈希值与校验文件中记录的值一致，则表明数据自导出后未被篡改。",
	}
	for _, line := range integrityLines {
		pdf.SetX(pdfMarginLeft)
		pdf.SetY(curY)
		pdf.Cell(nil, line)
		curY += pdfLineHeight
	}

	// === Signature Area ===
	curY += pdfLineHeight * 3

	signatureLines := []string{
		"三、签名区域",
		"",
		"导出操作人签名: ____________________",
		"",
		"日期: ____________________",
		"",
		"备注: ____________________",
	}

	if err := pdf.SetFont("chinese", "", pdfTitleSize); err != nil {
		return nil, fmt.Errorf("设置字体失败: %w", err)
	}
	pdf.SetX(pdfMarginLeft)
	pdf.SetY(curY)
	pdf.Cell(nil, signatureLines[0])
	curY += pdfLineHeight * 2

	if err := pdf.SetFont("chinese", "", pdfFontSize); err != nil {
		return nil, fmt.Errorf("设置字体失败: %w", err)
	}
	for _, line := range signatureLines[2:] {
		pdf.SetX(pdfMarginLeft)
		pdf.SetY(curY)
		pdf.Cell(nil, line)
		curY += pdfLineHeight * 1.5
	}

	var buf bytes.Buffer
	if _, err := pdf.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("写入PDF失败: %w", err)
	}

	return buf.Bytes(), nil
}
