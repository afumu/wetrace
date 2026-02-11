package tts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// Client Whisper API 客户端（OpenAI 兼容）
type Client struct {
	APIKey  string
	BaseURL string
	Model   string
}

// NewClient 创建 TTS 客户端
func NewClient(apiKey, baseURL, model string) *Client {
	if model == "" {
		model = "whisper-1"
	}
	return &Client{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
	}
}

// Transcribe 将音频文件转为文字
func (c *Client) Transcribe(audioData []byte, filename string) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加 model 字段
	if err := writer.WriteField("model", c.Model); err != nil {
		return "", fmt.Errorf("write model field: %w", err)
	}

	// 添加 language 字段（中文优先）
	if err := writer.WriteField("language", "zh"); err != nil {
		return "", fmt.Errorf("write language field: %w", err)
	}

	// 添加音频文件
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return "", fmt.Errorf("write audio data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close writer: %w", err)
	}

	// 构建请求 URL
	url := fmt.Sprintf("%s/audio/transcriptions", c.BaseURL)

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respData))
	}

	// 解析响应
	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respData, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	return result.Text, nil
}
