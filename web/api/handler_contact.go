package api

import (
	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/types"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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
