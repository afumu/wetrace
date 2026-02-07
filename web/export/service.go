package export

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store"
	"github.com/afumu/wetrace/store/types"
	"github.com/afumu/wetrace/web/media"
)

//go:embed template/index.html
var exportTemplate string

type Service struct {
	Media    *media.Service
	Store    store.Store
	StaticFS fs.FS
}

func (s *Service) ExportChat(ctx context.Context, talker string, talkerName string, startTime, endTime time.Time) ([]byte, error) {
	query := types.MessageQuery{
		Talker:    talker,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     200000,
		Offset:    0,
	}
	messages, err := s.Store.GetMessages(ctx, query)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	fmt.Printf("[Export] Processing %d messages for %s...\n", len(messages), talkerName)

	for _, msg := range messages {
		s.processMedia(ctx, zw, msg)
	}

	msgJson, _ := json.Marshal(messages)
	dataJs := fmt.Sprintf("window.CHAT_DATA = %s;", string(msgJson))
	fData, _ := zw.Create("data.js")
	fData.Write([]byte(dataJs))

	html, err := s.buildHtml(talkerName)
	if err != nil {
		return nil, err
	}
	fHtml, _ := zw.Create("index.html")
	fHtml.Write([]byte(html))

	s.copyAssets(zw)

	zw.Close()
	return buf.Bytes(), nil
}

func (s *Service) ExportChatTxt(ctx context.Context, talker string, talkerName string, startTime, endTime time.Time) ([]byte, error) {
	query := types.MessageQuery{
		Talker:    talker,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     200000,
		Offset:    0,
	}
	messages, err := s.Store.GetMessages(ctx, query)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	fmt.Printf("[ExportTXT] Processing %d messages for %s...\n", len(messages), talkerName)

	for _, msg := range messages {
		// 简单的格式化: [昵称] 时间 \n 内容
		sender := msg.SenderName
		if sender == "" {
			sender = msg.Sender
		}

		// 设置 Host 以生成正确的本地链接
		msg.SetContent("host", "127.0.0.1:5200/api/v1/media")
		content := msg.PlainTextContent()

		line := fmt.Sprintf("[%s] %s\n%s\n\n",
			sender,
			msg.Time.Format("2006-01-02 15:04:05"),
			content,
		)
		buf.WriteString(line)
	}

	return buf.Bytes(), nil
}

func (s *Service) buildHtml(talkerName string) (string, error) {
	html := exportTemplate
	styles := s.getStyles()
	cssTag := fmt.Sprintf("<style>%s</style>", styles)
	html = strings.Replace(html, "<!-- CSS_PLACEHOLDER -->", cssTag, 1)
	html = strings.Replace(html, "聊天记录导出", talkerName+" 的聊天记录", 1)
	html = strings.Replace(html, "聊天记录", talkerName, 1)
	return html, nil
}

