package model

import (
	"path/filepath"
	"regexp"
	"strings"
)

type MediaV4 struct {
	Type        string `json:"type"`
	Key         string `json:"key"`
	Dir1        string `json:"dir1"`
	Dir2        string `json:"dir2"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	ModifyTime  int64  `json:"modifyTime"`
	ExtraBuffer []byte `json:"extraBuffer"`
}

func (m *MediaV4) Wrap() *Media {

	var path string
	switch m.Type {
	case "image":
		path = filepath.Join("msg", "attach", m.Dir1, m.Dir2, "Img", m.Name)
	case "image_merge":
		// 合并转发的图片，实际文件在 Img 目录下，但没有 .dat 后缀
		// 路径格式：msg/attach/{Dir1}/{Dir2}/Rec/{ExtraBufferRegexExtracted}/Img/{NameNoExt}
		realName := strings.TrimSuffix(m.Name, ".dat")
		recID := parseRecID(m.ExtraBuffer)
		path = filepath.Join("msg", "attach", m.Dir1, m.Dir2, "Rec", recID, "Img", realName)
	case "video":
		path = filepath.Join("msg", "video", m.Dir1, m.Name)
	case "file":
		path = filepath.Join("msg", "file", m.Dir1, m.Name)
	}

	return &Media{
		Type:       m.Type,
		Key:        m.Key,
		Path:       path,
		Name:       m.Name,
		Size:       m.Size,
		ModifyTime: m.ModifyTime,
	}
}

func parseRecID(extra []byte) string {
	if len(extra) == 0 {
		return ""
	}

	// 暴力提取：找出所有连续的字母和数字
	// 通常 ID 是其中最长的一段（如 16 位 Hex 串）
	re := regexp.MustCompile(`[a-zA-Z0-9]+`)
	matches := re.FindAll(extra, -1)

	if len(matches) > 0 {
		longest := ""
		for _, m := range matches {
			if len(m) > len(longest) {
				longest = string(m)
			}
		}
		return longest
	}

	return ""
}
