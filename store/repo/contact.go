package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/types"
)

func (r *Repository) GetContacts(ctx context.Context, q types.ContactQuery) ([]*model.Contact, error) {
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
	_ = db.QueryRowContext(ctx, "SELECT 1 FROM sqlite_master WHERE type='table' AND name='contact'").Scan(&exists)
	if exists {
		return r.queryV4Contacts(ctx, db, q)
	}

	// 2. 尝试 V3 模式
	return r.queryV3Contacts(ctx, db, q)
}

func (r *Repository) queryV4Contacts(ctx context.Context, db *sql.DB, q types.ContactQuery) ([]*model.Contact, error) {
	query := `SELECT username, local_type, alias, remark, nick_name, COALESCE(small_head_url,''), COALESCE(big_head_url,'') FROM contact`
	var args []interface{}

	if q.Keyword != "" {
		query += ` WHERE username = ? OR alias = ? OR remark = ? OR nick_name = ?`
		args = append(args, q.Keyword, q.Keyword, q.Keyword, q.Keyword)
	}

	query += ` ORDER BY username`

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

	var contacts []*model.Contact
	for rows.Next() {
		var c model.ContactV4
		err := rows.Scan(&c.UserName, &c.LocalType, &c.Alias, &c.Remark, &c.NickName, &c.SmallHeadURL, &c.BigHeadURL)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, c.Wrap())
	}
	return contacts, nil
}

func (r *Repository) queryV3Contacts(ctx context.Context, db *sql.DB, q types.ContactQuery) ([]*model.Contact, error) {
	query := `SELECT UserName, Alias, Remark, NickName, Reserved1, COALESCE(SmallHeadImgUrl,''), COALESCE(BigHeadImgUrl,'') FROM Contact`
	var args []interface{}

	if q.Keyword != "" {
		query += ` WHERE UserName = ? OR Alias = ? OR Remark = ? OR NickName = ?`
		args = append(args, q.Keyword, q.Keyword, q.Keyword, q.Keyword)
	}

	query += ` ORDER BY UserName`

	if q.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", q.Limit)
		if q.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", q.Offset)
		}
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query contacts failed: %w", err)
	}
	defer rows.Close()

	var contacts []*model.Contact
	for rows.Next() {
		var c model.ContactV3
		err := rows.Scan(&c.UserName, &c.Alias, &c.Remark, &c.NickName, &c.Reserved1, &c.SmallHeadImgUrl, &c.BigHeadImgUrl)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, c.Wrap())
	}
	return contacts, nil
}
