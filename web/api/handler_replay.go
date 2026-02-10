package api

import (
	"time"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/types"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ReplayRequest 回放消息查询参数
type ReplayRequest struct {
	TalkerID  string `form:"talker_id" binding:"required"`
	StartDate string `form:"start_date"` // YYYY-MM-DD
	EndDate   string `form:"end_date"`   // YYYY-MM-DD
	Limit     int    `form:"limit,default=200"`
	Offset    int    `form:"offset,default=0"`
}

// ReplayResponse 回放消息响应
type ReplayResponse struct {
	Total    int              `json:"total"`
	Messages []*model.Message `json:"messages"`
}

// GetReplayMessages 批量获取指定会话的消息用于回放
func (a *API) GetReplayMessages(c *gin.Context) {
	var req ReplayRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		transport.BadRequest(c, "无效的回放查询参数: "+err.Error())
		return
	}

	// 限制 limit 最大值为 1000
	if req.Limit > 1000 {
		req.Limit = 1000
	}
	if req.Limit <= 0 {
		req.Limit = 200
	}

	// 解析日期范围
	start, end := parseDateRange(req.StartDate, req.EndDate)

	// 先查询总数：用一个大 limit 获取总条数
	countQuery := types.MessageQuery{
		Talker:    req.TalkerID,
		StartTime: start,
		EndTime:   end,
		Limit:     0, // 用于计数
		Offset:    0,
	}

	// 获取全部消息以计算总数（使用较大的 limit）
	countQuery.Limit = 200000
	allMessages, err := a.Store.GetMessages(c.Request.Context(), countQuery)
	if err != nil {
		log.Error().Err(err).Msg("获取回放消息总数失败")
		transport.InternalServerError(c, "获取回放消息失败")
		return
	}
	total := len(allMessages)

	// 获取分页消息
	query := types.MessageQuery{
		Talker:    req.TalkerID,
		StartTime: start,
		EndTime:   end,
		Limit:     req.Limit,
		Offset:    req.Offset,
		Reverse:   false, // 回放按时间正序
	}

	messages, err := a.Store.GetMessages(c.Request.Context(), query)
	if err != nil {
		log.Error().Err(err).Msg("获取回放消息失败")
		transport.InternalServerError(c, "获取回放消息失败")
		return
	}

	if messages == nil {
		messages = make([]*model.Message, 0)
	}

	transport.SendSuccess(c, ReplayResponse{
		Total:    total,
		Messages: messages,
	})
}

// parseDateRange 解析日期范围字符串，返回 start 和 end 时间
func parseDateRange(startDate, endDate string) (time.Time, time.Time) {
	var start, end time.Time

	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			start = t
		}
	}
	if start.IsZero() {
		start = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	}

	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			// 结束日期设为当天 23:59:59
			end = t.Add(24*time.Hour - time.Second)
		}
	}
	if end.IsZero() {
		end = time.Now().Add(24 * time.Hour)
	}

	return start, end
}
