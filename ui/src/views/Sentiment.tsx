import { useState } from "react"
import { sentimentApi, type SentimentResponse } from "@/api/sentiment"
import { useSessions } from "@/hooks/useSession"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  BrainCircuit,
  Loader2,
  TrendingUp,
  Smile,
  Meh,
  Frown,
  Calendar,
  Activity,
} from "lucide-react"
import { cn } from "@/lib/utils"
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from "recharts"

export default function SentimentView() {
  const { data: sessions = [] } = useSessions()
  const [selectedTalker, setSelectedTalker] = useState("")
  const [startDate, setStartDate] = useState("")
  const [endDate, setEndDate] = useState("")
  const [isAnalyzing, setIsAnalyzing] = useState(false)
  const [result, setResult] = useState<SentimentResponse | null>(null)
  const [error, setError] = useState("")

  const handleAnalyze = async () => {
    if (!selectedTalker) return
    setIsAnalyzing(true)
    setError("")
    setResult(null)
    try {
      const timeRange =
        startDate && endDate ? `${startDate}~${endDate}` : undefined
      const res = await sentimentApi.analyze({
        talker: selectedTalker,
        time_range: timeRange,
      })
      setResult(res)
    } catch (err: any) {
      setError(err.message || "分析失败，请检查AI配置")
    } finally {
      setIsAnalyzing(false)
    }
  }

  const displayName =
    sessions.find((s) => s.talker === selectedTalker)?.name ||
    selectedTalker

  return (
    <ScrollArea className="h-full">
      <div className="max-w-5xl mx-auto p-6 space-y-6 pb-20">
        {/* Header */}
        <div>
          <h2 className="text-3xl font-bold tracking-tight mb-2">
            AI 情感分析
          </h2>
          <p className="text-muted-foreground">
            基于AI分析对话情绪倾向与关系变化趋势
          </p>
        </div>

        {/* Config Panel */}
        <SentimentConfigPanel
          sessions={sessions}
          selectedTalker={selectedTalker}
          onSelectTalker={setSelectedTalker}
          startDate={startDate}
          endDate={endDate}
          onStartDateChange={setStartDate}
          onEndDateChange={setEndDate}
          onAnalyze={handleAnalyze}
          isAnalyzing={isAnalyzing}
        />

        {/* Loading */}
        {isAnalyzing && (
          <div className="flex flex-col items-center justify-center py-20 gap-4">
            <Loader2 className="w-12 h-12 animate-spin text-primary" />
            <p className="text-muted-foreground animate-pulse font-medium">
              AI 正在分析对话情感...
            </p>
          </div>
        )}

        {/* Error */}
        {error && (
          <Card className="border-destructive/50 bg-destructive/5">
            <CardContent className="pt-4 pb-4">
              <p className="text-destructive text-sm">{error}</p>
            </CardContent>
          </Card>
        )}

        {/* Results */}
        {result && !isAnalyzing && (
          <SentimentResults result={result} displayName={displayName} />
        )}
      </div>
    </ScrollArea>
  )
}

