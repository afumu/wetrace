package core

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestConnectionPool(t *testing.T) {
	// 1. 设置测试环境
	tmpDir, err := os.MkdirTemp("", "wetrace_pool_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 2. 创建一个虚拟的 SQLite 文件
	dbPath := filepath.Join(tmpDir, "test.db")
	setupDB(t, dbPath)

	// 3. 初始化连接池
	pool := NewConnectionPool(tmpDir)
	defer pool.CloseAll()

	// 4. 测试: 获取连接
	db, err := pool.GetConnection(dbPath)
	if err != nil {
		t.Fatalf("GetConnection 失败: %v", err)
	}

	// 5. 测试: 执行查询
	var result string
	err = db.QueryRow("SELECT content FROM test_table LIMIT 1").Scan(&result)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	if result != "hello world" {
		t.Errorf("期望 'hello world', 实际得到 '%s'", result)
	}

	// 6. 测试: 连接复用
	db2, err := pool.GetConnection(dbPath)
	if err != nil {
		t.Fatalf("获取缓存连接失败: %v", err)
	}
	if db != db2 {
		t.Error("对于相同的路径，连接池应该返回相同的实例")
	}

	// 7. 测试: 关闭连接
	err = pool.CloseConnection(dbPath)
	if err != nil {
		t.Fatalf("CloseConnection 失败: %v", err)
	}

	// 验证已关闭
	if err := db.Ping(); err == nil {
		t.Error("数据库连接应该已关闭")
	}
}

// setupDB 创建一个包含数据的真实 SQLite 文件
func setupDB(t *testing.T, path string) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("setupDB 打开失败: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY, content TEXT)")
	if err != nil {
		t.Fatalf("setupDB 建表失败: %v", err)
	}

	_, err = db.Exec("INSERT INTO test_table (content) VALUES (?)", "hello world")
	if err != nil {
		t.Fatalf("setupDB 插入失败: %v", err)
	}
}
