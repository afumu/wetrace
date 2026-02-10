package api

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/afumu/wetrace/decrypt"
	"github.com/afumu/wetrace/internal/ai"
	"github.com/afumu/wetrace/internal/backup"
	"github.com/afumu/wetrace/internal/monitor"
	intsync "github.com/afumu/wetrace/internal/sync"
	"github.com/afumu/wetrace/store"
	"github.com/afumu/wetrace/store/types"
	"github.com/afumu/wetrace/web/export"
	"github.com/afumu/wetrace/web/media"
	"github.com/spf13/viper"
)

// API 封装了 API 处理器所需的所有依赖。
type API struct {
	Store           store.Store
	Media           *media.Service
	Export          *export.Service
	Conf            *Config
	AI              *ai.Client
	Password        *PasswordManager
	SyncScheduler   *intsync.Scheduler
	BackupScheduler *backup.Scheduler
	ReplayExporter  *replay.Exporter
	Monitor         *monitor.Store
	mu              sync.Mutex
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

	exportSvc := &export.Service{
		Media:    m,
		Store:    s,
		StaticFS: staticFS,
	}

	a := &API{
		Store:    s,
		Media:    m,
		Export:   exportSvc,
		Conf:     conf,
		AI:       aiClient,
		Password: NewPasswordManager(),
	}

	// Initialize sync scheduler
	syncFunc := func() error {
		_, _, err := decrypt.RunTask(conf.WechatDbSrcPath, conf.WechatDbKey)
		if err != nil {
			return err
		}
		return s.Reload()
	}
	a.SyncScheduler = intsync.NewScheduler(syncFunc)

	// Restore sync config from viper
	if viper.GetBool("SYNC_ENABLED") {
		interval := viper.GetInt("SYNC_INTERVAL_MINUTES")
		if interval < 5 {
			interval = 30
		}
		a.SyncScheduler.Configure(true, interval)
	}

	// Initialize backup scheduler
	backupFunc := a.createBackupFunc(exportSvc)
	historyFile := filepath.Join(viper.GetString("config_path"), "backup_history.json")
	if historyFile == "backup_history.json" {
		home, _ := os.UserHomeDir()
		historyFile = filepath.Join(home, ".wetrace", "backup_history.json")
	}
	a.BackupScheduler = backup.NewScheduler(backupFunc, historyFile)

	// Restore backup config from viper
	if viper.GetBool("BACKUP_ENABLED") {
		hours := viper.GetInt("BACKUP_INTERVAL_HOURS")
		if hours < 1 {
			hours = 24
		}
		bPath := viper.GetString("BACKUP_PATH")
		bFormat := viper.GetString("BACKUP_FORMAT")
		if bFormat == "" {
			bFormat = "html"
		}
		a.BackupScheduler.Configure(true, hours, bPath, bFormat)
	}

	// Initialize replay exporter
	a.ReplayExporter = replay.NewExporter(s)

	// Initialize monitor store
	monitorDir := filepath.Join(viper.GetString("config_path"), "monitor")
	if monitorDir == "monitor" {
		home, _ := os.UserHomeDir()
		monitorDir = filepath.Join(home, ".wetrace")
	}
	monitorStore, err := monitor.NewStore(monitorDir)
	if err == nil {
		a.Monitor = monitorStore
	}

	return a
}

// createBackupFunc creates the backup function that exports all sessions.
func (a *API) createBackupFunc(exportSvc *export.Service) backup.BackupFunc {
	return func(backupPath, format string) (string, int, error) {
		ctx := context.Background()

		sessions, err := a.Store.GetSessions(ctx, types.SessionQuery{Limit: 100000})
		if err != nil {
			return "", 0, fmt.Errorf("获取会话列表失败: %w", err)
		}

		if err := os.MkdirAll(backupPath, 0755); err != nil {
			return "", 0, fmt.Errorf("创建备份目录失败: %w", err)
		}

		timestamp := time.Now().Format("20060102_150405")
		outputFile := filepath.Join(backupPath, fmt.Sprintf("backup_%s.zip", timestamp))

		count := 0
		for _, sess := range sessions {
			talker := sess.UserName
			name := sess.NickName
			if name == "" {
				name = talker
			}

			var data []byte
			switch format {
			case "txt":
				data, err = exportSvc.ExportChatTxt(ctx, talker, name, time.Time{}, time.Time{})
			default:
				data, err = exportSvc.ExportChat(ctx, talker, name, time.Time{}, time.Time{})
			}
			if err != nil {
				continue
			}

			ext := ".zip"
			if format == "txt" {
				ext = ".txt"
			}
			fname := filepath.Join(backupPath, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
			if writeErr := os.WriteFile(fname, data, 0644); writeErr != nil {
				continue
			}
			count++
		}

		// Write a summary marker file as the "output"
		summary := fmt.Sprintf("Backup completed at %s, %d sessions exported", timestamp, count)
		_ = os.WriteFile(outputFile+".txt", []byte(summary), 0644)

		return outputFile + ".txt", count, nil
	}
}
