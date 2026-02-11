import { request } from "@/lib/request"
import type { MessageResponse, Message } from "@/types"

export interface ReplayData {
  total: number
  messages: MessageResponse[]
}

function transformReplayMessage(response: MessageResponse): Message {
  const createTime = Math.floor(new Date(response.time).getTime() / 1000)
  return {
    id: response.seq,
    seq: response.seq,
    time: response.time,
    createTime,
    talker: response.talker,
    talkerName: response.talkerName,
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

export const replayApi = {
  getMessages: async (params: {
    talker_id: string
    start_date?: string
    end_date?: string
    limit?: number
    offset?: number
  }): Promise<{ total: number; messages: Message[] }> => {
    const data = await request.get<ReplayData>("/api/v1/messages/replay", params)
    return {
      total: data.total,
      messages: (data.messages || []).map(transformReplayMessage),
    }
  },
}
