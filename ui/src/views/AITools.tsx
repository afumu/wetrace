import { useState } from "react"
import { aiApi, type AITodosResponse, type AIExtractResponse } from "@/api/ai"
import { useSessions } from "@/hooks/useSession"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  BrainCircuit,
  Loader2,
  ListTodo,
  FileSearch,
  FileText,
  Calendar,
  MapPin,
  Clock,
  DollarSign,
  Phone,
  AlertCircle,
} from "lucide-react"
import { cn } from "@/lib/utils"

type TabKey = "summary" | "todos" | "extract"

export default function AIToolsView() {
  const { data: sessions = [] } = useSessions()
  const [activeTab, setActiveTab] = useState<TabKey>("summary")
  const [selectedTalker, setSelectedTalker] = useState("")
  const [timeRange, setTimeRange] = useState("last_week")

  const displayName =
    sessions.find((s) => s.talker === selectedTalker)?.name || selectedTalker

  const tabs = [
    { key: "summary" as const, label: "对话摘要", icon: FileText },
    { key: "todos" as const, label: "待办提取", icon: ListTodo },
    { key: "extract" as const, label: "关键信息", icon: FileSearch },
  ]

  return (
    <ScrollArea className="h-full">
      <div className="max-w-5xl mx-auto p-6 space-y-6 pb-20">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">AI 工具箱</h2>
          <p className="text-sm text-muted-foreground mt-1">
            基于AI的对话分析工具：摘要生成、待办提取、关键信息抽取
          </p>
        </div>

        <SessionSelector
          sessions={sessions}
          selectedTalker={selectedTalker}
          onSelectTalker={setSelectedTalker}
          timeRange={timeRange}
          onTimeRangeChange={setTimeRange}
        />

        <div className="flex gap-2 border-b pb-0">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={cn(
                "flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors",
                activeTab === tab.key
                  ? "border-primary text-primary"
                  : "border-transparent text-muted-foreground hover:text-foreground"
              )}
            >
              <tab.icon className="w-4 h-4" />
              {tab.label}
            </button>
          ))}
        </div>

        {activeTab === "summary" && (
          <SummaryTab
            selectedTalker={selectedTalker}
            timeRange={timeRange}
            displayName={displayName}
          />
        )}
        {activeTab === "todos" && (
          <TodosTab
            selectedTalker={selectedTalker}
            timeRange={timeRange}
          />
        )}
        {activeTab === "extract" && (
          <ExtractTab
            selectedTalker={selectedTalker}
            timeRange={timeRange}
          />
        )}
      </div>
    </ScrollArea>
  )
}

