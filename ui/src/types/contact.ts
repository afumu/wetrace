/**
 * 联系人类型枚举
 */
export enum ContactType {
  Friend = 'friend',
  Chatroom = 'chatroom',
  Official = 'official',
  Enterprise = 'enterprise',
}

/**
 * 联系人接口
 */
export interface Contact {
  wxid: string
  nickname: string
  remark: string
  alias: string
  avatar: string
  type: ContactType
  gender?: number
  province?: string
  city?: string
  signature?: string
  labelIds?: string[]
  labels?: string[]
  isStarred?: boolean
  starredAt?: number
  lastContactTime?: number
  memberList?: string[]
  // 后端新增字段
  isPinned?: boolean
  isMinimized?: boolean
  bigHeadImgUrl?: string
  smallHeadImgUrl?: string
  headImgMd5?: string
  // 索引和排序相关字段（缓存）
  pinyinInitial?: string
  sortKey?: string
}

/**
 * 群聊接口
 */
export interface Chatroom {
  chatroomId: string
  name: string
  avatar: string
  memberCount: number
  owner: string
  members: ChatroomMember[]
}

/**
 * 群聊成员接口
 */
export interface ChatroomMember {
  wxid: string
  nickname: string
  displayName: string
  avatar?: string
}

/**
 * 联系人标签接口
 */
export interface ContactLabel {
  id: string
  name: string
  color?: string
}

/**
 * 性别枚举
 */
export enum Gender {
  Unknown = 0,
  Male = 1,
  Female = 2,
}

/**
 * 性别显示名称
 */
export const GenderNames: Record<Gender, string> = {
  [Gender.Unknown]: '未知',
  [Gender.Male]: '男',
  [Gender.Female]: '女',
}

/**
 * 群聊 API 响应接口
 */
export interface ChatroomApiResponse {
  items: ChatroomApiItem[]
}

/**
 * 群聊 API 单项数据
 */
export interface ChatroomApiItem {
  name: string
  owner: string
  users: ChatroomUserApiItem[]
  remark: string
  nickName: string
}

/**
 * 群聊用户 API 数据
 */
export interface ChatroomUserApiItem {
  userName: string
  displayName: string
}