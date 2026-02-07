package model

import "strings"

// CREATE TABLE GroupContact(
// m_nsUsrName TEXT PRIMARY KEY ASC,
// m_uiConType INTEGER,
// nickname TEXT,
// m_nsFullPY TEXT,
// m_nsShortPY TEXT,
// m_nsRemark TEXT,
// m_nsRemarkPYFull TEXT,
// m_nsRemarkPYShort TEXT,
// m_uiCertificationFlag INTEGER,
// m_uiSex INTEGER,
// m_uiType INTEGER,
// m_nsImgStatus TEXT,
// m_uiImgKey INTEGER,
// m_nsHeadImgUrl TEXT,
// m_nsHeadHDImgUrl TEXT,
// m_nsHeadHDMd5 TEXT,
// m_nsChatRoomMemList TEXT,
// m_nsChatRoomAdminList TEXT,
// m_uiChatRoomStatus INTEGER,
// m_nsChatRoomDesc TEXT,
// m_nsDraft TEXT,
// m_nsBrandIconUrl TEXT,
// m_nsGoogleContactName TEXT,
// m_nsAliasName TEXT,
// m_nsEncodeUserName TEXT,
// m_uiChatRoomVersion INTEGER,
// m_uiChatRoomMaxCount INTEGER,
// m_uiChatRoomType INTEGER,
// m_patSuffix TEXT,
// richChatRoomDesc TEXT,
// _packed_WCContactData BLOB,
// openIMInfo BLOB
// )
type ChatRoomDarwinV3 struct {
	M_nsUsrName           string `json:"m_nsUsrName"`
	Nickname              string `json:"nickname"`
	M_nsRemark            string `json:"m_nsRemark"`
	M_nsChatRoomMemList   string `json:"m_nsChatRoomMemList"`
	M_nsChatRoomAdminList string `json:"m_nsChatRoomAdminList"`

	// M_uiConType           int    `json:"m_uiConType"`
	// M_nsFullPY            string `json:"m_nsFullPY"`
	// M_nsShortPY           string `json:"m_nsShortPY"`
	// M_nsRemarkPYFull      string `json:"m_nsRemarkPYFull"`
	// M_nsRemarkPYShort     string `json:"m_nsRemarkPYShort"`
	// M_uiCertificationFlag int    `json:"m_uiCertificationFlag"`
	// M_uiSex               int    `json:"m_uiSex"`
	// M_uiType              int    `json:"m_uiType"`
	// M_nsImgStatus         string `json:"m_nsImgStatus"`
	// M_uiImgKey            int    `json:"m_uiImgKey"`
	// M_nsHeadImgUrl        string `json:"m_nsHeadImgUrl"`
	// M_nsHeadHDImgUrl      string `json:"m_nsHeadHDImgUrl"`
	// M_nsHeadHDMd5         string `json:"m_nsHeadHDMd5"`
	// M_uiChatRoomStatus    int    `json:"m_uiChatRoomStatus"`
	// M_nsChatRoomDesc      string `json:"m_nsChatRoomDesc"`
	// M_nsDraft             string `json:"m_nsDraft"`
	// M_nsBrandIconUrl      string `json:"m_nsBrandIconUrl"`
	// M_nsGoogleContactName string `json:"m_nsGoogleContactName"`
	// M_nsAliasName         string `json:"m_nsAliasName"`
	// M_nsEncodeUserName    string `json:"m_nsEncodeUserName"`
	// M_uiChatRoomVersion   int    `json:"m_uiChatRoomVersion"`
	// M_uiChatRoomMaxCount  int    `json:"m_uiChatRoomMaxCount"`
	// M_uiChatRoomType      int    `json:"m_uiChatRoomType"`
	// M_patSuffix           string `json:"m_patSuffix"`
	// RichChatRoomDesc      string `json:"richChatRoomDesc"`
	// Packed_WCContactData  []byte `json:"_packed_WCContactData"`
	// OpenIMInfo            []byte `json:"openIMInfo"`
}

func (c *ChatRoomDarwinV3) Wrap(user2DisplayName map[string]string) *ChatRoom {

	split := strings.Split(c.M_nsChatRoomMemList, ";")
	users := make([]ChatRoomUser, 0, len(split))
	_user2DisplayName := make(map[string]string)
	for _, v := range split {
		users = append(users, ChatRoomUser{
			UserName: v,
		})
		if name, ok := user2DisplayName[v]; ok {
			_user2DisplayName[v] = name
		}
	}

	return &ChatRoom{
		Name:             c.M_nsUsrName,
		Owner:            c.M_nsChatRoomAdminList,
		Remark:           c.M_nsRemark,
		NickName:         c.Nickname,
		Users:            users,
		User2DisplayName: _user2DisplayName,
	}
}