function SessionSelector({
  sessions,
  selectedTalker,
  onSelectTalker,
  timeRange,
  onTimeRangeChange,
}: {
  sessions: any[]
  selectedTalker: string
  onSelectTalker: (v: string) => void
  timeRange: string
  onTimeRangeChange: (v: string) => void
}) {
  const [searchText, setSearchText] = useState("")

  const filteredSessions = sessions.filter(
    (s) =>
      s.name?.includes(searchText) || s.talker.includes(searchText)
  )

  const timeRangeOptions = [
    { value: "last_week", label: "最近一周" },
    { value: "last_month", label: "最近一月" },
    { value: "last_year", label: "最近一年" },
  ]

  return (
    <Card>
      <CardContent className="pt-6 space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium leading-none">选择会话</label>
            <Input
              placeholder="搜索联系人或群聊..."
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

          <div className="space-y-2">
            <label className="text-sm font-medium leading-none">
              <Clock className="w-3 h-3 inline mr-1" />
              时间范围
            </label>
            <div className="flex gap-2">
              {timeRangeOptions.map((opt) => (
                <Button
                  key={opt.value}
                  variant={timeRange === opt.value ? "default" : "outline"}
                  size="sm"
                  onClick={() => onTimeRangeChange(opt.value)}
                >
                  {opt.label}
                </Button>
              ))}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function SummaryTab({
  selectedTalker,
  timeRange,
  displayName,
}: {
  selectedTalker: string
  timeRange: string
  displayName: string
}) {
  const [isLoading, setIsLoading] = useState(false)
  const [result, setResult] = useState<string | null>(null)
  const [error, setError] = useState("")

  const handleAnalyze = async () => {
    if (!selectedTalker) return
    setIsLoading(true)
    setError("")
    setResult(null)
    try {
      const res = await aiApi.summarize({ talker: selectedTalker, time_range: timeRange })
      setResult(res)
    } catch (err: any) {
      setError(err.message || "摘要生成失败")
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="space-y-4">
      <Button
        onClick={handleAnalyze}
        disabled={!selectedTalker || isLoading}
        className="gap-2"
      >
        {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <FileText className="w-4 h-4" />}
        生成对话摘要
      </Button>

      {isLoading && <LoadingIndicator text="AI 正在生成对话摘要..." />}
      {error && <ErrorCard message={error} />}

      {result && !isLoading && (
        <Card className="animate-in fade-in duration-300">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <BrainCircuit className="w-4 h-4 text-primary" />
              与 {displayName} 的对话摘要
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm leading-relaxed text-foreground/80 whitespace-pre-wrap">
              {result}
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

function TodosTab({
  selectedTalker,
  timeRange,
}: {
  selectedTalker: string
  timeRange: string
}) {
  const [isLoading, setIsLoading] = useState(false)
  const [result, setResult] = useState<AITodosResponse | null>(null)
  const [error, setError] = useState("")

  const handleExtract = async () => {
    if (!selectedTalker) return
    setIsLoading(true)
    setError("")
    setResult(null)
    try {
      const res = await aiApi.extractTodos({ talker: selectedTalker, time_range: timeRange })
      setResult(res)
    } catch (err: any) {
      setError(err.message || "待办提取失败")
    } finally {
      setIsLoading(false)
    }
  }

  const priorityStyles: Record<string, string> = {
    high: "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400",
    medium: "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400",
    low: "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400",
  }

  const priorityLabels: Record<string, string> = {
    high: "高",
    medium: "中",
    low: "低",
  }

  return (
    <div className="space-y-4">
      <Button
        onClick={handleExtract}
        disabled={!selectedTalker || isLoading}
        className="gap-2"
      >
        {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <ListTodo className="w-4 h-4" />}
        提取待办事项
      </Button>

      {isLoading && <LoadingIndicator text="AI 正在提取待办事项..." />}
      {error && <ErrorCard message={error} />}

      {result && !isLoading && (
        <div className="space-y-3 animate-in fade-in duration-300">
          {result.todos.length === 0 ? (
            <EmptyState icon={ListTodo} text="未发现待办事项" />
          ) : (
            result.todos.map((todo, i) => (
              <Card key={i} className="hover:shadow-md transition-all">
                <CardContent className="p-4">
                  <div className="flex items-start gap-3">
                    <div className="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center shrink-0 mt-0.5">
                      <span className="text-xs font-bold text-primary">{i + 1}</span>
                    </div>
                    <div className="flex-1 space-y-2">
                      <div className="flex items-center gap-2">
                        <p className="text-sm font-medium">{todo.content}</p>
                        <span className={cn(
                          "text-xs px-2 py-0.5 rounded-full font-medium",
                          priorityStyles[todo.priority] || priorityStyles.medium
                        )}>
                          {priorityLabels[todo.priority] || todo.priority}
                        </span>
                      </div>
                      {todo.deadline && (
                        <div className="flex items-center gap-1 text-xs text-muted-foreground">
                          <Calendar className="w-3 h-3" />
                          <span>截止: {todo.deadline}</span>
                        </div>
                      )}
                      {todo.source_msg && (
                        <p className="text-xs text-muted-foreground bg-muted/50 rounded px-2 py-1">
                          原文: {todo.source_msg}
                        </p>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))
          )}
        </div>
      )}
    </div>
  )
}

function ExtractTab({
  selectedTalker,
  timeRange,
}: {
  selectedTalker: string
  timeRange: string
}) {
  const [isLoading, setIsLoading] = useState(false)
  const [result, setResult] = useState<AIExtractResponse | null>(null)
  const [error, setError] = useState("")

  const handleExtract = async () => {
    if (!selectedTalker) return
    setIsLoading(true)
    setError("")
    setResult(null)
    try {
      const res = await aiApi.extractInfo({
        talker: selectedTalker,
        time_range: timeRange,
        types: ["address", "time", "amount", "phone"],
      })
      setResult(res)
    } catch (err: any) {
      setError(err.message || "信息提取失败")
    } finally {
      setIsLoading(false)
    }
  }

  const typeConfig: Record<string, { icon: typeof MapPin; label: string; color: string }> = {
    address: { icon: MapPin, label: "地址", color: "text-blue-500" },
    time: { icon: Clock, label: "时间", color: "text-amber-500" },
    amount: { icon: DollarSign, label: "金额", color: "text-green-500" },
    phone: { icon: Phone, label: "电话", color: "text-purple-500" },
  }

  return (
    <div className="space-y-4">
      <Button
        onClick={handleExtract}
        disabled={!selectedTalker || isLoading}
        className="gap-2"
      >
        {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <FileSearch className="w-4 h-4" />}
        提取关键信息
      </Button>

      {isLoading && <LoadingIndicator text="AI 正在提取关键信息..." />}
      {error && <ErrorCard message={error} />}

      {result && !isLoading && (
        <div className="space-y-3 animate-in fade-in duration-300">
          {result.extractions.length === 0 ? (
            <EmptyState icon={FileSearch} text="未发现关键信息" />
          ) : (
            result.extractions.map((item, i) => {
              const config = typeConfig[item.type] || {
                icon: FileSearch,
                label: item.type,
                color: "text-muted-foreground",
              }
              const Icon = config.icon
              return (
                <Card key={i} className="hover:shadow-md transition-all">
                  <CardContent className="p-4">
                    <div className="flex items-start gap-3">
                      <div className={cn("mt-0.5 shrink-0", config.color)}>
                        <Icon className="w-4 h-4" />
                      </div>
                      <div className="flex-1 space-y-1">
                        <div className="flex items-center gap-2">
                          <span className="text-xs text-muted-foreground font-medium">
                            {config.label}
                          </span>
                          {item.time && (
                            <span className="text-xs text-muted-foreground">
                              {item.time}
                            </span>
                          )}
                        </div>
                        <p className="text-sm font-medium">{item.value}</p>
                        {item.context && (
                          <p className="text-xs text-muted-foreground bg-muted/50 rounded px-2 py-1">
                            {item.context}
                          </p>
                        )}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )
            })
          )}
        </div>
      )}
    </div>
  )
}

function LoadingIndicator({ text }: { text: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 gap-4">
      <Loader2 className="w-10 h-10 animate-spin text-primary" />
      <p className="text-muted-foreground animate-pulse text-sm">{text}</p>
    </div>
  )
}

function ErrorCard({ message }: { message: string }) {
  return (
    <Card className="border-destructive/50 bg-destructive/5">
      <CardContent className="p-4 flex items-center gap-2">
        <AlertCircle className="w-4 h-4 text-destructive shrink-0" />
        <p className="text-destructive text-sm">{message}</p>
      </CardContent>
    </Card>
  )
}

function EmptyState({ icon: Icon, text }: { icon: typeof FileSearch; text: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 gap-4">
      <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center">
        <Icon className="w-8 h-8 text-muted-foreground/30" />
      </div>
      <p className="text-muted-foreground text-sm font-medium">{text}</p>
    </div>
  )
}
