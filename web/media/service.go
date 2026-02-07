package media

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/pkg/util/dat2img"
	"github.com/afumu/wetrace/pkg/util/silk"
	"github.com/rs/zerolog/log"
)

// Service 处理准备用于服务的媒体文件的业务逻辑。
type Service struct {
	DataDir         string
	ImageKey        string
	XorKey          string
	WechatDbSrcPath string

	// 任务状态
	cacheStatus struct {
		sync.RWMutex
		IsRunning bool
		Total     int32
		Processed int32
		Scope     string
	}
}

// NewService 创建一个新的媒体服务。
func NewService(dataDir, imageKey, xorKey, wechatDbSrcPath string) *Service {
	return &Service{
		DataDir:         dataDir,
		ImageKey:        imageKey,
		XorKey:          xorKey,
		WechatDbSrcPath: wechatDbSrcPath,
	}
}

// CacheStatus 缓存任务进度
type CacheStatus struct {
	IsRunning bool   `json:"isRunning"`
	Total     int    `json:"total"`
	Processed int    `json:"processed"`
	Scope     string `json:"scope"`
}

func (s *Service) GetCacheStatus() CacheStatus {
	s.cacheStatus.RLock()
	defer s.cacheStatus.RUnlock()
	return CacheStatus{
		IsRunning: s.cacheStatus.IsRunning,
		Total:     int(s.cacheStatus.Total),
		Processed: int(s.cacheStatus.Processed),
		Scope:     s.cacheStatus.Scope,
	}
}

func (s *Service) StartCacheTask(scope string, talker string) error {
	s.cacheStatus.Lock()
	if s.cacheStatus.IsRunning {
		s.cacheStatus.Unlock()
		return errors.New("已有任务正在运行中")
	}
	s.cacheStatus.IsRunning = true
	s.cacheStatus.Total = 0
	s.cacheStatus.Processed = 0
	s.cacheStatus.Scope = scope
	s.cacheStatus.Unlock()

	go s.runCacheTask(scope, talker)
	return nil
}

func (s *Service) runCacheTask(scope string, talker string) {
	defer func() {
		s.cacheStatus.Lock()
		s.cacheStatus.IsRunning = false
		s.cacheStatus.Unlock()
	}()

	var targetDirs []string
	baseAttachDir := filepath.Join(s.WechatDbSrcPath, "msg", "attach")

	if scope == "session" && talker != "" {
		// 计算 md5(talker)
		h := md5.Sum([]byte(talker))
		talkerMd5 := hex.EncodeToString(h[:])
		targetDirs = append(targetDirs, filepath.Join(baseAttachDir, talkerMd5))
	} else {
		// 全量模式，获取 attach 下所有子目录
		entries, err := os.ReadDir(baseAttachDir)
		if err != nil {
			log.Error().Err(err).Msg("读取 attach 目录失败")
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				targetDirs = append(targetDirs, filepath.Join(baseAttachDir, entry.Name()))
			}
		}
	}

	// 1. 扫描文件总数
	var files []string
	for _, dir := range targetDirs {
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".dat") {
				files = append(files, path)
			}
			return nil
		})
	}

	atomic.StoreInt32(&s.cacheStatus.Total, int32(len(files)))
	if len(files) == 0 {
		return
	}

	// 2. 并发解密 (限制并发数)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 4) // 限制 4 个并发

	for _, path := range files {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 检查缓存，如果不存在则解密（doPrepareFile 内部已包含此逻辑）
			_ = s.doPrepareFile(p, false)
			atomic.AddInt32(&s.cacheStatus.Processed, 1)
		}(path)
	}

	wg.Wait()
}

// PreparedMedia 保存媒体文件的最终内容和内容类型。
type PreparedMedia struct {
	Content     []byte
	ContentType string
	Error       error
}

