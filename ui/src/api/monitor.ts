import { request } from "@/lib/request"

export interface MonitorConfig {
  id: number
  name: string
  type: "keyword" | "ai"
  prompt: string
  keywords: string[]
  platform: "webhook" | "feishu"
  webhook_url: string
  feishu_url: string
  enabled: boolean
  session_ids: string[]
  interval_minutes: number
  last_check_time: number
  created_at: number
  updated_at: number
}

export interface MonitorConfigCreate {
  name: string
  type: "keyword" | "ai"
  prompt?: string
  keywords?: string[]
  platform: "webhook" | "feishu"
  webhook_url?: string
  feishu_url?: string
  enabled: boolean
  session_ids?: string[]
  interval_minutes?: number
}

export interface FeishuConfig {
  bot_webhook: string
  sign_secret: string
  enabled: boolean
  app_id: string
  app_secret: string
  app_token: string
  table_id: string
  push_type: "bot" | "bitable" | "both"
}

export interface FeishuConfigUpdate {
  bot_webhook: string
  sign_secret?: string
  enabled: boolean
  app_id?: string
  app_secret?: string
  app_token?: string
  table_id?: string
  push_type?: "bot" | "bitable" | "both"
}

export const monitorApi = {
  // Monitor configs CRUD
  getConfigs: () =>
    request.get<MonitorConfig[]>("/api/v1/monitor/configs"),

  createConfig: (data: MonitorConfigCreate) =>
    request.post<{ id: number }>("/api/v1/monitor/configs", data),

  updateConfig: (id: number, data: MonitorConfigCreate) =>
    request.put("/api/v1/monitor/configs/" + id, data),

  deleteConfig: (id: number) =>
    request.delete("/api/v1/monitor/configs/" + id),

  testPush: (data: { url: string; secret?: string }) =>
    request.post<{ status: string; response_code: number }>("/api/v1/monitor/test", data),

  // Feishu config
  getFeishuConfig: () =>
    request.get<FeishuConfig>("/api/v1/feishu/config"),

  updateFeishuConfig: (data: FeishuConfigUpdate) =>
    request.put("/api/v1/feishu/config", data),

  testFeishu: () =>
    request.post<{ status: string; message: string }>("/api/v1/feishu/test"),

  testFeishuBitable: () =>
    request.post<{ status: string; message: string }>("/api/v1/feishu/test_bitable"),
}
