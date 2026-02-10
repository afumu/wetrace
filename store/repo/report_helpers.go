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
	"github.com/afumu/wetrace/store/types"
)

// getAnnualOverview 获取年度概览统计
func (r *Repository) getAnnualOverview(ctx context.Context, start, end time.Time) (model.AnnualOverview, error) {
	var overview model.AnnualOverview
	var totalMsgs, sentMsgs, recvMsgs int
	contactSet := make(map[string]bool)
	chatroomSet := make(map[string]bool)
	daySet := make(map[string]bool)
	var firstDate, lastDate string

	for _, shard := range r.router.GetShards() {
		db, err := r.pool.GetConnection(shard.FilePath)
		if err != nil {
			continue
		}

		if r.isTableExist(db, "MSG") {
			r.overviewV3(ctx, db, start, end, &totalMsgs, &sentMsgs, &recvMsgs, contactSet, chatroomSet, daySet, &firstDate, &lastDate)
		} else {
			r.overviewV4(ctx, db, start, end, &totalMsgs, &sentMsgs, &recvMsgs, contactSet, chatroomSet, daySet, &firstDate, &lastDate)
		}
	}

	overview.TotalMessages = totalMsgs
	overview.SentMessages = sentMsgs
	overview.ReceivedMessages = recvMsgs
	overview.ActiveDays = len(daySet)
	overview.FirstMessageDate = firstDate
	overview.LastMessageDate = lastDate

	// 统计联系人和群聊
	activeContacts := 0
	activeChatrooms := 0
	for id := range contactSet {
		if strings.HasSuffix(id, "@chatroom") {
			activeChatrooms++
		} else {
			activeContacts++
		}
	}
	overview.ActiveContacts = activeContacts
	overview.ActiveChatrooms = activeChatrooms

	// 总联系人和群聊数从 session 获取
	sessions, _ := r.GetSessions(ctx, types.SessionQuery{Limit: 10000})
	totalContacts := 0
	totalChatrooms := 0
	for _, s := range sessions {
		if strings.HasSuffix(s.UserName, "@chatroom") {
			totalChatrooms++
		} else {
			totalContacts++
		}
	}
	overview.TotalContacts = totalContacts
	overview.TotalChatrooms = totalChatrooms

	return overview, nil
}

func (r *Repository) overviewV3(ctx context.Context, db *sql.DB, start, end time.Time,
	totalMsgs, sentMsgs, recvMsgs *int,
	contactSet, chatroomSet, daySet map[string]bool,
	firstDate, lastDate *string) {

	query := `SELECT COUNT(*),
		SUM(CASE WHEN COALESCE(IsSender, 0) = 1 THEN 1 ELSE 0 END),
		SUM(CASE WHEN COALESCE(IsSender, 0) != 1 THEN 1 ELSE 0 END)
		FROM MSG WHERE CreateTime >= ? AND CreateTime <= ?`
	var total, sent, recv sql.NullInt64
	if err := db.QueryRowContext(ctx, query, start.Unix()*1000, end.Unix()*1000).Scan(&total, &sent, &recv); err == nil {
		if total.Valid {
			*totalMsgs += int(total.Int64)
		}
		if sent.Valid {
			*sentMsgs += int(sent.Int64)
		}
		if recv.Valid {
			*recvMsgs += int(recv.Int64)
		}
	}

	// 活跃联系人
	rows, err := db.QueryContext(ctx, "SELECT DISTINCT StrTalker FROM MSG WHERE CreateTime >= ? AND CreateTime <= ?", start.Unix()*1000, end.Unix()*1000)
	if err == nil {
		for rows.Next() {
			var talker string
			rows.Scan(&talker)
			contactSet[talker] = true
		}
		rows.Close()
	}

	// 活跃天数和首末日期
	rows, err = db.QueryContext(ctx,
		"SELECT DISTINCT strftime('%Y-%m-%d', CreateTime/1000, 'unixepoch', 'localtime') as d FROM MSG WHERE CreateTime >= ? AND CreateTime <= ? ORDER BY d",
		start.Unix()*1000, end.Unix()*1000)
	if err == nil {
		for rows.Next() {
			var d string
			rows.Scan(&d)
			daySet[d] = true
			if *firstDate == "" || d < *firstDate {
				*firstDate = d
			}
			if d > *lastDate {
				*lastDate = d
			}
		}
		rows.Close()
	}
}