// DownloadAndDecryptEmoji 下载并解密表情包
func (s *Service) DownloadAndDecryptEmoji(url string, keyHex string) PreparedMedia {
	// 1. 下载文件
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return PreparedMedia{Error: fmt.Errorf("创建请求失败: %w", err)}
	}
	// 模拟微信 User-Agent，防止被拦截
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36 MicroMessenger/7.0.20.1781(0x6700143B)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return PreparedMedia{Error: fmt.Errorf("下载失败: %w", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PreparedMedia{Error: fmt.Errorf("下载返回状态码: %d", resp.StatusCode)}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return PreparedMedia{Error: fmt.Errorf("读取内容失败: %w", err)}
	}

	// 2. 检查是否已经是图片 (未加密)
	contentType := detectContentType(data)
	if contentType != "application/octet-stream" {
		return PreparedMedia{Content: data, ContentType: contentType}
	}

	// 3. AES 解密
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return PreparedMedia{Error: fmt.Errorf("密钥解码失败: %w", err)}
	}

	if len(key) < 16 {
		return PreparedMedia{Error: errors.New("密钥长度不足 16 字节")}
	}

	iv := key[:16] // 微信通常使用 Key 的前16位作为 IV

	block, err := aes.NewCipher(key)
	if err != nil {
		return PreparedMedia{Error: fmt.Errorf("创建 Cipher 失败: %w", err)}
	}

	if len(data)%aes.BlockSize != 0 {
		// 数据长度不是块大小的倍数，尝试直接返回（可能下载不完整或不是加密数据）
		return PreparedMedia{Content: data, ContentType: "application/octet-stream"}
	}

	decrypted := make([]byte, len(data))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, data)

	// 4. 去除 PKCS7 填充
	unpadded, err := pkcs7Unpad(decrypted, aes.BlockSize)
	if err != nil {
		// 填充错误，尝试使用解密后的原始数据（有时尾部数据不影响显示）
		unpadded = decrypted
	}

	// 5. 再次检测类型
	contentType = detectContentType(unpadded)

	return PreparedMedia{
		Content:     unpadded,
		ContentType: contentType,
	}
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("data is empty")
	}
	if length%blockSize != 0 {
		return nil, errors.New("data length is not a multiple of block size")
	}
	paddingLen := int(data[length-1])
	if paddingLen == 0 || paddingLen > blockSize {
		return nil, errors.New("invalid padding length")
	}
	// check padding
	for i := 0; i < paddingLen; i++ {
		if data[length-1-i] != byte(paddingLen) {
			return nil, errors.New("invalid padding bytes")
		}
	}
	return data[:length-paddingLen], nil
}

func detectContentType(data []byte) string {
	if len(data) > 4 && string(data[:4]) == "GIF8" {
		return "image/gif"
	}
	if len(data) > 8 && string(data[:8]) == "\x89PNG\r\n\x1a\n" {
		return "image/png"
	}
	if len(data) > 2 && string(data[:2]) == "\xff\xd8" {
		return "image/jpeg"
	}
	return "application/octet-stream"
}

// Prepare 处理获取、读取和解码媒体文件的完整生命周期。
func (s *Service) Prepare(media *model.Media, isThumb bool) PreparedMedia {
	if media.Type == "voice" {
		return s.prepareVoice(media.Data)
	}

	if media.Path == "" {
		return PreparedMedia{Error: fmt.Errorf("媒体路径为空，key 为 %s", media.Key)}
	}

	// 对于图片类型，使用启发式路径查找
	if media.Type == "image" {
		return s.prepareImageWithFallback(media.Path, isThumb)
	}

	res := s.prepareFile(media.Path, media.Type == "video")

	// 如果是视频类型，强制设置为 video/mp4，确保前端可以播放
	if media.Type == "video" {
		res.ContentType = "video/mp4"
	}

	return res
}

