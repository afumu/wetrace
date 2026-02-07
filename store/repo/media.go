package repo

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/afumu/wetrace/internal/model"
	"github.com/afumu/wetrace/store/strategy"
)

func (r *Repository) GetMedia(ctx context.Context, mediaType string, key string) (*model.Media, error) {
	if key == "" {
		return nil, fmt.Errorf("key 为空")
	}

	// 确定 DB 类型
	var groupType strategy.GroupType
	switch mediaType {
	case "image", "video", "file", "image_merge":
		groupType = strategy.Image // V4 中这些都在 hardlink.db
	case "voice":
		return r.getVoice(ctx, key)
	default:
		return nil, fmt.Errorf("不支持的媒体类型: %s", mediaType)
	}

	// 解析 DB
	dbPath, err := r.router.GetMediaDBPath(groupType)
	if err != nil {
		return nil, err
	}

	db, err := r.pool.GetConnection(dbPath)
	if err != nil {
		return nil, err
	}

	// 1. 尝试 V4 模式
	var table string
	switch mediaType {
	case "image", "image_merge":
		table = "image_hardlink_info_v4"
		if !r.isTableExist(db, table) {
			table = "image_hardlink_info_v3"
		}
	case "video":
		table = "video_hardlink_info_v4"
		if !r.isTableExist(db, table) {
			table = "video_hardlink_info_v3"
		}
	case "file":
		table = "file_hardlink_info_v4"
		if !r.isTableExist(db, table) {
			table = "file_hardlink_info_v3"
		}
	}

	if table != "" && (r.isTableExist(db, table)) {
		return r.queryV4Media(ctx, db, table, mediaType, key)
	}

	// 2. 尝试 V3 模式
	return r.queryV3Media(ctx, db, mediaType, key)
}

func (r *Repository) isTableExist(db *sql.DB, table string) bool {
	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
	return err == nil
}

func (r *Repository) queryV4Media(ctx context.Context, db *sql.DB, table, mediaType, key string) (*model.Media, error) {
	query := fmt.Sprintf(`
	SELECT 
		f.md5,
		f.file_name,
		f.file_size,
		f.modify_time,
		f.extra_buffer,
		IFNULL(d1.username,""),
		IFNULL(d2.username,"")
	FROM 
		%s f
	LEFT JOIN 
		dir2id d1 ON d1.rowid = f.dir1
	LEFT JOIN 
		dir2id d2 ON d2.rowid = f.dir2
	`, table)
	query += " WHERE f.md5 = ? OR f.file_name LIKE ? || '%'"

	rows, err := db.QueryContext(ctx, query, key, key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media *model.Media
	for rows.Next() {
		var m model.MediaV4
		err := rows.Scan(&m.Key, &m.Name, &m.Size, &m.ModifyTime, &m.ExtraBuffer, &m.Dir1, &m.Dir2)
		if err != nil {
			return nil, err
		}
		m.Type = mediaType
		media = m.Wrap()
		if mediaType == "image" && strings.HasSuffix(m.Name, "_h.dat") {
			break
		}
	}
	if media == nil {
		return nil, fmt.Errorf("media not found")
	}
	return media, nil
}

func (r *Repository) queryV3Media(ctx context.Context, db *sql.DB, mediaType, key string) (*model.Media, error) {
	var table1, table2 string
	switch mediaType {
	case "image":
		table1 = "HardLinkImageAttribute"
		table2 = "HardLinkImageID"
	case "video":
		table1 = "HardLinkVideoAttribute"
		table2 = "HardLinkVideoID"
	case "file":
		table1 = "HardLinkFileAttribute"
		table2 = "HardLinkFileID"
	default:
		return nil, fmt.Errorf("unsupported V3 media type")
	}

	// 解码 Key
	md5key, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("解码 key 失败: %w", err)
	}

	query := fmt.Sprintf(`
        SELECT 
            a.FileName,
            a.ModifyTime,
            IFNULL(d1.Dir,"") AS Dir1,
            IFNULL(d2.Dir,"") AS Dir2
        FROM 
            %s a
        LEFT JOIN 
            %s d1 ON a.DirID1 = d1.DirId
        LEFT JOIN 
            %s d2 ON a.DirID2 = d2.DirId
        WHERE 
            a.Md5 = ?
    `, table1, table2, table2)

	row := db.QueryRowContext(ctx, query, md5key)

	var m model.MediaV3
	err = row.Scan(&m.Name, &m.ModifyTime, &m.Dir1, &m.Dir2)
	if err != nil {
		return nil, err
	}

	m.Type = mediaType
	m.Key = key
	return m.Wrap(), nil
}

func (r *Repository) getVoice(ctx context.Context, key string) (*model.Media, error) {
	// 获取所有 Voice 类型的数据库路径
	dbPaths, err := r.router.GetAllDBPaths(strategy.Voice)
	if err != nil {
		return nil, fmt.Errorf("查找语音数据库失败: %w", err)
	}

	query := `
	SELECT voice_data
	FROM VoiceInfo
	WHERE svr_id = ? 
	`

	// 遍历所有数据库进行查询
	for _, dbPath := range dbPaths {
		db, err := r.pool.GetConnection(dbPath)
		if err != nil {
			// 如果获取连接失败，记录日志但不中断，继续尝试下一个
			// 这里可以使用 log.Warn，但由于 Repository 好像没有 log 引用，先忽略或返回错误
			// 为了简单起见，如果连接失败，我们跳过
			continue
		}

		rows, err := db.QueryContext(ctx, query, key)
		if err != nil {
			// 如果查询出错（比如表不存在），继续下一个
			continue
		}
		defer rows.Close()

		for rows.Next() {
			var voiceData []byte
			err := rows.Scan(&voiceData)
			if err != nil {
				return nil, fmt.Errorf("读取语音数据失败: %w", err)
			}

			if len(voiceData) > 0 {
				return &model.Media{
					Type: "voice",
					Key:  key,
					Data: voiceData,
				}, nil
			}
		}
		rows.Close()
	}

	return nil, fmt.Errorf("voice not found: %s", key)
}
