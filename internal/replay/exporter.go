package replay

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/afumu/wetrace/internal/model"
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

// renderFrames 使用 chromedp 将消息渲染为 PNG 帧序列
func (e *Exporter) renderFrames(messages []*model.Message, framesDir string, width, height int, task *ExportTask) error {
	renderer, err := NewRenderer()
	if err != nil {
		return err
	}

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

	for i, msg := range messages {
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
			return fmt.Errorf("渲染第 %d 帧失败: %w", i, err)
		}

		// 截图
		framePath := filepath.Join(framesDir, fmt.Sprintf("frame_%06d.png", i))
		var buf []byte
		err = chromedp.Run(ctx,
			chromedp.Navigate("about:blank"),
			chromedp.ActionFunc(func(ctx context.Context) error {
				return chromedp.Evaluate(fmt.Sprintf(`document.open(); document.write(%q); document.close();`, html), nil).Do(ctx)
			}),
			chromedp.Sleep(100*time.Millisecond),
			chromedp.FullScreenshot(&buf, 90),
		)
		if err != nil {
			return fmt.Errorf("截图第 %d 帧失败: %w", i, err)
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

	return nil
}

// composeOutput 使用 FFmpeg 将帧序列合成为视频或 GIF
func (e *Exporter) composeOutput(framesDir string, task *ExportTask) (string, error) {
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

	inputPattern := filepath.Join(framesDir, "frame_%06d.png")
	var outputFile string
	var args []string

	switch format {
	case "gif":
		outputFile = filepath.Join(e.OutputDir, task.TaskID+".gif")
		args = []string{
			"-framerate", fmt.Sprintf("%d", fps),
			"-i", inputPattern,
			"-vf", "fps=10,scale=480:-1:flags=lanczos",
			"-y", outputFile,
		}
	default:
		outputFile = filepath.Join(e.OutputDir, task.TaskID+".mp4")
		args = []string{
			"-framerate", fmt.Sprintf("%d", fps),
			"-i", inputPattern,
			"-c:v", "libx264",
			"-pix_fmt", "yuv420p",
			"-y", outputFile,
		}
	}

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("FFmpeg 合成失败: %v, output: %s", err, string(output))
	}

	return outputFile, nil
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
