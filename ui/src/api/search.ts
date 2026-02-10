import { request } from "@/lib/request";

export interface SearchItem {
  seq: number;
  time: string;
  talker: string;
  talkerName: string;
  sender: string;
  senderName: string;
  isChatRoom: boolean;
  type: number;
  content: string;
  highlight: string;
}

export interface SearchResponse {
  total: number;
  items: SearchItem[];
}

export interface SearchContextResponse {
  messages: SearchItem[];
  anchor_index: number;
}

export interface SearchParams {
  keyword: string;
  talker?: string;
  sender?: string;
  type?: number;
  time_range?: string;
  limit?: number;
  offset?: number;
}

export const searchApi = {
  search: (params: SearchParams) =>
    request.get<SearchResponse>("/api/v1/search", params),

  getContext: (talker: string, seq: number, before?: number, after?: number) =>
    request.get<SearchContextResponse>("/api/v1/search/context", {
      talker,
      seq,
      before: before ?? 10,
      after: after ?? 10,
    }),
};
