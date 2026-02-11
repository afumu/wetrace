package api

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/types"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

// GetContacts 处理获取联系人列表的请求。
func (a *API) GetContacts(c *gin.Context) {
	var pageQuery transport.PaginationQuery
	if err := c.ShouldBindQuery(&pageQuery); err != nil {
		transport.BadRequest(c, "无效的分页参数: "+err.Error())
		return
	}

	var keywordQuery transport.KeywordQuery
	if err := c.ShouldBindQuery(&keywordQuery); err != nil {
		transport.BadRequest(c, "无效的关键字参数: "+err.Error())
		return
	}

	query := types.ContactQuery{
		Keyword: keywordQuery.Keyword,
		Limit:   pageQuery.Limit,
		Offset:  pageQuery.Offset,
	}

	contacts, err := a.Store.GetContacts(c.Request.Context(), query)
	if err != nil {
		log.Error().Err(err).Msg("从 store 获取联系人失败")
		transport.InternalServerError(c, "获取联系人列表失败。")
		return
	}

	if contacts == nil {
		contacts = make([]*model.Contact, 0)
	}

	transport.SendSuccess(c, contacts)
}

// GetContactByID 处理通过 ID 获取单个联系人信息的请求。
func (a *API) GetContactByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		transport.BadRequest(c, "联系人 ID 是必需的。")
		return
	}

	query := types.ContactQuery{
		Keyword: id,
		Limit:   1,
	}

	contacts, err := a.Store.GetContacts(c.Request.Context(), query)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("从 store 通过 ID 获取联系人失败")
		transport.InternalServerError(c, "获取联系人信息失败。")
		return
	}

	if len(contacts) == 0 {
		transport.NotFound(c, "未找到联系人。")
		return
	}

	transport.SendSuccess(c, contacts[0])
}

// ExportContacts 处理联系人导出请求，支持 CSV 和 XLSX 格式。
func (a *API) ExportContacts(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	keyword := c.Query("keyword")

	query := types.ContactQuery{
		Keyword: keyword,
		Limit:   100000,
		Offset:  0,
	}

	contacts, err := a.Store.GetContacts(c.Request.Context(), query)
	if err != nil {
		log.Error().Err(err).Msg("导出联系人时获取数据失败")
		transport.InternalServerError(c, "获取联系人数据失败。")
		return
	}

	if contacts == nil {
		contacts = make([]*model.Contact, 0)
	}

	dateStr := time.Now().Format("20060102")

	switch format {
	case "xlsx":
		data, err := a.exportContactsXLSX(contacts)
		if err != nil {
			log.Error().Err(err).Msg("生成联系人 XLSX 失败")
			transport.InternalServerError(c, "生成 XLSX 文件失败。")
			return
		}
		fileName := fmt.Sprintf("contacts_export_%s.xlsx", dateStr)
		contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		c.Data(http.StatusOK, contentType, data)
	default:
		data, err := a.exportContactsCSV(contacts)
		if err != nil {
			log.Error().Err(err).Msg("生成联系人 CSV 失败")
			transport.InternalServerError(c, "生成 CSV 文件失败。")
			return
		}
		fileName := fmt.Sprintf("contacts_export_%s.csv", dateStr)
		contentType := "text/csv; charset=utf-8"
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		c.Data(http.StatusOK, contentType, data)
	}
}

// exportContactsCSV 将联系人列表生成 CSV 格式的字节数据。
func (a *API) exportContactsCSV(contacts []*model.Contact) ([]byte, error) {
	var buf bytes.Buffer
	// UTF-8 BOM，确保 Excel 正确识别编码
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(&buf)

	header := []string{"微信ID", "微信号", "昵称", "备注", "是否好友"}
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("写入CSV表头失败: %w", err)
	}

	for _, c := range contacts {
		isFriend := "否"
		if c.IsFriend {
			isFriend = "是"
		}
		row := []string{c.UserName, c.Alias, c.NickName, c.Remark, isFriend}
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

// exportContactsXLSX 将联系人列表生成 XLSX 格式的字节数据。
func (a *API) exportContactsXLSX(contacts []*model.Contact) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "联系人"
	f.SetSheetName("Sheet1", sheetName)

	headers := []string{"微信ID", "微信号", "昵称", "备注", "是否好友"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellStyle(sheetName, "A1", "E1", headerStyle)

	f.SetColWidth(sheetName, "A", "A", 25)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "C", "C", 20)
	f.SetColWidth(sheetName, "D", "D", 20)
	f.SetColWidth(sheetName, "E", "E", 10)

	for i, ct := range contacts {
		rowNum := i + 2
		isFriend := "否"
		if ct.IsFriend {
			isFriend = "是"
		}
		vals := []string{ct.UserName, ct.Alias, ct.NickName, ct.Remark, isFriend}
		for j, val := range vals {
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

// GetNeedContactList 处理获取需要联系的客户列表的请求。
func (a *API) GetNeedContactList(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days := 7
	if _, err := fmt.Sscanf(daysStr, "%d", &days); err != nil || days <= 0 {
		days = 7
	}

	items, err := a.Store.GetNeedContactList(c.Request.Context(), days)
	if err != nil {
		log.Error().Err(err).Int("days", days).Msg("获取需要联系的客户列表失败")
		transport.InternalServerError(c, "获取客户联系提醒列表失败。")
		return
	}

	if items == nil {
		items = make([]*model.NeedContactItem, 0)
	}

	transport.SendSuccess(c, items)
}
