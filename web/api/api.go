package api

import (
	"io/fs"
	"sync"

	"github.com/afumu/wetrace/internal/ai"
	"github.com/afumu/wetrace/store"
	"github.com/afumu/wetrace/web/export"
	"github.com/afumu/wetrace/web/media"
)

// API 封装了 API 处理器所需的所有依赖。
type API struct {
	Store  store.Store
	Media  *media.Service
	Export *export.Service
	Conf   *Config
	AI     *ai.Client
	mu     sync.Mutex
}

type Config struct {
	WechatDbSrcPath string
	WechatDbKey     string
	WxKeyDllPath    string
	WechatPath      string
	WechatDataPath  string
	ImageKey        string
	XorKey          string
	AIEnabled       bool
	AIProvider      string
	AIAPIKey        string
	AIBaseURL       string
	AIModel         string
}

// NewAPI 创建一个新的 API 处理器。
func NewAPI(s store.Store, m *media.Service, conf *Config, staticFS fs.FS) *API {
	var aiClient *ai.Client
	if conf.AIEnabled {
		aiClient = ai.NewClient(conf.AIAPIKey, conf.AIBaseURL, conf.AIModel)
	}

	return &API{
		Store: s,
		Media: m,
		Export: &export.Service{
			Media:    m,
			Store:    s,
			StaticFS: staticFS,
		},
		Conf: conf,
		AI:   aiClient,
	}
}
