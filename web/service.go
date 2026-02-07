package web

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/afumu/wetrace/store"
	"github.com/afumu/wetrace/web/api"
	"github.com/afumu/wetrace/web/media"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Service 定义了 web 服务。
type Service struct {
	store    store.Store
	router   *gin.Engine
	server   *http.Server
	conf     *Config
	api      *api.API
	media    *media.Service
	staticFS fs.FS
}

// Config 保存 web 服务的配置。
type Config struct {
	ListenAddr      string
	DataDir         string
	ImageKey        string
	XorKey          string
	WechatDbSrcPath string
	WechatDbKey     string
	WxKeyDllPath    string
	WechatPath      string
	WechatDataPath  string
}

// NewService 创建一个新的 web 服务。
func NewService(store store.Store, conf *Config, staticFS fs.FS) *Service {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	mediaService := media.NewService(conf.DataDir, conf.ImageKey, conf.XorKey, conf.WechatDbSrcPath)

	// 创建共享的 API 配置指针
	apiConf := &api.Config{
		WechatDbSrcPath: conf.WechatDbSrcPath,
		WechatDbKey:     conf.WechatDbKey,
		WxKeyDllPath:    conf.WxKeyDllPath,
		WechatPath:      conf.WechatPath,
		WechatDataPath:  conf.WechatDataPath,
		ImageKey:        conf.ImageKey,
		XorKey:          conf.XorKey,
	}

	apiHandler := api.NewAPI(store, mediaService, apiConf, staticFS)

	s := &Service{
		store:    store,
		router:   router,
		conf:     conf,
		api:      apiHandler,
		media:    mediaService,
		staticFS: staticFS,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// Start 开始提供 web 应用服务。
func (s *Service) Start() error {
	s.server = &http.Server{
		Addr:    s.conf.ListenAddr,
		Handler: s.router,
	}

	log.Info().Msg(fmt.Sprintf("在 %s 上启动 web 服务", s.conf.ListenAddr))

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Web 服务启动失败")
		}
	}()

	return nil
}

// Stop 优雅地关闭 web 服务器。
func (s *Service) Stop() error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("优雅关闭 web 服务器失败")
		return err
	}

	log.Info().Msg("Web 服务已停止")
	return nil
}

func (s *Service) GetRouter() *gin.Engine {
	return s.router
}
