package monitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookPayload Webhook推送载荷
type WebhookPayload struct {
	Event     string         `json:"event"`
	Timestamp int64          `json:"timestamp"`
	Keyword   string         `json:"keyword"`
	Message   WebhookMessage `json:"message"`
}

// WebhookMessage Webhook消息详情
type WebhookMessage struct {
	Talker     string `json:"talker"`
	TalkerName string `json:"talker_name"`
	Sender     string `json:"sender"`
	SenderName string `json:"sender_name"`
	Content    string `json:"content"`
	Time       string `json:"time"`
	IsChatroom bool   `json:"is_chatroom"`
}

// SendWebhook 发送Webhook通知
func SendWebhook(url string, payload WebhookPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

// TestWebhookURL 测试Webhook连通性
func TestWebhookURL(url string) (int, error) {
	payload := WebhookPayload{
		Event:     "test",
		Timestamp: time.Now().Unix(),
		Message: WebhookMessage{
			Content: "WeTrace Webhook 连通性测试",
			Time:    time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal payload: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("send test webhook: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
