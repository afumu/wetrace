package backup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// BackupFunc is the function called to perform a backup operation.
// It receives the backup path and format, returns the output file path.
type BackupFunc func(backupPath, format string) (string, int, error)

// Record represents a single backup history entry.
type Record struct {
	ID            string `json:"id"`
	Time          string `json:"time"`
	Status        string `json:"status"`
	FilePath      string `json:"file_path"`
	FileSize      int64  `json:"file_size"`
	SessionsCount int    `json:"sessions_count"`
	ErrorMsg      string `json:"error_msg,omitempty"`
}

// Scheduler manages automatic backup scheduling.
type Scheduler struct {
	mu               sync.Mutex
	enabled          bool
	intervalHours    int
	backupPath       string
	format           string
	lastBackupTime   time.Time
	lastBackupStatus string
	isRunning        bool
	backupFunc       BackupFunc
	ticker           *time.Ticker
	stopCh           chan struct{}
	history          []Record
	historyFile      string
}

// NewScheduler creates a new backup scheduler.
func NewScheduler(backupFunc BackupFunc, historyFile string) *Scheduler {
	s := &Scheduler{
		backupFunc:  backupFunc,
		intervalHours: 24,
		format:      "html",
		historyFile: historyFile,
	}
	s.loadHistory()
	return s
}

// Status represents the current backup configuration and state.
type Status struct {
	Enabled          bool   `json:"enabled"`
	IntervalHours    int    `json:"interval_hours"`
	BackupPath       string `json:"backup_path"`
	Format           string `json:"format"`
	LastBackupTime   string `json:"last_backup_time"`
	LastBackupStatus string `json:"last_backup_status"`
}

// GetStatus returns the current backup scheduler status.
func (s *Scheduler) GetStatus() Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastTime := ""
	if !s.lastBackupTime.IsZero() {
		lastTime = s.lastBackupTime.Format(time.RFC3339)
	}
	return Status{
		Enabled:          s.enabled,
		IntervalHours:    s.intervalHours,
		BackupPath:       s.backupPath,
		Format:           s.format,
		LastBackupTime:   lastTime,
		LastBackupStatus: s.lastBackupStatus,
	}
}

// Configure updates the backup scheduler settings.
func (s *Scheduler) Configure(enabled bool, intervalHours int, backupPath, format string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.enabled = enabled
	if intervalHours >= 1 {
		s.intervalHours = intervalHours
	}
	if backupPath != "" {
		s.backupPath = backupPath
	}
	if format != "" {
		s.format = format
	}

	s.stopTicker()
	if s.enabled {
		s.startTicker()
	}
}

func (s *Scheduler) stopTicker() {
	if s.ticker != nil {
		s.ticker.Stop()
		close(s.stopCh)
		s.ticker = nil
		s.stopCh = nil
	}
}

func (s *Scheduler) startTicker() {
	s.ticker = time.NewTicker(time.Duration(s.intervalHours) * time.Hour)
	s.stopCh = make(chan struct{})

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.RunBackup()
			case <-s.stopCh:
				return
			}
		}
	}()
	log.Info().Int("interval_hours", s.intervalHours).Msg("auto backup scheduler started")
}

// RunBackup executes a backup operation. Safe to call concurrently.
func (s *Scheduler) RunBackup() {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	backupPath := s.backupPath
	format := s.format
	s.mu.Unlock()

	if backupPath == "" {
		s.mu.Lock()
		s.isRunning = false
		s.lastBackupStatus = "failed"
		s.lastBackupTime = time.Now()
		s.mu.Unlock()
		log.Error().Msg("backup path not configured")
		return
	}

	log.Info().Str("path", backupPath).Msg("backup started")
	now := time.Now()
	backupID := "backup_" + now.Format("20060102_150405")

	filePath, sessionsCount, err := s.backupFunc(backupPath, format)

	record := Record{
		ID:            backupID,
		Time:          now.Format(time.RFC3339),
		SessionsCount: sessionsCount,
	}

	s.mu.Lock()
	s.isRunning = false
	s.lastBackupTime = now
	if err != nil {
		s.lastBackupStatus = "failed"
		record.Status = "failed"
		record.ErrorMsg = err.Error()
		log.Error().Err(err).Msg("backup failed")
	} else {
		s.lastBackupStatus = "success"
		record.Status = "success"
		record.FilePath = filePath
		if info, statErr := os.Stat(filePath); statErr == nil {
			record.FileSize = info.Size()
		}
		log.Info().Str("file", filePath).Msg("backup completed")
	}
	s.history = append([]Record{record}, s.history...)
	s.mu.Unlock()

	s.saveHistory()
}

// GetHistory returns backup history records with pagination.
func (s *Scheduler) GetHistory(limit, offset int) []Record {
	s.mu.Lock()
	defer s.mu.Unlock()

	if offset >= len(s.history) {
		return []Record{}
	}
	end := offset + limit
	if end > len(s.history) {
		end = len(s.history)
	}
	return s.history[offset:end]
}

// Stop shuts down the scheduler.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopTicker()
}

func (s *Scheduler) loadHistory() {
	if s.historyFile == "" {
		return
	}
	data, err := os.ReadFile(s.historyFile)
	if err != nil {
		return
	}
	var records []Record
	if err := json.Unmarshal(data, &records); err != nil {
		return
	}
	s.history = records
}

func (s *Scheduler) saveHistory() {
	s.mu.Lock()
	records := make([]Record, len(s.history))
	copy(records, s.history)
	s.mu.Unlock()

	if s.historyFile == "" {
		return
	}
	dir := filepath.Dir(s.historyFile)
	_ = os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal backup history")
		return
	}
	if err := os.WriteFile(s.historyFile, data, 0644); err != nil {
		log.Error().Err(err).Msg("failed to save backup history")
	}
}
