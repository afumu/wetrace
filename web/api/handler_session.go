package api

import (
	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/types"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetSessions 处理获取聊天会话列表的请求。
func (a *API) GetSessions(c *gin.Context) {
	// 1. 绑定查询参数
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

	// 2. 构建 store 查询
	query := types.SessionQuery{
		Keyword: keywordQuery.Keyword,
		Limit:   pageQuery.Limit,
		Offset:  pageQuery.Offset,
	}

	// 3. 调用 store
	sessions, err := a.Store.GetSessions(c.Request.Context(), query)
	if err != nil {
		log.Error().Err(err).Msg("从 store 获取会话失败")
		transport.InternalServerError(c, "获取会话列表失败。")
		return
	}

	// 确保总是返回一个列表，而不是 nil
	if sessions == nil {
		sessions = make([]*model.Session, 0)
	}

	// 4. 发送成功响应
	transport.SendSuccess(c, sessions)
}

// DeleteSession 处理删除会话的请求。
func (a *API) DeleteSession(c *gin.Context) {
	username := c.Param("id")
	if username == "" {
		transport.BadRequest(c, "缺少会话 ID")
		return
	}

	if err := a.Store.DeleteSession(c.Request.Context(), username); err != nil {
		log.Error().Err(err).Str("username", username).Msg("删除会话失败")
		transport.InternalServerError(c, "删除会话失败")
		return
	}

	transport.SendSuccess(c, nil)
}
