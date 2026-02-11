import { useState, useEffect, useRef, useCallback } from "react"
import { useQuery } from "@tanstack/react-query"
import { replayApi } from "@/api/replay"
import { sessionApi } from "@/api/session"
import type { Session, Message } from "@/types"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  Play,
  Pause,
  Search,
  MessageSquare,
  Loader2,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { MessageBubble } from "@/components/chat/MessageBubble"
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar"
import { mediaApi } from "@/api/media"

const SPEED_OPTIONS = [1, 2, 4, 8] as const
const MAX_INTERVAL_MS = 2000 // cap real interval at 2s

/* ============================================================
 * useReplay hook — manages replay state machine
 * ============================================================ */
function useReplay(messages: Message[]) {
  const [visibleCount, setVisibleCount] = useState(0)
  const [playing, setPlaying] = useState(false)
  const [speed, setSpeed] = useState<number>(1)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const total = messages.length

  const clearTimer = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current)
      timerRef.current = null
    }
  }, [])

  const scheduleNext = useCallback(() => {
    if (visibleCount >= total) {
      setPlaying(false)
      return
    }
    const cur = messages[visibleCount]
    const next = visibleCount + 1 < total ? messages[visibleCount + 1] : null
    let delay = 500
    if (cur && next) {
      const curTime = new Date(cur.time).getTime()
      const nextTime = new Date(next.time).getTime()
      const realDiff = Math.max(nextTime - curTime, 0)
      delay = Math.min(realDiff / speed, MAX_INTERVAL_MS)
    }
    delay = Math.max(delay, 80)
    timerRef.current = setTimeout(() => {
      setVisibleCount((c) => c + 1)
    }, delay)
  }, [visibleCount, total, messages, speed])

  useEffect(() => {
    if (playing && visibleCount < total) {
      scheduleNext()
    }
    return clearTimer
  }, [playing, visibleCount, total, scheduleNext, clearTimer])

  const play = () => setPlaying(true)
  const pause = () => {
    setPlaying(false)
    clearTimer()
  }
  const jumpTo = (idx: number) => {
    clearTimer()
    setVisibleCount(Math.max(0, Math.min(idx, total)))
  }
  const reset = () => {
    pause()
    setVisibleCount(0)
  }

  return {
    visibleCount,
    total,
    playing,
    speed,
    setSpeed,
    play,
    pause,
    jumpTo,
    reset,
    progress: total > 0 ? (visibleCount / total) * 100 : 0,
  }
}

/* ============================================================
 * Replay Controls Bar
 * ============================================================ */
function ReplayControls({
  playing,
  speed,
  visibleCount,
  total,
  onPlay,
  onPause,
  onSpeedChange,
  onSeek,
}: {
  playing: boolean
  speed: number
  progress: number
  visibleCount: number
  total: number
  onPlay: () => void
  onPause: () => void
  onSpeedChange: (s: number) => void
  onSeek: (idx: number) => void
}) {
  return (
    <div className="border-t bg-background px-4 py-3 flex items-center gap-3">
      {/* Play / Pause */}
      <Button
        size="icon"
        variant="ghost"
        onClick={playing ? onPause : onPlay}
        disabled={total === 0}
      >
        {playing ? (
          <Pause className="w-5 h-5" />
        ) : (
          <Play className="w-5 h-5" />
        )}
      </Button>

      {/* Speed selector */}
      <div className="flex items-center gap-1">
        {SPEED_OPTIONS.map((s) => (
          <button
            key={s}
            onClick={() => onSpeedChange(s)}
            className={cn(
              "px-2 py-0.5 rounded text-xs font-medium transition-colors",
              speed === s
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground hover:bg-muted"
            )}
          >
            {s}x
          </button>
        ))}
      </div>

      {/* Progress bar */}
      <div className="flex-1 mx-2">
        <input
          type="range"
          min={0}
          max={total}
          value={visibleCount}
          onChange={(e) => onSeek(Number(e.target.value))}
          className="w-full h-1.5 accent-primary cursor-pointer"
        />
      </div>

      {/* Counter */}
      <span className="text-xs text-muted-foreground whitespace-nowrap">
        {visibleCount} / {total}
      </span>
    </div>
  )
}

/* ============================================================
 * Session Selector — full-page session picker
 * ============================================================ */