func (s *Service) processMedia(ctx context.Context, zw *zip.Writer, msg *model.Message) {
	if msg.Contents == nil {
		return
	}

	mediaType := ""
	subDir := ""
	switch msg.Type {
	case 3:
		mediaType, subDir = "image", "images"
	case 43:
		mediaType, subDir = "video", "videos"
	case 34:
		mediaType, subDir = "voice", "voice"
	case 47:
		mediaType, subDir = "image", "emojis"
	case 49:
		mediaType, subDir = "file", "files"
	default:
		return
	}

	key, _ := msg.Contents["md5"].(string)
	if key == "" {
		if v, ok := msg.Contents["voice"].(string); ok {
			key = v
		}
		if v, ok := msg.Contents["fileid"].(string); ok {
			key = v
		}
	}

	path, _ := msg.Contents["path"].(string)
	if key == "" && path == "" && msg.Type != 47 {
		return
	}

	// 1. 针对商店表情 (Type 47 SubType 0) 的特殊下载逻辑
	var prepared media.PreparedMedia
	if msg.Type == 47 {
		url, _ := msg.Contents["cdnurl"].(string)
		md5Key, _ := msg.Contents["aeskey"].(string)
		if url != "" && md5Key != "" {
			// 直接调用底层服务方法，与 GetEmoji 接口逻辑保持 100% 一致
			prepared = s.Media.DownloadAndDecryptEmoji(url, md5Key)
		}
	}

	var mediaInfo *model.Media
	// 2. 如果不是商店表情，或者下载失败，尝试常规 GetMedia 逻辑 (针对本地自定义表情或文件)
	if len(prepared.Content) == 0 {
		if key != "" {
			mediaInfo, _ = s.Store.GetMedia(ctx, mediaType, key)
		}

		if mediaInfo == nil {
			mediaInfo = &model.Media{
				Type: mediaType,
				Key:  key,
				Path: path,
			}
		} else if mediaInfo.Path == "" && path != "" {
			mediaInfo.Path = path
		}

		if mediaInfo.Path != "" {
			mediaInfo.Path = filepath.ToSlash(mediaInfo.Path)
		}

		prepared = s.Media.Prepare(mediaInfo, false)
	}

	// 3. 针对视频进行深度保底查找 (尝试添加扩展名)
	if (prepared.Error != nil || len(prepared.Content) == 0) && msg.Type == 43 && mediaInfo != nil && mediaInfo.Path != "" {
		exts := []string{".mp4", ".dat", ".MP4", ".DAT"}
		for _, e := range exts {
			mCopy := *mediaInfo
			mCopy.Path = mediaInfo.Path + e
			res := s.Media.Prepare(&mCopy, false)
			if res.Error == nil && len(res.Content) > 0 {
				prepared = res
				break
			}
		}
	}

	if (prepared.Error != nil || len(prepared.Content) == 0) && mediaInfo.Path != "" {
		roots := []string{s.Media.WechatDbSrcPath, filepath.Join(s.Media.WechatDbSrcPath, "Msg")}
		for _, root := range roots {
			fullPath := filepath.Join(root, mediaInfo.Path)
			if data, err := os.ReadFile(fullPath); err == nil {
				prepared = media.PreparedMedia{
					Content:     data,
					ContentType: "application/octet-stream",
				}
				break
			}
		}
	}

	if prepared.Error != nil || len(prepared.Content) == 0 {
		return
	}

	ext := ""
	contentType := strings.ToLower(prepared.ContentType)
	if strings.Contains(contentType, "jpeg") {
		ext = ".jpg"
	} else if strings.Contains(contentType, "png") {
		ext = ".png"
	} else if strings.Contains(contentType, "gif") {
		ext = ".gif"
	} else if strings.Contains(contentType, "mp4") {
		ext = ".mp4"
	} else if strings.Contains(contentType, "mpeg") || strings.Contains(contentType, "mp3") {
		ext = ".mp3"
	} else if msg.Type == 43 {
		ext = ".mp4"
	} else if msg.Type == 47 {
		// 表情包保底使用 .gif (微信表情大多是 gif 或 png)
		ext = ".gif"
	}

	// 7. 文件名生成
	saveName := key
	if saveName == "" {
		if len(prepared.Content) > 0 {
			// 如果 key 为空但有内容，直接计算内容的 MD5
			saveName = fmt.Sprintf("%x", md5.Sum(prepared.Content))
		} else if mediaInfo != nil && mediaInfo.Path != "" {
			saveName = fmt.Sprintf("%x", md5.Sum([]byte(mediaInfo.Path)))
		} else {
			saveName = fmt.Sprintf("msg_%d", msg.Seq)
		}
	}

	if msg.Type == 49 {
		if title, ok := msg.Contents["title"].(string); ok && title != "" {
			saveName = title
			ext = ""
		}
	}

	relPath := fmt.Sprintf("media/%s/%s%s", subDir, saveName, ext)
	f, err := zw.Create(relPath)
	if err == nil {
		f.Write(prepared.Content)
		msg.Contents["_url"] = relPath
	}
}

func (s *Service) getStyles() string {
	if s.StaticFS == nil {
		return ""
	}
	assetsDir := "assets"
	entries, err := fs.ReadDir(s.StaticFS, assetsDir)
	if err != nil {
		return ""
	}

	var sb strings.Builder
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".css") {
			c, err := fs.ReadFile(s.StaticFS, assetsDir+"/"+entry.Name())
			if err == nil {
				sb.Write(c)
			}
		}
	}
	return sb.String()
}

func (s *Service) copyAssets(zw *zip.Writer) {
	if s.StaticFS == nil {
		return
	}
	baseDir := "assets"
	fs.WalkDir(s.StaticFS, baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js") {
			return nil
		}
		f, err := zw.Create(path)
		if err != nil {
			return nil
		}
		c, err := fs.ReadFile(s.StaticFS, path)
		if err == nil {
			f.Write(c)
		}
		return nil
	})
}
