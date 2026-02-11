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
)

// GetNeedContactList 获取需要联系的客户列表
// days: 超过多少天未联系的客户
func (r *Repository) GetNeedContactList(ctx context.Context, days int) ([]*model.NeedContactItem, error) {
	// 1. 获取所有会话
	sessions, err := r.GetSessions(ctx, types.SessionQuery{Limit: 100000})
	if err != nil {
		return nil, fmt.Errorf("获取会话列表失败: %w", err)
	}

	// 2. 计算每个联系人的最后消息时间
	lastContactMap := make(map[string]int64)

	for _, session := range sessions {
		talker := session.UserName
		// 排除群聊和公众号
		if strings.HasSuffix(talker, "@chatroom") || strings.HasPrefix(talker, "gh_") {
			continue
		}

		lastTime := r.getLastMessageTime(ctx, talker)
		if lastTime > 0 {
			lastContactMap[talker] = lastTime
		}
	}

	// 3. 过滤出超过指定天数未联系的联系人
	now := time.Now()
	threshold := now.AddDate(0, 0, -days).Unix()

	var needContactTalkers []string
	for talker, lastTime := range lastContactMap {
		if lastTime < threshold {
			needContactTalkers = append(needContactTalkers, talker)
		}
	}

	// 4. 获取联系人信息
	profiles, _ := r.getContactProfiles(ctx, needContactTalkers)

	// 5. 构建结果
	var result []*model.NeedContactItem
	for _, talker := range needContactTalkers {
		lastTime := lastContactMap[talker]
		daysSince := int(now.Sub(time.Unix(lastTime, 0)).Hours() / 24)

		item := &model.NeedContactItem{
			UserName:         talker,
			NickName:         talker,
			Remark:           "",
			LastContactTime:  lastTime,
			DaysSinceContact: daysSince,
		}

		if p, ok := profiles[talker]; ok {
			item.NickName = p.NickName
			item.Remark = p.Remark
			item.SmallHeadURL = p.SmallHeadURL
			item.BigHeadURL = p.BigHeadURL
		}

		result = append(result, item)
	}

	// 6. 按距今天数降序排列（最久未联系的排前面）
	sort.Slice(result, func(i, j int) bool {
		return result[i].DaysSinceContact > result[j].DaysSinceContact
	})

	return result, nil
}

// getLastMessageTime 获取指定联系人的最后消息时间（unix秒）
func (r *Repository) getLastMessageTime(ctx context.Context, talker string) int64 {
	var maxTime int64

	for _, shard := range r.router.GetShards() {
		db, err := r.pool.GetConnection(shard.FilePath)
		if err != nil {
			continue
		}

		t := r.queryLastTimeFromShard(ctx, db, shard, talker)
		if t > maxTime {
			maxTime = t
		}
	}

	return maxTime
}

// queryLastTimeFromShard 从单个分片查询指定联系人的最后消息时间
func (r *Repository) queryLastTimeFromShard(ctx context.Context, db *sql.DB, shard *bind.DatabaseShard, talker string) int64 {
	// 尝试 V4 模式
	hash := md5.Sum([]byte(talker))
	tableName := "Msg_" + hex.EncodeToString(hash[:])

	if r.isTableExist(db, tableName) {
		var maxTime sql.NullInt64
		query := fmt.Sprintf("SELECT MAX(create_time) FROM %s", tableName)
		if err := db.QueryRowContext(ctx, query).Scan(&maxTime); err == nil && maxTime.Valid {
			return maxTime.Int64
		}
		return 0
	}

	// V3 模式
	if r.isTableExist(db, "MSG") {
		var maxTime sql.NullInt64
		query := "SELECT MAX(CreateTime/1000) FROM MSG WHERE StrTalker = ?"
		// 优先使用 TalkerID
		if id, ok := shard.TalkerMap[talker]; ok {
			query = "SELECT MAX(CreateTime/1000) FROM MSG WHERE TalkerId = ?"
			if err := db.QueryRowContext(ctx, query, id).Scan(&maxTime); err == nil && maxTime.Valid {
				return maxTime.Int64
			}
		} else {
			if err := db.QueryRowContext(ctx, query, talker).Scan(&maxTime); err == nil && maxTime.Valid {
				return maxTime.Int64
			}
		}
	}

	return 0
}
