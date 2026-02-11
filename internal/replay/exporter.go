package replay

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/pkg/util/dat2img"
	"github.com/afumu/wetrace/store"
	"github.com/afumu/wetrace/store/types"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
)

// Exporter 回放导出服务
type Exporter struct {
	Store       store.Store
	TaskManager *TaskManager
	OutputDir   string
}

// NewExporter 创建导出服务
func NewExporter(s store.Store) *Exporter {
	outputDir := filepath.Join(os.TempDir(), "wetrace_replay_export")
	os.MkdirAll(outputDir, 0755)

	e := &Exporter{
		Store:       s,
		TaskManager: NewTaskManager(),
		OutputDir:   outputDir,
	}

	// 启动定期清理
	go e.cleanupLoop()

	return e
}

// CreateTask 创建导出任务并异步执行
func (e *Exporter) CreateTask(req ExportRequest) *ExportTask {
	taskID := fmt.Sprintf("export_%d", time.Now().UnixNano())

	task := &ExportTask{
		TaskID:     taskID,
		TalkerID:   req.TalkerID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Format:     req.Format,
		Speed:      req.Speed,
		Resolution: req.Resolution,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
	}

	e.TaskManager.AddTask(task)

	go e.executeTask(task)

	return task
}

// ExportRequest 导出请求参数
type ExportRequest struct {
	TalkerID   string `json:"talker_id" binding:"required"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Format     string `json:"format"`
	Speed      int    `json:"speed"`
	Resolution string `json:"resolution"`
}

// executeTask 异步执行导出任务
func (e *Exporter) executeTask(task *ExportTask) {
	e.TaskManager.UpdateTask(task.TaskID, func(t *ExportTask) {
		t.Status = StatusProcessing
	})

	// 1. 获取消息
	messages, err := e.fetchMessages(task)
	if err != nil {
		e.failTask(task.TaskID, fmt.Sprintf("获取消息失败: %v", err))
		return
	}
	if len(messages) == 0 {
		e.failTask(task.TaskID, "没有找到消息")
		return
	}

	// 2. 渲染帧
	width, height := ResolutionSize(task.Resolution)
	framesDir := filepath.Join(e.OutputDir, task.TaskID+"_frames")
	os.MkdirAll(framesDir, 0755)
	defer os.RemoveAll(framesDir)

	totalFrames := len(messages)
	e.TaskManager.UpdateTask(task.TaskID, func(t *ExportTask) {
		t.TotalFrames = totalFrames
	})

	err = e.renderFrames(messages, framesDir, width, height, task)
	if err != nil {
		e.failTask(task.TaskID, fmt.Sprintf("渲染帧失败: %v", err))
		return
	}

	// 3. 合成视频/GIF
	outputFile, err := e.composeOutput(framesDir, task)
	if err != nil {
		e.failTask(task.TaskID, fmt.Sprintf("合成失败: %v", err))
		return
	}

	e.TaskManager.UpdateTask(task.TaskID, func(t *ExportTask) {
		t.Status = StatusCompleted
		t.Progress = 100
		t.ProcessedFrames = t.TotalFrames
		t.FilePath = outputFile
	})

	log.Info().Str("task_id", task.TaskID).Msg("回放导出完成")
}

// fetchMessages 获取导出任务所需的消息
func (e *Exporter) fetchMessages(task *ExportTask) ([]*model.Message, error) {
	var start, end time.Time

	if task.StartDate != "" {
		if t, err := time.Parse("2006-01-02", task.StartDate); err == nil {
			start = t
		}
	}
	if start.IsZero() {
		start = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	}

	if task.EndDate != "" {
		if t, err := time.Parse("2006-01-02", task.EndDate); err == nil {
			end = t.Add(24*time.Hour - time.Second)
		}
	}
	if end.IsZero() {
		end = time.Now().Add(24 * time.Hour)
	}

	query := types.MessageQuery{
		Talker:    task.TalkerID,
		StartTime: start,
		EndTime:   end,
		Limit:     200000,
		Offset:    0,
	}

	return e.Store.GetMessages(context.Background(), query)
}

// renderFrames 使用 chromedp 将消息渲染为 JPEG 帧序列
// 使用 file:// 协议加载 HTML（避免 document.write 对大 HTML 的 CDP 限制）
func (e *Exporter) renderFrames(messages []*model.Message, framesDir string, width, height int, task *ExportTask) error {
	renderer, err := NewRenderer()
	if err != nil {
		return err
	}

	// 创建临时目录存放 HTML 文件
	htmlDir := filepath.Join(e.OutputDir, task.TaskID+"_html")
	os.MkdirAll(htmlDir, 0755)
	defer os.RemoveAll(htmlDir)

	// 创建 chromedp 上下文
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(width, height),
		chromedp.Flag("headless", true),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 每帧显示最近 N 条消息（模拟聊天窗口滚动）
	windowSize := 10

	// 分批处理：每批最多 200 帧，避免 chromedp 长时间运行导致内存泄漏
	batchSize := 200

	for i, msg := range messages {
		// 每批重建 chromedp 上下文，释放内存
		if i > 0 && i%batchSize == 0 {
			cancel()
			allocCancel()

			allocCtx, allocCancel = chromedp.NewExecAllocator(context.Background(), opts...)
			ctx, cancel = chromedp.NewContext(allocCtx)

			log.Info().Str("task_id", task.TaskID).Int("frame", i).Msg("重建 chromedp 上下文")
		}

		// 计算窗口范围
		startIdx := 0
		if i-windowSize+1 > 0 {
			startIdx = i - windowSize + 1
		}
		windowMsgs := messages[startIdx : i+1]

		html, err := renderer.RenderFrame(FrameData{
			Messages:   windowMsgs,
			Width:      width,
			Height:     height,
			Background: "#ebebeb",
		})
		if err != nil {
			log.Error().Err(err).Int("frame", i).Msg("渲染帧 HTML 失败")
			return fmt.Errorf("渲染第 %d 帧失败: %w", i, err)
		}

		// 将 HTML 写入临时文件，使用 file:// 协议加载
		htmlPath := filepath.Join(htmlDir, fmt.Sprintf("frame_%06d.html", i))
		if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
			return fmt.Errorf("写入第 %d 帧 HTML 失败: %w", i, err)
		}

		fileURL := "file://" + htmlPath

		// 截图：通过 Navigate 加载本地文件（避免 document.write 大小限制）
		framePath := filepath.Join(framesDir, fmt.Sprintf("frame_%06d.jpg", i))
		var buf []byte
		err = chromedp.Run(ctx,
			chromedp.Navigate(fileURL),
			chromedp.Sleep(150*time.Millisecond),
			chromedp.FullScreenshot(&buf, 90),
		)
		if err != nil {
			log.Error().Err(err).Int("frame", i).Str("html_path", htmlPath).Msg("chromedp 截图失败")
			return fmt.Errorf("截图第 %d 帧失败: %w", i, err)
		}

		if len(buf) == 0 {
			log.Warn().Int("frame", i).Msg("截图结果为空，跳过")
			continue
		}

		if err := os.WriteFile(framePath, buf, 0644); err != nil {
			return fmt.Errorf("保存第 %d 帧失败: %w", i, err)
		}

		// 计算帧间延迟用于后续合成
		_ = msg // delay 在合成阶段处理

		// 更新进度
		processed := i + 1
		progress := processed * 90 / len(messages) // 90% 给渲染，10% 给合成
		e.TaskManager.UpdateTask(task.TaskID, func(t *ExportTask) {
			t.ProcessedFrames = processed
			t.Progress = progress
		})
	}

	// 清理最后一批的 chromedp 上下文（defer 会处理，但显式调用更清晰）
	cancel()
	allocCancel()

	return nil
}

// composeOutput 使用 FFmpeg 将帧序列合成为视频或 GIF
func (e *Exporter) composeOutput(framesDir string, task *ExportTask) (string, error) {
	// 合成前验证帧文件有效性
	validCount, err := e.validateFrames(framesDir)
	if err != nil {
		return "", fmt.Errorf("帧文件验证失败: %w", err)
	}
	if validCount == 0 {
		return "", fmt.Errorf("没有有效的帧文件可供合成")
	}
	log.Info().Str("task_id", task.TaskID).Int("valid_frames", validCount).Msg("帧文件验证通过")

	format := task.Format
	if format == "" {
		format = "mp4"
	}

	speed := task.Speed
	if speed <= 0 {
		speed = 4
	}

	// 帧率根据速度调整：速度越快帧率越高
	fps := speed * 2
	if fps < 2 {
		fps = 2
	}
	if fps > 30 {
		fps = 30
	}

	inputPattern := filepath.Join(framesDir, "frame_%06d.jpg")
	var outputFile string
	var args []string

	switch format {
	case "gif":
		outputFile = filepath.Join(e.OutputDir, task.TaskID+".gif")
		args = []string{
			"-f", "image2",
			"-framerate", fmt.Sprintf("%d", fps),
			"-i", inputPattern,
			"-vf", "fps=10,scale=480:-1:flags=lanczos",
			"-y", outputFile,
		}
	default:
		outputFile = filepath.Join(e.OutputDir, task.TaskID+".mp4")
		args = []string{
			"-f", "image2",
			"-framerate", fmt.Sprintf("%d", fps),
			"-i", inputPattern,
			"-c:v", "libx264",
			"-pix_fmt", "yuv420p",
			"-y", outputFile,
		}
	}

	ffmpegBin := dat2img.FFMpegPath
	if ffmpegBin == "" {
		ffmpegBin = "ffmpeg"
	}

	log.Info().Str("task_id", task.TaskID).Str("ffmpeg", ffmpegBin).Strs("args", args).Msg("开始 FFmpeg 合成")

	cmd := exec.Command(ffmpegBin, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Str("output", string(output)).Msg("FFmpeg 合成失败")
		return "", fmt.Errorf("FFmpeg 合成失败: %v, output: %s", err, string(output))
	}

	// 验证输出文件
	info, err := os.Stat(outputFile)
	if err != nil || info.Size() == 0 {
		return "", fmt.Errorf("合成输出文件无效: %s", outputFile)
	}

	log.Info().Str("task_id", task.TaskID).Int64("size", info.Size()).Msg("FFmpeg 合成完成")
	return outputFile, nil
}

// validateFrames 检查帧目录中的 JPEG 文件是否有效（非空且具有 JPEG 头）
func (e *Exporter) validateFrames(framesDir string) (int, error) {
	entries, err := os.ReadDir(framesDir)
	if err != nil {
		return 0, err
	}

	validCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		// 跳过空文件
		if info.Size() < 100 {
			log.Warn().Str("file", entry.Name()).Int64("size", info.Size()).Msg("帧文件过小，可能无效")
			continue
		}
		validCount++
	}
	return validCount, nil
}

// failTask 标记任务失败
func (e *Exporter) failTask(taskID string, errMsg string) {
	log.Error().Str("task_id", taskID).Msg(errMsg)
	e.TaskManager.UpdateTask(taskID, func(t *ExportTask) {
		t.Status = StatusFailed
		t.Error = errMsg
	})
}

// cleanupLoop 定期清理过期任务和文件
func (e *Exporter) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		e.TaskManager.CleanExpired()
	}
}
