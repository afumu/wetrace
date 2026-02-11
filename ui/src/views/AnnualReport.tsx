import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { reportApi, type AnnualReport } from "@/api/report"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  CalendarDays,
  MessageSquare,
  Users,
  TrendingUp,
  Moon,
  Clock,
  Flame,
  Trophy,
} from "lucide-react"
import { formatNumber } from "@/lib/utils"
import { mediaApi } from "@/api/media"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
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
  Pie,
  LineChart,
  Line,
} from "recharts"

const WEEKDAY_NAMES = ["周日", "周一", "周二", "周三", "周四", "周五", "周六"]

export default function AnnualReportView() {
  const currentYear = new Date().getFullYear()
  const [year, setYear] = useState(currentYear)
  const [inputYear, setInputYear] = useState(String(currentYear))

  const { data, isLoading, error } = useQuery({
    queryKey: ["annual-report", year],
    queryFn: () => reportApi.getAnnualReport(year),
  })

  const handleYearChange = () => {
    const y = parseInt(inputYear)
    if (y > 2000 && y <= currentYear) setYear(y)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="flex flex-col items-center gap-4">
          <div className="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin" />
          <p className="text-muted-foreground animate-pulse text-sm">
            正在生成 {year} 年度报告...
          </p>
        </div>
      </div>
    )
  }

  if (error || !data) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center space-y-4">
          <p className="text-destructive font-medium text-sm">加载年度报告失败</p>
          <div className="flex items-center gap-2 justify-center">
            <Input
              type="number"
              value={inputYear}
              onChange={(e) => setInputYear(e.target.value)}
              className="w-24 h-9"
              min={2000}
              max={currentYear}
            />
            <Button size="sm" onClick={handleYearChange}>
              重新加载
            </Button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <ScrollArea className="h-full">
      <div className="max-w-5xl mx-auto p-6 space-y-6 pb-20">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold tracking-tight">
              {data.year} 年度社交报告
            </h2>
            <p className="text-sm text-muted-foreground mt-1">
              你的微信年度数据回顾
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Input
              type="number"
              value={inputYear}
              onChange={(e) => setInputYear(e.target.value)}
              className="w-24 h-9"
              min={2000}
              max={currentYear}
            />
            <Button size="sm" onClick={handleYearChange}>
              切换年份
            </Button>
          </div>
        </div>

        {/* Overview Cards */}
        <AnnualOverviewCards overview={data.overview} />

        {/* Highlights */}
        <AnnualHighlightsSection highlights={data.highlights} />

        {/* Monthly Trend */}
        <MonthlyTrendChart data={data.monthly_trend} />

        {/* Top Contacts */}
        <TopContactsSection contacts={data.top_contacts} />

        {/* Distribution Charts */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <WeekdayChart data={data.weekday_distribution} />
          <HourlyChart data={data.hourly_distribution} />
        </div>

        {/* Message Types */}
        <MessageTypesChart types={data.message_types} />
      </div>
    </ScrollArea>
  )
}

