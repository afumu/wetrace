package monitor

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// FeishuCardMessage 飞书卡片消息
type FeishuCardMessage struct {
	MsgType string      `json:"msg_type"`
	Card    FeishuCard  `json:"card"`
}

// FeishuCard 飞书卡片
type FeishuCard struct {
	Header   FeishuCardHeader    `json:"header"`
	Elements []FeishuCardElement `json:"elements"`
}

// FeishuCardHeader 飞书卡片头部
type FeishuCardHeader struct {
	Title    FeishuText `json:"title"`
	Template string     `json:"template"`
}

// FeishuText 飞书文本
type FeishuText struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// FeishuCardElement 飞书卡片元素
type FeishuCardElement struct {
	Tag    string       `json:"tag"`
	Text   *FeishuText  `json:"text,omitempty"`
	Fields []FeishuField `json:"fields,omitempty"`
}

// FeishuField 飞书字段
type FeishuField struct {
	IsShort bool       `json:"is_short"`
	Text    FeishuText `json:"text"`
}

// feishuSign 生成飞书签名
func feishuSign(secret string, timestamp int64) (string, error) {
	stringToSign := strconv.FormatInt(timestamp, 10) + "\n" + secret
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write([]byte{})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// sendFeishuRequest 发送飞书请求
func sendFeishuRequest(webhookURL, secret string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal feishu message: %w", err)
	}

	// 如果有签名密钥，添加签名
	if secret != "" {
		ts := time.Now().Unix()
		sign, err := feishuSign(secret, ts)
		if err != nil {
			return fmt.Errorf("generate feishu sign: %w", err)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		m["timestamp"] = strconv.FormatInt(ts, 10)
		m["sign"] = sign
		data, err = json.Marshal(m)
		if err != nil {
			return err
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("send feishu message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feishu returned status %d", resp.StatusCode)
	}
	return nil
}

// SendFeishuAlert 发送飞书告警卡片
func SendFeishuAlert(webhookURL, secret string, msg WebhookMessage, ruleName string) error {
	card := FeishuCardMessage{
		MsgType: "interactive",
		Card: FeishuCard{
			Header: FeishuCardHeader{
				Title:    FeishuText{Tag: "plain_text", Content: "WeTrace 消息监控告警"},
				Template: "red",
			},
			Elements: []FeishuCardElement{
				{
					Tag: "div",
					Fields: []FeishuField{
						{IsShort: true, Text: FeishuText{Tag: "lark_md", Content: "**发送人：**\n" + msg.SenderName}},
						{IsShort: true, Text: FeishuText{Tag: "lark_md", Content: "**会话：**\n" + msg.TalkerName}},
					},
				},
				{
					Tag:  "div",
					Text: &FeishuText{Tag: "lark_md", Content: "**消息内容：**\n" + msg.Content},
				},
				{
					Tag:  "div",
					Text: &FeishuText{Tag: "lark_md", Content: "**触发规则：**\n" + ruleName},
				},
			},
		},
	}
	return sendFeishuRequest(webhookURL, secret, card)
}

// TestFeishuBot 测试飞书机器人连通性
func TestFeishuBot(webhookURL, secret string) error {
	card := FeishuCardMessage{
		MsgType: "interactive",
		Card: FeishuCard{
			Header: FeishuCardHeader{
				Title:    FeishuText{Tag: "plain_text", Content: "WeTrace 连通性测试"},
				Template: "blue",
			},
			Elements: []FeishuCardElement{
				{
					Tag:  "div",
					Text: &FeishuText{Tag: "lark_md", Content: "飞书机器人连通性测试成功"},
				},
			},
		},
	}
	return sendFeishuRequest(webhookURL, secret, card)
}
