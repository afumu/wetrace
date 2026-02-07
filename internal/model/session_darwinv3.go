package model

import "time"

// CREATE TABLE SessionAbstract(
// m_nsUserName TEXT PRIMARY KEY,
// m_uUnReadCount INTEGER,
// m_bShowUnReadAsRedDot INTEGER,
// m_bMarkUnread INTEGER,
// m_uLastTime INTEGER,
// strRes1 TEXT,
// strRes2 TEXT,
// strRes3 TEXT,
// intRes1 INTEGER,
// intRes2 INTEGER,
// intRes3 INTEGER,
// _packed_MMSessionInfo BLOB
// )
type SessionDarwinV3 struct {
	M_nsUserName string `json:"m_nsUserName"`
	M_uLastTime  int    `json:"m_uLastTime"`

	// M_uUnReadCount        int    `json:"m_uUnReadCount"`
	// M_bShowUnReadAsRedDot int    `json:"m_bShowUnReadAsRedDot"`
	// M_bMarkUnread         int    `json:"m_bMarkUnread"`
	// StrRes1               string `json:"strRes1"`
	// StrRes2               string `json:"strRes2"`
	// StrRes3               string `json:"strRes3"`
	// IntRes1               int    `json:"intRes1"`
	// IntRes2               int    `json:"intRes2"`
	// IntRes3               int    `json:"intRes3"`
	// PackedMMSessionInfo   string `json:"_packed_MMSessionInfo"` // TODO: decode
}

func (s *SessionDarwinV3) Wrap() *Session {
	return &Session{
		UserName: s.M_nsUserName,
		NOrder:   s.M_uLastTime,
		NTime:    time.Unix(int64(s.M_uLastTime), 0),
	}
}