func (r *Repository) overviewV4(ctx context.Context, db *sql.DB, start, end time.Time,
	totalMsgs, sentMsgs, recvMsgs *int,
	contactSet, chatroomSet, daySet map[string]bool,
	firstDate, lastDate *string) {

	talkerMD5Map := r.getTalkerMD5Map(ctx)

	tables, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'Msg_%%'")
	if err != nil {
		return
	}
	defer tables.Close()

	for tables.Next() {
		var tableName string
		tables.Scan(&tableName)

		talker := "unknown"
		if strings.HasPrefix(tableName, "Msg_") {
			md5Hash := strings.TrimPrefix(tableName, "Msg_")
			if t, ok := talkerMD5Map[md5Hash]; ok {
				talker = t
			} else {
				// talkerMD5Map 查不到时，通过 distinct real_sender_id 数量判断是否为群聊
				// 群聊有多个不同发送者（real_sender_id > 0 的去重数 >= 2）
				var distinctSenders int
				cntQuery := fmt.Sprintf("SELECT COUNT(DISTINCT real_sender_id) FROM %s WHERE real_sender_id > 0", tableName)
				if db.QueryRowContext(ctx, cntQuery).Scan(&distinctSenders) == nil && distinctSenders >= 2 {
					talker = md5Hash + "@chatroom"
				}
			}
		}

		// 统计消息数: real_sender_id = 0 表示自己发送的消息
		query := fmt.Sprintf(`SELECT COUNT(*),
			SUM(CASE WHEN m.real_sender_id = 0 THEN 1 ELSE 0 END),
			SUM(CASE WHEN m.real_sender_id != 0 THEN 1 ELSE 0 END)
			FROM %s m WHERE m.create_time >= ? AND m.create_time <= ?`, tableName)
		var total, sent, recv sql.NullInt64
		if err := db.QueryRowContext(ctx, query, start.Unix(), end.Unix()).Scan(&total, &sent, &recv); err == nil {
			if total.Valid && total.Int64 > 0 {
				*totalMsgs += int(total.Int64)
				if sent.Valid {
					*sentMsgs += int(sent.Int64)
				}
				if recv.Valid {
					*recvMsgs += int(recv.Int64)
				}
				contactSet[talker] = true
			}
		}

		// 活跃天数
		daysQuery := fmt.Sprintf(
			"SELECT DISTINCT strftime('%%Y-%%m-%%d', create_time, 'unixepoch', 'localtime') as d FROM %s WHERE create_time >= ? AND create_time <= ? ORDER BY d",
			tableName)
		rows, err := db.QueryContext(ctx, daysQuery, start.Unix(), end.Unix())
		if err == nil {
			for rows.Next() {
				var d string
				rows.Scan(&d)
				daySet[d] = true
				if *firstDate == "" || d < *firstDate {
					*firstDate = d
				}
				if d > *lastDate {
					*lastDate = d
				}
			}
			rows.Close()
		}
	}
}

