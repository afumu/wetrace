import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { wordcloudApi, type WordCloudResponse, type WordItem } from "@/api/wordcloud"
import { useSessions } from "@/hooks/useSession"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Cloud, Hash, MessageSquare, Loader2 } from "lucide-react"
import { cn } from "@/lib/utils"
import { formatNumber } from "@/lib/utils"

const COLORS = [
  "#ec4899", "#6366f1", "#06b6d4", "#10b981", "#f59e0b",
  "#8b5cf6", "#ef4444", "#14b8a6", "#f97316", "#3b82f6",
  "#d946ef", "#84cc16", "#e11d48", "#0ea5e9", "#a855f7",
]

export default function WordCloudView() {
  const { data: sessions = [] } = useSessions()
  const [mode, setMode] = useState<"global" | "session">("global")
  const [selectedTalker, setSelectedTalker] = useState("")
  const [searchText, setSearchText] = useState("")
  const [startDate, setStartDate] = useState("")
  const [endDate, setEndDate] = useState("")
  const [wordLimit, setWordLimit] = useState(100)

  const timeRange = startDate && endDate ? `${startDate}~${endDate}` : undefined

  const { data, isLoading, error } = useQuery({
    queryKey: ["wordcloud", mode, selectedTalker, timeRange, wordLimit],
    queryFn: () => {
      if (mode === "global") {
        return wordcloudApi.getGlobalWordCloud({ time_range: timeRange, limit: wordLimit })
      }
      return wordcloudApi.getWordCloud(selectedTalker, { time_range: timeRange, limit: wordLimit })
    },
    enabled: mode === "global" || !!selectedTalker,
  })

  const filteredSessions = sessions.filter(
    (s) => s.name?.includes(searchText) || s.talker.includes(searchText)
  )

  return (
    <ScrollArea className="h-full">
      <div className="max-w-5xl mx-auto p-6 space-y-6 pb-20">
        {/* Header */}
        <div>
          <h2 className="text-3xl font-bold tracking-tight mb-2">词云分析</h2>
          <p className="text-muted-foreground">高频词汇可视化，发现聊天中的关键词</p>
        </div>

        {/* Config */}
        <WordCloudConfig
          mode={mode}
          onModeChange={setMode}
          sessions={sessions}
          filteredSessions={filteredSessions}
          selectedTalker={selectedTalker}
          onSelectTalker={setSelectedTalker}
          searchText={searchText}
          onSearchTextChange={setSearchText}
          startDate={startDate}
          endDate={endDate}
          onStartDateChange={setStartDate}
          onEndDateChange={setEndDate}
          wordLimit={wordLimit}
          onWordLimitChange={setWordLimit}
        />

        {/* Loading */}
        {isLoading && (
          <div className="flex flex-col items-center justify-center py-20 gap-4">
            <Loader2 className="w-10 h-10 animate-spin text-primary" />
            <p className="text-muted-foreground animate-pulse">正在分析词频...</p>
          </div>
        )}

        {/* Error */}
        {error && (
          <Card className="border-destructive/50 bg-destructive/5">
            <CardContent className="pt-4 pb-4">
              <p className="text-destructive text-sm">加载词云数据失败</p>
            </CardContent>
          </Card>
        )}

        {/* Results */}
        {data && !isLoading && (
          <WordCloudResults data={data} />
        )}
      </div>
    </ScrollArea>
  )
}

