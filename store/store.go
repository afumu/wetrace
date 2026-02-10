package store

import (
	"context"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/types"
	"github.com/fsnotify/fsnotify"
)

// Store 定义了数据访问的统一接口。
// 它屏蔽了底层的文结构和平台差异。
type Store interface {
	// 消息操作
	GetMessages(ctx context.Context, query types.MessageQuery) ([]*model.Message, error)
	SearchGlobalMessages(ctx context.Context, query types.MessageQuery) ([]*model.Message, error)

	// 联系人操作
	GetContacts(ctx context.Context, query types.ContactQuery) ([]*model.Contact, error)
	GetChatRooms(ctx context.Context, query types.ChatRoomQuery) ([]*model.ChatRoom, error)
	GetSessions(ctx context.Context, query types.SessionQuery) ([]*model.Session, error)
	DeleteSession(ctx context.Context, username string) error

	// 媒体操作
	GetMedia(ctx context.Context, mediaType string, key string) (*model.Media, error)

	// 分析操作
	GetHourlyActivity(ctx context.Context, sessionID string) ([]*model.HourlyStat, error)
	GetDailyActivity(ctx context.Context, sessionID string) ([]*model.DailyStat, error)
	GetWeekdayActivity(ctx context.Context, sessionID string) ([]*model.WeekdayStat, error)
	GetMonthlyActivity(ctx context.Context, sessionID string) ([]*model.MonthlyStat, error)
	GetMessageTypeDistribution(ctx context.Context, sessionID string) ([]*model.MessageTypeStat, error)
	GetMemberActivity(ctx context.Context, sessionID string) ([]*model.MemberActivity, error)
	GetRepeatAnalysis(ctx context.Context, sessionID string) ([]*model.RepeatStat, error)
	GetPersonalTopContacts(ctx context.Context, limit int) ([]*model.PersonalTopContact, error)
	GetDashboardData(ctx context.Context) (*model.DashboardData, error)

	// 搜索操作
	SearchMessages(ctx context.Context, query types.MessageQuery) (*model.SearchResult, error)
	GetMessageContext(ctx context.Context, talker string, seq int64, before, after int) ([]*model.Message, error)

	// 年度报告
	GetAnnualReport(ctx context.Context, year int) (*model.AnnualReport, error)

	// Watch 注册文件系统事件的回调函数
	Watch(group string, callback func(event fsnotify.Event) error) error

	// Reload 重新加载存储（重建索引、刷新连接等）
	Reload() error

	// 生命周期管理
	Close() error
}