function AnnualOverviewCards({ overview }: { overview: AnnualReport["overview"] }) {
  const stats = [
    { label: "消息总数", value: formatNumber(overview.total_messages), icon: MessageSquare, color: "text-pink-500" },
    { label: "发送消息", value: formatNumber(overview.sent_messages), icon: TrendingUp, color: "text-blue-500" },
    { label: "接收消息", value: formatNumber(overview.received_messages), icon: TrendingUp, color: "text-violet-500" },
    { label: "活跃联系人", value: String(overview.active_contacts), icon: Users, color: "text-green-500" },
    { label: "活跃群聊", value: String(overview.active_chatrooms), icon: Users, color: "text-orange-500" },
    { label: "活跃天数", value: `${overview.active_days} 天`, icon: CalendarDays, color: "text-cyan-500" },
  ]

  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
      {stats.map((s) => (
        <Card key={s.label}>
          <CardContent className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <s.icon className={`w-4 h-4 ${s.color}`} />
              <span className="text-xs text-muted-foreground">{s.label}</span>
            </div>
            <div className="text-xl font-bold">{s.value}</div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}

function AnnualHighlightsSection({ highlights }: { highlights: AnnualReport["highlights"] }) {
  const items = [
    { label: "最忙碌的一天", value: highlights.busiest_day.date, sub: `${formatNumber(highlights.busiest_day.count)} 条消息`, icon: Flame, color: "from-red-50 dark:from-red-950/20" },
    { label: "最安静的一天", value: highlights.quietest_day.date, sub: `${highlights.quietest_day.count} 条消息`, icon: Moon, color: "from-blue-50 dark:from-blue-950/20" },
    { label: "最长连续活跃", value: `${highlights.longest_streak} 天`, sub: "连续聊天记录", icon: Trophy, color: "from-yellow-50 dark:from-yellow-950/20" },
    { label: "深夜消息", value: formatNumber(highlights.late_night_count), sub: "23:00 - 05:00", icon: Moon, color: "from-purple-50 dark:from-purple-950/20" },
    { label: "最早消息", value: highlights.earliest_message_time, sub: "当日最早一条", icon: Clock, color: "from-green-50 dark:from-green-950/20" },
    { label: "最晚消息", value: highlights.latest_message_time, sub: "当日最晚一条", icon: Clock, color: "from-indigo-50 dark:from-indigo-950/20" },
  ]

  return (
    <div>
      <h3 className="text-lg font-bold mb-4 flex items-center gap-2">
        <Flame className="w-5 h-5 text-orange-500" />
        年度亮点
      </h3>
      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
        {items.map((item) => (
          <Card key={item.label} className={`border-none shadow-sm bg-gradient-to-br ${item.color} to-white dark:to-background`}>
            <CardContent className="p-4">
              <div className="flex items-center gap-2 mb-2">
                <item.icon className="w-4 h-4 text-muted-foreground" />
                <span className="text-xs text-muted-foreground font-medium">{item.label}</span>
              </div>
              <div className="text-lg font-bold">{item.value}</div>
              <div className="text-xs text-muted-foreground">{item.sub}</div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}

function MonthlyTrendChart({ data }: { data: AnnualReport["monthly_trend"] }) {
  const chartData = (data || []).map((d) => ({
    name: `${d.month}月`,
    count: d.count,
  }))

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <TrendingUp className="w-4 h-4 text-pink-500" />
          月度消息趋势
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-[300px] w-full" style={{ minWidth: 0, minHeight: 0 }}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" opacity={0.3} />
              <XAxis dataKey="name" style={{ fontSize: "12px" }} />
              <YAxis style={{ fontSize: "12px" }} />
              <Tooltip
                contentStyle={{
                  borderRadius: "8px",
                  border: "none",
                  boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
                }}
              />
              <Line
                type="monotone"
                dataKey="count"
                stroke="#ec4899"
                strokeWidth={2}
                dot={{ fill: "#ec4899", r: 4 }}
                name="消息数"
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}

function TopContactsSection({ contacts }: { contacts: AnnualReport["top_contacts"] }) {
  const top10 = (contacts || []).slice(0, 10)

  return (
    <div>
      <h3 className="text-lg font-bold mb-4 flex items-center gap-2">
        <Trophy className="w-5 h-5 text-yellow-500" />
        亲密度排行 TOP 10
      </h3>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        {top10.map((contact, idx) => (
          <Card key={contact.talker} className="border-none shadow-sm hover:shadow-md transition-shadow">
            <CardContent className="p-4 flex items-center gap-4">
              <div className="text-lg font-black text-muted-foreground/30 w-6 text-center">
                {idx + 1}
              </div>
              <Avatar className="h-10 w-10 border shadow-sm">
                <AvatarImage src={contact.avatar && (contact.avatar.startsWith('http') ? contact.avatar : mediaApi.getAvatarUrl(`avatar/${contact.talker}`))} />
                <AvatarFallback>{contact.name?.substring(0, 1) || "?"}</AvatarFallback>
              </Avatar>
              <div className="flex-1 min-w-0">
                <div className="font-medium text-sm truncate">{contact.name}</div>
                <div className="text-xs text-muted-foreground">
                  发送 {formatNumber(contact.sentCount)} / 接收 {formatNumber(contact.recvCount)}
                </div>
              </div>
              <div className="text-right">
                <div className="text-lg font-bold text-primary">
                  {formatNumber(contact.messageCount)}
                </div>
                <div className="text-[10px] text-muted-foreground">总消息</div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}

function WeekdayChart({ data }: { data: AnnualReport["weekday_distribution"] }) {
  const chartData = (data || []).map((d) => ({
    name: WEEKDAY_NAMES[d.weekday] || `Day ${d.weekday}`,
    count: d.count,
  }))

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">星期分布</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-[250px] w-full" style={{ minWidth: 0, minHeight: 0 }}>
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" opacity={0.3} />
              <XAxis dataKey="name" style={{ fontSize: "12px" }} />
              <YAxis style={{ fontSize: "12px" }} />
              <Tooltip />
              <Bar dataKey="count" radius={[4, 4, 0, 0]} name="消息数">
                {chartData.map((_, i) => (
                  <Cell key={i} fill={i === 0 || i === 6 ? "#ec4899" : "#6366f1"} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}

function HourlyChart({ data }: { data: AnnualReport["hourly_distribution"] }) {
  const chartData = (data || []).map((d) => ({
    name: `${d.hour}时`,
    count: d.count,
  }))

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">24小时分布</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-[250px] w-full" style={{ minWidth: 0, minHeight: 0 }}>
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" opacity={0.3} />
              <XAxis dataKey="name" style={{ fontSize: "10px" }} interval={2} />
              <YAxis style={{ fontSize: "12px" }} />
              <Tooltip />
              <Bar dataKey="count" radius={[2, 2, 0, 0]} fill="#6366f1" name="消息数" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}

function MessageTypesChart({ types }: { types: Record<string, number> }) {
  const TYPE_LABELS: Record<string, string> = {
    text: "文本",
    image: "图片",
    voice: "语音",
    video: "视频",
    link: "链接",
    other: "其他",
  }
  const COLORS = ["#ec4899", "#6366f1", "#06b6d4", "#f59e0b", "#10b981", "#8b5cf6"]

  const pieData = Object.entries(types || {}).map(([key, value], i) => ({
    name: TYPE_LABELS[key] || key,
    value,
    color: COLORS[i % COLORS.length],
  }))

  const total = pieData.reduce((s, d) => s + d.value, 0)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">消息类型分布</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col md:flex-row items-center gap-8">
          <div className="h-[250px] w-[250px]" style={{ minWidth: 0, minHeight: 0 }}>
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={90}
                  paddingAngle={3}
                  dataKey="value"
                >
                  {pieData.map((entry, i) => (
                    <Cell key={i} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="grid grid-cols-2 gap-x-8 gap-y-3">
            {pieData.map((item) => (
              <div key={item.name} className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full shrink-0" style={{ backgroundColor: item.color }} />
                <span className="text-sm">{item.name}</span>
                <span className="text-sm font-bold ml-auto">{formatNumber(item.value)}</span>
                <span className="text-xs text-muted-foreground">
                  ({total > 0 ? ((item.value / total) * 100).toFixed(1) : 0}%)
                </span>
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