function SentimentConfigPanel({
  sessions,
  selectedTalker,
  onSelectTalker,
  startDate,
  endDate,
  onStartDateChange,
  onEndDateChange,
  onAnalyze,
  isAnalyzing,
}: {
  sessions: any[]
  selectedTalker: string
  onSelectTalker: (v: string) => void
  startDate: string
  endDate: string
  onStartDateChange: (v: string) => void
  onEndDateChange: (v: string) => void
  onAnalyze: () => void
  isAnalyzing: boolean
}) {
  const [searchText, setSearchText] = useState("")

  const filteredSessions = sessions.filter(
    (s) =>
      !s.talker.endsWith("@chatroom") &&
      (s.name?.includes(searchText) || s.talker.includes(searchText))
  )

  return (
    <Card>
      <CardContent className="pt-6 space-y-4">
        <div className="space-y-2">
          <Label>选择聊天对象</Label>
          <Input
            placeholder="搜索联系人..."
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            className="h-9"
          />
          {searchText && (
            <div className="max-h-40 overflow-y-auto border rounded-lg">
              {filteredSessions.slice(0, 20).map((s) => (
                <div
                  key={s.talker}
                  className={cn(
                    "px-3 py-2 text-sm cursor-pointer hover:bg-muted/50 transition-colors",
                    selectedTalker === s.talker && "bg-primary/10 text-primary"
                  )}
                  onClick={() => {
                    onSelectTalker(s.talker)
                    setSearchText("")
                  }}
                >
                  {s.name || s.talker}
                  <span className="text-xs text-muted-foreground ml-2">
                    {s.talker}
                  </span>
                </div>
              ))}
            </div>
          )}
          {selectedTalker && (
            <div className="text-sm text-primary font-medium">
              已选择: {sessions.find((s) => s.talker === selectedTalker)?.name || selectedTalker}
            </div>
          )}
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <Label className="text-xs text-muted-foreground">
              <Calendar className="w-3 h-3 inline mr-1" />
              开始日期 (可选)
            </Label>
            <Input
              type="date"
              value={startDate}
              onChange={(e) => onStartDateChange(e.target.value)}
              className="h-9"
            />
          </div>
          <div className="space-y-1.5">
            <Label className="text-xs text-muted-foreground">
              <Calendar className="w-3 h-3 inline mr-1" />
              结束日期 (可选)
            </Label>
            <Input
              type="date"
              value={endDate}
              onChange={(e) => onEndDateChange(e.target.value)}
              className="h-9"
            />
          </div>
        </div>

        <Button
          className="w-full gap-2"
          onClick={onAnalyze}
          disabled={!selectedTalker || isAnalyzing}
        >
          {isAnalyzing ? (
            <Loader2 className="w-4 h-4 animate-spin" />
          ) : (
            <BrainCircuit className="w-4 h-4" />
          )}
          开始情感分析
        </Button>
      </CardContent>
    </Card>
  )
}

