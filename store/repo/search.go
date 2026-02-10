package repo

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/bind"
	"github.com/afumu/wetrace/store/types"
	"github.com/rs/zerolog/log"
)

// SearchMessages 高级搜索（带总数统计）
func (r *Repository) SearchMessages(ctx context.Context, q types.MessageQuery) (*model.SearchResult, error) {
	var allMessages []*model.Message
	var talkerMD5Map map[string]string

	for _, shard := range r.router.GetShards() {
		db, err := r.pool.GetConnection(shard.FilePath)
		if err != nil {
			continue
		}

		if r.isTableExist(db, "MSG") {
			msgs, _ := r.searchV3Advanced(ctx, db, q)
			allMessages = append(allMessages, msgs...)
		} else {
			if talkerMD5Map == nil {
				talkerMD5Map = r.getTalkerMD5Map(ctx)
			}
			msgs, _ := r.searchV4Advanced(ctx, db, q, talkerMD5Map)
			allMessages = append(allMessages, msgs...)
		}
	}

	// 丰富信息
	if len(allMessages) > 0 {
		r.enrichMessages(ctx, allMessages)
	}

	// 排序：按时间降序
	sort.Slice(allMessages, func(i, j int) bool {
		return allMessages[i].Seq > allMessages[j].Seq
	})

	total := len(allMessages)

	// 分页
	start := q.Offset
	if start > total {
		start = total
	}
	end := start + q.Limit
	if q.Limit == 0 || end > total {
		end = total
	}
	paged := allMessages[start:end]

	// 构建搜索结果
	items := make([]*model.SearchItem, 0, len(paged))
	for _, msg := range paged {
		items = append(items, &model.SearchItem{
			Message: msg,
		})
	}

	return &model.SearchResult{
		Total: total,
		Items: items,
	}, nil
}

