package api

import (
	"time"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/pkg/util"
	"github.com/afumu/wetrace/store/types"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type MessageRequest struct {
	TalkerID  string `form:"talker_id"` // Remove binding:"required" for global search
	SenderID  string `form:"sender_id"`
	Keyword   string `form:"keyword"`
	TimeRange string `form:"time_range"` // e.g., "2023-01-01~2023-01-31"
	Reverse   bool   `form:"reverse"`    // 是否倒序
	transport.PaginationQuery
}

// GetMessages 处理获取指定对话者聊天消息的请求。
func (a *API) GetMessages(c *gin.Context) {
	// 1. 绑定并验证查询参数
	var req MessageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		transport.BadRequest(c, "无效的消息查询参数: "+err.Error())
		return
	}

	// 2. 区分全局搜索和会话搜索
	if req.TalkerID == "" && req.Keyword != "" {
		query := types.MessageQuery{
			Keyword: req.Keyword,
			Limit:   req.Limit,
			Offset:  req.Offset,
		}
		messages, err := a.Store.SearchGlobalMessages(c.Request.Context(), query)
		if err != nil {
			transport.InternalServerError(c, "全局搜索失败")
			return
		}
		transport.SendSuccess(c, messages)
		return
	}

	if req.TalkerID == "" {
		transport.BadRequest(c, "必须指定 talker_id 或 keyword")
		return
	}

	start, end, ok := util.TimeRangeOf(req.TimeRange)
	if !ok {
		// 如果未指定或无效，则默认为一个很宽的范围
		end = time.Now()
		start = end.AddDate(-10, 0, 0) // 10 年前
	}

	// 2. 构建 store 查询
	query := types.MessageQuery{
		Talker:    req.TalkerID,
		Sender:    req.SenderID,
		Keyword:   req.Keyword,
		StartTime: start,
		EndTime:   end,
		Limit:     req.Limit,
		Offset:    req.Offset,
		Reverse:   req.Reverse,
	}

	// 3. 调用 store
	messages, err := a.Store.GetMessages(c.Request.Context(), query)
	if err != nil {
		log.Error().Err(err).Msg("从 store 获取消息失败")
		transport.InternalServerError(c, "获取消息失败。")
		return
	}

	if messages == nil {
		messages = make([]*model.Message, 0)
	}

	// 4. 发送成功响应
	transport.SendSuccess(c, messages)
}
