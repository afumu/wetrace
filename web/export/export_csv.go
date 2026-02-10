package export

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/afumu/wetrace/store/types"
	"github.com/rs/zerolog/log"
)

// ExportChatCSV 导出聊天记录为 CSV 格式
func (s *Service) ExportChatCSV(ctx context.Context, talker string, talkerName string, startTime, endTime time.Time) ([]byte, error) {
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

	log.Info().Int("count", len(messages)).Str("talker", talkerName).Msg("ExportCSV processing")

	var buf bytes.Buffer

	// 写入 UTF-8 BOM，确保 Excel 正确识别编码
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(&buf)

	// 写入表头
	header := []string{"时间", "发送人昵称", "发送人ID", "聊天对象昵称", "聊天对象ID", "内容"}
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("写入CSV表头失败: %w", err)
	}

	// 写入数据行
	for _, msg := range messages {
		row := msg.CSV("127.0.0.1:5200/api/v1/media")
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("写入CSV数据失败: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("CSV写入错误: %w", err)
	}

	return buf.Bytes(), nil
}
