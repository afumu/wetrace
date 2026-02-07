package api

import (
	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/types"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetChatRooms 处理获取群聊列表的请求。
func (a *API) GetChatRooms(c *gin.Context) {
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

	query := types.ChatRoomQuery{
		Keyword: keywordQuery.Keyword,
		Limit:   pageQuery.Limit,
		Offset:  pageQuery.Offset,
	}

	chatrooms, err := a.Store.GetChatRooms(c.Request.Context(), query)
	if err != nil {
		log.Error().Err(err).Msg("从 store 获取群聊失败")
		transport.InternalServerError(c, "获取群聊列表失败。")
		return
	}

	if chatrooms == nil {
		chatrooms = make([]*model.ChatRoom, 0)
	}

	transport.SendSuccess(c, chatrooms)
}

// GetChatRoomByID 处理通过 ID 获取单个群聊信息的请求。
func (a *API) GetChatRoomByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		transport.BadRequest(c, "群聊 ID 是必需的。")
		return
	}

	query := types.ChatRoomQuery{
		Keyword: id,
		Limit:   1,
	}

	chatrooms, err := a.Store.GetChatRooms(c.Request.Context(), query)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("从 store 通过 ID 获取群聊失败")
		transport.InternalServerError(c, "获取群聊信息失败。")
		return
	}

	if len(chatrooms) == 0 {
		transport.NotFound(c, "未找到群聊。")
		return
	}

	transport.SendSuccess(c, chatrooms[0])
}
