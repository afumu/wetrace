import { request, getApiBaseUrl } from "@/lib/request";
import type { AxiosRequestConfig } from "axios";

export interface ComplianceStatus {
  agreed: boolean;
  agreed_at: string;
  version: string;
}

export interface AIConfig {
  enabled: boolean;
  provider: string;
  model: string;
  base_url: string;
  api_key_masked: string;
}

export interface AIConfigUpdate {
  enabled: boolean;
  provider?: string;
  model?: string;
  base_url?: string;
  api_key?: string;
}

export interface SyncConfig {
  enabled: boolean;
  interval_minutes: number;
  last_sync_time: string;
  last_sync_status: string;
  is_syncing: boolean;
}

export interface SyncConfigUpdate {
  enabled: boolean;
  interval_minutes: number;
}

export interface PasswordStatus {
  enabled: boolean;
  is_locked: boolean;
}

export interface BackupConfig {
  enabled: boolean;
  interval_hours: number;
  backup_path: string;
  format: string;
  last_backup_time: string;
  last_backup_status: string;
}

export interface BackupConfigUpdate {
  enabled: boolean;
  interval_hours: number;
  backup_path: string;
  format?: string;
}

export interface BackupHistoryItem {
  id: string;
  time: string;
  status: string;
  file_path: string;
  file_size: number;
  sessions_count: number;
}

export interface TTSConfig {
  enabled: boolean;
  provider: string;
  base_url: string;
  api_key_masked: string;
  model: string;
  voice: string;
  speed: number;
  format: string;
}

export interface TTSConfigUpdate {
  enabled: boolean;
  provider?: string;
  base_url?: string;
  api_key?: string;
  model?: string;
  voice?: string;
  speed?: number;
  format?: string;
}

export const systemApi = {
  decrypt: () => request.post("/api/v1/system/decrypt"),
  getStatus: () => request.get("/api/v1/system/status"),
  getWeChatDbKey: (config?: AxiosRequestConfig) =>
    request.get("/api/v1/system/wxkey/db", {}, { timeout: 180000, ...config }),
  getWeChatImageKey: (config?: AxiosRequestConfig) =>
    request.get("/api/v1/system/wxkey/image", {}, { timeout: 300000, ...config }),
  activate: (license: string) => request.post("/api/v1/system/activate", { license }),
  detectWeChatPath: () => request.get("/api/v1/system/detect/wechat_path"),
  detectDbPath: () => request.get("/api/v1/system/detect/db_path"),
  selectPath: (type: 'file' | 'folder') => request.post("/api/v1/system/select_path", { type }),
  updateConfig: (data: Record<string, string>) => request.post("/api/v1/system/config", data),
  getCompliance: () => request.get<ComplianceStatus>("/api/v1/system/compliance"),
  agreeCompliance: (version: string) => request.post("/api/v1/system/compliance/agree", { version }),

  // AI Config
  getAIConfig: () => request.get<AIConfig>("/api/v1/system/ai_config"),
  updateAIConfig: (data: AIConfigUpdate) => request.post("/api/v1/system/ai_config", data),
  testAIConfig: () => request.post<{ status: string; model: string; latency_ms: number }>("/api/v1/system/ai_config/test"),

  // Sync Config
  getSyncConfig: () => request.get<SyncConfig>("/api/v1/system/sync_config"),
  updateSyncConfig: (data: SyncConfigUpdate) => request.post("/api/v1/system/sync_config", data),
  triggerSync: () => request.post("/api/v1/system/sync"),

  // Password
  getPasswordStatus: () => request.get<PasswordStatus>("/api/v1/system/password/status"),
  setPassword: (old_password: string, new_password: string) =>
    request.post("/api/v1/system/password/set", { old_password, new_password }),
  verifyPassword: (password: string) =>
    request.post("/api/v1/system/password/verify", { password }),
  disablePassword: (password: string) =>
    request.post("/api/v1/system/password/disable", { password }),

  // Backup Config
  getBackupConfig: () => request.get<BackupConfig>("/api/v1/system/backup_config"),
  updateBackupConfig: (data: BackupConfigUpdate) => request.post("/api/v1/system/backup_config", data),
  runBackup: () => request.post("/api/v1/system/backup/run"),
  getBackupHistory: (limit = 20, offset = 0) =>
    request.get<BackupHistoryItem[]>("/api/v1/system/backup/history", { limit, offset }),

  // TTS Config
  getTTSConfig: () => request.get<TTSConfig>("/api/v1/system/tts_config"),
  updateTTSConfig: (data: TTSConfigUpdate) => request.post("/api/v1/system/tts_config", data),
  testTTSConfigUrl: (text?: string, voice?: string, speed?: number): string => {
    const baseURL = getApiBaseUrl();
    const params = new URLSearchParams();
    if (text) params.set("text", text);
    if (voice) params.set("voice", voice);
    if (speed !== undefined) params.set("speed", String(speed));
    return `${baseURL}/api/v1/system/tts_config/test?${params.toString()}`;
  },
};