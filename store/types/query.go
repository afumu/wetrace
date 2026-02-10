package types

import "time"

// MessageQuery 封装了查询消息的参数
type MessageQuery struct {
	StartTime time.Time
	EndTime   time.Time
	Talker    string // 多个对话者可以用逗号分隔
	Sender    string // 多个发送者可以用逗号分隔
	Keyword   string
	MsgType   int // 消息类型筛选，0 表示不限
	Limit     int
	Offset    int
	Reverse   bool // 是否按时间倒序排列
}

// ContactQuery 封装了查询联系人的参数
type ContactQuery struct {
	Keyword string
	Limit   int
	Offset  int
}

// ChatRoomQuery 封装了查询群聊的参数
type ChatRoomQuery struct {
	Keyword string
	Limit   int
	Offset  int
}

// SessionQuery 封装了查询会话的参数
type SessionQuery struct {
	Keyword string
	Limit   int
	Offset  int
}
