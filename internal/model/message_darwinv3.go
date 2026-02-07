package model

import (
	"strings"
	"time"
)

// CREATE TABLE Chat_md5(talker)(
// mesLocalID INTEGER PRIMARY KEY AUTOINCREMENT,
// mesSvrID INTEGER,msgCreateTime INTEGER,
// msgContent TEXT,msgStatus INTEGER,
// msgImgStatus INTEGER,
// messageType INTEGER,
// mesDes INTEGER,
// msgSource TEXT,
// IntRes1 INTEGER,
// IntRes2 INTEGER,
// StrRes1 TEXT,
// StrRes2 TEXT,
// msgVoiceText TEXT,
// msgSeq INTEGER,
// CompressContent BLOB,
// ConBlob BLOB
// )
type MessageDarwinV3 struct {
	MsgCreateTime int64  `json:"msgCreateTime"`
	MsgContent    string `json:"msgContent"`
	MessageType   int64  `json:"messageType"`
	MesDes        int    `json:"mesDes"` // 0: 发送, 1: 接收
}

func (m *MessageDarwinV3) Wrap(talker string) *Message {

	_m := &Message{
		Time:       time.Unix(m.MsgCreateTime, 0),
		Type:       m.MessageType,
		Talker:     talker,
		IsChatRoom: strings.HasSuffix(talker, "@chatroom"),
		IsSelf:     m.MesDes == 0,
		Version:    WeChatDarwinV3,
	}

	content := m.MsgContent
	if _m.IsChatRoom {
		split := strings.SplitN(content, ":\n", 2)
		if len(split) == 2 {
			_m.Sender = split[0]
			content = split[1]
		}
	} else if !_m.IsSelf {
		_m.Sender = talker
	}

	_m.ParseMediaInfo(content)

	return _m
}
