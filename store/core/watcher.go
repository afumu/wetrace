package core

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

type Watcher struct {
	watcher   *fsnotify.Watcher
	base      string
	callbacks []func(event fsnotify.Event)
	mu        sync.RWMutex
	done      chan struct{}
}

func NewWatcher(basePath string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建 watcher 失败: %w", err)
	}

	if err := w.Add(basePath); err != nil {
		w.Close()
		return nil, fmt.Errorf("监控路径 %s 失败: %w", basePath, err)
	}

	return &Watcher{
		watcher:   w,
		base:      basePath,
		callbacks: make([]func(event fsnotify.Event), 0),
		done:      make(chan struct{}),
	}, nil
}

func (w *Watcher) Start() {
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				w.dispatch(event)
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Error().Err(err).Msg("Watcher 错误")
			case <-w.done:
				return
			}
		}
	}()
}

func (w *Watcher) Stop() error {
	close(w.done)
	return w.watcher.Close()
}

func (w *Watcher) AddCallback(cb func(event fsnotify.Event)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks = append(w.callbacks, cb)
}

func (w *Watcher) dispatch(event fsnotify.Event) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, cb := range w.callbacks {
		// 异步还是同步运行回调？
		// 同步对于排序更安全，但缓慢的回调会阻塞 watcher。
		// 目前，同步运行。
		go cb(event)
	}
}
