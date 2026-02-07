package store

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/afumu/wetrace/store/types"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestDefaultStore_Integration(t *testing.T) {
	// 1. 初始化 Store
	// 注意：确保该路径下的数据库文件已解密，否则会报错 "file is not a database"
	tmpDir := "C:\\Users\\Administrator\\Documents\\chatlog\\wxid_dcn1p0w7ipne22_50ce\\db_storage"

	// 如果您想使用 Mock 数据进行测试，请取消下面两行的注释，并注释掉上面的 tmpDir 赋值
	// tmpDir := t.TempDir()
	// setupMockData(t, tmpDir)

	s, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer s.Close()

	ctx := context.Background()

	// 2. 测试获取联系人
	contacts, err := s.GetContacts(ctx, types.ContactQuery{})
	if err != nil {
		t.Fatalf("GetContacts failed: %v", err)
	}

	// 3. 将联系人转换为 JSON 并打印
	data, err := json.MarshalIndent(contacts, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	fmt.Printf("\n--- 检索到的联系人 (JSON) ---\n%s\n----------------------------\n\n", string(data))

	if len(contacts) == 0 {
		t.Log("警告: 未找到任何联系人")
	}
}

func setupMockData(t *testing.T, dir string) {
	// contact.db (Contacts, ChatRooms, Sessions)
	p := filepath.Join(dir, "contact.db")
	f, _ := os.Create(p)
	f.Close()
	db, _ := sql.Open("sqlite3", p)
	defer db.Close()

	db.Exec("CREATE TABLE contact (username TEXT, local_type INTEGER, alias TEXT, remark TEXT, nick_name TEXT)")
	db.Exec("INSERT INTO contact VALUES (?, ?, ?, ?, ?)", "user1", 0, "alias1", "remark1", "nick1")

	db.Exec("CREATE TABLE chat_room (username TEXT, owner TEXT, ext_buffer BLOB)")
	db.Exec("INSERT INTO chat_room VALUES (?, ?, ?)", "room1", "owner1", nil)

	db.Exec("CREATE TABLE SessionTable (username TEXT, summary TEXT, last_timestamp INTEGER, last_msg_sender TEXT, last_sender_display_name TEXT, sort_timestamp INTEGER)")
	db.Exec("INSERT INTO SessionTable VALUES (?, ?, ?, ?, ?, ?)", "user1", "hi", 123456, "sender1", "display1", 123456)

	// message_0.db
	p2 := filepath.Join(dir, "message_0.db")
	f2, _ := os.Create(p2)
	f2.Close()
	db2, _ := sql.Open("sqlite3", p2)
	defer db2.Close()

	db2.Exec("CREATE TABLE Timestamp (timestamp INTEGER)")
	db2.Exec("INSERT INTO Timestamp VALUES (?)", 1672531200) // 2023-01-01

	hash := md5.Sum([]byte("alice"))
	tableName := "Msg_" + hex.EncodeToString(hash[:])
	db2.Exec(fmt.Sprintf(`CREATE TABLE %s (
		sort_seq INTEGER, server_id INTEGER, local_type INTEGER, 
		real_sender_id INTEGER, create_time INTEGER, 
		message_content TEXT, packed_info_data BLOB, status INTEGER
	)`, tableName))

	db2.Exec("CREATE TABLE Name2Id (user_name TEXT)")
	db2.Exec("INSERT INTO Name2Id VALUES (?)", "alice")

	db2.Exec(fmt.Sprintf(`INSERT INTO %s VALUES (
		?, 1, 1, 1, ?, 'hello', NULL, 0
	)`, tableName), 1672531200000, 1672531200)
}
