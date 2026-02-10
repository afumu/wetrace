package replay

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/afumu/wetrace/internal/model"
)

// Renderer 聊天界面 HTML 渲染器
type Renderer struct {
	tmpl *template.Template
}

// NewRenderer 创建渲染器
func NewRenderer() (*Renderer, error) {
	tmpl, err := template.New("chat").Parse(chatTemplate)
	if err != nil {
		return nil, fmt.Errorf("解析聊天模板失败: %w", err)
	}
	return &Renderer{tmpl: tmpl}, nil
}

// FrameData 单帧渲染数据
type FrameData struct {
	Messages   []*model.Message
	Width      int
	Height     int
	Background string
}

// RenderFrame 渲染单帧 HTML
func (r *Renderer) RenderFrame(data FrameData) (string, error) {
	type msgView struct {
		SenderName string
		Content    string
		Time       string
		IsSelf     bool
		AvatarText string
	}

	views := make([]msgView, 0, len(data.Messages))
	for _, msg := range data.Messages {
		name := msg.SenderName
		if name == "" {
			name = msg.Sender
		}
		avatar := "?"
		if len([]rune(name)) > 0 {
			avatar = string([]rune(name)[0])
		}
		views = append(views, msgView{
			SenderName: name,
			Content:    msg.Content,
			Time:       msg.Time.Format("15:04:05"),
			IsSelf:     msg.IsSelf,
			AvatarText: avatar,
		})
	}

	templateData := struct {
		Messages   []msgView
		Width      int
		Height     int
		Background string
	}{
		Messages:   views,
		Width:      data.Width,
		Height:     data.Height,
		Background: data.Background,
	}

	var buf bytes.Buffer
	if err := r.tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("渲染帧失败: %w", err)
	}
	return buf.String(), nil
}

// ResolutionSize 根据分辨率字符串返回宽高
func ResolutionSize(resolution string) (int, int) {
	switch resolution {
	case "1080p":
		return 1080, 1920
	default:
		return 720, 1280
	}
}

// CalculateDelay 计算两条消息之间的延迟（毫秒），考虑速度倍率
func CalculateDelay(prev, curr time.Time, speed int) int {
	if speed <= 0 {
		speed = 4
	}
	diff := curr.Sub(prev).Milliseconds()
	if diff <= 0 {
		return 100
	}
	delay := int(diff) / speed
	// 最大间隔 2 秒（避免长时间等待）
	if delay > 2000 {
		delay = 2000
	}
	// 最小间隔 100ms
	if delay < 100 {
		delay = 100
	}
	return delay
}
