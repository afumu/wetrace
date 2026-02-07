package model

import (
	"github.com/afumu/wetrace/internal/model/wxproto"

	"google.golang.org/protobuf/proto"
)

type ChatRoom struct {
	Name  string         `json:"name"`
	Owner string         `json:"owner"`
	Users []ChatRoomUser `json:"users"`

	// Extra From Contact
	Remark   string `json:"remark"`
	NickName string `json:"nickName"`

	User2DisplayName map[string]string `json:"-"`
}

type ChatRoomUser struct {
	UserName    string `json:"userName"`
	DisplayName string `json:"displayName"`
}

// CREATE TABLE ChatRoom(
// ChatRoomName TEXT PRIMARY KEY,
// UserNameList TEXT,
// DisplayNameList TEXT,
// ChatRoomFlag int Default 0,
// Owner INTEGER DEFAULT 0,
// IsShowName INTEGER DEFAULT 0,
// SelfDisplayName TEXT,
// Reserved1 INTEGER DEFAULT 0,
// Reserved2 TEXT,
// Reserved3 INTEGER DEFAULT 0,
// Reserved4 TEXT,
// Reserved5 INTEGER DEFAULT 0,
// Reserved6 TEXT,
// RoomData BLOB,
// Reserved7 INTEGER DEFAULT 0,
// Reserved8 TEXT
// )
type ChatRoomV3 struct {
	ChatRoomName string `json:"ChatRoomName"`
	Reserved2    string `json:"Reserved2"` // Creator
	RoomData     []byte `json:"RoomData"`

	// // 非关键信息，暂时忽略
	// UserNameList    string `json:"UserNameList"`
	// DisplayNameList string `json:"DisplayNameList"`
	// ChatRoomFlag    int    `json:"ChatRoomFlag"`
	// Owner           int    `json:"Owner"`
	// IsShowName      int    `json:"IsShowName"`
	// SelfDisplayName string `json:"SelfDisplayName"`
	// Reserved1       int    `json:"Reserved1"`
	// Reserved3       int    `json:"Reserved3"`
	// Reserved4       string `json:"Reserved4"`
	// Reserved5       int    `json:"Reserved5"`
	// Reserved6       string `json:"Reserved6"`
	// Reserved7       int    `json:"Reserved7"`
	// Reserved8       string `json:"Reserved8"`
}

func (c *ChatRoomV3) Wrap() *ChatRoom {

	var users []ChatRoomUser
	if len(c.RoomData) != 0 {
		users = ParseRoomData(c.RoomData)
	}

	user2DisplayName := make(map[string]string, len(users))
	for _, user := range users {
		if user.DisplayName != "" {
			user2DisplayName[user.UserName] = user.DisplayName
		}
	}

	return &ChatRoom{
		Name:             c.ChatRoomName,
		Owner:            c.Reserved2,
		Users:            users,
		User2DisplayName: user2DisplayName,
	}
}

func ParseRoomData(b []byte) (users []ChatRoomUser) {
	var pbMsg wxproto.RoomData
	if err := proto.Unmarshal(b, &pbMsg); err != nil {
		return
	}
	if pbMsg.Users == nil {
		return
	}

	users = make([]ChatRoomUser, 0, len(pbMsg.Users))
	for _, user := range pbMsg.Users {
		u := ChatRoomUser{UserName: user.UserName}
		if user.DisplayName != nil {
			u.DisplayName = *user.DisplayName
		}
		users = append(users, u)
	}
	return users
}

func (c *ChatRoom) DisplayName() string {
	switch {
	case c.Remark != "":
		return c.Remark
	case c.NickName != "":
		return c.NickName
	}
	return ""
}