// getAnnualTopContacts 获取年度亲密度排行
func (r *Repository) getAnnualTopContacts(ctx context.Context, start, end time.Time, limit int) ([]*model.PersonalTopContact, error) {
	sessions, err := r.GetSessions(ctx, types.SessionQuery{Limit: 1000})
	if err != nil {
		return nil, err
	}

	type stats struct {
		sent     int
		recv     int
		lastTime int64
	}
	aggStats := make(map[string]*stats)

	for _, session := range sessions {
		talker := session.UserName
		if strings.HasSuffix(talker, "@chatroom") {
			continue
		}

		targets := r.router.Resolve(start, end, talker)
		s := &stats{}

		for _, target := range targets {
			db, err := r.pool.GetConnection(target.FilePath)
			if err != nil {
				continue
			}
			hash := md5.Sum([]byte(talker))
			tableName := "Msg_" + hex.EncodeToString(hash[:])

			if r.isTableExist(db, tableName) {
				// real_sender_id = 0 表示自己发送的消息
				query := fmt.Sprintf(
					"SELECT CASE WHEN real_sender_id = 0 THEN 1 ELSE 0 END as is_self, COUNT(*), MAX(create_time) FROM %s WHERE create_time >= ? AND create_time <= ? GROUP BY is_self",
					tableName)
				rows, err := db.QueryContext(ctx, query, start.Unix(), end.Unix())
				if err == nil {
					for rows.Next() {
						var isSelf, count int
						var maxTime int64
						if rows.Scan(&isSelf, &count, &maxTime) == nil {
							if isSelf == 1 {
								s.sent += count
							} else {
								s.recv += count
							}
							if maxTime > s.lastTime {
								s.lastTime = maxTime
							}
						}
					}
					rows.Close()
				}
			} else {
				query := "SELECT IsSender, COUNT(*), MAX(CreateTime/1000) FROM MSG WHERE CreateTime >= ? AND CreateTime <= ?"
				var args []interface{}
				args = append(args, start.Unix()*1000, end.Unix()*1000)
				if target.TalkerID != 0 {
					query += " AND TalkerId = ?"
					args = append(args, target.TalkerID)
				} else {
					query += " AND StrTalker = ?"
					args = append(args, target.Talker)
				}
				query += " GROUP BY IsSender"
				rows, err := db.QueryContext(ctx, query, args...)
				if err == nil {
					for rows.Next() {
						var isSelf, count int
						var maxTime int64
						if rows.Scan(&isSelf, &count, &maxTime) == nil {
							if isSelf == 1 {
								s.sent += count
							} else {
								s.recv += count
							}
							if maxTime > s.lastTime {
								s.lastTime = maxTime
							}
						}
					}
					rows.Close()
				}
			}
		}
		if s.sent > 0 || s.recv > 0 {
			aggStats[talker] = s
		}
	}

	var talkers []string
	for t := range aggStats {
		talkers = append(talkers, t)
	}
	profiles, _ := r.getContactProfiles(ctx, talkers)

	var result []*model.PersonalTopContact
	for t, s := range aggStats {
		name := t
		var avatar string
		if p, ok := profiles[t]; ok {
			if p.Remark != "" {
				name = p.Remark
			} else if p.NickName != "" {
				name = p.NickName
			}
			avatar = p.SmallHeadURL
		}
		result = append(result, &model.PersonalTopContact{
			Talker:       t,
			Name:         name,
			Avatar:       avatar,
			MessageCount: s.sent + s.recv,
			SentCount:    s.sent,
			RecvCount:    s.recv,
			LastTime:     s.lastTime,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].MessageCount > result[j].MessageCount
	})

	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// getAnnualMonthlyTrend 获取年度月度趋势
func (r *Repository) getAnnualMonthlyTrend(ctx context.Context, start, end time.Time) []*model.MonthlyStat {
	monthlyStats := make(map[int]int)

	for _, shard := range r.router.GetShards() {
		db, err := r.pool.GetConnection(shard.FilePath)
		if err != nil {
			continue
		}

		if r.isTableExist(db, "MSG") {
			query := "SELECT CAST(strftime('%m', CreateTime/1000, 'unixepoch', 'localtime') AS INTEGER) as month, COUNT(*) as count FROM MSG WHERE CreateTime >= ? AND CreateTime <= ? GROUP BY month"
			rows, err := db.QueryContext(ctx, query, start.Unix()*1000, end.Unix()*1000)
			if err == nil {
				for rows.Next() {
					var m, c int
					rows.Scan(&m, &c)
					monthlyStats[m] += c
				}
				rows.Close()
			}
		} else {
			tables, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'Msg_%%'")
			if err != nil {
				continue
			}
			for tables.Next() {
				var tableName string
				tables.Scan(&tableName)
				query := fmt.Sprintf("SELECT CAST(strftime('%%m', create_time, 'unixepoch', 'localtime') AS INTEGER) as month, COUNT(*) as count FROM %s WHERE create_time >= ? AND create_time <= ? GROUP BY month", tableName)
				rows, err := db.QueryContext(ctx, query, start.Unix(), end.Unix())
				if err == nil {
					for rows.Next() {
						var m, c int
						rows.Scan(&m, &c)
						monthlyStats[m] += c
					}
					rows.Close()
				}
			}
			tables.Close()
		}
	}

	var result []*model.MonthlyStat
	for i := 1; i <= 12; i++ {
		result = append(result, &model.MonthlyStat{Month: i, Count: monthlyStats[i]})
	}
	return result
}

// getAnnualWeekdayDist 获取年度星期分布
func (r *Repository) getAnnualWeekdayDist(ctx context.Context, start, end time.Time) []*model.WeekdayStat {
	weekdayStats := make(map[int]int)

	for _, shard := range r.router.GetShards() {
		db, err := r.pool.GetConnection(shard.FilePath)
		if err != nil {
			continue
		}

		if r.isTableExist(db, "MSG") {
			query := "SELECT CASE WHEN CAST(strftime('%w', CreateTime/1000, 'unixepoch', 'localtime') AS INTEGER) = 0 THEN 7 ELSE CAST(strftime('%w', CreateTime/1000, 'unixepoch', 'localtime') AS INTEGER) END as weekday, COUNT(*) as count FROM MSG WHERE CreateTime >= ? AND CreateTime <= ? GROUP BY weekday"
			rows, err := db.QueryContext(ctx, query, start.Unix()*1000, end.Unix()*1000)
			if err == nil {
				for rows.Next() {
					var w, c int
					rows.Scan(&w, &c)
					weekdayStats[w] += c
				}
				rows.Close()
			}
		} else {
			r.weekdayV4Shards(ctx, db, start, end, weekdayStats)
		}
	}

	var result []*model.WeekdayStat
	for i := 1; i <= 7; i++ {
		result = append(result, &model.WeekdayStat{Weekday: i, Count: weekdayStats[i]})
	}
	return result
}

func (r *Repository) weekdayV4Shards(ctx context.Context, db *sql.DB, start, end time.Time, weekdayStats map[int]int) {
	tables, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'Msg_%%'")
	if err != nil {
		return
	}
	defer tables.Close()

	for tables.Next() {
		var tableName string
		tables.Scan(&tableName)
		query := fmt.Sprintf("SELECT CASE WHEN CAST(strftime('%%w', create_time, 'unixepoch', 'localtime') AS INTEGER) = 0 THEN 7 ELSE CAST(strftime('%%w', create_time, 'unixepoch', 'localtime') AS INTEGER) END as weekday, COUNT(*) as count FROM %s WHERE create_time >= ? AND create_time <= ? GROUP BY weekday", tableName)
		rows, err := db.QueryContext(ctx, query, start.Unix(), end.Unix())
		if err == nil {
			for rows.Next() {
				var w, c int
				rows.Scan(&w, &c)
				weekdayStats[w] += c
			}
			rows.Close()
		}
	}
}

// getAnnualHourlyDist 获取年度小时分布
func (r *Repository) getAnnualHourlyDist(ctx context.Context, start, end time.Time) []*model.HourlyStat {
	hourlyStats := make(map[int]int)

	for _, shard := range r.router.GetShards() {
		db, err := r.pool.GetConnection(shard.FilePath)
		if err != nil {
			continue
		}

		if r.isTableExist(db, "MSG") {
			query := "SELECT CAST(strftime('%H', CreateTime/1000, 'unixepoch', 'localtime') AS INTEGER) as hour, COUNT(*) as count FROM MSG WHERE CreateTime >= ? AND CreateTime <= ? GROUP BY hour"
			rows, err := db.QueryContext(ctx, query, start.Unix()*1000, end.Unix()*1000)
			if err == nil {
				for rows.Next() {
					var h, c int
					rows.Scan(&h, &c)
					hourlyStats[h] += c
				}
				rows.Close()
			}
		} else {
			r.hourlyV4Shards(ctx, db, start, end, hourlyStats)
		}
	}

	var result []*model.HourlyStat
	for i := 0; i < 24; i++ {
		result = append(result, &model.HourlyStat{Hour: i, Count: hourlyStats[i]})
	}
	return result
}

func (r *Repository) hourlyV4Shards(ctx context.Context, db *sql.DB, start, end time.Time, hourlyStats map[int]int) {
	tables, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'Msg_%%'")
	if err != nil {
		return
	}
	defer tables.Close()

	for tables.Next() {
		var tableName string
		tables.Scan(&tableName)
		query := fmt.Sprintf("SELECT CAST(strftime('%%H', create_time, 'unixepoch', 'localtime') AS INTEGER) as hour, COUNT(*) as count FROM %s WHERE create_time >= ? AND create_time <= ? GROUP BY hour", tableName)
		rows, err := db.QueryContext(ctx, query, start.Unix(), end.Unix())
		if err == nil {
			for rows.Next() {
				var h, c int
				rows.Scan(&h, &c)
				hourlyStats[h] += c
			}
			rows.Close()
		}
	}
}

// getAnnualMessageTypes 获取年度消息类型分布
func (r *Repository) getAnnualMessageTypes(ctx context.Context, start, end time.Time) map[string]int {
	typeStats := make(map[int]int)

	for _, shard := range r.router.GetShards() {
		db, err := r.pool.GetConnection(shard.FilePath)
		if err != nil {
			continue
		}

		if r.isTableExist(db, "MSG") {
			query := "SELECT Type, COUNT(*) FROM MSG WHERE CreateTime >= ? AND CreateTime <= ? GROUP BY Type"
			rows, err := db.QueryContext(ctx, query, start.Unix()*1000, end.Unix()*1000)
			if err == nil {
				for rows.Next() {
					var t, c int
					rows.Scan(&t, &c)
					typeStats[t] += c
				}
				rows.Close()
			}
		} else {
			r.messageTypesV4Shards(ctx, db, start, end, typeStats)
		}
	}

	// 转换为可读名称
	result := make(map[string]int)
	for t, c := range typeStats {
		name := messageTypeName(t)
		result[name] += c
	}
	return result
}

func (r *Repository) messageTypesV4Shards(ctx context.Context, db *sql.DB, start, end time.Time, typeStats map[int]int) {
	tables, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'Msg_%%'")
	if err != nil {
		return
	}
	defer tables.Close()

	for tables.Next() {
		var tableName string
		tables.Scan(&tableName)
		query := fmt.Sprintf("SELECT local_type, COUNT(*) FROM %s WHERE create_time >= ? AND create_time <= ? GROUP BY local_type", tableName)
		rows, err := db.QueryContext(ctx, query, start.Unix(), end.Unix())
		if err == nil {
			for rows.Next() {
				var t, c int
				rows.Scan(&t, &c)
				typeStats[t] += c
			}
			rows.Close()
		}
	}
}

// messageTypeName 将消息类型数字转换为可读名称
func messageTypeName(t int) string {
	switch t {
	case 1:
		return "text"
	case 3:
		return "image"
	case 34:
		return "voice"
	case 43:
		return "video"
	case 49:
		return "link"
	case 47:
		return "emoji"
	case 42:
		return "card"
	case 48:
		return "location"
	case 50:
		return "voip"
	case 10000:
		return "system"
	default:
		return "other"
	}
}

// getAnnualHighlights 获取年度亮点数据
func (r *Repository) getAnnualHighlights(ctx context.Context, start, end time.Time) model.AnnualHighlights {
	highlights := model.AnnualHighlights{}
	dailyCounts := make(map[string]int)
	var lateNightCount int
	var earliestMinute, latestMinute int
	earliestMinute = 24 * 60 // 初始化为最大值
	latestMinute = -1

	for _, shard := range r.router.GetShards() {
		db, err := r.pool.GetConnection(shard.FilePath)
		if err != nil {
			continue
		}

		if r.isTableExist(db, "MSG") {
			r.highlightsV3(ctx, db, start, end, dailyCounts, &lateNightCount, &earliestMinute, &latestMinute)
		} else {
			r.highlightsV4(ctx, db, start, end, dailyCounts, &lateNightCount, &earliestMinute, &latestMinute)
		}
	}

	// 找出最忙和最闲的一天
	var busiestDay, quietestDay model.DayCount
	quietestDay.Count = int(^uint(0) >> 1) // MaxInt

	for date, count := range dailyCounts {
		if count > busiestDay.Count {
			busiestDay = model.DayCount{Date: date, Count: count}
		}
		if count < quietestDay.Count {
			quietestDay = model.DayCount{Date: date, Count: count}
		}
	}
	if quietestDay.Count == int(^uint(0)>>1) {
		quietestDay = model.DayCount{}
	}

	highlights.BusiestDay = busiestDay
	highlights.QuietestDay = quietestDay
	highlights.LateNightCount = lateNightCount
	highlights.LongestStreak = calcLongestStreak(dailyCounts)

	if earliestMinute < 24*60 {
		highlights.EarliestMessageTime = fmt.Sprintf("%02d:%02d", earliestMinute/60, earliestMinute%60)
	}
	if latestMinute >= 0 {
		highlights.LatestMessageTime = fmt.Sprintf("%02d:%02d", latestMinute/60, latestMinute%60)
	}

	return highlights
}

func (r *Repository) highlightsV3(ctx context.Context, db *sql.DB, start, end time.Time,
	dailyCounts map[string]int, lateNightCount *int, earliestMinute, latestMinute *int) {

	// 每日消息数
	query := "SELECT strftime('%Y-%m-%d', CreateTime/1000, 'unixepoch', 'localtime') as d, COUNT(*) as c FROM MSG WHERE CreateTime >= ? AND CreateTime <= ? GROUP BY d"
	rows, err := db.QueryContext(ctx, query, start.Unix()*1000, end.Unix()*1000)
	if err == nil {
		for rows.Next() {
			var d string
			var c int
			rows.Scan(&d, &c)
			dailyCounts[d] += c
		}
		rows.Close()
	}

	// 深夜消息数 (0-5点)
	var lnc int
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM MSG WHERE CreateTime >= ? AND CreateTime <= ? AND CAST(strftime('%H', CreateTime/1000, 'unixepoch', 'localtime') AS INTEGER) < 5",
		start.Unix()*1000, end.Unix()*1000).Scan(&lnc)
	if err == nil {
		*lateNightCount += lnc
	}

	// 最早和最晚消息时间
	var minHour, minMin sql.NullInt64
	err = db.QueryRowContext(ctx,
		"SELECT CAST(strftime('%H', CreateTime/1000, 'unixepoch', 'localtime') AS INTEGER), CAST(strftime('%M', CreateTime/1000, 'unixepoch', 'localtime') AS INTEGER) FROM MSG WHERE CreateTime >= ? AND CreateTime <= ? ORDER BY strftime('%H%M', CreateTime/1000, 'unixepoch', 'localtime') ASC LIMIT 1",
		start.Unix()*1000, end.Unix()*1000).Scan(&minHour, &minMin)
	if err == nil && minHour.Valid {
		m := int(minHour.Int64)*60 + int(minMin.Int64)
		if m < *earliestMinute {
			*earliestMinute = m
		}
	}

	var maxHour, maxMin sql.NullInt64
	err = db.QueryRowContext(ctx,
		"SELECT CAST(strftime('%H', CreateTime/1000, 'unixepoch', 'localtime') AS INTEGER), CAST(strftime('%M', CreateTime/1000, 'unixepoch', 'localtime') AS INTEGER) FROM MSG WHERE CreateTime >= ? AND CreateTime <= ? ORDER BY strftime('%H%M', CreateTime/1000, 'unixepoch', 'localtime') DESC LIMIT 1",
		start.Unix()*1000, end.Unix()*1000).Scan(&maxHour, &maxMin)
	if err == nil && maxHour.Valid {
		m := int(maxHour.Int64)*60 + int(maxMin.Int64)
		if m > *latestMinute {
			*latestMinute = m
		}
	}
}

