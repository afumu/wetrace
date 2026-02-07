import { request } from "@/lib/request"
import type { Message, MessageResponse, ChatlogParams, SearchParams } from "@/types"
import { format } from "date-fns"

function transformMessage(response: MessageResponse): Message {
  const createTime = Math.floor(new Date(response.time).getTime() / 1000)
  const id = response.seq
  
  return {
    id,
    seq: response.seq,
    time: response.time,
    createTime,
    talker: response.talker,
    talkerName: response.talkerName,
    talkerAvatar: undefined,
    sender: response.sender,
    senderName: response.senderName,
    isSelf: response.isSelf,
    isSend: response.isSelf ? 1 : 0,
    isChatRoom: response.isChatRoom,
    type: response.type,
    subType: response.subType,
    content: response.content,
    contents: response.contents,
    duration: response.contents?.duration,
    fileName: response.contents?.title,
    fileUrl: response.contents?.url,
    smallHeadURL: response.smallHeadURL,
    bigHeadURL: response.bigHeadURL,
  }
}

function getToday(): string {
  return format(new Date(), 'yyyy-MM-dd')
}

function getDateRange(startDate: Date, endDate: Date): string {
  return `${format(startDate, 'yyyy-MM-dd')}~${format(endDate, 'yyyy-MM-dd')}`
}

export const chatlogApi = {
  getChatlog: async (params: ChatlogParams): Promise<Message[]> => {
    const responses = await request.get<MessageResponse[]>('/api/v1/messages', params)
    return responses.map(transformMessage)
  },

  searchMessages: async (params: SearchParams): Promise<Message[]> => {
    const responses = await request.get<MessageResponse[]>('/api/v1/messages', params)
    return responses.map(transformMessage)
  },

  getSessionMessages: async (talker: string, time?: string, limit = 50, offset = 0, bottom = 0): Promise<Message[]> => {
    return chatlogApi.getChatlog({
      talker_id: talker,
      time: time || getToday(),
      limit,
      offset,
      bottom,
    })
  },

  getMessagesByTime: async (time: string, talker?: string, limit = 50): Promise<Message[]> => {
    return chatlogApi.getChatlog({
      time,
      talker_id: talker,
      limit,
    })
  },

  getMessagesBySender: async (sender: string, time?: string, talker?: string, limit = 50): Promise<Message[]> => {
    return chatlogApi.getChatlog({
      sender,
      time: time || getToday(),
      talker_id: talker,
      limit,
    })
  },

  getTodayMessages: async (talker?: string, limit = 50): Promise<Message[]> => {
    return chatlogApi.getChatlog({
      time: getToday(),
      talker_id: talker,
      limit,
    })
  },

  getRecentMessages: async (days: number, talker?: string, limit = 50): Promise<Message[]> => {
    const endDate = new Date()
    const startDate = new Date()
    startDate.setDate(startDate.getDate() - days)
    
    return chatlogApi.getChatlog({
      time: getDateRange(startDate, endDate),
      talker_id: talker,
      limit,
    })
  },

  getMessagesByDateRange: async (startDate: Date, endDate: Date, talker?: string, limit = 50): Promise<Message[]> => {
    return chatlogApi.getChatlog({
      time: getDateRange(startDate, endDate),
      talker_id: talker,
      limit,
    })
  },

  searchInSession: async (keyword: string, talker: string, limit = 50): Promise<Message[]> => {
    return chatlogApi.searchMessages({
      keyword,
      talker_id: talker,
      limit,
    })
  },

  globalSearch: async (keyword: string, type?: number, limit = 50): Promise<Message[]> => {
    return chatlogApi.searchMessages({
      keyword,
      type,
      limit,
    })
  },

  searchByType: async (type: number, talker?: string, limit = 50): Promise<Message[]> => {
    return chatlogApi.searchMessages({
      keyword: '',
      type,
      talker_id: talker,
      limit,
    })
  }
}
