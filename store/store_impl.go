package store

import (
	"context"
	"fmt"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/bind"
	"github.com/afumu/wetrace/store/core"
	"github.com/afumu/wetrace/store/repo"
	"github.com/afumu/wetrace/store/strategy"
	"github.com/afumu/wetrace/store/types"
	"github.com/fsnotify/fsnotify"
)

// DefaultStore 是 Store 接口的默认实现
type DefaultStore struct {
	pool    *core.ConnectionPool
	router  *bind.TimelineRouter
	watcher *core.Watcher
	repo    *repo.Repository
}

// NewStore 初始化一个新的存储实例
func NewStore(baseDir string) (*DefaultStore, error) {
	// 1. 初始化核心组件
	pool := core.NewConnectionPool(baseDir)
	watcher, err := core.NewWatcher(baseDir)
	if err != nil {
		pool.CloseAll()
		return nil, err
	}

	// 2. 策略层
	strat := strategy.NewV4()
	router := bind.NewTimelineRouter(baseDir, pool, strat)

	// 3. 构建索引 (这一步可能比较耗时，但必须在启动时完成)
	if err := router.RebuildIndex(context.Background()); err != nil {
		pool.CloseAll()
		watcher.Stop()
		return nil, fmt.Errorf("构建时间线索引失败: %w", err)
	}

	// 4. 初始化仓储
	r := repo.New(router, pool)

	// 5. 启动文件监听
	watcher.Start()

	// 注册自动刷新逻辑：当有新文件生成时，重建索引
	watcher.AddCallback(func(event fsnotify.Event) {
		if event.Op&fsnotify.Create == fsnotify.Create {
			// 只有当新文件被策略识别为消息数据库时，才重建索引
			if meta, ok := strat.Identify(event.Name); ok && meta.Type == strategy.Message {
				_ = router.RebuildIndex(context.Background())
			}
		}
	})

	return &DefaultStore{
		pool:    pool,
		router:  router,
		watcher: watcher,
		repo:    r,
	}, nil
}

func (s *DefaultStore) Close() error {
	s.watcher.Stop()
	return s.pool.CloseAll()
}

// --- 下面是 Store 接口的代理实现 ---

func (s *DefaultStore) GetMessages(ctx context.Context, query types.MessageQuery) ([]*model.Message, error) {
	return s.repo.GetMessages(ctx, query)
}

func (s *DefaultStore) SearchGlobalMessages(ctx context.Context, query types.MessageQuery) ([]*model.Message, error) {
	return s.repo.SearchGlobalMessages(ctx, query)
}

func (s *DefaultStore) GetContacts(ctx context.Context, query types.ContactQuery) ([]*model.Contact, error) {
	return s.repo.GetContacts(ctx, query)
}

func (s *DefaultStore) GetChatRooms(ctx context.Context, query types.ChatRoomQuery) ([]*model.ChatRoom, error) {
	return s.repo.GetChatRooms(ctx, query)
}

func (s *DefaultStore) GetSessions(ctx context.Context, query types.SessionQuery) ([]*model.Session, error) {
	return s.repo.GetSessions(ctx, query)
}

func (s *DefaultStore) DeleteSession(ctx context.Context, username string) error {
	return s.repo.DeleteSession(ctx, username)
}

func (s *DefaultStore) GetMedia(ctx context.Context, mediaType string, key string) (*model.Media, error) {
	return s.repo.GetMedia(ctx, mediaType, key)
}

func (s *DefaultStore) GetHourlyActivity(ctx context.Context, sessionID string) ([]*model.HourlyStat, error) {
	return s.repo.GetHourlyActivity(ctx, sessionID)
}

func (s *DefaultStore) GetDailyActivity(ctx context.Context, sessionID string) ([]*model.DailyStat, error) {
	return s.repo.GetDailyActivity(ctx, sessionID)
}

func (s *DefaultStore) GetWeekdayActivity(ctx context.Context, sessionID string) ([]*model.WeekdayStat, error) {
	return s.repo.GetWeekdayActivity(ctx, sessionID)
}

func (s *DefaultStore) GetMonthlyActivity(ctx context.Context, sessionID string) ([]*model.MonthlyStat, error) {
	return s.repo.GetMonthlyActivity(ctx, sessionID)
}

func (s *DefaultStore) GetMessageTypeDistribution(ctx context.Context, sessionID string) ([]*model.MessageTypeStat, error) {
	return s.repo.GetMessageTypeDistribution(ctx, sessionID)
}

func (s *DefaultStore) GetMemberActivity(ctx context.Context, sessionID string) ([]*model.MemberActivity, error) {
	return s.repo.GetMemberActivity(ctx, sessionID)
}

func (s *DefaultStore) GetRepeatAnalysis(ctx context.Context, sessionID string) ([]*model.RepeatStat, error) {
	return s.repo.GetRepeatAnalysis(ctx, sessionID)
}

func (s *DefaultStore) GetPersonalTopContacts(ctx context.Context, limit int) ([]*model.PersonalTopContact, error) {
	return s.repo.GetPersonalTopContacts(ctx, limit)
}

func (s *DefaultStore) GetDashboardData(ctx context.Context) (*model.DashboardData, error) {
	return s.repo.GetDashboardData(ctx)
}

func (s *DefaultStore) Watch(group string, callback func(event fsnotify.Event) error) error {
	s.watcher.AddCallback(func(event fsnotify.Event) {
		_ = callback(event)
	})
	return nil
}

// Reload 重新加载存储（重建索引、刷新连接等）
func (s *DefaultStore) Reload() error {
	// 1. 关闭所有现有连接（这将强制下次查询时重新打开连接）
	if err := s.pool.CloseAll(); err != nil {
		return fmt.Errorf("reload: close all connections failed: %w", err)
	}

	// 2. 重新构建时间线索引（扫描目录，重新发现文件）
	if err := s.router.RebuildIndex(context.Background()); err != nil {
		return fmt.Errorf("reload: rebuild index failed: %w", err)
	}

	return nil
}
