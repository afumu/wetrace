package api

import (
	"fmt"
	"time"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/types"
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

// imageListQuery 图片列表请求参数
type imageListQuery struct {
	Talker    string `form:"talker"`
	TimeRange string `form:"time_range"`
	Limit     int    `form:"limit,default=50"`
	Offset    int    `form:"offset,default=0"`
}

// imageListItem 图片列表响应项
type imageListItem struct {
	Key          string `json:"key"`
	Talker       string `json:"talker"`
	TalkerName   string `json:"talkerName"`
	Time         string `json:"time"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Seq          int64  `json:"seq"`
}

// imageListResponse 图片列表响应
type imageListResponse struct {
	Total int              `json:"total"`
	Items []*imageListItem `json:"items"`
}

// GetImageList 获取图片列表，支持按会话筛选和时间范围筛选。
func (a *API) GetImageList(c *gin.Context) {
	var q imageListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		transport.BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	// 解析时间范围
	var startTime, endTime time.Time
	startTime, endTime = parseImageTimeRange(q.TimeRange)

	// 构建消息查询：MsgType=3 表示图片消息
	msgQuery := types.MessageQuery{
		Talker:    q.Talker,
		MsgType:   model.MessageTypeImage,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     200000,
		Offset:    0,
	}

	messages, err := a.Store.GetMessages(c.Request.Context(), msgQuery)
	if err != nil {
		log.Error().Err(err).Msg("获取图片消息列表失败")
		transport.InternalServerError(c, "获取图片列表失败。")
		return
	}

	// 从消息中提取图片信息
	allItems := make([]*imageListItem, 0, len(messages))
	for _, msg := range messages {
		key := ""
		if msg.Contents != nil {
			if md5, ok := msg.Contents["md5"].(string); ok {
				key = md5
			}
		}
		if key == "" {
			continue
		}

		item := &imageListItem{
			Key:          key,
			Talker:       msg.Talker,
			TalkerName:   msg.TalkerName,
			Time:         msg.Time.Format(time.RFC3339),
			ThumbnailURL: fmt.Sprintf("/api/v1/media/image/%s", key),
			Seq:          msg.Seq,
		}
		allItems = append(allItems, item)
	}

	total := len(allItems)

	// 分页
	start := q.Offset
	if start > total {
		start = total
	}
	end := start + q.Limit
	if end > total {
		end = total
	}
	pageItems := allItems[start:end]

	transport.SendSuccess(c, imageListResponse{
		Total: total,
		Items: pageItems,
	})
}

// parseImageTimeRange 将前端传入的时间范围字符串转换为起止时间。
func parseImageTimeRange(timeRange string) (start, end time.Time) {
	now := time.Now()
	end = now.Add(24 * time.Hour)

	switch timeRange {
	case "last_week":
		start = now.AddDate(0, 0, -7)
	case "last_month":
		start = now.AddDate(0, -1, 0)
	case "last_year":
		start = now.AddDate(-1, 0, 0)
	default:
		// "all" 或空值，查询全部
		start = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return
}
