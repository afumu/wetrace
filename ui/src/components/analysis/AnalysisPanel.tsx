import { useAnalysis } from "@/hooks/useAnalysis";
import { X, BarChart3, TrendingUp, Users, MessageSquare } from "lucide-react";
import { Button } from "@/components/ui/button";
import { HourlyChart } from "./HourlyChart";
import { DailyChart } from "./DailyChart";
import { MemberRank } from "./MemberRank";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";

interface Props {
  talker: string;
  onClose: () => void;
}

export function AnalysisPanel({ talker, onClose }: Props) {
  const { hourly, daily, members, isLoading } = useAnalysis(talker);

  return (
    <div className="fixed inset-0 z-50 bg-background flex flex-col animate-in fade-in slide-in-from-bottom-4 duration-300">
      <div className="h-14 border-b flex items-center justify-between px-6 bg-card/50 backdrop-blur-md sticky top-0 z-10">
        <div className="flex items-center gap-2">
          <BarChart3 className="w-5 h-5 text-pink-500" />
          <h2 className="font-semibold text-lg">数据分析报告</h2>
        </div>
        <Button variant="ghost" size="icon" onClick={onClose} className="rounded-full">
          <X className="w-5 h-5" />
        </Button>
      </div>

      <ScrollArea className="flex-1">
        <div className="max-w-6xl mx-auto p-6 space-y-6 pb-20">
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <div className="flex flex-col items-center gap-4">
                <div className="w-10 h-10 border-4 border-pink-500 border-t-transparent rounded-full animate-spin" />
                <p className="text-muted-foreground animate-pulse">正在深度挖掘数据中...</p>
              </div>
            </div>
          ) : (
            <>
              {/* Top Grid: Basic Stats */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">总消息数</CardTitle>
                    <MessageSquare className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {(daily.data?.reduce((sum, d) => sum + d.count, 0) || 0).toLocaleString()}
                    </div>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">活跃天数</CardTitle>
                    <TrendingUp className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{daily.data?.length || 0} 天</div>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">参与人数</CardTitle>
                    <Users className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{members.data?.length || 0} 位</div>
                  </CardContent>
                </Card>
              </div>

              <div className="grid grid-cols-1 gap-6">
                {/* Hourly Activity */}
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <BarChart3 className="w-4 h-4 text-pink-500" />
                      24小时活跃分布
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <HourlyChart data={hourly.data || []} />
                  </CardContent>
                </Card>
              </div>

              {/* Daily Trend */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <TrendingUp className="w-4 h-4 text-pink-500" />
                    每日消息趋势
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <DailyChart data={daily.data || []} />
                </CardContent>
              </Card>

              <div className="grid grid-cols-1 gap-6">
                {/* Member Ranking */}
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <Users className="w-4 h-4 text-pink-500" />
                      活跃度排行榜 (Top 10)
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <MemberRank data={members.data?.slice(0, 10) || []} />
                  </CardContent>
                </Card>
              </div>
            </>
          )}
        </div>
      </ScrollArea>
    </div>
  );
}
