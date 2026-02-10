import { request } from "@/lib/request";

export interface SentimentRequest {
  talker: string;
  time_range?: string;
}

export interface EmotionTimelineItem {
  period: string;
  score: number;
  label: string;
  keywords: string[];
}

export interface SentimentResponse {
  overall_score: number;
  overall_label: string;
  relationship_health: string;
  summary: string;
  emotion_timeline: EmotionTimelineItem[];
  sentiment_distribution: {
    positive: number;
    neutral: number;
    negative: number;
  };
  relationship_indicators: {
    initiative_ratio: number;
    response_speed: string;
    intimacy_trend: string;
  };
}

export const sentimentApi = {
  analyze: (data: SentimentRequest) =>
    request.post<SentimentResponse>("/api/v1/ai/sentiment", data),
};