func (s *Service) prepareImageWithFallback(relativePath string, isThumb bool) PreparedMedia {
	var candidates []string
	if isThumb {
		// 缩略图模式优先级：_t.dat -> .dat -> 原路径
		candidates = []string{
			relativePath + "_t.dat",
			relativePath + ".dat",
			relativePath,
		}
	} else {
		// 原图模式优先级：.dat -> 原路径 -> _t.dat (回退)
		candidates = []string{
			relativePath + ".dat",
			relativePath,
			relativePath + "_t.dat",
		}
	}

	for _, c := range candidates {
		abs := filepath.Join(s.WechatDbSrcPath, c)
		res := s.doPrepareFile(abs, false)
		if res.Error == nil {
			return res
		}
	}

	return PreparedMedia{Error: fmt.Errorf("图片文件不存在 (磁盘及缓存均未找到): %s", relativePath)}
}

func (s *Service) prepareFile(relativePath string, isVideo bool) PreparedMedia {
	if strings.Contains(relativePath, "..") {
		return PreparedMedia{Error: fmt.Errorf("无效的文件路径: %s", relativePath)}
	}

	baseDir := s.WechatDbSrcPath
	absolutePath := filepath.Join(baseDir, relativePath)

	return s.doPrepareFile(absolutePath, isVideo)
}

func (s *Service) doPrepareFile(absolutePath string, isVideo bool) PreparedMedia {
	// 1. 检查缓存 (仅针对图片/解密类文件)
	// 计算相对于微信根目录的路径，用于建立缓存镜像
	relPath, err := filepath.Rel(s.WechatDbSrcPath, absolutePath)
	isDat := false
	if err == nil && !isVideo {
		ext := strings.ToLower(filepath.Ext(absolutePath))
		isDat = strings.HasSuffix(ext, ".dat") || strings.Contains(strings.ToLower(filepath.ToSlash(absolutePath)), "/img/")

		if isDat {
			cachePath := filepath.Join(s.DataDir, "cache", "images", relPath)
			if cacheContent, err := os.ReadFile(cachePath); err == nil {
				return PreparedMedia{
					Content:     cacheContent,
					ContentType: detectContentType(cacheContent),
				}
			}
		}
	}

	// 2. 如果缓存不存在，则检查原文件是否存在
	if _, err := os.Stat(absolutePath); os.IsNotExist(err) {
		return PreparedMedia{Error: fmt.Errorf("文件在磁盘上不存在: %s", absolutePath)}
	}

	// 3. 如果是视频，尝试转码
	if isVideo {
		transcodedPath, err := s.ensureVideoTranscoded(absolutePath)
		if err == nil {
			absolutePath = transcodedPath
		}
	}

	// 3. 处理解密或直接读取
	var res PreparedMedia
	if isDat {
		res = s.prepareDatFile(absolutePath)
		// 解密成功后，异步写入缓存
		if res.Error == nil {
			cachePath := filepath.Join(s.DataDir, "cache", "images", relPath)
			go func(path string, content []byte) {
				os.MkdirAll(filepath.Dir(path), 0755)
				_ = os.WriteFile(path, content, 0644)
			}(cachePath, res.Content)
		}
	} else {
		ext := strings.ToLower(filepath.Ext(absolutePath))
		contentType := getMimeTypeByExtension(ext)
		content, err := os.ReadFile(absolutePath)
		if err != nil {
			return PreparedMedia{Error: fmt.Errorf("读取文件失败: %w", err)}
		}
		res = PreparedMedia{
			Content:     content,
			ContentType: contentType,
		}
	}

	return res
}