function WordCloudConfig({
  mode, onModeChange,
  sessions, filteredSessions,
  selectedTalker, onSelectTalker,
  searchText, onSearchTextChange,
  startDate, endDate,
  onStartDateChange, onEndDateChange,
  wordLimit, onWordLimitChange,
}: {
  mode: "global" | "session"
  onModeChange: (m: "global" | "session") => void
  sessions: any[]
  filteredSessions: any[]
  selectedTalker: string
  onSelectTalker: (v: string) => void
  searchText: string
  onSearchTextChange: (v: string) => void
  startDate: string
  endDate: string
  onStartDateChange: (v: string) => void
  onEndDateChange: (v: string) => void
  wordLimit: number
  onWordLimitChange: (v: number) => void
}) {
  return (
    <Card>
      <CardContent className="pt-6 space-y-4">
        {/* Mode Toggle */}
        <div className="space-y-2">
          <Label>分析范围</Label>
          <div className="flex gap-3">
            <Button
              variant={mode === "global" ? "default" : "outline"}
              className="flex-1"
              onClick={() => onModeChange("global")}
            >
              全局词云
            </Button>
            <Button
              variant={mode === "session" ? "default" : "outline"}
              className="flex-1"
              onClick={() => onModeChange("session")}
            >
              指定会话
            </Button>
          </div>
        </div>

        {/* Session Selector */}
        {mode === "session" && (
          <div className="space-y-2">
            <Label>选择会话</Label>
            <Input
              placeholder="搜索联系人或群聊..."
              value={searchText}
              onChange={(e) => onSearchTextChange(e.target.value)}
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
                      onSearchTextChange("")
                    }}
                  >
                    {s.name || s.talker}
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
        )}

        {/* Time Range & Limit */}
        <div className="grid grid-cols-3 gap-4">
          <div className="space-y-1.5">
            <Label className="text-xs text-muted-foreground">开始日期</Label>
            <Input type="date" value={startDate} onChange={(e) => onStartDateChange(e.target.value)} className="h-9" />
          </div>
          <div className="space-y-1.5">
            <Label className="text-xs text-muted-foreground">结束日期</Label>
            <Input type="date" value={endDate} onChange={(e) => onEndDateChange(e.target.value)} className="h-9" />
          </div>
          <div className="space-y-1.5">
            <Label className="text-xs text-muted-foreground">词汇数量</Label>
            <Input
              type="number"
              value={wordLimit}
              onChange={(e) => onWordLimitChange(Number(e.target.value) || 100)}
              className="h-9"
              min={10}
              max={500}
            />
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function WordCloudResults({ data }: { data: WordCloudResponse }) {
  return (
    <div className="space-y-6 animate-in fade-in duration-300">
      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardContent className="pt-4 pb-4 flex items-center gap-3">
            <MessageSquare className="w-5 h-5 text-pink-500" />
            <div>
              <div className="text-xs text-muted-foreground">分析消息数</div>
              <div className="text-xl font-bold">{formatNumber(data.total_messages)}</div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4 pb-4 flex items-center gap-3">
            <Hash className="w-5 h-5 text-violet-500" />
            <div>
              <div className="text-xs text-muted-foreground">总词汇数</div>
              <div className="text-xl font-bold">{formatNumber(data.total_words)}</div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4 pb-4 flex items-center gap-3">
            <Cloud className="w-5 h-5 text-cyan-500" />
            <div>
              <div className="text-xs text-muted-foreground">高频词数</div>
              <div className="text-xl font-bold">{data.words.length}</div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Word Cloud Visual */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Cloud className="w-4 h-4 text-primary" />
            词云
          </CardTitle>
        </CardHeader>
        <CardContent>
          <WordCloudCanvas words={data.words} />
        </CardContent>
      </Card>

      {/* Word Frequency Table */}
      <WordFrequencyTable words={data.words} />
    </div>
  )
}

function WordCloudCanvas({ words }: { words: WordItem[] }) {
  if (!words.length) {
    return (
      <div className="h-[300px] flex items-center justify-center text-muted-foreground">
        暂无数据
      </div>
    )
  }

  const maxCount = words[0]?.count || 1

  return (
    <div className="min-h-[300px] flex flex-wrap items-center justify-center gap-2 p-4">
      {words.map((word, i) => {
        const ratio = word.count / maxCount
        const fontSize = Math.max(12, Math.min(48, ratio * 48))
        const color = COLORS[i % COLORS.length]
        const opacity = 0.6 + ratio * 0.4

        return (
          <span
            key={word.text}
            className="inline-block cursor-default hover:opacity-80 transition-opacity"
            style={{
              fontSize: `${fontSize}px`,
              color,
              opacity,
              fontWeight: ratio > 0.5 ? 700 : ratio > 0.3 ? 600 : 400,
              lineHeight: 1.2,
            }}
            title={`${word.text}: ${word.count} 次`}
          >
            {word.text}
          </span>
        )
      })}
    </div>
  )
}

function WordFrequencyTable({ words }: { words: WordItem[] }) {
  const maxCount = words[0]?.count || 1

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Hash className="w-4 h-4 text-primary" />
          词频排行
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {words.slice(0, 30).map((word, idx) => (
            <div key={word.text} className="flex items-center gap-3">
              <span className="text-xs font-bold text-muted-foreground/40 w-6 text-right">
                {idx + 1}
              </span>
              <span className="text-sm font-medium w-20 truncate">{word.text}</span>
              <div className="flex-1 h-5 bg-muted/30 rounded-full overflow-hidden">
                <div
                  className="h-full rounded-full transition-all"
                  style={{
                    width: `${(word.count / maxCount) * 100}%`,
                    backgroundColor: COLORS[idx % COLORS.length],
                  }}
                />
              </div>
              <span className="text-xs font-bold text-muted-foreground w-16 text-right">
                {formatNumber(word.count)}
              </span>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
