package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/types"
)

// contactProfile 定义了跨 Session 和 Message 复用的联系人简要信息
type contactProfile struct {
	Remark       string
	NickName     string
	SmallHeadURL string
	BigHeadURL   string
}

// GetSessions 获取会话列表
func (r *Repository) GetSessions(ctx context.Context, q types.SessionQuery) ([]*model.Session, error) {
	dbPath, err := r.router.GetSessionDBPath()
	if err != nil {
		return nil, err
	}
	db, err := r.pool.GetConnection(dbPath)
	if err != nil {
		return nil, err
	}

	// 1. 获取基础会话数据
	sessions, err := r.queryRawSessions(ctx, db, q)
	if err != nil {
		return nil, err
	}

	// 2. 丰富会话信息（头像、昵称等）
	if len(sessions) > 0 {
		if err := r.enrichSessions(ctx, sessions); err != nil {
			return nil, err // 或者记录日志并降级返回
		}
	}

	return sessions, nil
}

// DeleteSession 删除会话
func (r *Repository) DeleteSession(ctx context.Context, username string) error {
	dbPath, err := r.router.GetSessionDBPath()
	if err != nil {
		return err
	}
	db, err := r.pool.GetConnection(dbPath)
	if err != nil {
		return err
	}

	var query string
	if r.isTableExist(db, "SessionTable") {
		// V4
		query = "DELETE FROM SessionTable WHERE username = ?"
	} else {
		// V3
		query = "DELETE FROM Session WHERE strUsrName = ?"
	}

	_, err = db.ExecContext(ctx, query, username)
	return err
}

// queryRawSessions 根据数据库版本路由并执行查询
func (r *Repository) queryRawSessions(ctx context.Context, db *sql.DB, q types.SessionQuery) ([]*model.Session, error) {
	if r.isTableExist(db, "SessionTable") {
		return r.queryV4Sessions(ctx, db, q)
	}
	return r.queryV3Sessions(ctx, db, q)
}

// enrichSessions 批量填充会话的头像和昵称信息
func (r *Repository) enrichSessions(ctx context.Context, sessions []*model.Session) error {
	usernames := make([]string, 0, len(sessions))
	for _, s := range sessions {
		usernames = append(usernames, s.UserName)
	}

	// 调用包内共享方法获取联系人信息
	profiles, err := r.getContactProfiles(ctx, usernames)
	if err != nil {
		return err
	}

	for _, s := range sessions {
		if profile, ok := profiles[s.UserName]; ok {
			s.SmallHeadURL = profile.SmallHeadURL
			s.BigHeadURL = profile.BigHeadURL

			// 昵称显示逻辑：优先显示备注，其次是微信昵称
			if profile.Remark != "" {
				s.NickName = profile.Remark
			} else if profile.NickName != "" {
				s.NickName = profile.NickName
			}
		}
	}
	return nil
}

// getContactProfiles 包内共享方法：根据用户名批量获取联系人信息
// 该方法封装了连接 Contact 库、判断版本、执行查询的复杂性
func (r *Repository) getContactProfiles(ctx context.Context, usernames []string) (map[string]contactProfile, error) {
	if len(usernames) == 0 {
		return nil, nil
	}

	dbPath, err := r.router.GetContactDBPath()
	if err != nil {
		return nil, err
	}
	db, err := r.pool.GetConnection(dbPath)
	if err != nil {
		return nil, err
	}

	// 构建查询
	placeholders := strings.Repeat("?,", len(usernames))
	placeholders = placeholders[:len(placeholders)-1]

	args := make([]interface{}, len(usernames))
	for i, u := range usernames {
		args[i] = u
	}

	var query string
	if r.isTableExist(db, "contact") {
		// V4
		query = fmt.Sprintf("SELECT username, remark, nick_name, small_head_url, big_head_url FROM contact WHERE username IN (%s)", placeholders)
	} else {
		// V3
		query = fmt.Sprintf("SELECT UserName, Remark, NickName, SmallHeadImgUrl, BigHeadImgUrl FROM Contact WHERE UserName IN (%s)", placeholders)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make(map[string]contactProfile)
	for rows.Next() {
		var username string
		var remark, nickName, smallHeadURL, bigHeadURL sql.NullString

		if err := rows.Scan(&username, &remark, &nickName, &smallHeadURL, &bigHeadURL); err != nil {
			return nil, err
		}

		profiles[username] = contactProfile{
			Remark:       remark.String,
			NickName:     nickName.String,
			SmallHeadURL: smallHeadURL.String,
			BigHeadURL:   bigHeadURL.String,
		}
	}
	return profiles, nil
}

// queryV4Sessions V4 版本查询逻辑
func (r *Repository) queryV4Sessions(ctx context.Context, db *sql.DB, q types.SessionQuery) ([]*model.Session, error) {
	baseQuery := `SELECT username, summary, last_timestamp, last_msg_sender, last_sender_display_name FROM SessionTable WHERE username != '@placeholder_foldgroup'`
	orderBy := `ORDER BY sort_timestamp DESC`

	var query string
	var args []interface{}

	if q.Keyword != "" {
		query = fmt.Sprintf("%s AND (username = ? OR last_sender_display_name = ?) %s", baseQuery, orderBy)
		args = []interface{}{q.Keyword, q.Keyword}
	} else {
		query = fmt.Sprintf("%s %s", baseQuery, orderBy)
	}

	query = r.appendPagination(query, q.Limit, q.Offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*model.Session
	for rows.Next() {
		var s model.SessionV4
		if err := rows.Scan(&s.Username, &s.Summary, &s.LastTimestamp, &s.LastMsgSender, &s.LastSenderDisplayName); err != nil {
			return nil, err
		}
		sessions = append(sessions, s.Wrap())
	}
	return sessions, nil
}

// queryV3Sessions V3 版本查询逻辑
func (r *Repository) queryV3Sessions(ctx context.Context, db *sql.DB, q types.SessionQuery) ([]*model.Session, error) {
	baseQuery := `SELECT strUsrName, nOrder, strNickName, strContent, nTime FROM Session WHERE strUsrName != '@placeholder_foldgroup'`
	orderBy := `ORDER BY nOrder DESC`

	var query string
	var args []interface{}

	if q.Keyword != "" {
		query = fmt.Sprintf("%s AND (strUsrName = ? OR strNickName = ?) %s", baseQuery, orderBy)
		args = []interface{}{q.Keyword, q.Keyword}
	} else {
		query = fmt.Sprintf("%s %s", baseQuery, orderBy)
	}

	query = r.appendPagination(query, q.Limit, q.Offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*model.Session
	for rows.Next() {
		var s model.SessionV3
		if err := rows.Scan(&s.StrUsrName, &s.NOrder, &s.StrNickName, &s.StrContent, &s.NTime); err != nil {
			return nil, err
		}
		sessions = append(sessions, s.Wrap())
	}
	return sessions, nil
}

// appendPagination 辅助方法：追加分页 SQL
func (r *Repository) appendPagination(query string, limit, offset int) string {
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
		if offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", offset)
		}
	}
	return query
}