// ensureVideoTranscoded 确保视频被转码为兼容性好的格式（H.264/AAC MP4）。
// 返回转码后文件的绝对路径。如果已存在缓存，直接返回。
func (s *Service) ensureVideoTranscoded(srcPath string) (string, error) {
	// 简单的缓存键生成策略：基于文件名或路径 hash
	// 这里简单使用文件名加后缀，保存在系统临时目录的 chatlog_video_cache 子目录下
	fileName := filepath.Base(srcPath)
	cacheDir := filepath.Join(os.TempDir(), "chatlog_video_cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("创建缓存目录失败: %w", err)
	}

	// 目标文件：文件名 + .transcoded.mp4
	// 注意：如果不同目录下有同名文件，这里会冲突。
	// 更严谨的做法是 hash(srcPath)。这里为了演示简单处理。
	// 改进：使用 srcPath 的 hash
	hashName := hex.EncodeToString([]byte(srcPath)) // 简单的 path hash，实际可用 md5
	// 或者是文件名 + hash 的组合以便调试
	dstPath := filepath.Join(cacheDir, fmt.Sprintf("%s_%s.mp4", fileName, hashName[:8]))

	// 1. 检查缓存是否存在
	if _, err := os.Stat(dstPath); err == nil {
		// 缓存存在，直接返回
		// log.Debug().Str("cache", dstPath).Msg("命中视频转码缓存")
		return dstPath, nil
	}

	// 2. 调用 ffmpeg 转码
	// 命令：ffmpeg -i <src> -c:v libx264 -c:a aac -strict experimental <dst>
	// -y 覆盖输出
	// -preset ultrafast 加速转码（牺牲压缩率）
	log.Info().Str("src", srcPath).Msg("开始视频转码 (HEVC -> H.264)...")

	cmd := exec.Command(dat2img.FFMpegPath,
		"-y",
		"-i", srcPath,
		"-c:v", "libx264",
		"-preset", "ultrafast", // 追求速度
		"-c:a", "aac",
		dstPath,
	)

	// 捕获输出以便调试
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg 转码失败: %w, output: %s", err, string(output))
	}

	log.Info().Str("dst", dstPath).Msg("视频转码成功")
	return dstPath, nil
}

func (s *Service) prepareDatFile(path string) PreparedMedia {
	b, err := os.ReadFile(path)
	if err != nil {
		return PreparedMedia{Error: fmt.Errorf("读取 .dat 文件失败: %w", err)}
	}

	// 使用配置的密钥
	foundKey := hex.EncodeToString([]byte(s.ImageKey))
	dat2img.SetAesKey(foundKey)
	_ = dat2img.SetV4XorKey(s.XorKey)

	out, ext, err := dat2img.Dat2Image(b)
	if err != nil {
		// 如果解码失败，则回退到提供原始数据
		log.Warn().Err(err).Str("path", path).Msg("解码 .dat 文件失败，提供原始数据。")
		return PreparedMedia{Content: b, ContentType: "application/octet-stream"}
	}

	contentType := getMimeTypeByExtension(ext)
	return PreparedMedia{Content: out, ContentType: contentType}
}

func getMimeTypeByExtension(ext string) string {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	switch ext {
	case "mp4", "mov", "m4v", "3gp", "mkv", "avi", "wmv", "flv", "webm":
		return "video/mp4"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "bmp":
		return "image/bmp"
	case "ico":
		return "image/x-icon"
	case "svg":
		return "image/svg+xml"
	case "mp3":
		return "audio/mpeg"
	case "wav":
		return "audio/wav"
	case "m4a":
		return "audio/mp4"
	case "aac":
		return "audio/aac"
	case "flac":
		return "audio/flac"
	case "ogg":
		return "audio/ogg"
	case "pdf":
		return "application/pdf"
	case "doc":
		return "application/msword"
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "xls":
		return "application/vnd.ms-excel"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "ppt":
		return "application/vnd.ms-powerpoint"
	case "pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case "txt", "md", "log", "json", "xml":
		return "text/plain; charset=utf-8"
	case "csv":
		return "text/csv"
	case "zip":
		return "application/zip"
	case "rar":
		return "application/x-rar-compressed"
	case "7z":
		return "application/x-7z-compressed"
	case "tar":
		return "application/x-tar"
	case "gz":
		return "application/gzip"
	default:
		return "application/octet-stream"
	}
}

func (s *Service) prepareVoice(data []byte) PreparedMedia {
	if len(data) == 0 {
		return PreparedMedia{Error: fmt.Errorf("语音数据为空")}
	}

	out, err := silk.Silk2MP3(data)
	if err != nil {
		log.Warn().Err(err).Msg("解码 .silk 音频失败，提供原始数据。")
		return PreparedMedia{Content: data, ContentType: "audio/silk"} // 回退
	}

	return PreparedMedia{Content: out, ContentType: "audio/mp3"}
}
