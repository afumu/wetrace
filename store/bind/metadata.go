package bind

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// MetadataReader 封装了读取数据库元信息的 SQL 操作
type MetadataReader struct{}

// ReadStartTime 从数据库中读取开始时间 (兼容 V3 和 V4)
func (r MetadataReader) ReadStartTime(ctx context.Context, db *sql.DB) (time.Time, error) {
	// 1. 尝试 V4 的 Timestamp 表
	var ts int64
	err := db.QueryRowContext(ctx, "SELECT timestamp FROM Timestamp LIMIT 1").Scan(&ts)
	if err == nil {
		return time.Unix(ts, 0), nil
	}

	// 2. 尝试 V3 的 DBInfo 表
	query := "SELECT tableIndex, tableVersion, tableDesc FROM DBInfo"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return time.Time{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			idx     int
			version int64
			desc    string
		)
		if err := rows.Scan(&idx, &version, &desc); err != nil {
			continue
		}

		if strings.Contains(desc, "Start Time") {
			return time.Unix(version/1000, (version%1000)*1000000), nil
		}
	}
	return time.Time{}, fmt.Errorf("未在数据库中找到开始时间记录")
}

// ReadTalkerMap 从 Name2ID 表中读取用户名到内部 ID 的映射 (主要用于 V3)
func (r MetadataReader) ReadTalkerMap(ctx context.Context, db *sql.DB) (map[string]int, error) {
	// V4 逻辑不同，可能不需要此映射，或者表名不同
	query := "SELECT UsrName FROM Name2ID"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		// 如果表不存在，返回空映射而不是错误，以支持 V4
		return make(map[string]int), nil
	}
	defer rows.Close()

	mapping := make(map[string]int)
	id := 1

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		mapping[name] = id
		id++
	}
	return mapping, nil
}
