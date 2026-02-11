package api

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
	FullURL      string `json:"fullUrl"`
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

		path := ""
		if msg.Contents != nil {
			if p, ok := msg.Contents["path"].(string); ok {
				path = p
			}
		}
		thumbnailURL := fmt.Sprintf("/api/v1/media/image/%s?thumb=1", key)
		if path != "" {
			thumbnailURL += "&path=" + url.QueryEscape(path)
		}

		item := &imageListItem{
			Key:          key,
			Talker:       msg.Talker,
			TalkerName:   msg.TalkerName,
			Time:         msg.Time.Format(time.RFC3339),
			ThumbnailURL: thumbnailURL,
			Seq:          msg.Seq,
		}
		allItems = append(allItems, item)
	}

	// 当数据库查询结果为空时，扫描本地缓存目录获取图片列表
	if len(allItems) == 0 {
		cacheItems := a.scanCacheImages(q.Talker)
		allItems = append(allItems, cacheItems...)
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

// cacheImageExtensions 缓存目录中可能包含的图片文件扩展名。
// 缓存文件保留原始 .dat 扩展名，也可能是已解码的图片格式。
var cacheImageExtensions = map[string]bool{
	".dat": true,
	".jpg": true, ".jpeg": true, ".png": true,
	".gif": true, ".bmp": true, ".webp": true,
}

// scanCacheImages 扫描本地缓存目录获取图片列表。
// 当数据库中没有图片消息记录时，作为回退方案使用。
func (a *API) scanCacheImages(talker string) []*imageListItem {
	cacheBaseDir := filepath.Join(a.Media.DataDir, "cache", "images", "msg", "attach")

	// 检查缓存目录是否存在
	if _, err := os.Stat(cacheBaseDir); os.IsNotExist(err) {
		log.Debug().Str("dir", cacheBaseDir).Msg("缓存图片目录不存在，跳过扫描")
		return nil
	}

	// 确定要扫描的目录列表
	scanDirs := a.getCacheScanDirs(cacheBaseDir, talker)
	if len(scanDirs) == 0 {
		return nil
	}

	// 遍历目录收集图片文件
	var items []*imageListItem
	for _, dir := range scanDirs {
		dirItems := a.scanSingleCacheDir(dir, cacheBaseDir)
		items = append(items, dirItems...)
	}

	log.Info().Int("count", len(items)).Msg("从缓存目录扫描到图片")
	return items
}

// getCacheScanDirs 根据 talker 参数确定需要扫描的目录列表。
func (a *API) getCacheScanDirs(cacheBaseDir, talker string) []string {
	if talker != "" {
		// 按会话筛选：计算 talker 的 md5 作为子目录名
		h := fmt.Sprintf("%x", md5Sum([]byte(talker)))
		targetDir := filepath.Join(cacheBaseDir, h)
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			return nil
		}
		return []string{targetDir}
	}

	// 全量模式：扫描 attach 下所有子目录
	entries, err := os.ReadDir(cacheBaseDir)
	if err != nil {
		log.Error().Err(err).Msg("读取缓存 attach 目录失败")
		return nil
	}

	dirs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(cacheBaseDir, entry.Name()))
		}
	}
	return dirs
}

// scanSingleCacheDir 扫描单个缓存子目录中的图片文件。
func (a *API) scanSingleCacheDir(dir, cacheBaseDir string) []*imageListItem {
	var items []*imageListItem

	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !cacheImageExtensions[ext] {
			return nil
		}

		item := a.buildCacheImageItem(path, cacheBaseDir, info)
		if item != nil {
			items = append(items, item)
		}
		return nil
	})

	return items
}

// buildCacheImageItem 根据缓存文件路径构造 imageListItem。
func (a *API) buildCacheImageItem(path, cacheBaseDir string, info os.FileInfo) *imageListItem {
	fileName := info.Name()

	// 跳过缩略图文件（以 _t.dat 结尾），避免重复
	if strings.HasSuffix(strings.ToLower(fileName), "_t.dat") {
		return nil
	}

	// 计算相对于 cacheBaseDir 的路径
	relPath, err := filepath.Rel(cacheBaseDir, path)
	if err != nil {
		return nil
	}

	// 使用文件名（不含扩展名）作为 key
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// 从相对路径中提取 talker hash（第一级目录）
	parts := strings.SplitN(filepath.ToSlash(relPath), "/", 2)
	talkerHash := ""
	if len(parts) > 0 {
		talkerHash = parts[0]
	}

	// 构造 path 参数（相对于 WechatDbSrcPath）。
	// 缓存文件镜像源文件路径，扩展名可能是 .dat 或已解码的图片格式。
	// prepareImageWithFallback 会尝试 path+".dat"、path、path+"_t.dat"，
	// 所以这里去掉扩展名，让它自动匹配。
	datRelPath := filepath.ToSlash(filepath.Join("msg", "attach", relPath))
	datRelPath = strings.TrimSuffix(datRelPath, filepath.Ext(datRelPath))

	// 构造缩略图 URL，使用 path 参数让 GetMedia 能找到文件
	thumbnailURL := fmt.Sprintf("/api/v1/media/image/%s?thumb=1&path=%s",
		url.PathEscape(baseName),
		url.QueryEscape(datRelPath),
	)

	return &imageListItem{
		Key:          baseName,
		Talker:       talkerHash,
		TalkerName:   talkerHash,
		Time:         info.ModTime().Format(time.RFC3339),
		ThumbnailURL: thumbnailURL,
		Seq:          info.ModTime().UnixMilli(),
	}
}

