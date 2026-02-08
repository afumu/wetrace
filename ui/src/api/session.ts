import { request } from "@/lib/request"
import type { Session, SessionParams } from "@/types"

interface SessionApiResponse {
  userName: string
  nOrder: number
  nickName: string
  content: string
  nTime: string
  smallHeadURL: string
}

function transformSession(apiData: SessionApiResponse): Session {
  // Check if fields exist before using them, providing defaults if missing
  const userName = apiData.userName || ''
  const nickName = apiData.nickName || ''
  
  const isChatRoom = userName.includes('@chatroom')
  const isOfficialAccount = userName.startsWith('gh_')
  const isHolder = userName.includes("@placeholder_foldgroup") || userName.includes("brandsessionholder") || userName.includes("brandservicesessionholder")
  const isPrivate = !isChatRoom && !isOfficialAccount && !isHolder

  let session_type: 'group' | 'private' | 'official' | 'unknown' = 'unknown'
  if (isChatRoom) {
    session_type = 'group'
  } else if (isPrivate) {
    session_type = 'private'
  } else if(isOfficialAccount) {
    session_type = 'official'
  } else {
    session_type = 'unknown'
  }

  function getSessionName(userName: string, nickName: string, session_type: Session['type']): string {
    if (userName.includes("@placeholder_foldgroup")) {
      return '【折叠群聊】'
    }
    if (userName === 'brandsessionholder') {
      return '【公众号】'
    }
    if (userName === 'brandservicesessionholder') {
      return '服务号'
    }

    switch (session_type) {
      case 'group':
        return `${nickName}` || `群聊(无名)`
      case 'official':
        return nickName || `${userName}`
      case 'private':
        return nickName || userName
      default:
        return nickName || userName
    }
  }

  return {
    id: userName,
    talker: userName,
    talkerName: nickName || userName,
    name: getSessionName(userName, nickName, session_type),
    avatar: apiData.smallHeadURL || '',
    smallHeadURL: apiData.smallHeadURL || '',
    remark: '',
    type: session_type,
    lastMessage: (apiData.content || nickName) ? {
      nickName: nickName,
      content: apiData.content,
      createTime: apiData.nTime ? new Date(apiData.nTime).getTime() / 1000 : 0,
      type: 1,
    } : undefined,
    lastTime: apiData.nTime,
    lastMessageType: 1,
    unreadCount: 0,
    isPinned: false,
    isMinimized: false,
    isChatRoom: isChatRoom,
    messageCount: 0,
  }
}

export const sessionApi = {
  getSessions: async (params?: SessionParams) => {
    const response = await request.get<SessionApiResponse[]>('/api/v1/sessions', params)
    
    let items: Session[] = []
    let total = 0

    if (Array.isArray(response)) {
      items = response.map(transformSession)
      total = items.length
    }

    return { items, total }
  },

  getSessionDetail: async (talker: string) => {
    // Assuming REST convention for detail
    const response = await request.get<SessionApiResponse>(`/api/v1/sessions/${encodeURIComponent(talker)}`)
    return transformSession(response)
  },

  deleteSession: async (talker: string) => {
    return await request.delete(`/api/v1/sessions/${encodeURIComponent(talker)}`)
  }
}
