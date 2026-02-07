import { useState } from "react"
import { usePersonalAnalysis } from "@/hooks/useAnalysis"
import { X, Trophy, MessageCircle, Heart, Calendar } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { mediaApi } from "@/api/media"
import { format } from "date-fns"
import { cn } from "@/lib/utils"
import { 
  BarChart, 
  Bar, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  Cell,
  PieChart,
  Pie
} from 'recharts'

interface Props {
  onClose: () => void
}

export function PersonalInsight({ onClose }: Props) {
  const { topContacts, isLoading } = usePersonalAnalysis()
  const [showAllContacts, setShowAllContacts] = useState(false)

  if (isLoading) {
    return (
      <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-md flex items-center justify-center">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin" />
          <p className="text-muted-foreground animate-pulse font-medium">深度解析你的社交网络中...</p>
        </div>
      </div>
    )
  }

  const data = topContacts.data || []
  const totalMsgs = data.reduce((sum, item) => sum + item.messageCount, 0)
  const totalSent = data.reduce((sum, item) => sum + item.sentCount, 0)
  const totalRecv = data.reduce((sum, item) => sum + item.recvCount, 0)

  const top3 = data.slice(0, 3)
  const chartData = data.slice(0, 10).map(item => ({
    name: item.name.length > 6 ? item.name.substring(0, 6) + '...' : item.name,
    count: item.messageCount,
    sent: item.sentCount,
    recv: item.recvCount
  }))

  const pieData = [
    { name: '我发送的', value: totalSent, color: '#ec4899' },
    { name: '我接收的', value: totalRecv, color: '#6366f1' }
  ]

  return (
    <div className="fixed inset-0 z-50 bg-background flex flex-col animate-in fade-in slide-in-from-right-4 duration-300">
      {/* Header */}
      <div className="h-16 border-b flex items-center justify-between px-8 bg-card/50 backdrop-blur-md sticky top-0 z-10">
        <div className="flex items-center gap-3">
          <div className="bg-gradient-to-br from-pink-500 to-violet-500 p-2 rounded-lg text-white shadow-lg shadow-pink-500/20">
            <Trophy className="w-5 h-5" />
          </div>
          <div>
            <h2 className="font-bold text-xl tracking-tight">年度社交洞察报告</h2>
            <p className="text-xs text-muted-foreground">基于本地所有历史记录的深度分析</p>
          </div>
        </div>
        <Button variant="ghost" size="icon" onClick={onClose} className="rounded-full hover:bg-muted/80">
          <X className="w-5 h-5" />
        </Button>
      </div>

      <ScrollArea className="flex-1 bg-slate-50/50 dark:bg-slate-950/20">
        <div className="max-w-5xl mx-auto p-8 space-y-8 pb-20">
          
          {/* Summary Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <Card className="border-none shadow-md bg-gradient-to-br from-pink-50 to-white dark:from-pink-950/20">
              <CardContent className="pt-6">
                <div className="flex items-center gap-4">
                  <div className="p-3 bg-pink-100 dark:bg-pink-900/30 rounded-xl text-pink-600">
                    <MessageCircle className="w-6 h-6" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">累计互动消息</p>
                    <h3 className="text-3xl font-black text-pink-600">{totalMsgs.toLocaleString()}</h3>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card 
              className="border-none shadow-md bg-gradient-to-br from-violet-50 to-white dark:from-violet-950/20 cursor-pointer hover:ring-2 hover:ring-violet-500/30 transition-all active:scale-95 group"
              onClick={() => setShowAllContacts(true)}
            >
              <CardContent className="pt-6">
                <div className="flex items-center gap-4">
                  <div className="p-3 bg-violet-100 dark:bg-violet-900/30 rounded-xl text-violet-600 group-hover:bg-violet-600 group-hover:text-white transition-colors">
                    <Heart className="w-6 h-6" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">深度互动联系人</p>
                    <div className="flex items-baseline gap-2">
                      <h3 className="text-3xl font-black text-violet-600">{data.length} 位</h3>
                      <span className="text-[10px] text-violet-400 font-bold uppercase tracking-wider">查看全部</span>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-none shadow-md bg-gradient-to-br from-blue-50 to-white dark:from-blue-950/20">
              <CardContent className="pt-6">
                <div className="flex items-center gap-4">
                  <div className="p-3 bg-blue-100 dark:bg-blue-900/30 rounded-xl text-blue-600">
                    <Calendar className="w-6 h-6" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">最近一次活跃</p>
                    <h3 className="text-xl font-bold text-blue-600">
                      {data.length > 0 ? format(data[0].lastTime * 1000, 'yyyy-MM-dd') : '暂无'}
                    </h3>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-5 gap-8">
            {/* Top Contacts Ranking */}
            <Card className="lg:col-span-3 shadow-lg border-none">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Trophy className="w-5 h-5 text-yellow-500" />
                  亲密度排行榜 (TOP 10)
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="h-[350px]">
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={chartData} layout="vertical" margin={{ left: 20, right: 30 }}>
                      <CartesianGrid strokeDasharray="3 3" horizontal={true} vertical={false} opacity={0.3} />
                      <XAxis type="number" hide />
                      <YAxis 
                        dataKey="name" 
                        type="category" 
                        axisLine={false} 
                        tickLine={false} 
                        width={80}
                        style={{ fontSize: '12px', fontWeight: 500 }}
                      />
                      <Tooltip 
                        cursor={{ fill: 'rgba(0,0,0,0.05)' }}
                        contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}
                      />
                      <Bar dataKey="count" radius={[0, 4, 4, 0]} barSize={20}>
                        {chartData.map((_, index) => (
                          <Cell key={`cell-${index}`} fill={index < 3 ? '#ec4899' : '#6366f1'} />
                        ))}
                      </Bar>
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </CardContent>
            </Card>

            {/* Sent/Recv Balance */}
            <Card className="lg:col-span-2 shadow-lg border-none">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <MessageCircle className="w-4 h-4 text-primary" />
                  社交平衡度
                </CardTitle>
              </CardHeader>
              <CardContent className="flex flex-col items-center">
                <div className="h-[250px] w-full">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={pieData}
                        cx="50%"
                        cy="50%"
                        innerRadius={60}
                        outerRadius={80}
                        paddingAngle={5}
                        dataKey="value"
                      >
                        {pieData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={entry.color} />
                        ))}
                      </Pie>
                      <Tooltip />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
                <div className="flex gap-8 mt-2">
                  {pieData.map(item => (
                    <div key={item.name} className="flex flex-col items-center">
                      <div className="flex items-center gap-2 text-sm font-medium">
                        <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
                        {item.name}
                      </div>
                      <p className="text-xl font-black">
                        {((item.value / (totalSent + totalRecv)) * 100).toFixed(1)}%
                      </p>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Top 3 Spotlight */}
          <div className="space-y-4">
            <h3 className="text-lg font-bold flex items-center gap-2 ml-1">
              <Heart className="w-5 h-5 text-red-500 fill-red-500" />
              社交核心圈
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {top3.map((contact, index) => (
                <Card key={contact.talker} className="relative overflow-hidden border-none shadow-md hover:shadow-xl transition-all group">
                  <div className={cn(
                    "absolute top-0 left-0 w-1 h-full",
                    index === 0 ? "bg-pink-500" : index === 1 ? "bg-violet-500" : "bg-blue-500"
                  )} />
                  <CardContent className="pt-6">
                    <div className="flex flex-col items-center gap-3">
                      <div className="relative">
                        <Avatar className="w-20 h-20 border-4 border-background shadow-lg">
                          <AvatarImage src={mediaApi.getImageUrl(contact.avatar)} />
                          <AvatarFallback className="text-2xl font-black bg-muted">
                            {contact.name.substring(0, 1)}
                          </AvatarFallback>
                        </Avatar>
                        <div className="absolute -bottom-1 -right-1 bg-yellow-400 text-white w-7 h-7 rounded-full flex items-center justify-center font-bold text-sm border-2 border-background shadow-sm">
                          {index + 1}
                        </div>
                      </div>
                      <div className="text-center">
                        <h4 className="font-bold text-lg line-clamp-1">{contact.name}</h4>
                        <p className="text-xs text-muted-foreground font-medium mb-4">累计互动 {contact.messageCount.toLocaleString()} 条</p>
                        
                        <div className="grid grid-cols-2 gap-4 w-full border-t pt-4">
                          <div className="text-center">
                            <p className="text-[10px] text-muted-foreground uppercase font-bold">我发送</p>
                            <p className="text-sm font-black text-pink-500">{contact.sentCount.toLocaleString()}</p>
                          </div>
                          <div className="text-center">
                            <p className="text-[10px] text-muted-foreground uppercase font-bold">对方发</p>
                            <p className="text-sm font-black text-violet-500">{contact.recvCount.toLocaleString()}</p>
                          </div>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </div>
      </ScrollArea>
      
      {/* Visual background elements */}
      <div className="fixed top-20 right-[-100px] w-[300px] h-[300px] bg-pink-500/5 rounded-full blur-[100px] -z-10" />
      <div className="fixed bottom-20 left-[-100px] w-[300px] h-[300px] bg-blue-500/5 rounded-full blur-[100px] -z-10" />

      {/* All Contacts Modal */}
      {showAllContacts && (
        <div className="fixed inset-0 z-[160] flex items-center justify-center bg-black/60 backdrop-blur-md animate-in fade-in duration-200 p-4">
          <div className="bg-background border shadow-2xl rounded-2xl w-full max-w-lg max-h-[80vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
            <div className="p-6 border-b flex items-center justify-between bg-muted/20">
              <div>
                <h3 className="font-bold text-xl">全部互动联系人</h3>
                <p className="text-xs text-muted-foreground mt-1">按消息往来总数排序</p>
              </div>
              <Button variant="ghost" size="icon" onClick={() => setShowAllContacts(false)} className="rounded-full">
                <X className="h-5 w-5" />
              </Button>
            </div>
            
            <ScrollArea className="flex-1 p-4">
              <div className="space-y-3">
                {data.map((contact, idx) => (
                  <div key={contact.talker} className="flex items-center justify-between p-3 rounded-xl border bg-card/50 hover:bg-muted/50 transition-colors group">
                    <div className="flex items-center gap-4">
                      <div className="text-xs font-black text-muted-foreground/30 w-5">
                        {idx + 1}
                      </div>
                      <Avatar className="h-10 w-10 border shadow-sm">
                        <AvatarImage src={mediaApi.getImageUrl(contact.avatar)} />
                        <AvatarFallback>{contact.name.substring(0,1)}</AvatarFallback>
                      </Avatar>
                      <div className="flex flex-col">
                        <span className="text-sm font-bold line-clamp-1">{contact.name}</span>
                        <span className="text-[10px] text-muted-foreground font-mono truncate max-w-[150px]">{contact.talker}</span>
                      </div>
                    </div>
                    <div className="text-right flex flex-col items-end">
                      <div className="text-sm font-black text-primary group-hover:scale-110 transition-transform">
                        {contact.messageCount.toLocaleString()}
                      </div>
                      <div className="text-[9px] text-muted-foreground uppercase font-bold">总消息</div>
                    </div>
                  </div>
                ))}
              </div>
            </ScrollArea>
            <div className="p-4 bg-muted/10 border-t text-center text-[10px] text-muted-foreground">
              列表仅包含私聊记录，已自动排除群聊消息。
            </div>
          </div>
          <div className="absolute inset-0 -z-10" onClick={() => setShowAllContacts(false)} />
        </div>
      )}
    </div>
  )
}
