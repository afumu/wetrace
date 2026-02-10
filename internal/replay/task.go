package replay

import (
	"sync"
	"time"
)

// TaskStatus 导出任务状态
type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
)

// ExportTask 回放导出任务
type ExportTask struct {
	TaskID          string     `json:"task_id"`
	TalkerID        string     `json:"talker_id"`
	StartDate       string     `json:"start_date,omitempty"`
	EndDate         string     `json:"end_date,omitempty"`
	Format          string     `json:"format"`
	Speed           int        `json:"speed"`
	Resolution      string     `json:"resolution"`
	Status          TaskStatus `json:"status"`
	Progress        int        `json:"progress"`
	TotalFrames     int        `json:"total_frames"`
	ProcessedFrames int        `json:"processed_frames"`
	FilePath        string     `json:"-"`
	Error           string     `json:"error,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// TaskManager 管理异步导出任务
type TaskManager struct {
	mu    sync.RWMutex
	tasks map[string]*ExportTask
}

// NewTaskManager 创建任务管理器
func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[string]*ExportTask),
	}
}

// AddTask 添加新任务
func (tm *TaskManager) AddTask(task *ExportTask) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.tasks[task.TaskID] = task
}

// GetTask 获取任务
func (tm *TaskManager) GetTask(taskID string) *ExportTask {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.tasks[taskID]
}

// UpdateTask 更新任务状态
func (tm *TaskManager) UpdateTask(taskID string, fn func(*ExportTask)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if task, ok := tm.tasks[taskID]; ok {
		fn(task)
	}
}

// CleanExpired 清理过期任务（超过1小时的已完成/失败任务）
func (tm *TaskManager) CleanExpired() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	cutoff := time.Now().Add(-1 * time.Hour)
	for id, task := range tm.tasks {
		if (task.Status == StatusCompleted || task.Status == StatusFailed) && task.CreatedAt.Before(cutoff) {
			delete(tm.tasks, id)
		}
	}
}
