import { useState } from "react"
import { useDashboard } from "@/hooks/useDashboard"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { MessageSquare, Users, Database, Clock, Sparkles } from "lucide-react"
import { formatNumber, formatFileSize } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { PersonalInsight } from "@/components/analysis/PersonalInsight"

function StatCard({ title, value, icon: Icon, subtext }: { title: string, value: string, icon: any, subtext?: string }) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">
          {title}
        </CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {subtext && <p className="text-xs text-muted-foreground">{subtext}</p>}
      </CardContent>
    </Card>
  )
}

export default function Dashboard() {
  const { data, isLoading, error } = useDashboard()
  const [showInsight, setShowInsight] = useState(false)

  if (isLoading) {
    return <div className="p-8 text-center text-muted-foreground">加载中...</div>
  }

  if (error || !data) {
    return <div className="p-8 text-center text-destructive">加载失败</div>
  }

  const { overview } = data

  // Helper to format file size from MB
  const formatMB = (mb: number) => formatFileSize(mb * 1024 * 1024)

  return (
    <ScrollArea className="h-full">
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <h2 className="text-3xl font-bold tracking-tight">数据总览</h2>
          <Button 
            onClick={() => setShowInsight(true)}
            className="bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white shadow-lg shadow-pink-500/20 gap-2"
          >
            <Sparkles className="w-4 h-4" />
            生成个人社交报告
          </Button>
        </div>
        
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <StatCard 
            title="消息总数" 
            value={formatNumber(overview.msgStats.total_msgs)} 
            icon={MessageSquare}
            subtext={`发送: ${formatNumber(overview.msgStats.sent_msgs)} / 接收: ${formatNumber(overview.msgStats.received_msgs)}`}
          />
          <StatCard 
            title="群聊/联系人" 
            value={formatNumber(overview.groups.length)} 
            icon={Users}
            subtext="活跃群聊数"
          />
          <StatCard 
            title="存储占用" 
            value={formatMB(overview.dbStats.db_size_mb + overview.dbStats.dir_size_mb)} 
            icon={Database}
            subtext={`DB: ${formatMB(overview.dbStats.db_size_mb)}`}
          />
          <StatCard 
            title="时间跨度" 
            value={`${overview.timeline.duration_days} 天`} 
            icon={Clock}
            subtext="记录覆盖天数"
          />
        </div>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-1">
          <Card className="col-span-1">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Users className="h-5 w-5 text-primary" />
                活跃群聊 Top 10
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-2">
                {overview.groups
                  .sort((a, b) => b.message_count - a.message_count)
                  .slice(0, 10)
                  .map(group => (
                    <div key={group.ChatRoomName} className="flex items-center justify-between p-3 rounded-lg bg-muted/30 border border-border/50">
                      <div className="space-y-1">
                        <p className="text-sm font-bold truncate max-w-[200px]">
                          {group.NickName || group.ChatRoomName}
                        </p>
                        <p className="text-xs text-muted-foreground truncate max-w-[200px]">
                          {group.ChatRoomName}
                        </p>
                      </div>
                      <div className="flex flex-col items-end">
                        <div className="text-lg font-black text-primary">
                          {formatNumber(group.message_count)}
                        </div>
                        <div className="text-[10px] text-muted-foreground uppercase">消息</div>
                      </div>
                    </div>
                  ))
                }
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
      
      {showInsight && (
        <PersonalInsight onClose={() => setShowInsight(false)} />
      )}
    </ScrollArea>
  )
}
