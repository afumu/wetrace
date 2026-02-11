package monitor

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/afumu/wetrace/internal/ai"
	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store"
	"github.com/afumu/wetrace/store/types"
	"github.com/rs/zerolog/log"
)

// Checker 定时检查消息并触发告警
type Checker struct {
	mu      sync.Mutex
	store   store.Store
	monitor *Store
	ai      *ai.Client
	stopCh  chan struct{}
	running bool
}

// NewChecker 创建消息检查器
func NewChecker(s store.Store, m *Store, aiClient *ai.Client) *Checker {
	return &Checker{
		store:   s,
		monitor: m,
		ai:      aiClient,
	}
}

// Start 启动定时检查
func (c *Checker) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.running {
		return
	}
	c.running = true
	c.stopCh = make(chan struct{})
	go c.loop()
}

// Stop 停止定时检查
func (c *Checker) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.running {
		return
	}
	close(c.stopCh)
	c.running = false
}

// IsRunning 是否正在运行
func (c *Checker) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}

// loop 主循环，每分钟检查一次是否有配置需要执行
func (c *Checker) loop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// 启动时立即执行一次
	c.checkAll()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.checkAll()
		}
	}
}

// checkAll 检查所有启用的监控配置
func (c *Checker) checkAll() {
	configs := c.monitor.GetEnabledConfigs()
	now := time.Now().Unix()

	for _, cfg := range configs {
		interval := cfg.IntervalMinutes
		if interval <= 0 {
			interval = 5 // 默认5分钟
		}

		// 判断是否到达检查时间
		if cfg.LastCheckTime > 0 && now-cfg.LastCheckTime < int64(interval*60) {
			continue
		}

		c.checkConfig(cfg)
	}
}

// checkConfig 检查单个监控配置
func (c *Checker) checkConfig(cfg MonitorConfig) {
	now := time.Now()
	ctx := context.Background()

	// 确定检查的时间范围
	startTime := time.Unix(cfg.LastCheckTime, 0)
	if cfg.LastCheckTime == 0 {
		startTime = now.Add(-time.Duration(cfg.IntervalMinutes) * time.Minute)
		if cfg.IntervalMinutes <= 0 {
			startTime = now.Add(-5 * time.Minute)
		}
	}

	// 查询指定会话的新消息
	sessions := cfg.SessionIDs
	if len(sessions) == 0 {
		// 无指定会话则跳过（避免全量扫描）
		_ = c.monitor.UpdateLastCheckTime(cfg.ID, now.Unix())
		return
	}

	for _, talker := range sessions {
		c.checkSession(ctx, cfg, talker, startTime, now)
	}

	// 更新检查时间
	_ = c.monitor.UpdateLastCheckTime(cfg.ID, now.Unix())
}

// checkSession 检查单个会话的新消息
func (c *Checker) checkSession(ctx context.Context, cfg MonitorConfig, talker string, startTime, endTime time.Time) {
	query := types.MessageQuery{
		Talker:    talker,
		StartTime: startTime,
		EndTime:   endTime,
		MsgType:   model.MessageTypeText,
		Limit:     500,
	}

	messages, err := c.store.GetMessages(ctx, query)
	if err != nil {
		log.Error().Err(err).Str("talker", talker).Msg("监控检查：查询消息失败")
		return
	}

	for _, msg := range messages {
		if msg.IsSelf {
			continue
		}
		content := msg.Content
		if content == "" {
			continue
		}

		matched, matchInfo := c.matchMessage(cfg, content)
		if matched {
			c.sendAlert(cfg, msg, matchInfo)
		}
	}
}

// matchMessage 检查消息是否匹配规则
func (c *Checker) matchMessage(cfg MonitorConfig, content string) (bool, string) {
	switch cfg.Type {
	case "keyword":
		lower := strings.ToLower(content)
		for _, kw := range cfg.Keywords {
			if strings.Contains(lower, strings.ToLower(kw)) {
				return true, "关键词: " + kw
			}
		}
		return false, ""
	case "ai":
		if c.ai == nil || cfg.Prompt == "" {
			return false, ""
		}
		resp, err := c.ai.Chat([]ai.Message{
			{Role: "system", Content: cfg.Prompt},
			{Role: "user", Content: content},
		})
		if err != nil {
			log.Error().Err(err).Msg("监控AI匹配失败")
			return false, ""
		}
		lower := strings.ToLower(resp)
		if strings.Contains(lower, "yes") || strings.Contains(lower, "是") || strings.Contains(lower, "匹配") {
			return true, "AI判定: " + resp
		}
		return false, ""
	}
	return false, ""
}

// sendAlert 发送告警通知
func (c *Checker) sendAlert(cfg MonitorConfig, msg *model.Message, matchInfo string) {
	webhookMsg := WebhookMessage{
		Talker:     msg.Talker,
		TalkerName: msg.TalkerName,
		Sender:     msg.Sender,
		SenderName: msg.SenderName,
		Content:    msg.Content,
		Time:       msg.Time.Format("2006-01-02 15:04:05"),
		IsChatroom: strings.HasSuffix(msg.Talker, "@chatroom"),
	}

	ruleName := cfg.Name + " (" + matchInfo + ")"

	// 根据平台发送告警
	switch cfg.Platform {
	case "feishu":
		c.sendFeishuAlert(cfg, webhookMsg, ruleName)
	case "webhook":
		if cfg.WebhookURL != "" {
			c.sendWebhookAlert(cfg, webhookMsg, ruleName)
		}
	}
}

// sendFeishuAlert 通过飞书发送告警
func (c *Checker) sendFeishuAlert(cfg MonitorConfig, msg WebhookMessage, ruleName string) {
	feishuCfg := c.monitor.GetFeishuConfig()
	pushType := feishuCfg.PushType
	if pushType == "" {
		pushType = "bot"
	}

	// 机器人推送
	if pushType == "bot" || pushType == "both" {
		webhook := cfg.FeishuURL
		if webhook == "" {
			webhook = feishuCfg.BotWebhook
		}
		if webhook != "" {
			if err := SendFeishuAlert(webhook, feishuCfg.SignSecret, msg, ruleName); err != nil {
				log.Error().Err(err).Msg("飞书机器人告警发送失败")
			}
		}
	}

	// 多维表格推送
	if pushType == "bitable" || pushType == "both" {
		if feishuCfg.AppID != "" && feishuCfg.AppToken != "" && feishuCfg.TableID != "" {
			if err := SendBitableRecord(feishuCfg.AppID, feishuCfg.AppSecret, feishuCfg.AppToken, feishuCfg.TableID, msg, ruleName); err != nil {
				log.Error().Err(err).Msg("飞书多维表格告警写入失败")
			}
		}
	}
}

// sendWebhookAlert 通过Webhook发送告警
func (c *Checker) sendWebhookAlert(cfg MonitorConfig, msg WebhookMessage, ruleName string) {
	payload := WebhookPayload{
		Event:     "monitor_alert",
		Timestamp: time.Now().Unix(),
		Keyword:   ruleName,
		Message:   msg,
	}
	if err := SendWebhook(cfg.WebhookURL, payload); err != nil {
		log.Error().Err(err).Msg("Webhook告警发送失败")
	}
}
