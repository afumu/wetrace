import { request } from "@/lib/request";

export interface AISummarizeRequest {
  talker: string;
  time_range?: string;
}

export interface AISimulateRequest {
  talker: string;
  message: string;
}

export const aiApi = {
  summarize: (data: AISummarizeRequest) =>
    request.post<string>('/api/v1/ai/summarize', data),
  simulate: (data: AISimulateRequest) =>
    request.post<string>('/api/v1/ai/simulate', data),
};