// searchV3Advanced V3 高级搜索
func (r *Repository) searchV3Advanced(ctx context.Context, db *sql.DB, q types.MessageQuery) ([]*model.Message, error) {
	var sb strings.Builder
	var args []interface{}

	sb.WriteString("SELECT MsgSvrID, Sequence, CreateTime, StrTalker, IsSender, Type, SubType, StrContent, CompressContent, BytesExtra FROM MSG WHERE StrContent LIKE ?")
	args = append(args, "%"+q.Keyword+"%")

	if q.Talker != "" {
		talkers := strings.Split(q.Talker, ",")
		placeholders := make([]string, len(talkers))
		for i, t := range talkers {
			placeholders[i] = "?"
			args = append(args, strings.TrimSpace(t))
		}
		sb.WriteString(" AND StrTalker IN (" + strings.Join(placeholders, ",") + ")")
	}

	if !q.StartTime.IsZero() {
		sb.WriteString(" AND CreateTime >= ?")
		args = append(args, q.StartTime.Unix()*1000)
	}
	if !q.EndTime.IsZero() {
		sb.WriteString(" AND CreateTime <= ?")
		args = append(args, q.EndTime.Unix()*1000)
	}

	if q.MsgType > 0 {
		sb.WriteString(" AND Type = ?")
		args = append(args, q.MsgType)
	}

	sb.WriteString(" ORDER BY CreateTime DESC LIMIT 1000")

	rows, err := db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*model.Message
	for rows.Next() {
		var msg model.MessageV3
		var compressContent, bytesExtra []byte
		if err := rows.Scan(&msg.MsgSvrID, &msg.Sequence, &msg.CreateTime, &msg.StrTalker, &msg.IsSender, &msg.Type, &msg.SubType, &msg.StrContent, &compressContent, &bytesExtra); err != nil {
			continue
		}
		msg.CompressContent = compressContent
		msg.BytesExtra = bytesExtra
		wrapped := msg.Wrap()

		if q.Sender != "" {
			senders := strings.Split(q.Sender, ",")
			matched := false
			for _, s := range senders {
				if strings.TrimSpace(s) == wrapped.Sender {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		msgs = append(msgs, wrapped)
	}
	return msgs, nil
}

// searchV4Advanced V4 高级搜索
func (r *Repository) searchV4Advanced(ctx context.Context, db *sql.DB, q types.MessageQuery, talkerMD5Map map[string]string) ([]*model.Message, error) {
	tables, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'Msg_%%'")
	if err != nil {
		return nil, err
	}
	defer tables.Close()

	var msgs []*model.Message
	keywordParam := "%" + q.Keyword + "%"

	// 如果指定了 talker，只搜索对应的表
	talkerFilter := make(map[string]bool)
	if q.Talker != "" {
		for _, t := range strings.Split(q.Talker, ",") {
			t = strings.TrimSpace(t)
			h := md5.Sum([]byte(t))
			talkerFilter["Msg_"+hex.EncodeToString(h[:])] = true
		}
	}

	for tables.Next() {
		var tableName string
		tables.Scan(&tableName)

		if len(talkerFilter) > 0 && !talkerFilter[tableName] {
			continue
		}

		var sb strings.Builder
		var args []interface{}

		sb.WriteString(fmt.Sprintf(`
			SELECT m.sort_seq, m.server_id, m.local_type, n.user_name, m.create_time, m.message_content, m.packed_info_data, m.status
			FROM %s m
			LEFT JOIN Name2Id n ON m.real_sender_id = n.rowid
			WHERE m.message_content LIKE ?`, tableName))
		args = append(args, keywordParam)

		if !q.StartTime.IsZero() {
			sb.WriteString(" AND m.create_time >= ?")
			args = append(args, q.StartTime.Unix())
		}
		if !q.EndTime.IsZero() {
			sb.WriteString(" AND m.create_time <= ?")
			args = append(args, q.EndTime.Unix())
		}
		if q.MsgType > 0 {
			sb.WriteString(" AND m.local_type = ?")
			args = append(args, q.MsgType)
		}

		sb.WriteString(" ORDER BY m.create_time DESC LIMIT 1000")

		rows, err := db.QueryContext(ctx, sb.String(), args...)
		if err != nil {
			continue
		}

		talker := "unknown"
		if strings.HasPrefix(tableName, "Msg_") {
			md5Hash := strings.TrimPrefix(tableName, "Msg_")
			if t, ok := talkerMD5Map[md5Hash]; ok {
				talker = t
			}
		}

		for rows.Next() {
			var msg model.MessageV4
			if err := rows.Scan(&msg.SortSeq, &msg.ServerID, &msg.LocalType, &msg.UserName, &msg.CreateTime, &msg.MessageContent, &msg.PackedInfoData, &msg.Status); err != nil {
				continue
			}
			wrapped := msg.Wrap(talker)

			if q.Sender != "" {
				senders := strings.Split(q.Sender, ",")
				matched := false
				for _, s := range senders {
					if strings.TrimSpace(s) == wrapped.Sender {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}

			msgs = append(msgs, wrapped)
		}
		rows.Close()
	}
	return msgs, nil
}

// GetMessageContext 获取某条消息前后的上下文消息
func (r *Repository) GetMessageContext(ctx context.Context, talker string, seq int64, before, after int) ([]*model.Message, error) {
	targets := r.router.Resolve(time.Unix(0, 0), time.Now(), talker)
	if len(targets) == 0 {
		return []*model.Message{}, nil
	}

	var allMessages []*model.Message

	for _, target := range targets {
		db, err := r.pool.GetConnection(target.FilePath)
		if err != nil {
			continue
		}

		hash := md5.Sum([]byte(talker))
		tableName := "Msg_" + hex.EncodeToString(hash[:])

		if r.isTableExist(db, tableName) {
			msgs, err := r.queryV4Context(ctx, db, tableName, talker, seq, before, after)
			if err != nil {
				log.Warn().Err(err).Msg("查询V4上下文失败")
				continue
			}
			allMessages = append(allMessages, msgs...)
		} else {
			msgs, err := r.queryV3Context(ctx, db, target, seq, before, after)
			if err != nil {
				log.Warn().Err(err).Msg("查询V3上下文失败")
				continue
			}
			allMessages = append(allMessages, msgs...)
		}
	}

	// 丰富信息
	if len(allMessages) > 0 {
		r.enrichMessages(ctx, allMessages)
	}

	// 排序
	sort.Slice(allMessages, func(i, j int) bool {
		return allMessages[i].Seq < allMessages[j].Seq
	})

	return allMessages, nil
}

func (r *Repository) queryV4Context(ctx context.Context, db *sql.DB, tableName, talker string, seq int64, before, after int) ([]*model.Message, error) {
	var msgs []*model.Message

	// 查询锚点之前的消息
	if before > 0 {
		query := fmt.Sprintf(`
			SELECT m.sort_seq, m.server_id, m.local_type, n.user_name, m.create_time, m.message_content, m.packed_info_data, m.status
			FROM %s m
			LEFT JOIN Name2Id n ON m.real_sender_id = n.rowid
			WHERE m.sort_seq < ?
			ORDER BY m.sort_seq DESC LIMIT ?`, tableName)
		rows, err := db.QueryContext(ctx, query, seq, before)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var msg model.MessageV4
			rows.Scan(&msg.SortSeq, &msg.ServerID, &msg.LocalType, &msg.UserName, &msg.CreateTime, &msg.MessageContent, &msg.PackedInfoData, &msg.Status)
			msgs = append(msgs, msg.Wrap(talker))
		}
		rows.Close()
	}

	// 查询锚点消息本身
	queryAnchor := fmt.Sprintf(`
		SELECT m.sort_seq, m.server_id, m.local_type, n.user_name, m.create_time, m.message_content, m.packed_info_data, m.status
		FROM %s m
		LEFT JOIN Name2Id n ON m.real_sender_id = n.rowid
		WHERE m.sort_seq = ?`, tableName)
	row := db.QueryRowContext(ctx, queryAnchor, seq)
	var anchorMsg model.MessageV4
	if err := row.Scan(&anchorMsg.SortSeq, &anchorMsg.ServerID, &anchorMsg.LocalType, &anchorMsg.UserName, &anchorMsg.CreateTime, &anchorMsg.MessageContent, &anchorMsg.PackedInfoData, &anchorMsg.Status); err == nil {
		msgs = append(msgs, anchorMsg.Wrap(talker))
	}

	// 查询锚点之后的消息
	if after > 0 {
		query := fmt.Sprintf(`
			SELECT m.sort_seq, m.server_id, m.local_type, n.user_name, m.create_time, m.message_content, m.packed_info_data, m.status
			FROM %s m
			LEFT JOIN Name2Id n ON m.real_sender_id = n.rowid
			WHERE m.sort_seq > ?
			ORDER BY m.sort_seq ASC LIMIT ?`, tableName)
		rows, err := db.QueryContext(ctx, query, seq, after)
		if err != nil {
			return msgs, nil
		}
		for rows.Next() {
			var msg model.MessageV4
			rows.Scan(&msg.SortSeq, &msg.ServerID, &msg.LocalType, &msg.UserName, &msg.CreateTime, &msg.MessageContent, &msg.PackedInfoData, &msg.Status)
			msgs = append(msgs, msg.Wrap(talker))
		}
		rows.Close()
	}

	return msgs, nil
}

func (r *Repository) queryV3Context(ctx context.Context, db *sql.DB, target bind.RouteResult, seq int64, before, after int) ([]*model.Message, error) {
	var msgs []*model.Message

	talkerCond := "StrTalker = ?"
	talkerArg := interface{}(target.Talker)
	if target.TalkerID != 0 {
		talkerCond = "TalkerId = ?"
		talkerArg = target.TalkerID
	}

	// 查询锚点之前的消息
	if before > 0 {
		query := fmt.Sprintf("SELECT MsgSvrID, Sequence, CreateTime, StrTalker, IsSender, Type, SubType, StrContent, CompressContent, BytesExtra FROM MSG WHERE %s AND Sequence < ? ORDER BY Sequence DESC LIMIT ?", talkerCond)
		rows, err := db.QueryContext(ctx, query, talkerArg, seq, before)
		if err == nil {
			for rows.Next() {
				var msg model.MessageV3
				var compressContent, bytesExtra []byte
				rows.Scan(&msg.MsgSvrID, &msg.Sequence, &msg.CreateTime, &msg.StrTalker, &msg.IsSender, &msg.Type, &msg.SubType, &msg.StrContent, &compressContent, &bytesExtra)
				msg.CompressContent = compressContent
				msg.BytesExtra = bytesExtra
				msgs = append(msgs, msg.Wrap())
			}
			rows.Close()
		}
	}

	// 查询锚点消息
	queryAnchor := fmt.Sprintf("SELECT MsgSvrID, Sequence, CreateTime, StrTalker, IsSender, Type, SubType, StrContent, CompressContent, BytesExtra FROM MSG WHERE %s AND Sequence = ?", talkerCond)
	row := db.QueryRowContext(ctx, queryAnchor, talkerArg, seq)
	var anchorMsg model.MessageV3
	var cc, be []byte
	if err := row.Scan(&anchorMsg.MsgSvrID, &anchorMsg.Sequence, &anchorMsg.CreateTime, &anchorMsg.StrTalker, &anchorMsg.IsSender, &anchorMsg.Type, &anchorMsg.SubType, &anchorMsg.StrContent, &cc, &be); err == nil {
		anchorMsg.CompressContent = cc
		anchorMsg.BytesExtra = be
		msgs = append(msgs, anchorMsg.Wrap())
	}

	// 查询锚点之后的消息
	if after > 0 {
		query := fmt.Sprintf("SELECT MsgSvrID, Sequence, CreateTime, StrTalker, IsSender, Type, SubType, StrContent, CompressContent, BytesExtra FROM MSG WHERE %s AND Sequence > ? ORDER BY Sequence ASC LIMIT ?", talkerCond)
		rows, err := db.QueryContext(ctx, query, talkerArg, seq, after)
		if err == nil {
			for rows.Next() {
				var msg model.MessageV3
				var compressContent, bytesExtra []byte
				rows.Scan(&msg.MsgSvrID, &msg.Sequence, &msg.CreateTime, &msg.StrTalker, &msg.IsSender, &msg.Type, &msg.SubType, &msg.StrContent, &compressContent, &bytesExtra)
				msg.CompressContent = compressContent
				msg.BytesExtra = bytesExtra
				msgs = append(msgs, msg.Wrap())
			}
			rows.Close()
		}
	}

	return msgs, nil
}
