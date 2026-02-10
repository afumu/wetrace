import { request } from "@/lib/request";

export interface AnnualOverview {
  total_messages: number;
  sent_messages: number;
  received_messages: number;
  total_contacts: number;
  active_contacts: number;
  total_chatrooms: number;
  active_chatrooms: number;
  first_message_date: string;
  last_message_date: string;
  active_days: number;
}

export interface AnnualHighlights {
  busiest_day: { date: string; count: number };
  quietest_day: { date: string; count: number };
  longest_streak: number;
  late_night_count: number;
  earliest_message_time: string;
  latest_message_time: string;
}

export interface AnnualReport {
  year: number;
  overview: AnnualOverview;
  top_contacts: Array<{
    talker: string;
    name: string;
    avatar: string;
    messageCount: number;
    sentCount: number;
    recvCount: number;
  }>;
  monthly_trend: Array<{ month: number; count: number }>;
  weekday_distribution: Array<{ weekday: number; count: number }>;
  hourly_distribution: Array<{ hour: number; count: number }>;
  message_types: Record<string, number>;
  highlights: AnnualHighlights;
}

export const reportApi = {
  getAnnualReport: (year?: number) =>
    request.get<AnnualReport>("/api/v1/report/annual", year ? { year } : {}),
};
