package bind

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/afumu/wetrace/store/core"
	"github.com/afumu/wetrace/store/strategy"

	_ "github.com/mattn/go-sqlite3"
)

func TestTimelineRouter_Resolve(t *testing.T) {
	// 1. 设置环境
	tmpDir := t.TempDir()
	pool := core.NewConnectionPool(tmpDir)
	defer pool.CloseAll()
	strat := strategy.NewV4()

	// 2. 准备数据
	t1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)

	createMockDB(t, filepath.Join(tmpDir, "message_0.db"), t1, "alice", 1)
	createMockDB(t, filepath.Join(tmpDir, "message_1.db"), t2, "alice", 2)
	createMockFile(t, filepath.Join(tmpDir, "contact.db")) // 空文件

	// 3. 初始化 Router
	router := NewTimelineRouter(tmpDir, pool, strat)
	err := router.RebuildIndex(context.Background())
	if err != nil {
		t.Fatalf("RebuildIndex 失败: %v", err)
	}

	// 4. 测试路由: 跨两个 DB 的查询
	results := router.Resolve(t1, t2.Add(time.Hour), "alice")
	if len(results) != 2 {
		t.Errorf("期望 2 个结果, 实际得到 %d", len(results))
	}

	// 5. 测试路由: 只包含第一个 DB
	results = router.Resolve(t1, t1.Add(time.Hour), "alice")
	if len(results) != 1 {
		t.Errorf("期望 1 个结果, 实际得到 %d", len(results))
	}

	// 6. 测试路由: 只包含第二个 DB
	results = router.Resolve(t2, t2.Add(time.Hour), "alice")
	if len(results) != 1 {
		t.Errorf("期望 1 个结果, 实际得到 %d", len(results))
	}

	// 7. 测试 GetContactDB
	contactPath, err := router.GetContactDBPath()
	if err != nil {
		t.Fatalf("GetContactDBPath 失败: %v", err)
	}
	if filepath.Base(contactPath) != "contact.db" {
		t.Errorf("期望 contact.db, 实际得到 %s", contactPath)
	}
}

func createMockDB(t *testing.T, path string, startTime time.Time, user string, id int) {
	// 确保文件存在
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("打开 db 失败: %v", err)
	}
	defer db.Close()

	// V4 使用 Timestamp 表
	_, err = db.Exec("CREATE TABLE Timestamp (timestamp INTEGER)")
	if err != nil {
		t.Fatalf("创建 Timestamp 失败: %v", err)
	}
	_, err = db.Exec("INSERT INTO Timestamp VALUES (?)", startTime.Unix())
	if err != nil {
		t.Fatalf("插入 Timestamp 失败: %v", err)
	}
}

func createMockFile(t *testing.T, path string) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
}
