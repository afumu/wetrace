package repo

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/bind"
	"github.com/afumu/wetrace/store/types"
	"github.com/rs/zerolog/log"
)

// GetMessages 获取符合条件的消息列表
// 步骤：路由 -> 多分片查询 -> 聚合 -> 排序 -> 分页 -> 填充信息
func (r *Repository) GetMessages(ctx context.Context, q types.MessageQuery) ([]*model.Message, error) {
	// 1. 路由：确定目标分片
	targets := r.router.Resolve(q.StartTime, q.EndTime, q.Talker)
	if len(targets) == 0 {
		return []*model.Message{}, nil
	}

	// 2. 查询：从所有分片获取原始数据
	allMessages, err := r.queryAllMessageShards(ctx, targets, q)
	if err != nil {
		return nil, err
	}

	// 3. 排序：按序列号升序
	r.sortMessages(allMessages)

	// 4. 分页：内存切片
	msgs := r.paginateMessages(allMessages, q.Limit, q.Offset)

	// 5. 丰富：填充头像等信息
	if len(msgs) > 0 {
		if err := r.enrichMessages(ctx, msgs); err != nil {
			log.Warn().Err(err).Msg("填充消息头像失败")
			// 降级策略：不阻断主流程
		}
	}

	return msgs, nil
}

// queryAllMessageShards 遍历所有分片执行查询并聚合结果
func (r *Repository) queryAllMessageShards(ctx context.Context, targets []bind.RouteResult, q types.MessageQuery) ([]*model.Message, error) {
	var allMessages []*model.Message

	for _, target := range targets {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		msgs, err := r.querySingleShard(ctx, target, q)
		if err != nil {
			// 单个分片查询失败只记录警告，不中断整体流程
			log.Warn().Err(err).Str("db", target.FilePath).Msg("查询数据库分片失败，跳过")
			continue
		}
		allMessages = append(allMessages, msgs...)
	}
	return allMessages, nil
}

// querySingleShard 在单个数据库分片上执行查询
func (r *Repository) querySingleShard(ctx context.Context, target bind.RouteResult, q types.MessageQuery) ([]*model.Message, error) {
	db, err := r.pool.GetConnection(target.FilePath)
	if err != nil {
		return nil, err
	}

	// 计算 V4 表名
	hash := md5.Sum([]byte(target.Talker))
	tableName := "Msg_" + hex.EncodeToString(hash[:])

	if r.isTableExist(db, tableName) {
		return r.queryV4Messages(ctx, db, tableName, target.Talker, q)
	}

	return r.queryV3Messages(ctx, db, target, q)
}

// enrichMessages 填充消息的发送者头像信息
// enrichMessages 填充消息的发送者头像信息和会话对象名称
func (r *Repository) enrichMessages(ctx context.Context, msgs []*model.Message) error {
	ids := make(map[string]struct{})
	for _, m := range msgs {
		if m.Sender != "" {
			ids[m.Sender] = struct{}{}
		}
		if m.Talker != "" {
			ids[m.Talker] = struct{}{}
		}
	}

	if len(ids) == 0 {
		return nil
	}

	idList := make([]string, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}

	// 使用包内共享方法获取联系人信息
	profiles, err := r.getContactProfiles(ctx, idList)
	if err != nil {
		return err
	}

	for _, m := range msgs {
		// 填充发送者信息
		if profile, ok := profiles[m.Sender]; ok {
			m.BigHeadURL = profile.BigHeadURL
			m.SmallHeadURL = profile.SmallHeadURL
			if profile.Remark == "" {
				m.SenderName = profile.NickName
			} else {
				m.SenderName = profile.Remark
			}
		}
		// 填充会话对象信息
		if profile, ok := profiles[m.Talker]; ok {
			if profile.Remark == "" {
				m.TalkerName = profile.NickName
			} else {
				m.TalkerName = profile.Remark
			}
		}
		if m.TalkerName == "" {
			m.TalkerName = m.Talker
		}
	}
	return nil
}

func (r *Repository) queryV4Messages(ctx context.Context, db *sql.DB, tableName, talker string, q types.MessageQuery) ([]*model.Message, error) {
	query := fmt.Sprintf(`
		SELECT m.sort_seq, m.server_id, m.local_type, n.user_name, m.create_time, m.message_content, m.packed_info_data, m.status
		FROM %s m
		LEFT JOIN Name2Id n ON m.real_sender_id = n.rowid
		WHERE m.create_time >= ? AND m.create_time <= ?
		ORDER BY m.sort_seq ASC
	`, tableName)

	rows, err := db.QueryContext(ctx, query, q.StartTime.Unix(), q.EndTime.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*model.Message
	for rows.Next() {
		var msg model.MessageV4
		err := rows.Scan(
			&msg.SortSeq,
			&msg.ServerID,
			&msg.LocalType,
			&msg.UserName,
			&msg.CreateTime,
			&msg.MessageContent,
			&msg.PackedInfoData,
			&msg.Status,
		)
		if err != nil {
			return nil, err
		}

		wrapped := msg.Wrap(talker)
		if q.Sender != "" && wrapped.Sender != q.Sender {
			continue
		}
		msgs = append(msgs, wrapped)
	}
	return msgs, nil
}

func (r *Repository) queryV3Messages(ctx context.Context, db *sql.DB, target bind.RouteResult, q types.MessageQuery) ([]*model.Message, error) {
	query, args := r.buildV3MessageQuery(target, q)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "no such table") {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var msgs []*model.Message
	for rows.Next() {
		var msg model.MessageV3
		var compressContent []byte
		var bytesExtra []byte

		err := rows.Scan(
			&msg.MsgSvrID,
			&msg.Sequence,
			&msg.CreateTime,
			&msg.StrTalker,
			&msg.IsSender,
			&msg.Type,
			&msg.SubType,
			&msg.StrContent,
			&compressContent,
			&bytesExtra,
		)
		if err != nil {
			return nil, err
		}
		msg.CompressContent = compressContent
		msg.BytesExtra = bytesExtra

		wrapped := msg.Wrap()
		if q.Sender != "" && wrapped.Sender != q.Sender {
			continue
		}
		msgs = append(msgs, wrapped)
	}
	return msgs, nil
}

func (r *Repository) buildV3MessageQuery(target bind.RouteResult, q types.MessageQuery) (string, []interface{}) {
	var sb strings.Builder
	var args []interface{}

	sb.WriteString(`
		SELECT MsgSvrID, Sequence, CreateTime, StrTalker, IsSender, 
			Type, SubType, StrContent, CompressContent, BytesExtra
		FROM MSG 
		WHERE Sequence >= ? AND Sequence <= ?
	`)
	args = append(args, q.StartTime.Unix()*1000, q.EndTime.Unix()*1000)

	if target.TalkerID != 0 {
		sb.WriteString(" AND TalkerId = ?")
		args = append(args, target.TalkerID)
	} else {
		sb.WriteString(" AND StrTalker = ?")
		args = append(args, target.Talker)
	}

	sb.WriteString(" ORDER BY Sequence ASC")

	return sb.String(), args
}

func (r *Repository) sortMessages(msgs []*model.Message) {
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Seq < msgs[j].Seq
	})
}

func (r *Repository) paginateMessages(msgs []*model.Message, limit, offset int) []*model.Message {
	total := len(msgs)
	start := offset

	if start >= total {
		return []*model.Message{}
	}

	end := start + limit
	if limit == 0 || end > total {
		end = total
	}

	return msgs[start:end]
}
