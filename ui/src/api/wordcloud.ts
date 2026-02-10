import { request } from "@/lib/request";

export interface WordItem {
  text: string;
  count: number;
}

export interface WordCloudResponse {
  total_messages: number;
  total_words: number;
  words: WordItem[];
}

export const wordcloudApi = {
  getWordCloud: (id: string, params?: { time_range?: string; limit?: number; sender?: string }) =>
    request.get<WordCloudResponse>(`/api/v1/analysis/wordcloud/${id}`, params),

  getGlobalWordCloud: (params?: { time_range?: string; limit?: number }) =>
    request.get<WordCloudResponse>("/api/v1/analysis/wordcloud/global", params),
};