function SessionSelector({
  sessions,
  loading,
  search,
  onSearchChange,
  onSelect,
}: {
  sessions: Session[]
  loading: boolean
  search: string
  onSearchChange: (v: string) => void
  onSelect: (s: Session) => void
}) {
  return (
    <ScrollArea className="h-full w-full">
      <div className="max-w-3xl mx-auto p-6 space-y-6 pb-20">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">对话回放</h2>
          <p className="text-sm text-muted-foreground mt-1">
            选择一个会话开始回放
          </p>
        </div>

        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="搜索会话..."
            className="h-10 pl-9"
          />
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
          </div>
        ) : sessions.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 gap-4">
            <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center">
              <MessageSquare className="w-8 h-8 text-muted-foreground/30" />
            </div>
            <p className="text-muted-foreground text-sm font-medium">
              未找到会话
            </p>
          </div>
        ) : (
          <div className="space-y-1">
            {sessions.map((s) => (
              <button
                key={s.id}
                onClick={() => onSelect(s)}
                className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-muted/50 transition-colors text-left"
              >
                <Avatar className="w-9 h-9 rounded-full shrink-0">
                  <AvatarImage src={s.smallHeadURL || (s.avatar && (s.avatar.startsWith('http') ? s.avatar : mediaApi.getAvatarUrl(`avatar/${s.talker}`)))} />
                  <AvatarFallback>{(s.name || s.talker).charAt(0)}</AvatarFallback>
                </Avatar>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">{s.name || s.talker}</p>
                  <p className="text-xs text-muted-foreground truncate">
                    {s.talker}
                  </p>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>
    </ScrollArea>
  )
}

/* ============================================================
 * Main ReplayView
 * ============================================================ */
export default function ReplayView() {
  const [selectedSession, setSelectedSession] = useState<Session | null>(null)
  const [sessionSearch, setSessionSearch] = useState("")
  const [startDate, setStartDate] = useState("")
  const scrollRef = useRef<HTMLDivElement>(null)

  // Fetch sessions
  const { data: sessionsData, isLoading: sessionsLoading } = useQuery({
    queryKey: ["sessions-replay"],
    queryFn: () => sessionApi.getSessions({ limit: 500, offset: 0 }),
  })

  const sessions = sessionsData?.items || []
  const filteredSessions = sessionSearch
    ? sessions.filter(
        (s) =>
          (s.name || "").toLowerCase().includes(sessionSearch.toLowerCase()) ||
          s.talker.toLowerCase().includes(sessionSearch.toLowerCase())
      )
    : sessions

  // Fetch replay messages when session selected
  const { data: replayData, isLoading: messagesLoading } = useQuery({
    queryKey: ["replay-messages", selectedSession?.talker, startDate],
    queryFn: () =>
      replayApi.getMessages({
        talker_id: selectedSession!.talker,
        start_date: startDate || undefined,
        limit: 1000,
        offset: 0,
      }),
    enabled: !!selectedSession,
  })

  const messages = replayData?.messages || []
  const replay = useReplay(messages)

  // Auto-scroll on new visible message
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [replay.visibleCount])

  // --- RENDER: no session selected ---
  if (!selectedSession) {
    return (
      <div className="flex h-full">
        <SessionSelector
          sessions={filteredSessions}
          loading={sessionsLoading}
          search={sessionSearch}
          onSearchChange={setSessionSearch}
          onSelect={setSelectedSession}
        />
      </div>
    )
  }

  // --- RENDER: session selected, replay mode ---
  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="border-b px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Button
            size="sm"
            variant="ghost"
            onClick={() => {
              replay.reset()
              setSelectedSession(null)
            }}
          >
            返回
          </Button>
          <span className="text-sm font-medium">{selectedSession.name}</span>
          <span className="text-xs text-muted-foreground">
            共 {replay.total} 条消息
          </span>
        </div>
        <div className="flex items-center gap-2">
          <Input
            type="date"
            value={startDate}
            onChange={(e) => {
              replay.reset()
              setStartDate(e.target.value)
            }}
            className="h-8 w-36 text-xs"
          />
        </div>
      </div>

      {/* Message area */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto bg-muted/30 py-4">
        {messagesLoading ? (
          <div className="flex items-center justify-center h-full">
            <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
          </div>
        ) : messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full gap-4">
            <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center">
              <MessageSquare className="w-8 h-8 text-muted-foreground/30" />
            </div>
            <p className="text-muted-foreground text-sm">暂无消息</p>
          </div>
        ) : (
          <>
            {messages.slice(0, replay.visibleCount).map((msg, i) => (
              <MessageBubble key={msg.seq || i} message={msg} showAvatar showName />
            ))}
          </>
        )}
      </div>

      {/* Controls */}
      <ReplayControls
        playing={replay.playing}
        speed={replay.speed}
        progress={replay.progress}
        visibleCount={replay.visibleCount}
        total={replay.total}
        onPlay={replay.play}
        onPause={replay.pause}
        onSpeedChange={replay.setSpeed}
        onSeek={replay.jumpTo}
      />
    </div>
  )
}
