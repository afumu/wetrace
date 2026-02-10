package export

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/types"
	"github.com/gomutex/godocx"
	"github.com/gomutex/godocx/docx"
	"github.com/rs/zerolog/log"
)

// ExportChatDOCX 导出聊天记录为 DOCX 格式
func (s *Service) ExportChatDOCX(ctx context.Context, talker string, talkerName string, startTime, endTime time.Time) ([]byte, error) {
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

	log.Info().Int("count", len(messages)).Str("talker", talkerName).Msg("ExportDOCX processing")

	doc, err := godocx.NewDocument()
	if err != nil {
		return nil, fmt.Errorf("创建DOCX文档失败: %w", err)
	}
	defer doc.Close()

	// 添加标题
	doc.AddHeading(talkerName+" 的聊天记录", 1)
	doc.AddEmptyParagraph()

	// 按日期分段写入消息
	writeMessages(doc, messages)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("写入DOCX失败: %w", err)
	}

	return buf.Bytes(), nil
}

// writeMessages 按日期分段写入消息到 DOCX 文档
func writeMessages(doc *docx.RootDoc, messages []*model.Message) {
	currentDate := ""

	for _, msg := range messages {
		dateStr := msg.Time.Format("2006-01-02")

		// 新的日期段落
		if dateStr != currentDate {
			currentDate = dateStr
			doc.AddEmptyParagraph()
			doc.AddHeading(dateStr, 2)
		}

		// 发送人
		sender := msg.SenderName
		if sender == "" {
			sender = msg.Sender
		}

		// 设置 Host 以生成正确的本地链接
		msg.SetContent("host", "127.0.0.1:5200/api/v1/media")
		content := msg.PlainTextContent()

		line := fmt.Sprintf("[%s] %s\n%s",
			sender,
			msg.Time.Format("15:04:05"),
			content,
		)
		doc.AddParagraph(line)
	}
}
