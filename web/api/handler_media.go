package api

import (
	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetMedia 处理媒体文件（如图片、视频、语音等）的请求。
func (a *API) GetMedia(c *gin.Context) {
	mediaType := c.Param("type")
	key := c.Param("key")
	path := c.Query("path")
	isThumb := c.Query("thumb") == "1"

	if mediaType == "" || key == "" {
		transport.BadRequest(c, "媒体类型和 key 是必需的。")
		return
	}

	// 1. 从 store 获取媒体元数据
	mediaInfo, err := a.Store.GetMedia(c.Request.Context(), mediaType, key)
	if err != nil {
		// 如果提供了 path，我们可以创建一个虚拟的 mediaInfo 继续处理
		if path != "" {
			mediaInfo = &model.Media{
				Type: mediaType,
				Key:  key,
				Path: path,
			}
		} else {
			log.Warn().Err(err).Str("type", mediaType).Str("key", key).Msg("从 store 获取媒体失败")
			transport.NotFound(c, "未找到媒体文件。")
			return
		}
	} else if path != "" {
		// 如果数据库中有数据，但前端传了 path，以传参为准
		mediaInfo.Path = path
	}

	// 2. 使用媒体服务准备内容
	preparedMedia := a.Media.Prepare(mediaInfo, isThumb)

	// 3. 发送响应
	transport.SendMedia(c, preparedMedia)
}

// GetEmoji 处理表情包的下载和解密请求。
func (a *API) GetEmoji(c *gin.Context) {
	url := c.Query("url")
	key := c.Query("key")

	if url == "" || key == "" {
		transport.BadRequest(c, "url 和 key 参数是必需的。")
		return
	}

	// 调用 Media Service 进行下载和解密
	preparedMedia := a.Media.DownloadAndDecryptEmoji(url, key)

	// 发送响应
	transport.SendMedia(c, preparedMedia)
}

// HandleStartCache 启动图片缓存预加载任务
func (a *API) HandleStartCache(c *gin.Context) {
	var req struct {
		Scope  string `json:"scope"`  // "all" 或 "session"
		Talker string `json:"talker"` // 仅当 scope 为 session 时需要
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "无效的请求参数")
		return
	}

	err := a.Media.StartCacheTask(req.Scope, req.Talker)
	if err != nil {
		transport.InternalServerError(c, err.Error())
		return
	}

	transport.SendSuccess(c, "任务已启动")
}

// GetCacheStatus 获取当前缓存任务的进度
func (a *API) GetCacheStatus(c *gin.Context) {
	status := a.Media.GetCacheStatus()
	transport.SendSuccess(c, status)
}
