package api

import (
	"context"
	"strconv"
	"time"

	"github.com/afumu/wetrace/pkg/util"
	"github.com/afumu/wetrace/pkg/wordcloud"
	"github.com/afumu/wetrace/store/types"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
)

// GetWordCloud 获取指定会话的词云数据
func (a *API) GetWordCloud(c *gin.Context) {
	talker := c.Param("id")
	if talker == "" {
		transport.BadRequest(c, "会话 ID 不能为空")
		return
	}

	// 解析时间范围
	var start, end time.Time
	var ok bool
	timeRange := c.Query("time_range")
	if timeRange != "" {
		start, end, ok = util.TimeRangeOf(timeRange)
	}
	if !ok {
		end = time.Now()
		start = end.AddDate(-20, 0, 0)
	}

	// 解析 limit
	limit := 100
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	// 构建查询
	query := types.MessageQuery{
		Talker:    talker,
		StartTime: start,
		EndTime:   end,
		Limit:     5000,
	}

	// 限定发送人
	if sender := c.Query("sender"); sender != "" {
		query.Sender = sender
	}

	msgs, err := a.Store.GetMessages(context.Background(), query)
	if err != nil {
		transport.InternalServerError(c, err.Error())
		return
	}

	// 提取文本消息内容
	texts := make([]string, 0, len(msgs))
	for _, m := range msgs {
		if m.Type == 1 {
			texts = append(texts, m.Content)
		}
	}

	if len(texts) == 0 {
		transport.SendSuccess(c, &wordcloud.WordCloudResult{
			Words: []*wordcloud.WordItem{},
		})
		return
	}

	result := wordcloud.Analyze(texts, limit)
	transport.SendSuccess(c, result)
}

// GetWordCloudGlobal 获取全局词云数据（不限定会话）
func (a *API) GetWordCloudGlobal(c *gin.Context) {
	// 解析时间范围
	var start, end time.Time
	var ok bool
	timeRange := c.Query("time_range")
	if timeRange != "" {
		start, end, ok = util.TimeRangeOf(timeRange)
	}
	if !ok {
		end = time.Now()
		start = end.AddDate(-20, 0, 0)
	}

	// 解析 limit
	limit := 100
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	msgs, err := a.Store.SearchGlobalMessages(context.Background(), types.MessageQuery{
		StartTime: start,
		EndTime:   end,
		Limit:     5000,
	})
	if err != nil {
		transport.InternalServerError(c, err.Error())
		return
	}

	// 提取文本消息内容
	texts := make([]string, 0, len(msgs))
	for _, m := range msgs {
		if m.Type == 1 {
			texts = append(texts, m.Content)
		}
	}

	if len(texts) == 0 {
		transport.SendSuccess(c, &wordcloud.WordCloudResult{
			Words: []*wordcloud.WordItem{},
		})
		return
	}

	result := wordcloud.Analyze(texts, limit)
	transport.SendSuccess(c, result)
}