function SentimentResults({
  result,
  displayName,
}: {
  result: SentimentResponse
  displayName: string
}) {
  return (
    <div className="space-y-6 animate-in fade-in duration-300">
      {/* Overall Score */}
      <OverallScoreCard result={result} displayName={displayName} />

      {/* Summary */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <BrainCircuit className="w-4 h-4 text-primary" />
            AI 分析总结
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm leading-relaxed text-foreground/80 whitespace-pre-wrap">
            {result.summary}
          </p>
        </CardContent>
      </Card>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <SentimentDistributionChart distribution={result.sentiment_distribution} />
        <RelationshipIndicators indicators={result.relationship_indicators} />
      </div>

      {/* Emotion Timeline */}
      {result.emotion_timeline?.length > 0 && (
        <EmotionTimelineChart timeline={result.emotion_timeline} />
      )}
    </div>
  )
}

function OverallScoreCard({
  result,
  displayName,
}: {
  result: SentimentResponse
  displayName: string
}) {
  const scorePercent = Math.round(result.overall_score * 100)
  const ScoreIcon =
    result.overall_score >= 0.6
      ? Smile
      : result.overall_score >= 0.4
        ? Meh
        : Frown
  const scoreColor =
    result.overall_score >= 0.6
      ? "text-green-500"
      : result.overall_score >= 0.4
        ? "text-yellow-500"
        : "text-red-500"

  return (
    <Card className="border-none shadow-lg bg-gradient-to-br from-primary/5 to-background">
      <CardContent className="pt-6">
        <div className="flex items-center gap-6">
          <div className="relative">
            <div
              className={cn(
                "w-24 h-24 rounded-full border-4 flex items-center justify-center",
                scoreColor,
                "border-current"
              )}
            >
              <div className="text-center">
                <div className="text-2xl font-black">{scorePercent}</div>
                <div className="text-[10px] font-medium">分</div>
              </div>
            </div>
            <ScoreIcon
              className={cn("w-6 h-6 absolute -bottom-1 -right-1", scoreColor)}
            />
          </div>
          <div className="flex-1">
            <h3 className="text-xl font-bold mb-1">
              与 {displayName} 的情感评估
            </h3>
            <div className="flex items-center gap-3 text-sm">
              <span className={cn("font-bold", scoreColor)}>
                {result.overall_label}
              </span>
              <span className="text-muted-foreground">|</span>
              <span className="text-muted-foreground">
                关系健康度:{" "}
                <span className="font-medium text-foreground">
                  {result.relationship_health}
                </span>
              </span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function SentimentDistributionChart({
  distribution,
}: {
  distribution: SentimentResponse["sentiment_distribution"]
}) {
  const COLORS = { positive: "#10b981", neutral: "#6b7280", negative: "#ef4444" }
  const LABELS = { positive: "积极", neutral: "中性", negative: "消极" }

  const pieData = Object.entries(distribution).map(([key, value]) => ({
    name: LABELS[key as keyof typeof LABELS] || key,
    value: Math.round(value * 100),
    color: COLORS[key as keyof typeof COLORS] || "#6b7280",
  }))

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">情感分布</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-[200px]">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={pieData}
                cx="50%"
                cy="50%"
                innerRadius={50}
                outerRadius={75}
                paddingAngle={5}
                dataKey="value"
              >
                {pieData.map((entry, i) => (
                  <Cell key={i} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip formatter={(v) => `${v}%`} />
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="flex justify-center gap-6 mt-2">
          {pieData.map((item) => (
            <div key={item.name} className="flex items-center gap-2 text-sm">
              <div
                className="w-3 h-3 rounded-full"
                style={{ backgroundColor: item.color }}
              />
              <span>{item.name}</span>
              <span className="font-bold">{item.value}%</span>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

function RelationshipIndicators({
  indicators,
}: {
  indicators: SentimentResponse["relationship_indicators"]
}) {
  const initiativePercent = Math.round(indicators.initiative_ratio * 100)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Activity className="w-4 h-4 text-primary" />
          关系指标
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <div className="flex justify-between text-sm mb-1">
            <span className="text-muted-foreground">主动发起比例</span>
            <span className="font-bold">{initiativePercent}%</span>
          </div>
          <div className="h-2 bg-muted rounded-full overflow-hidden">
            <div
              className="h-full bg-primary rounded-full transition-all"
              style={{ width: `${initiativePercent}%` }}
            />
          </div>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-muted-foreground">回复速度</span>
          <span className="font-medium">{indicators.response_speed}</span>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-muted-foreground">亲密度趋势</span>
          <span className="font-medium">{indicators.intimacy_trend}</span>
        </div>
      </CardContent>
    </Card>
  )
}

function EmotionTimelineChart({
  timeline,
}: {
  timeline: SentimentResponse["emotion_timeline"]
}) {
  const chartData = timeline.map((item) => ({
    name: item.period,
    score: Math.round(item.score * 100),
    label: item.label,
  }))

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <TrendingUp className="w-4 h-4 text-pink-500" />
          情绪变化趋势
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-[250px]">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" opacity={0.3} />
              <XAxis dataKey="name" style={{ fontSize: "11px" }} />
              <YAxis
                domain={[0, 100]}
                style={{ fontSize: "12px" }}
                tickFormatter={(v) => `${v}`}
              />
              <Tooltip
                formatter={(v) => [`${v} 分`, "情感得分"]}
              />
              <Line
                type="monotone"
                dataKey="score"
                stroke="#ec4899"
                strokeWidth={2}
                dot={{ fill: "#ec4899", r: 4 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Keywords per period */}
        <div className="mt-4 space-y-2">
          {timeline.map((item) => (
            <div key={item.period} className="flex items-center gap-3 text-sm">
              <span className="text-muted-foreground w-20 shrink-0">
                {item.period}
              </span>
              <span
                className={cn(
                  "text-xs font-medium px-2 py-0.5 rounded-full",
                  item.score >= 0.6
                    ? "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"
                    : item.score >= 0.4
                      ? "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400"
                      : "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400"
                )}
              >
                {item.label}
              </span>
              <div className="flex gap-1 flex-wrap">
                {item.keywords.map((kw) => (
                  <span
                    key={kw}
                    className="text-xs bg-muted px-1.5 py-0.5 rounded"
                  >
                    {kw}
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
