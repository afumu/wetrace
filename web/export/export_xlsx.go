package export

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/afumu/wetrace/store/types"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

// ExportChatXLSX 导出聊天记录为 XLSX 格式
func (s *Service) ExportChatXLSX(ctx context.Context, talker string, talkerName string, startTime, endTime time.Time) ([]byte, error) {
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

	log.Info().Int("count", len(messages)).Str("talker", talkerName).Msg("ExportXLSX processing")

	f := excelize.NewFile()
	defer f.Close()

	// 使用聊天对象昵称作为 Sheet 名称
	sheetName := talkerName
	if len(sheetName) > 31 {
		sheetName = sheetName[:31]
	}
	f.SetSheetName("Sheet1", sheetName)

	// 写入表头
	headers := []string{"时间", "发送人", "发送人ID", "聊天对象", "聊天对象ID", "内容"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellStyle(sheetName, "A1", "F1", headerStyle)

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 20)
	f.SetColWidth(sheetName, "B", "B", 15)
	f.SetColWidth(sheetName, "C", "C", 20)
	f.SetColWidth(sheetName, "D", "D", 15)
	f.SetColWidth(sheetName, "E", "E", 20)
	f.SetColWidth(sheetName, "F", "F", 50)

	// 写入数据行
	for i, msg := range messages {
		row := msg.CSV("127.0.0.1:5200/api/v1/media")
		rowNum := i + 2
		for j, val := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowNum)
			f.SetCellValue(sheetName, cell, val)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("写入XLSX失败: %w", err)
	}

	return buf.Bytes(), nil
}
