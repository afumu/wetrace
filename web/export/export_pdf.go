package export

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/afumu/wetrace/store/types"
	"github.com/rs/zerolog/log"
	"github.com/signintech/gopdf"
)

const (
	pdfPageWidth  = 595.28 // A4 width in points
	pdfPageHeight = 841.89 // A4 height in points
	pdfMarginLeft = 50.0
	pdfMarginTop  = 60.0
	pdfMarginBot  = 60.0
	pdfLineHeight = 18.0
	pdfFontSize   = 10.0
	pdfTitleSize  = 16.0
	pdfDateSize   = 12.0
)

// chineseFontPath returns the path to a Chinese-capable font on the current OS.
func chineseFontPath() string {
	candidates := []string{}

	switch runtime.GOOS {
	case "darwin":
		candidates = []string{
			"/System/Library/Fonts/STHeiti Medium.ttc",
			"/System/Library/Fonts/PingFang.ttc",
			"/Library/Fonts/Arial Unicode.ttf",
		}
	case "windows":
		candidates = []string{
			"C:\\Windows\\Fonts\\msyh.ttc",
			"C:\\Windows\\Fonts\\simsun.ttc",
			"C:\\Windows\\Fonts\\simhei.ttf",
		}
	default: // linux
		candidates = []string{
			"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/noto-cjk/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		}
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// Return the first candidate as fallback; gopdf will produce a clear error if missing
	if len(candidates) > 0 {
		return candidates[0]
	}
	return ""
}

// ExportChatPDF exports chat messages as a PDF file.
func (s *Service) ExportChatPDF(ctx context.Context, talker string, talkerName string, startTime, endTime time.Time) ([]byte, error) {
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

	log.Info().Int("count", len(messages)).Str("talker", talkerName).Msg("ExportPDF processing")

	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	// Load Chinese font
	fontPath := chineseFontPath()
	if err := pdf.AddTTFFont("chinese", fontPath); err != nil {
		return nil, fmt.Errorf("加载字体失败: %w", err)
	}

	// Add title page
	pdf.AddPage()
	curY := pdfMarginTop

	if err := pdf.SetFont("chinese", "", pdfTitleSize); err != nil {
		return nil, fmt.Errorf("设置字体失败: %w", err)
	}
	pdf.SetX(pdfMarginLeft)
	pdf.SetY(curY)
	title := talkerName + " 的聊天记录"
	pdf.Cell(nil, title)
	curY += pdfLineHeight * 2

	// Subtitle with export info
	if err := pdf.SetFont("chinese", "", pdfFontSize); err != nil {
		return nil, fmt.Errorf("设置字体失败: %w", err)
	}
	pdf.SetX(pdfMarginLeft)
	pdf.SetY(curY)
	pdf.Cell(nil, fmt.Sprintf("导出时间: %s", time.Now().Format("2006-01-02 15:04:05")))
	curY += pdfLineHeight

	pdf.SetX(pdfMarginLeft)
	pdf.SetY(curY)
	pdf.Cell(nil, fmt.Sprintf("消息总数: %d", len(messages)))
	curY += pdfLineHeight * 2

	// Write messages grouped by date
	currentDate := ""
	for _, msg := range messages {
		dateStr := msg.Time.Format("2006-01-02")

		// New date header
		if dateStr != currentDate {
			currentDate = dateStr
			curY += pdfLineHeight * 0.5

			// Check page space for date header
			if curY+pdfLineHeight*2 > pdfPageHeight-pdfMarginBot {
				pdf.AddPage()
				curY = pdfMarginTop
			}

			if err := pdf.SetFont("chinese", "", pdfDateSize); err != nil {
				return nil, fmt.Errorf("设置字体失败: %w", err)
			}
			pdf.SetX(pdfMarginLeft)
			pdf.SetY(curY)
			pdf.Cell(nil, "--- "+dateStr+" ---")
			curY += pdfLineHeight * 1.5
		}

		// Message content
		sender := msg.SenderName
		if sender == "" {
			sender = msg.Sender
		}
		msg.SetContent("host", "127.0.0.1:5200/api/v1/media")
		content := msg.PlainTextContent()

		line := fmt.Sprintf("[%s] %s  %s",
			sender,
			msg.Time.Format("15:04:05"),
			content,
		)

		// Write message line(s) with word wrap
		curY, err = pdfWriteWrappedText(pdf, line, curY)
		if err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if _, err := pdf.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("写入PDF失败: %w", err)
	}

	return buf.Bytes(), nil
}

// pdfWriteWrappedText writes text with automatic line wrapping and page breaks.
func pdfWriteWrappedText(pdf *gopdf.GoPdf, text string, curY float64) (float64, error) {
	if err := pdf.SetFont("chinese", "", pdfFontSize); err != nil {
		return curY, fmt.Errorf("设置字体失败: %w", err)
	}

	maxWidth := pdfPageWidth - pdfMarginLeft*2
	runes := []rune(text)

	for len(runes) > 0 {
		// Check if we need a new page
		if curY+pdfLineHeight > pdfPageHeight-pdfMarginBot {
			pdf.AddPage()
			curY = pdfMarginTop
			if err := pdf.SetFont("chinese", "", pdfFontSize); err != nil {
				return curY, fmt.Errorf("设置字体失败: %w", err)
			}
		}

		// Find how many runes fit in one line
		lineEnd := len(runes)
		for i := 1; i <= len(runes); i++ {
			w, _ := pdf.MeasureTextWidth(string(runes[:i]))
			if w > maxWidth {
				lineEnd = i - 1
				if lineEnd < 1 {
					lineEnd = 1
				}
				break
			}
		}

		pdf.SetX(pdfMarginLeft)
		pdf.SetY(curY)
		pdf.Cell(nil, string(runes[:lineEnd]))
		curY += pdfLineHeight
		runes = runes[lineEnd:]
	}

	return curY, nil
}
