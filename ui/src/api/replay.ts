import { request, getApiBaseUrl } from "@/lib/request"

export interface ReplayMessage {
  seq: number
  time: string
  sender: string
  senderName: string
  isSelf: boolean
  type: number
  content: string
}

export interface ReplayData {
  total: number
  messages: ReplayMessage[]
}

export interface ReplayExportRequest {
  talker_id: string
  start_date?: string
  end_date?: string
  format?: "mp4" | "gif"
  speed?: number
  resolution?: "720p" | "1080p"
}

export interface ReplayExportStatus {
  task_id: string
  status: "pending" | "processing" | "completed" | "failed"
  progress: number
  total_frames: number
  processed_frames: number
}

export const replayApi = {
  getMessages: (params: {
    talker_id: string
    start_date?: string
    end_date?: string
    limit?: number
    offset?: number
  }) => request.get<ReplayData>("/api/v1/messages/replay", params),

  createExport: (data: ReplayExportRequest) =>
    request.post<{ task_id: string; status: string; message: string }>(
      "/api/v1/export/replay",
      data
    ),

  getExportStatus: (taskId: string) =>
    request.get<ReplayExportStatus>(
      "/api/v1/export/replay/status/" + taskId
    ),

  getExportDownloadUrl: (taskId: string): string => {
    const baseURL = getApiBaseUrl()
    return `${baseURL}/api/v1/export/replay/download/${taskId}`
  },
}
