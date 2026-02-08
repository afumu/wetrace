package core

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// ConnectionPool 负责管理 SQLite 数据库连接的生命周期。
// 它保证同一个文件只会被打开一次，并且是线程安全的。
type ConnectionPool struct {
	mu      sync.RWMutex
	connMap map[string]*sql.DB // 路径 -> 连接对象
	baseDir string             // 基础工作目录
}

// NewConnectionPool 创建一个新的连接池
func NewConnectionPool(baseDir string) *ConnectionPool {
	return &ConnectionPool{
		connMap: make(map[string]*sql.DB),
		baseDir: baseDir,
	}
}

// GetConnection 获取指定路径的数据库连接。
// 如果连接已存在且活跃，直接返回；否则创建新连接。
func (p *ConnectionPool) GetConnection(path string) (*sql.DB, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 1. 尝试从缓存获取
	if conn, ok := p.connMap[path]; ok {
		if err := conn.Ping(); err == nil {
			return conn, nil
		}
		// 连接失效，清理旧连接
		_ = conn.Close()
		delete(p.connMap, path)
	}

	// 2. 建立新连接
	conn, err := p.openNewConnection(path)
	if err != nil {
		return nil, err
	}

	p.connMap[path] = conn
	return conn, nil
}

// openNewConnection 封装底层的 SQL 打开逻辑 (单一职责：创建)
func (p *ConnectionPool) openNewConnection(path string) (*sql.DB, error) {
	// 使用读写模式 (mode=rw) 和共享缓存 (cache=shared)
	dsn := fmt.Sprintf("file:%s?mode=rw&cache=shared", path)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("无法打开数据库文件 %s: %w", path, err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("数据库文件 %s 连接测试失败: %w", path, err)
	}

	return db, nil
}

// CloseConnection 关闭并移除特定路径的连接
func (p *ConnectionPool) CloseConnection(path string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if conn, ok := p.connMap[path]; ok {
		err := conn.Close()
		delete(p.connMap, path)
		return err
	}
	return nil
}

// CloseAll 关闭池中所有连接
func (p *ConnectionPool) CloseAll() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error
	for path, conn := range p.connMap {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭 %s 失败: %w", path, err))
		}
	}
	// 重置 map
	p.connMap = make(map[string]*sql.DB)

	if len(errs) > 0 {
		return fmt.Errorf("关闭连接池时出现错误: %v", errs)
	}
	return nil
}
