package repo

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/afumu/wetrace/store/bind"
	"github.com/afumu/wetrace/store/core"
	"github.com/afumu/wetrace/store/strategy"
	"github.com/afumu/wetrace/store/types"

	_ "github.com/mattn/go-sqlite3"
)

func TestRepo_GetContacts(t *testing.T) {
	tmpDir := t.TempDir()
	pool := core.NewConnectionPool(tmpDir)
	defer pool.CloseAll()
	strat := strategy.NewV4()

	// 创建 contact.db
	createContactDB(t, filepath.Join(tmpDir, "contact.db"))

	// 初始化 Router
	router := bind.NewTimelineRouter(tmpDir, pool, strat)

	repo := New(router, pool)

	// 测试
	contacts, err := repo.GetContacts(context.Background(), types.ContactQuery{})
	if err != nil {
		t.Fatalf("GetContacts 失败: %v", err)
	}
	if len(contacts) != 1 {
		t.Errorf("期望 1 个联系人, 实际得到 %d", len(contacts))
	}
	if contacts[0].UserName != "user1" {
		t.Errorf("期望 user1, 实际得到 %s", contacts[0].UserName)
	}
}

func TestRepo_GetMessages(t *testing.T) {
	tmpDir := t.TempDir()
	pool := core.NewConnectionPool(tmpDir)
	defer pool.CloseAll()
	strat := strategy.NewV4()

	// 创建 message_0.db
	t1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	createMessageDB(t, filepath.Join(tmpDir, "message_0.db"), t1, "alice", 1)

	// 初始化 Router
	router := bind.NewTimelineRouter(tmpDir, pool, strat)
	if err := router.RebuildIndex(context.Background()); err != nil {
		t.Fatalf("RebuildIndex 失败: %v", err)
	}

	repo := New(router, pool)

	// 测试
	msgs, err := repo.GetMessages(context.Background(), types.MessageQuery{
		StartTime: t1,
		EndTime:   t1.Add(time.Hour),
		Talker:    "alice",
	})
	if err != nil {
		t.Fatalf("GetMessages 失败: %v", err)
	}
	if len(msgs) != 1 {
		t.Errorf("期望 1 条消息, 实际得到 %d", len(msgs))
	}
	if msgs[0].Content != "hello" {
		t.Errorf("期望 'hello', 实际得到 '%s'", msgs[0].Content)
	}
}

func createContactDB(t *testing.T, path string) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE contact (username TEXT, local_type INTEGER, alias TEXT, remark TEXT, nick_name TEXT)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO contact VALUES (?, ?, ?, ?, ?)", "user1", 0, "alias1", "remark1", "nick1")
	if err != nil {
		t.Fatal(err)
	}
}

func createMessageDB(t *testing.T, path string, startTime time.Time, user string, id int) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// V4 Metadata
	db.Exec("CREATE TABLE Timestamp (timestamp INTEGER)")
	db.Exec("INSERT INTO Timestamp VALUES (?)", startTime.Unix())

	// Msg_{md5} table
	hash := md5.Sum([]byte(user))
	tableName := "Msg_" + hex.EncodeToString(hash[:])

	// sort_seq, server_id, local_type, real_sender_id, create_time, message_content, packed_info_data, status
	sqlStmt := fmt.Sprintf(`
	CREATE TABLE %s (
		sort_seq INTEGER, server_id INTEGER, local_type INTEGER, 
		real_sender_id INTEGER, create_time INTEGER, 
		message_content TEXT, packed_info_data BLOB, status INTEGER
	)`, tableName)
	_, err = db.Exec(sqlStmt)
	if err != nil {
		t.Fatalf("创建 %s 失败: %v", tableName, err)
	}

	// Name2Id table for sender resolution
	db.Exec("CREATE TABLE Name2Id (user_name TEXT)")
	db.Exec("INSERT INTO Name2Id VALUES (?)", user)

	// 插入消息
	_, err = db.Exec(fmt.Sprintf(`INSERT INTO %s VALUES (
		?, 1, 1, 1, ?, 'hello', NULL, 0
	)`, tableName), startTime.Unix()*1000, startTime.Unix())
	if err != nil {
		t.Fatalf("插入消息失败: %v", err)
	}
}
