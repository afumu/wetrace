import { request } from "@/lib/request"
import { ContactType } from "@/types/contact"
import type { Contact } from "@/types/contact"

interface ChatRoomUser {
  userName: string
  displayName: string
}

interface ChatRoom {
  name: string
  owner: string
  users: ChatRoomUser[]
  remark: string
  nickName: string
}

interface ChatRoomParams {
  keyword?: string
  limit?: number
  offset?: number
}

function transformChatRoom(apiData: ChatRoom): Contact {
  return {
    wxid: apiData.name,
    nickname: apiData.nickName || apiData.name,
    remark: apiData.remark || '',
    alias: '',
    avatar: '', // Chatrooms don't have direct avatar URL in this API
    type: ContactType.Chatroom,
    isStarred: false,
    isPinned: false,
    isMinimized: false,
    bigHeadImgUrl: '',
    smallHeadImgUrl: '',
    headImgMd5: '',
  }
}

export const chatroomApi = {
  getChatRooms: async (params?: ChatRoomParams): Promise<Contact[]> => {
    const response = await request.get<ChatRoom[]>('/api/v1/chatrooms', params)
    
    if (Array.isArray(response)) {
      return response.map(transformChatRoom)
    }

    return []
  },

  getChatRoomDetail: async (id: string): Promise<ChatRoom> => {
    // Assuming REST convention for detail
    return await request.get<ChatRoom>(`/api/v1/chatrooms/${encodeURIComponent(id)}`)
  }
}
