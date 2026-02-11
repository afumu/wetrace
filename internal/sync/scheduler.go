package sync

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// SyncFunc is the function called to perform a sync operation.
type SyncFunc func() error

// Scheduler manages automatic sync scheduling.
type Scheduler struct {
	mu             sync.Mutex
	enabled        bool
	intervalMin    int
	lastSyncTime   time.Time
	lastSyncStatus string
	isSyncing      bool
	syncFunc       SyncFunc
	ticker         *time.Ticker
	stopCh         chan struct{}
}

// NewScheduler creates a new sync scheduler.
func NewScheduler(syncFunc SyncFunc) *Scheduler {
	return &Scheduler{
		syncFunc:       syncFunc,
		intervalMin:    30,
		lastSyncStatus: "",
	}
}

// Status returns the current sync status.
type Status struct {
	Enabled        bool   `json:"enabled"`
	IntervalMin    int    `json:"interval_minutes"`
	LastSyncTime   string `json:"last_sync_time"`
	LastSyncStatus string `json:"last_sync_status"`
	IsSyncing      bool   `json:"is_syncing"`
}

// GetStatus returns the current scheduler status.
func (s *Scheduler) GetStatus() Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastTime := ""
	if !s.lastSyncTime.IsZero() {
		lastTime = s.lastSyncTime.Format(time.RFC3339)
	}
	return Status{
		Enabled:        s.enabled,
		IntervalMin:    s.intervalMin,
		LastSyncTime:   lastTime,
		LastSyncStatus: s.lastSyncStatus,
		IsSyncing:      s.isSyncing,
	}
}

// Configure updates the scheduler settings and restarts if needed.
func (s *Scheduler) Configure(enabled bool, intervalMin int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.enabled = enabled
	if intervalMin >= 5 && intervalMin <= 1440 {
		s.intervalMin = intervalMin
	}

	// Stop existing ticker
	s.stopTicker()

	// Start new ticker if enabled
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
	s.ticker = time.NewTicker(time.Duration(s.intervalMin) * time.Minute)
	s.stopCh = make(chan struct{})

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.RunSync()
			case <-s.stopCh:
				return
			}
		}
	}()
	log.Info().Int("interval_minutes", s.intervalMin).Msg("auto sync scheduler started")
}

// StartSync marks the scheduler as syncing and returns true if successful.
// Call this before launching RunSync in a goroutine to avoid race conditions
// where a status poll arrives before RunSync has set isSyncing = true.
func (s *Scheduler) StartSync() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isSyncing {
		return false
	}
	s.isSyncing = true
	return true
}

// RunSync executes a sync operation. Safe to call concurrently.
func (s *Scheduler) RunSync() {
	s.mu.Lock()
	if !s.isSyncing {
		// Not pre-started via StartSync, set it now
		s.isSyncing = true
	}
	s.mu.Unlock()

	log.Info().Msg("sync started")
	err := s.syncFunc()

	s.mu.Lock()
	s.isSyncing = false
	s.lastSyncTime = time.Now()
	if err != nil {
		s.lastSyncStatus = "failed"
		log.Error().Err(err).Msg("sync failed")
	} else {
		s.lastSyncStatus = "success"
		log.Info().Msg("sync completed")
	}
	s.mu.Unlock()
}

// Stop shuts down the scheduler.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopTicker()
}