// md5Sum 计算字节数组的 MD5 哈希值。
func md5Sum(data []byte) [16]byte {
	return md5.Sum(data)
}

// TranscribeVoice 语音转文字
func (a *API) TranscribeVoice(c *gin.Context) {
	if a.TTS == nil {
		transport.BadRequest(c, "语音转文字功能未启用，请先在设置中配置")
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "参数错误")
		return
	}
	if req.ID == "" {
		transport.BadRequest(c, "语音ID不能为空")
		return
	}

	// 获取语音媒体信息
	mediaInfo, err := a.Store.GetMedia(c.Request.Context(), "voice", req.ID)
	if err != nil {
		transport.NotFound(c, "未找到语音文件")
		return
	}

	// 使用媒体服务准备语音内容
	prepared := a.Media.Prepare(mediaInfo, false)
	if prepared.Error != nil || len(prepared.Content) == 0 {
		transport.InternalServerError(c, "无法读取语音文件")
		return
	}

	// 调用 Whisper API 转文字
	text, err := a.TTS.Transcribe(prepared.Content, "voice.mp3")
	if err != nil {
		log.Error().Err(err).Str("id", req.ID).Msg("语音转文字失败")
		transport.InternalServerError(c, "语音转文字失败: "+err.Error())
		return
	}

	transport.SendSuccess(c, gin.H{"text": text})
}

// ExportVoices 一键导出会话中的所有语音消息为 ZIP 包（每条语音为独立 MP3 文件）
func (a *API) ExportVoices(c *gin.Context) {
	talker := c.Query("talker")
	if talker == "" {
		transport.BadRequest(c, "talker 参数不能为空")
		return
	}
	name := c.Query("name")
	if name == "" {
		name = talker
	}

	// 查询该会话的所有语音消息
	msgQuery := types.MessageQuery{
		Talker:  talker,
		MsgType: model.MessageTypeVoice,
		Limit:   100000,
		Offset:  0,
	}

	messages, err := a.Store.GetMessages(c.Request.Context(), msgQuery)
	if err != nil {
		log.Error().Err(err).Str("talker", talker).Msg("查询语音消息失败")
		transport.InternalServerError(c, "查询语音消息失败")
		return
	}

	if len(messages) == 0 {
		transport.BadRequest(c, "该会话没有语音消息")
		return
	}

	// 创建 ZIP 缓冲区
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	voiceCount := 0
	for i, msg := range messages {
		// 获取语音 key
		voiceKey := ""
		if msg.Contents != nil {
			if v, ok := msg.Contents["voice"]; ok {
				voiceKey = fmt.Sprint(v)
			}
		}
		if voiceKey == "" {
			continue
		}

		// 从 store 获取语音媒体数据
		mediaInfo, err := a.Store.GetMedia(c.Request.Context(), "voice", voiceKey)
		if err != nil {
			log.Warn().Err(err).Str("key", voiceKey).Msg("获取语音媒体失败，跳过")
			continue
		}

		// 使用媒体服务解码语音
		prepared := a.Media.Prepare(mediaInfo, false)
		if prepared.Error != nil || len(prepared.Content) == 0 {
			log.Warn().Str("key", voiceKey).Msg("语音解码失败或为空，跳过")
			continue
		}

		// 确定文件扩展名
		ext := ".mp3"
		if prepared.ContentType == "audio/silk" {
			ext = ".silk"
		}

		// 构造文件名：序号_发送者_时间.mp3
		timeStr := msg.Time.Format("20060102_150405")
		senderName := msg.SenderName
		if senderName == "" {
			senderName = msg.Sender
		}
		// 清理文件名中的非法字符
		senderName = sanitizeFileName(senderName)
		fileName := fmt.Sprintf("%03d_%s_%s%s", i+1, senderName, timeStr, ext)

		// 写入 ZIP
		w, err := zipWriter.Create(fileName)
		if err != nil {
			log.Warn().Err(err).Str("file", fileName).Msg("创建 ZIP 条目失败")
			continue
		}
		if _, err := w.Write(prepared.Content); err != nil {
			log.Warn().Err(err).Str("file", fileName).Msg("写入 ZIP 条目失败")
			continue
		}
		voiceCount++
	}

	if err := zipWriter.Close(); err != nil {
		transport.InternalServerError(c, "生成 ZIP 文件失败")
		return
	}

	if voiceCount == 0 {
		transport.BadRequest(c, "没有可导出的语音文件")
		return
	}

	// 设置响应头并发送 ZIP
	zipName := fmt.Sprintf("voices_%s_%s.zip", sanitizeFileName(name), time.Now().Format("20060102"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipName))
	c.Header("Content-Type", "application/zip")
	c.Data(200, "application/zip", buf.Bytes())
}

// sanitizeFileName 清理文件名中的非法字符
func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
	)
	result := replacer.Replace(name)
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}
