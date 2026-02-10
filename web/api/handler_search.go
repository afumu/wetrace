package api

import (
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/afumu/wetrace/pkg/util"
	"github.com/afumu/wetrace/store/types"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// SearchRequest 全文搜索请求参数
type SearchRequest struct {
	Keyword   string `form:"keyword" binding:"required"`
	Talker    string `form:"talker"`
	Sender    string `form:"sender"`
	MsgType   int    `form:"type"`
	TimeRange string `form:"time_range"`
	transport.PaginationQuery
}

// Search 全文搜索
func (a *API) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		transport.BadRequest(c, "无效的搜索参数: "+err.Error())
		return
	}

	start, end, ok := util.TimeRangeOf(req.TimeRange)
	if !ok {
		end = time.Now()
		start = end.AddDate(-10, 0, 0)
	}

	query := types.MessageQuery{
		Keyword:   req.Keyword,
		Talker:    req.Talker,
		Sender:    req.Sender,
		MsgType:   req.MsgType,
		StartTime: start,
		EndTime:   end,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	result, err := a.Store.SearchMessages(c.Request.Context(), query)
	if err != nil {
		log.Error().Err(err).Msg("全文搜索失败")
		transport.InternalServerError(c, "搜索失败")
		return
	}

	// 为搜索结果添加高亮
	for _, item := range result.Items {
		item.Highlight = highlightKeyword(item.Content, req.Keyword)
	}

	transport.SendSuccess(c, result)
}

// SearchContextRequest 搜索上下文请求参数
type SearchContextRequest struct {
	Talker string `form:"talker" binding:"required"`
	Seq    int64  `form:"seq" binding:"required"`
	Before int    `form:"before,default=10"`
	After  int    `form:"after,default=10"`
}

// SearchContext 获取搜索结果的上下文消息
func (a *API) SearchContext(c *gin.Context) {
	var req SearchContextRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		transport.BadRequest(c, "无效的上下文查询参数: "+err.Error())
		return
	}

	// 从 query 参数手动解析 seq（避免 int64 绑定问题）
	if req.Seq == 0 {
		seqStr := c.Query("seq")
		if seqStr != "" {
			if s, err := strconv.ParseInt(seqStr, 10, 64); err == nil {
				req.Seq = s
			}
		}
	}

	messages, err := a.Store.GetMessageContext(c.Request.Context(), req.Talker, req.Seq, req.Before, req.After)
	if err != nil {
		log.Error().Err(err).Msg("获取消息上下文失败")
		transport.InternalServerError(c, "获取消息上下文失败")
		return
	}

	// 计算锚点索引
	anchorIndex := -1
	for i, msg := range messages {
		if msg.Seq == req.Seq {
			anchorIndex = i
			break
		}
	}

	transport.SendSuccess(c, gin.H{
		"messages":     messages,
		"anchor_index": anchorIndex,
	})
}

// highlightKeyword 对内容中的关键词进行 HTML 高亮标记
func highlightKeyword(content, keyword string) string {
	if keyword == "" || content == "" {
		return content
	}
	escaped := html.EscapeString(content)
	escapedKeyword := html.EscapeString(keyword)
	return strings.ReplaceAll(escaped, escapedKeyword, "<em>"+escapedKeyword+"</em>")
}