func (r *Repository) highlightsV4(ctx context.Context, db *sql.DB, start, end time.Time,
	dailyCounts map[string]int, lateNightCount *int, earliestMinute, latestMinute *int) {

	tables, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'Msg_%%'")
	if err != nil {
		return
	}
	defer tables.Close()

	for tables.Next() {
		var tableName string
		tables.Scan(&tableName)

		// 每日消息数
		query := fmt.Sprintf("SELECT strftime('%%Y-%%m-%%d', create_time, 'unixepoch', 'localtime') as d, COUNT(*) as c FROM %s WHERE create_time >= ? AND create_time <= ? GROUP BY d", tableName)
		rows, err := db.QueryContext(ctx, query, start.Unix(), end.Unix())
		if err == nil {
			for rows.Next() {
				var d string
				var c int
				rows.Scan(&d, &c)
				dailyCounts[d] += c
			}
			rows.Close()
		}

		// 深夜消息数
		var lnc int
		err = db.QueryRowContext(ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE create_time >= ? AND create_time <= ? AND CAST(strftime('%%H', create_time, 'unixepoch', 'localtime') AS INTEGER) < 5", tableName),
			start.Unix(), end.Unix()).Scan(&lnc)
		if err == nil {
			*lateNightCount += lnc
		}

		// 最早消息时间
		var minH, minM sql.NullInt64
		err = db.QueryRowContext(ctx,
			fmt.Sprintf("SELECT CAST(strftime('%%H', create_time, 'unixepoch', 'localtime') AS INTEGER), CAST(strftime('%%M', create_time, 'unixepoch', 'localtime') AS INTEGER) FROM %s WHERE create_time >= ? AND create_time <= ? ORDER BY strftime('%%H%%M', create_time, 'unixepoch', 'localtime') ASC LIMIT 1", tableName),
			start.Unix(), end.Unix()).Scan(&minH, &minM)
		if err == nil && minH.Valid {
			m := int(minH.Int64)*60 + int(minM.Int64)
			if m < *earliestMinute {
				*earliestMinute = m
			}
		}

		// 最晚消息时间
		var maxH, maxM sql.NullInt64
		err = db.QueryRowContext(ctx,
			fmt.Sprintf("SELECT CAST(strftime('%%H', create_time, 'unixepoch', 'localtime') AS INTEGER), CAST(strftime('%%M', create_time, 'unixepoch', 'localtime') AS INTEGER) FROM %s WHERE create_time >= ? AND create_time <= ? ORDER BY strftime('%%H%%M', create_time, 'unixepoch', 'localtime') DESC LIMIT 1", tableName),
			start.Unix(), end.Unix()).Scan(&maxH, &maxM)
		if err == nil && maxH.Valid {
			m := int(maxH.Int64)*60 + int(maxM.Int64)
			if m > *latestMinute {
				*latestMinute = m
			}
		}
	}
}

// calcLongestStreak 计算最长连续活跃天数
func calcLongestStreak(dailyCounts map[string]int) int {
	if len(dailyCounts) == 0 {
		return 0
	}

	dates := make([]string, 0, len(dailyCounts))
	for d := range dailyCounts {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	longest := 1
	current := 1

	for i := 1; i < len(dates); i++ {
		prev, err1 := time.Parse("2006-01-02", dates[i-1])
		curr, err2 := time.Parse("2006-01-02", dates[i])
		if err1 != nil || err2 != nil {
			current = 1
			continue
		}
		if curr.Sub(prev).Hours() == 24 {
			current++
			if current > longest {
				longest = current
			}
		} else {
			current = 1
		}
	}
	return longest
}
