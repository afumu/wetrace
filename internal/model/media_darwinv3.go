package model

import "path/filepath"

// CREATE TABLE HlinkMediaRecord(
// mediaMd5 TEXT,
// mediaSize INTEGER,
// inodeNumber INTEGER,
// modifyTime INTEGER ,
// CONSTRAINT _Md5_Size UNIQUE (mediaMd5,mediaSize)
// )
// CREATE TABLE HlinkMediaDetail(
// localId INTEGER PRIMARY KEY AUTOINCREMENT,
// inodeNumber INTEGER,
// relativePath TEXT,
// fileName TEXT
// )
type MediaDarwinV3 struct {
	MediaMd5     string `json:"mediaMd5"`
	MediaSize    int64  `json:"mediaSize"`
	InodeNumber  int64  `json:"inodeNumber"`
	ModifyTime   int64  `json:"modifyTime"`
	RelativePath string `json:"relativePath"`
	FileName     string `json:"fileName"`
}

func (m *MediaDarwinV3) Wrap() *Media {

	path := filepath.Join("Message/MessageTemp", m.RelativePath, m.FileName)
	name := filepath.Base(path)

	return &Media{
		Type:       "",
		Key:        m.MediaMd5,
		Size:       m.MediaSize,
		ModifyTime: m.ModifyTime,
		Path:       path,
		Name:       name,
	}
}
