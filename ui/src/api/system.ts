import { request } from "@/lib/request";
import type { AxiosRequestConfig } from "axios";

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
};