package bind

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/afumu/wetrace/store/core"
	"github.com/afumu/wetrace/store/strategy"
	"github.com/rs/zerolog/log"
)

// DatabaseShard 代表一个物理数据库分片 (例如 MSG0.db)
type DatabaseShard struct {
	FilePath  string
	StartTime time.Time
	EndTime   time.Time
	TalkerMap map[string]int // 用户名 -> 内部ID 的缓存
}

// TimelineRouter 负责基于时间线将查询路由到具体的数据库文件
type TimelineRouter struct {
	shards   []*DatabaseShard     // 按时间排序的数据库分片列表
	baseDir  string               // 基础目录
	pool     *core.ConnectionPool // 连接池
	strategy strategy.Strategy    // 文件识别策略
	meta     MetadataReader       // 元数据读取器 (内部工具)
}

// NewTimelineRouter 创建一个新的路由器
func NewTimelineRouter(baseDir string, pool *core.ConnectionPool, strat strategy.Strategy) *TimelineRouter {
	return &TimelineRouter{
		baseDir:  baseDir,
		pool:     pool,
		strategy: strat,
		shards:   make([]*DatabaseShard, 0),
		meta:     MetadataReader{},
	}
}

// RebuildIndex 扫描目录并重建时间线索引
func (r *TimelineRouter) RebuildIndex(ctx context.Context) error {
	log.Info().Str("baseDir", r.baseDir).Msg("开始重建时间线索引...")

	// 1. 发现所有消息数据库文件
	msgFiles, err := r.discoverMessageFiles()
	if err != nil {
		log.Error().Err(err).Msg("发现消息文件失败")
		return err
	}

	log.Info().Int("count", len(msgFiles)).Msg("发现消息分片文件")

	// 2. 加载每个文件的元数据 (Start Time, Talker Map)
	var shards []*DatabaseShard
	for _, path := range msgFiles {
		shard, err := r.loadShardMetadata(ctx, path)
		if err != nil {
			log.Warn().Err(err).Str("file", path).Msg("加载分片元数据失败，跳过")
			continue
		}
		shards = append(shards, shard)
	}

	// 3. 按开始时间排序
	sort.Slice(shards, func(i, j int) bool {
		return shards[i].StartTime.Before(shards[j].StartTime)
	})

	// 4. 计算推导结束时间
	for i := range shards {
		if i == len(shards)-1 {
			shards[i].EndTime = time.Now()
		} else {
			shards[i].EndTime = shards[i+1].StartTime
		}
		log.Debug().
			Str("path", shards[i].FilePath).
			Time("start", shards[i].StartTime).
			Time("end", shards[i].EndTime).
			Msg("分片索引已建立")
	}

	r.shards = shards
	return nil
}

// GetBaseDir 返回基础目录
func (r *TimelineRouter) GetBaseDir() string {
	return r.baseDir
}

// GetShards 返回所有已加载的数据库分片
func (r *TimelineRouter) GetShards() []*DatabaseShard {
	return r.shards
}

// RouteResult 包含路由结果：去哪个文件查，用什么 ID 查
type RouteResult struct {
	FilePath string
	TalkerID int    // 如果在 Name2ID 表中找到了映射
	Talker   string // 原始 Talker (用于回退)
}

// Resolve 根据查询条件 (时间范围, 对话者) 计算出需要查询的目标数据库
func (r *TimelineRouter) Resolve(start, end time.Time, talker string) []RouteResult {
	var results []RouteResult

	for _, shard := range r.shards {
		// 判断时间是否有交集
		// 逻辑：Shard.Start < Query.End AND Shard.End > Query.Start
		if shard.StartTime.Before(end) && shard.EndTime.After(start) {

			result := RouteResult{
				FilePath: shard.FilePath,
				Talker:   talker,
			}

			// 尝试获取优化的 TalkerID
			if id, ok := shard.TalkerMap[talker]; ok {
				result.TalkerID = id
			}

			results = append(results, result)
		}
	}
	return results
}

// GetContactDBPath 获取联系人数据库的路径
func (r *TimelineRouter) GetContactDBPath() (string, error) {
	log.Debug().Msg("尝试查找联系人数据库路径")
	return r.findFileByType(strategy.Contact)
}

// GetSessionDBPath 获取会话数据库的路径
func (r *TimelineRouter) GetSessionDBPath() (string, error) {
	log.Debug().Msg("尝试查找会话数据库路径")
	return r.findFileByType(strategy.Session)
}

// GetMediaDBPath 获取特定类型媒体数据库的路径
func (r *TimelineRouter) GetMediaDBPath(mediaType strategy.GroupType) (string, error) {
	log.Debug().Interface("type", mediaType).Msg("尝试查找媒体数据库路径")
	return r.findFileByType(mediaType)
}

