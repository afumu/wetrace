import { SessionList } from "@/components/chat/SessionList"
import { MessageList } from "@/components/chat/MessageList"
import { useAppStore } from "@/stores/app"
import { cn } from "@/lib/utils"
import { useChat } from "@/hooks/useChat"
import { RefreshCw, ArrowLeft, Smile, PlusCircle, Mic, Download, Sparkles, ImageIcon, Images, BrainCircuit, MessageSquareQuote, MoreHorizontal } from "lucide-react"
import { Button } from "@/components/ui/button"
import { systemApi, mediaApi } from "@/api"
import { toast } from "sonner"
import { aiApi } from "@/api/ai"
import { createPortal } from "react-dom"
import { useState, useMemo, useRef, useCallback } from "react"
import { useSessions } from "@/hooks/useSession"
import { useMessages } from "@/hooks/useChatLog"
import { AnalysisPanel } from "@/components/analysis/AnalysisPanel"
import { ExportModal } from "@/components/chat/ExportModal"
import { AISummaryModal } from "@/components/ai/AISummaryModal"
import { AISimulateChat } from "@/components/ai/AISimulateChat"
import { SessionGalleryModal } from "@/components/chat/SessionGalleryModal"

export default function Chat() {
  const isMobile = useAppStore((state) => state.isMobile)
  const { activeTalker, setActiveTalker } = useChat()
  const [isSyncing, setIsSyncing] = useState(false)
  const [showAnalysis, setShowAnalysis] = useState(false)
  const [showExportModal, setShowExportModal] = useState(false)
  
  // AI States
  const [showAISummary, setShowAISummary] = useState(false)
  const [aiSummary, setAiSummary] = useState("")
  const [isSummarizing, setIsSummarizing] = useState(false)
  const [showAISimulate, setShowAISimulate] = useState(false)
  const [showSessionGallery, setShowSessionGallery] = useState(false)
  const [showMoreMenu, setShowMoreMenu] = useState(false)
  const moreButtonRef = useRef<HTMLButtonElement>(null)
  const [menuPos, setMenuPos] = useState({ top: 0, right: 0 })

  const openMoreMenu = useCallback(() => {
    if (moreButtonRef.current) {
      const rect = moreButtonRef.current.getBoundingClientRect()
      setMenuPos({
        top: rect.bottom + 4,
        right: window.innerWidth - rect.right,
      })
    }
    setShowMoreMenu(true)
  }, [])

  const isGroupChat = useMemo(() => {
    return activeTalker?.endsWith('@chatroom')
  }, [activeTalker])

  const handleAISummarize = async (timeRange?: string) => {
    if (!activeTalker) return
    setShowAISummary(true)
    setIsSummarizing(true)
    try {
      const res = await aiApi.summarize({ 
        talker: activeTalker,
        time_range: timeRange
      })
      setAiSummary(res)
    } catch (err) {
      console.error("AI Summarize failed:", err)
      setAiSummary("AI 总结失败，请检查后端 AI 配置是否正确。")
    } finally {
      setIsSummarizing(false)
    }
  }

  const handleSessionCache = async () => {
    if (!activeTalker) return
    try {
      await mediaApi.startCache('session', activeTalker)
      window.dispatchEvent(new CustomEvent('image-cache-start'))
      toast.success("会话图片预加载已启动。")
    } catch (err) {
      console.error("Failed to start session cache:", err)
      toast.error("启动会话缓存失败")
    }
  }

  const { data: sessions = [] } = useSessions()
  const { data: allMessages = [] } = useMessages(activeTalker)

  const contactAvatar = useMemo(() => {
    if (!activeTalker) return ""
    const session = sessions.find(s => s.talker === activeTalker)
    if (session?.avatar) return session.avatar
    
    // Fallback: find from messages
    const msg = allMessages.find(m => m.sender === activeTalker)
    return msg?.smallHeadURL || msg?.bigHeadURL || mediaApi.getAvatarUrl(`avatar/${activeTalker}`)
  }, [activeTalker, sessions, allMessages])

  const selfAvatar = useMemo(() => {
    const msg = allMessages.find(m => m.isSelf)
    if (msg) return msg.smallHeadURL || msg.bigHeadURL || mediaApi.getAvatarUrl(`avatar/${msg.sender}`)
    return ""
  }, [allMessages])
  
  const displayName = useMemo(() => {
    if (!activeTalker) return ""
    const session = sessions.find(s => s.talker === activeTalker)
    return session ? (session.name || session.talkerName) : activeTalker
  }, [activeTalker, sessions])

  const handleSync = async () => {
    try {
      setIsSyncing(true)
      await systemApi.decrypt()
      // Refresh the page or data? For now just alert success
      toast.success("同步成功！")
      window.location.reload()
    } catch (error: any) {
      console.error("Sync failed:", error)
      const message = error.message || "同步失败，请检查控制台。"
      toast.error(message)
    } finally {
      setIsSyncing(false)
    }
  }

  const handleExportRequest = (type: string, range: { type: 'all' | 'custom', start?: string, end?: string }) => {
    if (!activeTalker) return

    let timeRangeParam = ''
    if (range.type === 'custom' && range.start && range.end) {
      timeRangeParam = `&time_range=${range.start}~${range.end}`
    }

    if (type === 'json') {
      const url = `/api/v1/messages?talker_id=${activeTalker}&limit=1000000${timeRangeParam}`
      const a = document.createElement('a')
      a.href = url
      a.download = `messages_${activeTalker}.json`
      a.style.display = 'none'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
    } else if (type === 'forensic') {
      const url = `/api/v1/export/forensic?talker=${activeTalker}&name=${encodeURIComponent(displayName)}${timeRangeParam}`
      window.open(url, '_blank')
    } else {
      const formatParam = type !== 'html' ? `&format=${type}` : ''
      const url = `/api/v1/export/chat?talker=${activeTalker}&name=${encodeURIComponent(displayName)}${formatParam}${timeRangeParam}`
      window.open(url, '_blank')
    }
  }
  
  return (
    <div className="flex h-full w-full">
      <div className={cn(
        "flex-shrink-0 border-r border-border bg-background transition-all duration-300", 
        isMobile 
          ? (activeTalker ? "w-0 overflow-hidden" : "w-full") 
          : "w-[320px]"
      )}>
        <SessionList />
      </div>
      
      <div className={cn(
        "flex-1 bg-muted/30 flex flex-col h-full overflow-hidden",
        isMobile && !activeTalker && "hidden"
      )}>
        {activeTalker ? (
          <>
            <div className="h-14 flex-shrink-0 border-b border-border/30 bg-background/50 backdrop-blur-md flex items-center justify-between px-4">
              <div className="flex items-center gap-2">
                {isMobile && (
                  <Button 
                    variant="ghost" 
                    size="icon" 
                    className="w-8 h-8"
                    onClick={() => setActiveTalker('')}
                  >
                    <ArrowLeft className="w-5 h-5" />
                  </Button>
                )}
                <h2 className="font-medium text-sm truncate">{displayName}</h2>
              </div>
              <div className="flex items-center gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  className="gap-2 text-primary hover:bg-primary/10"
                  onClick={() => handleAISummarize()}
                  title="AI 总结最近对话"
                >
                  <BrainCircuit className="w-4 h-4" />
                  <span className="text-xs font-bold">AI 总结</span>
                </Button>

                <Button
                  variant="ghost"
                  size="sm"
                  className="gap-2 text-muted-foreground hover:text-primary"
                  onClick={() => setShowExportModal(true)}
                  title="导出聊天记录"
                >
                  <Download className="w-4 h-4" />
                  <span className="text-xs">导出</span>
                </Button>

                {/* More dropdown */}
                <Button
                  ref={moreButtonRef}
                  variant="ghost"
                  size="sm"
                  className="gap-1 text-muted-foreground hover:text-primary"
                  onClick={openMoreMenu}
                  title="更多操作"
                >
                  <MoreHorizontal className="w-4 h-4" />
                  <span className="text-xs">更多</span>
                </Button>
                {showMoreMenu && createPortal(
                  <>
                    <div className="fixed inset-0 z-[140]" onClick={() => setShowMoreMenu(false)} />
                    <div className="fixed z-[150] bg-card border rounded-lg shadow-lg py-1 w-44" style={{ top: menuPos.top, right: menuPos.right }}>
                        {/* AI 功能 */}
                        <div className="px-3 py-1 text-[10px] text-muted-foreground font-medium uppercase tracking-wider">AI 功能</div>
                        {!isGroupChat && (
                          <button
                            className="w-full px-3 py-2 text-sm text-left hover:bg-muted/50 transition-colors flex items-center gap-2"
                            onClick={() => { setShowAISimulate(true); setShowMoreMenu(false) }}
                          >
                            <MessageSquareQuote className="w-4 h-4 text-primary" />
                            模拟对话
                          </button>
                        )}
                        <button
                          className="w-full px-3 py-2 text-sm text-left hover:bg-muted/50 transition-colors flex items-center gap-2"
                          onClick={() => { setShowAnalysis(true); setShowMoreMenu(false) }}
                        >
                          <Sparkles className="w-4 h-4 text-primary" />
                          会话分析
                        </button>

                        <div className="border-t my-1" />
                        {/* 媒体 */}
                        <div className="px-3 py-1 text-[10px] text-muted-foreground font-medium uppercase tracking-wider">媒体</div>
                        <button
                          className="w-full px-3 py-2 text-sm text-left hover:bg-muted/50 transition-colors flex items-center gap-2"
                          onClick={() => { handleSessionCache(); setShowMoreMenu(false) }}
                        >
                          <ImageIcon className="w-4 h-4" />
                          加载图片
                        </button>
                        <button
                          className="w-full px-3 py-2 text-sm text-left hover:bg-muted/50 transition-colors flex items-center gap-2"
                          onClick={() => { setShowSessionGallery(true); setShowMoreMenu(false) }}
                        >
                          <Images className="w-4 h-4" />
                          查看图片
                        </button>
                        <button
                          className="w-full px-3 py-2 text-sm text-left hover:bg-muted/50 transition-colors flex items-center gap-2"
                          onClick={() => {
                            if (activeTalker) {
                              const url = mediaApi.getExportVoicesUrl(activeTalker, displayName)
                              window.open(url, '_blank')
                              toast.success("语音导出已开始，请稍候...")
                            }
                            setShowMoreMenu(false)
                          }}
                        >
                          <Mic className="w-4 h-4" />
                          导出语音
                        </button>

                        <div className="border-t my-1" />
                        {/* 其他 */}
                        <div className="px-3 py-1 text-[10px] text-muted-foreground font-medium uppercase tracking-wider">其他</div>
                        <button
                          className="w-full px-3 py-2 text-sm text-left hover:bg-muted/50 transition-colors flex items-center gap-2"
                          onClick={() => { handleSync(); setShowMoreMenu(false) }}
                          disabled={isSyncing}
                        >
                          <RefreshCw className={cn("w-4 h-4", isSyncing && "animate-spin")} />
                          {isSyncing ? '正在同步...' : '同步数据'}
                        </button>
                      </div>
                  </>,
                  document.body
                )}
              </div>
            </div>
            
            <div className="flex-1 overflow-hidden relative">
              <MessageList />
              {showAISimulate && (
                <AISimulateChat 
                  talker={activeTalker} 
                  displayName={displayName} 
                  contactAvatar={contactAvatar}
                  selfAvatar={selfAvatar}
                  onClose={() => setShowAISimulate(false)} 
                />
              )}
            </div>

            {/* Dummy Input Area */}
            {!showAISimulate && (
              <div className="flex-shrink-0 bg-background border-t border-border/30 px-6 py-6 pb-safe">
                <div className="flex items-center gap-4">
                  <Button variant="ghost" size="icon" className="shrink-0 text-muted-foreground rounded-full hover:bg-muted h-10 w-10">
                    <Mic className="w-6 h-6" />
                  </Button>
                  
                  <div className="flex-1 bg-muted/50 border border-border/50 rounded-lg h-12 px-4 flex items-center text-muted-foreground/60 text-base cursor-not-allowed select-none">
                    只读模式，无法发送消息
                  </div>

                  <Button variant="ghost" size="icon" className="shrink-0 text-muted-foreground rounded-full hover:bg-muted h-10 w-10">
                    <Smile className="w-6 h-6" />
                  </Button>
                  <Button variant="ghost" size="icon" className="shrink-0 text-muted-foreground rounded-full hover:bg-muted h-10 w-10">
                    <PlusCircle className="w-6 h-6" />
                  </Button>
                </div>
              </div>
            )}
          </>
        ) : (
          !isMobile && (
            <div className="flex items-center justify-center h-full text-muted-foreground flex-col gap-4">
              <div className="w-24 h-24 bg-muted rounded-full flex items-center justify-center">
                <svg xmlns="http://www.w3.org/2000/svg" className="w-12 h-12 text-muted-foreground/50" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
              </div>
              <p>选择一个会话开始浏览</p>
            </div>
          )
        )}
      </div>

      {showAnalysis && activeTalker && (
        <AnalysisPanel 
          talker={activeTalker} 
          onClose={() => setShowAnalysis(false)} 
        />
      )}

      <ExportModal 
        isOpen={showExportModal}
        onClose={() => setShowExportModal(false)}
        onExport={handleExportRequest}
      />

      <AISummaryModal
        isOpen={showAISummary}
        onClose={() => setShowAISummary(false)}
        summary={aiSummary}
        isLoading={isSummarizing}
        onSummarize={handleAISummarize}
      />

      {activeTalker && (
        <SessionGalleryModal
          talker={activeTalker}
          displayName={displayName}
          isOpen={showSessionGallery}
          onClose={() => setShowSessionGallery(false)}
        />
      )}
    </div>
  )
}