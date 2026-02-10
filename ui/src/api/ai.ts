import { request } from "@/lib/request";

export interface AISummarizeRequest {
  talker: string;
  time_range?: string;
}

export interface AISimulateRequest {
  talker: string;
  message: string;
}

export interface AITodosRequest {
  talker: string;
  time_range?: string;
}

export interface TodoItem {
  content: string;
  deadline: string;
  priority: string;
  source_msg: string;
  source_time: string;
}

export interface AITodosResponse {
  todos: TodoItem[];
}

export interface AIExtractRequest {
  talker: string;
  time_range?: string;
  types?: string[];
}

export interface ExtractionItem {
  type: string;
  value: string;
  context: string;
  time: string;
}

export interface AIExtractResponse {
  extractions: ExtractionItem[];
}

export const aiApi = {
  summarize: (data: AISummarizeRequest) =>
    request.post<string>('/api/v1/ai/summarize', data),
  simulate: (data: AISimulateRequest) =>
    request.post<string>('/api/v1/ai/simulate', data),
  extractTodos: (data: AITodosRequest) =>
    request.post<AITodosResponse>('/api/v1/ai/todos', data),
  extractInfo: (data: AIExtractRequest) =>
    request.post<AIExtractResponse>('/api/v1/ai/extract', data),
};