// GetAllDBPaths 获取特定类型的所有数据库路径
func (r *TimelineRouter) GetAllDBPaths(targetType strategy.GroupType) ([]string, error) {
	log.Debug().Interface("type", targetType).Msg("尝试查找所有符合类型的数据库路径")

	// 1. 尝试在对应的子目录中查找 (V4 结构)
	var subDirName string
	switch targetType {
	case strategy.Contact:
		subDirName = "contact"
	case strategy.Image, strategy.Video, strategy.File:
		subDirName = "hardlink"
	case strategy.Message:
		subDirName = "message"
	case strategy.Voice:
		subDirName = "message"
	case strategy.Session:
		subDirName = "session"
	}

	if subDirName != "" {
		subDirPath := filepath.Join(r.baseDir, subDirName)
		found, err := r.scanAllFiles(subDirPath, targetType)
		if err == nil && len(found) > 0 {
			return found, nil
		}
	}

	// 2. 尝试在根目录查找 (Legacy 结构)
	return r.scanAllFiles(r.baseDir, targetType)
}

// --- 内部辅助方法 ---

func (r *TimelineRouter) discoverMessageFiles() ([]string, error) {
	// V4: 检查 message 子目录
	subDir := filepath.Join(r.baseDir, "message")
	files, err := r.scanDirForMessages(subDir)
	if err == nil && len(files) > 0 {
		return files, nil
	}

	// Fallback: 检查根目录
	return r.scanDirForMessages(r.baseDir)
}

func (r *TimelineRouter) scanDirForMessages(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		// 如果目录不存在，不是错误，只是没找到
		return nil, nil
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		meta, match := r.strategy.Identify(entry.Name())
		if match && meta.Type == strategy.Message {
			fullPath := filepath.Join(dir, entry.Name())
			files = append(files, fullPath)
			log.Info().Str("file", entry.Name()).Msg("识别到消息数据库")
		}
	}
	return files, nil
}

func (r *TimelineRouter) loadShardMetadata(ctx context.Context, path string) (*DatabaseShard, error) {
	conn, err := r.pool.GetConnection(path)
	if err != nil {
		return nil, err
	}

	startTime, err := r.meta.ReadStartTime(ctx, conn)
	if err != nil {
		return nil, err
	}

	talkerMap, err := r.meta.ReadTalkerMap(ctx, conn)
	if err != nil {
		return nil, err
	}

	return &DatabaseShard{
		FilePath:  path,
		StartTime: startTime,
		TalkerMap: talkerMap,
	}, nil
}

func (r *TimelineRouter) findFileByType(targetType strategy.GroupType) (string, error) {
	// 1. 尝试在对应的子目录中查找 (V4 结构)
	// 映射关系: Contact -> contact/, Image -> hardlink/, etc.
	var subDirName string
	switch targetType {
	case strategy.Contact:
		subDirName = "contact"
	case strategy.Image, strategy.Video, strategy.File:
		subDirName = "hardlink"
	case strategy.Message:
		subDirName = "message"
	case strategy.Voice:
		subDirName = "message" // 或者 voice? 通常 voice 在 media_x.db
	case strategy.Session:
		subDirName = "session"
	}

	if subDirName != "" {
		subDirPath := filepath.Join(r.baseDir, subDirName)
		found, err := r.scanOneFile(subDirPath, targetType)
		if err == nil {
			return found, nil
		}
	}

	// 2. 尝试在根目录查找 (Legacy 结构)
	return r.scanOneFile(r.baseDir, targetType)
}

func (r *TimelineRouter) scanOneFile(dir string, targetType strategy.GroupType) (string, error) {
	log.Info().Str("dir", dir).Str("target", targetType.String()).Msg("正在扫描目录查找文件")

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		meta, match := r.strategy.Identify(entry.Name())

		log.Debug().
			Str("file", entry.Name()).
			Bool("isMatch", match).
			Str("identifiedType", meta.Type.String()).
			Msg("检查文件")

		if match && meta.Type == targetType {
			fullPath := filepath.Join(dir, entry.Name())
			log.Info().Str("path", fullPath).Msg("找到目标数据库文件")
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("未找到类型为 %s 的数据库文件 (扫描目录: %s)", targetType, dir)
}

func (r *TimelineRouter) scanAllFiles(dir string, targetType strategy.GroupType) ([]string, error) {
	log.Info().Str("dir", dir).Str("target", targetType.String()).Msg("正在扫描目录查找所有文件")

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		meta, match := r.strategy.Identify(entry.Name())

		if match && meta.Type == targetType {
			fullPath := filepath.Join(dir, entry.Name())
			results = append(results, fullPath)
			log.Debug().Str("path", fullPath).Msg("找到目标数据库文件")
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("未找到类型为 %s 的数据库文件 (扫描目录: %s)", targetType, dir)
	}

	return results, nil
}
