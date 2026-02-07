/**
 * 会话接口
 */
export interface Session {
  id: string
  talker: string
  talkerName: string
  name?: string
  avatar: string
  smallHeadURL?: string
  remark?: string
  type?: 'private' | 'group' | 'official' | 'unknown'
  lastMessage?: {
    nickName: string
    content: string
    createTime: number
    type: number
  }
  lastTime: string
  lastMessageType: number
  unreadCount: number
  isPinned: boolean
  isLocalPinned?: boolean
  isMinimized: boolean
  isChatRoom: boolean
  messageCount: number
}