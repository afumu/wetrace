package monitor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MonitorConfig 统一监控配置（关键词匹配 + AI匹配）
type MonitorConfig struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	Type            string   `json:"type"`             // "keyword" | "ai"
	Prompt          string   `json:"prompt"`            // AI提示词（type=ai时使用）
	Keywords        []string `json:"keywords"`          // 关键词列表（type=keyword时使用）
	Platform        string   `json:"platform"`          // "webhook" | "feishu"
	WebhookURL      string   `json:"webhook_url"`       // 通用Webhook URL
	FeishuURL       string   `json:"feishu_url"`        // 飞书机器人Webhook URL
	Secret          string   `json:"secret"`            // 签名密钥（可选）
	Enabled         bool     `json:"enabled"`
	SessionIDs      []string `json:"session_ids"`       // 监控哪些会话（空=全部）
	IntervalMinutes int      `json:"interval_minutes"`  // 监控间隔（分钟）
	LastCheckTime   int64    `json:"last_check_time"`   // 上次检查时间
	CreatedAt       int64    `json:"created_at"`
	UpdatedAt       int64    `json:"updated_at"`
}

// FeishuConfig 飞书平台全局配置
type FeishuConfig struct {
	BotWebhook string `json:"bot_webhook"`
	SignSecret string `json:"sign_secret"`
	Enabled    bool   `json:"enabled"`
	// 多维表格配置
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
	AppToken  string `json:"app_token"`
	TableID   string `json:"table_id"`
	PushType  string `json:"push_type"` // "bot" | "bitable" | "both"
}

// storeData 持久化数据结构
type storeData struct {
	Configs      []MonitorConfig `json:"configs"`
	FeishuConfig FeishuConfig    `json:"feishu_config"`
	NextID       int64           `json:"next_id"`
}

// Store 监控配置存储
type Store struct {
	mu       sync.RWMutex
	data     storeData
	filePath string
}

// NewStore 创建监控配置存储
func NewStore(dataDir string) (*Store, error) {
	filePath := filepath.Join(dataDir, "monitor_configs.json")
	s := &Store{
		filePath: filePath,
		data: storeData{
			NextID: 1,
		},
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// load 从文件加载数据
func (s *Store) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.data)
}

// save 保存数据到文件
func (s *Store) save() error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

// ListConfigs 获取所有监控配置
func (s *Store) ListConfigs() []MonitorConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]MonitorConfig, len(s.data.Configs))
	copy(result, s.data.Configs)
	return result
}

// CreateConfig 创建监控配置
func (s *Store) CreateConfig(cfg MonitorConfig) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg.ID = s.data.NextID
	s.data.NextID++
	now := time.Now().Unix()
	cfg.CreatedAt = now
	cfg.UpdatedAt = now
	s.data.Configs = append(s.data.Configs, cfg)
	return cfg.ID, s.save()
}

// UpdateConfig 更新监控配置
func (s *Store) UpdateConfig(id int64, cfg MonitorConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Configs {
		if s.data.Configs[i].ID == id {
			cfg.ID = id
			cfg.CreatedAt = s.data.Configs[i].CreatedAt
			cfg.UpdatedAt = time.Now().Unix()
			s.data.Configs[i] = cfg
			return s.save()
		}
	}
	return os.ErrNotExist
}

// DeleteConfig 删除监控配置
func (s *Store) DeleteConfig(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Configs {
		if s.data.Configs[i].ID == id {
			s.data.Configs = append(s.data.Configs[:i], s.data.Configs[i+1:]...)
			return s.save()
		}
	}
	return os.ErrNotExist
}

// GetFeishuConfig 获取飞书全局配置
func (s *Store) GetFeishuConfig() FeishuConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.FeishuConfig
}

// UpdateFeishuConfig 更新飞书全局配置
func (s *Store) UpdateFeishuConfig(cfg FeishuConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.FeishuConfig = cfg
	return s.save()
}

// GetEnabledConfigs 获取所有启用的监控配置
func (s *Store) GetEnabledConfigs() []MonitorConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []MonitorConfig
	for _, cfg := range s.data.Configs {
		if cfg.Enabled {
			result = append(result, cfg)
		}
	}
	return result
}

// UpdateLastCheckTime 更新指定配置的上次检查时间
func (s *Store) UpdateLastCheckTime(id int64, t int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Configs {
		if s.data.Configs[i].ID == id {
			s.data.Configs[i].LastCheckTime = t
			return s.save()
		}
	}
	return os.ErrNotExist
}
