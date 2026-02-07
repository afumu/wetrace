package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/types"
)

func (r *Repository) GetChatRooms(ctx context.Context, q types.ChatRoomQuery) ([]*model.ChatRoom, error) {

	dbPath, err := r.router.GetContactDBPath()

	if err != nil {

		return nil, err

	}

	db, err := r.pool.GetConnection(dbPath)

	if err != nil {

		return nil, err

	}

	// 1. 尝试 V4 模式

	var exists bool

	_ = db.QueryRowContext(ctx, "SELECT 1 FROM sqlite_master WHERE type='table' AND name='chat_room'").Scan(&exists)

	if exists {

		return r.queryV4ChatRooms(ctx, db, q)

	}

	// 2. 尝试 V3 模式

	return r.queryV3ChatRooms(ctx, db, q)

}

func (r *Repository) queryV4ChatRooms(ctx context.Context, db *sql.DB, q types.ChatRoomQuery) ([]*model.ChatRoom, error) {

	var query string

	var args []interface{}

	if q.Keyword != "" {

		query = `SELECT username, owner, ext_buffer FROM chat_room WHERE username = ?`

		args = []interface{}{q.Keyword}

	} else {

		query = `SELECT username, owner, ext_buffer FROM chat_room ORDER BY username`

	}

	if q.Limit > 0 {

		query += fmt.Sprintf(" LIMIT %d", q.Limit)

		if q.Offset > 0 {

			query += fmt.Sprintf(" OFFSET %d", q.Offset)

		}

	}

	rows, err := db.QueryContext(ctx, query, args...)

	if err != nil {

		return nil, err

	}

	defer rows.Close()

	var rooms []*model.ChatRoom

	for rows.Next() {

		var c model.ChatRoomV4

		if err := rows.Scan(&c.UserName, &c.Owner, &c.ExtBuffer); err != nil {

			return nil, err

		}

		rooms = append(rooms, c.Wrap())

	}

	// 处理后备逻辑 (单房查找)

	if q.Keyword != "" && len(rooms) == 0 {

		contacts, err := r.GetContacts(ctx, types.ContactQuery{Keyword: q.Keyword, Limit: 1})

		if err == nil && len(contacts) > 0 && strings.HasSuffix(contacts[0].UserName, "@chatroom") {

			userName := contacts[0].UserName

			rows2, err := db.QueryContext(ctx, `SELECT username, owner, ext_buffer FROM chat_room WHERE username = ?`, userName)

			if err == nil {

				defer rows2.Close()

				for rows2.Next() {

					var c model.ChatRoomV4

					rows2.Scan(&c.UserName, &c.Owner, &c.ExtBuffer)

					rooms = append(rooms, c.Wrap())

				}

			}

			if len(rooms) == 0 {

				rooms = append(rooms, &model.ChatRoom{

					Name: userName,

					Users: []model.ChatRoomUser{},

					User2DisplayName: make(map[string]string),
				})

			}

		}

	}

	return rooms, nil

}

func (r *Repository) queryV3ChatRooms(ctx context.Context, db *sql.DB, q types.ChatRoomQuery) ([]*model.ChatRoom, error) {

	var query string

	var args []interface{}

	if q.Keyword != "" {

		query = `SELECT ChatRoomName, Reserved2, RoomData FROM ChatRoom WHERE ChatRoomName = ?`

		args = []interface{}{q.Keyword}

	} else {

		query = `SELECT ChatRoomName, Reserved2, RoomData FROM ChatRoom ORDER BY ChatRoomName`

	}

	if q.Limit > 0 {

		query += fmt.Sprintf(" LIMIT %d", q.Limit)

		if q.Offset > 0 {

			query += fmt.Sprintf(" OFFSET %d", q.Offset)

		}

	}

	rows, err := db.QueryContext(ctx, query, args...)

	if err != nil {

		return nil, err

	}

	defer rows.Close()

	var rooms []*model.ChatRoom

	for rows.Next() {

		var c model.ChatRoomV3

		if err := rows.Scan(&c.ChatRoomName, &c.Reserved2, &c.RoomData); err != nil {

			return nil, err

		}

		rooms = append(rooms, c.Wrap())

	}

	// 单个房间查找的后备逻辑

	if q.Keyword != "" && len(rooms) == 0 {

		// 检查联系人表

		contacts, err := r.GetContacts(ctx, types.ContactQuery{Keyword: q.Keyword, Limit: 1})

		if err == nil && len(contacts) > 0 && strings.HasSuffix(contacts[0].UserName, "@chatroom") {

			userName := contacts[0].UserName

			rows2, err := db.QueryContext(ctx, `SELECT ChatRoomName, Reserved2, RoomData FROM ChatRoom WHERE ChatRoomName = ?`, userName)

			if err == nil {

				defer rows2.Close()

				for rows2.Next() {

					var c model.ChatRoomV3

					rows2.Scan(&c.ChatRoomName, &c.Reserved2, &c.RoomData)

					rooms = append(rooms, c.Wrap())

				}

			}

			// 如果仍然为空，创建假的

			if len(rooms) == 0 {

				rooms = append(rooms, &model.ChatRoom{

					Name: userName,

					Users: []model.ChatRoomUser{},

					User2DisplayName: make(map[string]string),
				})

			}

		}

	}

	return rooms, nil

}
