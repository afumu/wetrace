package monitor

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
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

// --- 飞书多维表格（Bitable）推送 ---

// tenantTokenCache 缓存 tenant_access_token
var tenantTokenCache struct {
	sync.RWMutex
	token   string
	expires time.Time
}

// getTenantAccessToken 获取飞书 tenant_access_token
func getTenantAccessToken(appID, appSecret string) (string, error) {
	tenantTokenCache.RLock()
	if tenantTokenCache.token != "" && time.Now().Before(tenantTokenCache.expires) {
		token := tenantTokenCache.token
		tenantTokenCache.RUnlock()
		return token, nil
	}
	tenantTokenCache.RUnlock()

	body := map[string]string{
		"app_id":     appID,
		"app_secret": appSecret,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal token request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(
		"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal",
		"application/json; charset=utf-8",
		bytes.NewReader(data),
	)
	if err != nil {
		return "", fmt.Errorf("request tenant_access_token: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	if err := json.Unmarshal(respData, &result); err != nil {
		return "", fmt.Errorf("unmarshal token response: %w", err)
	}
	if result.Code != 0 {
		return "", fmt.Errorf("get tenant_access_token failed: code=%d, msg=%s", result.Code, result.Msg)
	}

	tenantTokenCache.Lock()
	tenantTokenCache.token = result.TenantAccessToken
	tenantTokenCache.expires = time.Now().Add(time.Duration(result.Expire-60) * time.Second)
	tenantTokenCache.Unlock()

	return result.TenantAccessToken, nil
}

// SendBitableRecord 写入飞书多维表格记录
func SendBitableRecord(appID, appSecret, appToken, tableID string, msg WebhookMessage, ruleName string) error {
	token, err := getTenantAccessToken(appID, appSecret)
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	fields := map[string]interface{}{
		"发送人":  msg.SenderName,
		"会话":   msg.TalkerName,
		"消息内容": msg.Content,
		"触发规则": ruleName,
		"消息时间": msg.Time,
		"告警时间": time.Now().Format("2006-01-02 15:04:05"),
	}

	reqBody := map[string]interface{}{
		"fields": fields,
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal bitable record: %w", err)
	}

	url := fmt.Sprintf("https://open.feishu.cn/open-apis/bitable/v1/apps/%s/tables/%s/records", appToken, tableID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create bitable request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send bitable record: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read bitable response: %w", err)
	}

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(respData, &result); err != nil {
		return fmt.Errorf("unmarshal bitable response: %w", err)
	}
	if result.Code != 0 {
		return fmt.Errorf("bitable write failed: code=%d, msg=%s", result.Code, result.Msg)
	}
	return nil
}

// TestBitableConnection 测试飞书多维表格连通性
func TestBitableConnection(appID, appSecret, appToken, tableID string) error {
	token, err := getTenantAccessToken(appID, appSecret)
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	url := fmt.Sprintf("https://open.feishu.cn/open-apis/bitable/v1/apps/%s/tables/%s/records?page_size=1", appToken, tableID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("create test request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("test bitable connection: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read test response: %w", err)
	}

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(respData, &result); err != nil {
		return fmt.Errorf("unmarshal test response: %w", err)
	}
	if result.Code != 0 {
		return fmt.Errorf("bitable test failed: code=%d, msg=%s", result.Code, result.Msg)
	}
	return nil
}
