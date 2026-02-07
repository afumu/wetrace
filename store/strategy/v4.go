package strategy

import (
	"regexp"
	"strings"
)

type v4Pattern struct {
	group GroupType
	re    *regexp.Regexp
}

// V4 实现 WeChat V4 的策略接口
type V4 struct {
	patterns []v4Pattern
}

// NewV4 创建一个新的策略实例
func NewV4() *V4 {
	return &V4{
		patterns: []v4Pattern{
			{Message, regexp.MustCompile(`(?i)^message(_[0-9]?[0-9])?\.db$`)},
			{Contact, regexp.MustCompile(`(?i)^contact\.db$`)},
			{Image, regexp.MustCompile(`(?i)^hardlink\.db$`)},
			{Video, regexp.MustCompile(`(?i)^hardlink\.db$`)},
			{File, regexp.MustCompile(`(?i)^hardlink\.db$`)},
			{Voice, regexp.MustCompile(`(?i)^media(_[0-9]?[0-9])?\.db$`)},
			{Session, regexp.MustCompile(`(?i)^session\.db$`)},

			// --- 兼容旧版 Windows 文件名 (可选，但也设为不区分大小写) ---
			{Message, regexp.MustCompile(`(?i)^MSG([0-9]?[0-9])?\.db$`)},
			{Contact, regexp.MustCompile(`(?i)^MicroMsg\.db$`)},
		},
	}
}

// Identify 检查文件名是否匹配任何已知模式
func (s *V4) Identify(filename string) (FileMeta, bool) {
	for _, p := range s.patterns {
		matches := p.re.FindStringSubmatch(filename)
		if matches != nil {
			meta := FileMeta{
				Type: p.group,
			}
			// 如果存在索引（第一个捕获组），则提取它
			if len(matches) > 1 && matches[1] != "" {
				// 去掉开头的下划线
				meta.Index = strings.TrimPrefix(matches[1], "_")
			}
			return meta, true
		}
	}
	return FileMeta{Type: Unknown}, false
}
