import { useState, useEffect, useCallback } from "react"
import { mediaApi } from "@/api"
import { ImageIcon, Loader2, CheckCircle2, X } from "lucide-react"
import { Progress } from "@/components/ui/progress"
import { Button } from "@/components/ui/button"

export function ImageCacheManager() {
  const [status, setStatus] = useState<{
    isRunning: boolean;
    total: number;
    processed: number;
    scope: string;
  } | null>(null)
  
  const [visible, setVisible] = useState(false)

  const fetchStatus = useCallback(async (reason: string) => {
    console.log(`[ImageCache] Fetching status. Reason: ${reason}`)
    try {
      const res = await mediaApi.getCacheStatus()
      console.log(`[ImageCache] Response: isRunning=${res.isRunning}, progress=${res.processed}/${res.total}`)
      
      setStatus(res)
      
      // 只有在运行中，或者刚刚完成（processed > 0）时才显示
      if (res.isRunning || (res.processed > 0 && res.processed === res.total)) {
        setVisible(true)
      }
      return res.isRunning
    } catch (err) {
      console.error("[ImageCache] Fetch error:", err)
      return false
    }
  }, [])

  // 主轮询逻辑
  useEffect(() => {
    let timer: any = null

    const poll = async () => {
      const isRunning = await fetchStatus("Polling")
      if (isRunning) {
        console.log("[ImageCache] Task still running, scheduling next poll in 2s")
        timer = setTimeout(poll, 2000)
      } else {
        console.log("[ImageCache] Task stopped or not running, stopping poll.")
      }
    }

    poll()

    // 监听自定义事件，当其他组件启动任务时手动触发一次 poll
    const handleTaskStart = () => {
      console.log("[ImageCache] Task start event received, restarting poll")
      if (!timer) poll()
    }
    window.addEventListener('image-cache-start', handleTaskStart)

    return () => {
      if (timer) clearTimeout(timer)
      window.removeEventListener('image-cache-start', handleTaskStart)
    }
  }, [fetchStatus])

  if (!visible || !status) return null
  if (!status.isRunning && status.processed === 0) return null

  const progress = status.total > 0 ? Math.round((status.processed / status.total) * 100) : 0
  const isFinished = !status.isRunning && status.processed > 0 && status.processed >= status.total

  return (
    <div className="fixed bottom-4 right-4 z-[100] w-72 bg-card border border-border shadow-xl rounded-xl p-4 animate-in slide-in-from-right-4">
      <div className="flex items-center gap-3 mb-3">
        {status.isRunning ? (
          <Loader2 className="w-5 h-5 text-primary animate-spin" />
        ) : isFinished ? (
          <CheckCircle2 className="w-5 h-5 text-green-500" />
        ) : (
          <ImageIcon className="w-5 h-5 text-muted-foreground" />
        )}
        <div className="flex-1 min-w-0">
          <h4 className="text-sm font-semibold truncate">
            {status.isRunning ? "正在预加载图片..." : isFinished ? "预加载完成" : "任务已停止"}
          </h4>
          <p className="text-[10px] text-muted-foreground">
            {status.scope === 'all' ? '全局扫描' : '会话扫描'} • {status.processed} / {status.total}
          </p>
        </div>
        {!status.isRunning && (
            <Button variant="ghost" size="icon" className="w-6 h-6 rounded-full" onClick={() => setVisible(false)}>
                <X className="w-4 h-4" />
            </Button>
        )}
      </div>

      <Progress value={progress} className="h-1.5 mb-2" />
      
      {status.isRunning && (
        <div className="text-[10px] text-right text-muted-foreground">
          进度: {progress}%
        </div>
      )}
    </div>
  )
}