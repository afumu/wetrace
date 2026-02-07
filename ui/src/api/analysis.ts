import { request } from "@/lib/request";

export interface HourlyStat {
  hour: number;
  count: number;
}

export interface DailyStat {
  date: string;
  count: number;
}

export interface WeekdayStat {
  weekday: number;
  count: number;
}

export interface MonthlyStat {
  month: number;
  count: number;
}

export interface MessageTypeStat {
  type: number;
  count: number;
}

export interface MemberActivity {
  memberId: number;
  platformId: string;
  name: string;
  messageCount: number;
  avatar?: string;
}

export interface RepeatStat {
  content: string;
  count: number;
  memberName: string;
}

export interface PersonalTopContact {
  talker: string;
  name: string;
  avatar: string;
  messageCount: number;
  sentCount: number;
  recvCount: number;
  lastTime: number;
}

export const analysisApi = {
  getHourly: (id: string) => request.get<HourlyStat[]>(`/api/v1/analysis/hourly/${id}`),
  getDaily: (id: string) => request.get<DailyStat[]>(`/api/v1/analysis/daily/${id}`),
  getWeekday: (id: string) => request.get<WeekdayStat[]>(`/api/v1/analysis/weekday/${id}`),
  getMonthly: (id: string) => request.get<MonthlyStat[]>(`/api/v1/analysis/monthly/${id}`),
  getTypeDistribution: (id: string) => request.get<MessageTypeStat[]>(`/api/v1/analysis/type_distribution/${id}`),
  getMemberActivity: (id: string) => request.get<MemberActivity[]>(`/api/v1/analysis/member_activity/${id}`),
  getRepeat: (id: string) => request.get<RepeatStat[]>(`/api/v1/analysis/repeat/${id}`),
  getPersonalTopContacts: () => request.get<PersonalTopContact[]>('/api/v1/analysis/personal/top_contacts'),
};
