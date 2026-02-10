import { createPortal } from "react-dom"
import { X, Loader2, BrainCircuit, Calendar } from "lucide-react"
import { Button } from "../ui/button"
import { useState } from "react"
import { Input } from "../ui/input"
import { Label } from "../ui/label"

interface AISummaryModalProps {
  isOpen: boolean
  onClose: () => void
  summary: string
  isLoading: boolean
  onSummarize: (timeRange?: string) => void
}

export function AISummaryModal({ isOpen, onClose, summary, isLoading, onSummarize }: AISummaryModalProps) {
  const [startDate, setStartDate] = useState("")
  const [endDate, setEndDate] = useState("")
  const [showRange, setShowRange] = useState(false)

  if (!isOpen) return null

  const handleSummarize = () => {
    if (showRange && startDate && endDate) {
      onSummarize(`${startDate}~${endDate}`)
    } else {
      onSummarize()
    }
  }

  return (createPortal(
    <div className="fixed inset-0 z-[150] flex items-center justify-center bg-black/50 backdrop-blur-sm animate-in fade-in duration-200 p-4">
      <div 
        className="bg-background border shadow-2xl rounded-xl p-6 w-full max-w-[700px] max-h-[90vh] flex flex-col animate-in zoom-in-95 duration-200 relative"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between mb-4 border-b pb-4 shrink-0">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
              <BrainCircuit className="w-5 h-5 text-primary" />
            </div>
            <div>
              <h3 className="text-lg font-bold">AI 会话总结</h3>
              <p className="text-xs text-muted-foreground">根据历史聊天记录智能生成概要</p>
            </div>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose} className="rounded-full">
            <X className="h-5 w-5" />
          </Button>
        </div>

        <div className="mb-4 space-y-3 shrink-0">
          <div className="flex items-center justify-between">
            <Label className="text-sm font-medium">时间范围</Label>
            <Button 
              variant="ghost" 
              size="sm" 
              className="h-7 text-xs gap-1"
              onClick={() => setShowRange(!showRange)}
            >
              <Calendar className="w-3 h-3" />
              {showRange ? "取消自定义" : "自定义范围"}
            </Button>
          </div>
          
          {showRange && (
            <div className="grid grid-cols-2 gap-3 animate-in slide-in-from-top-2 duration-200">
              <div className="space-y-1">
                <Label className="text-[10px] text-muted-foreground uppercase">开始日期</Label>
                <Input 
                  type="date" 
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  className="h-8 text-xs"
                />
              </div>
              <div className="space-y-1">
                <Label className="text-[10px] text-muted-foreground uppercase">结束日期</Label>
                <Input 
                  type="date" 
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                  className="h-8 text-xs"
                />
              </div>
            </div>
          )}
          
          <Button 
            className="w-full h-9 gap-2" 
            onClick={handleSummarize}
            disabled={isLoading || (showRange && (!startDate || !endDate))}
          >
            {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <BrainCircuit className="w-4 h-4" />}
            {showRange ? "总结该时间段" : "总结最近内容"}
          </Button>
        </div>

        <div className="flex-1 min-h-[300px] overflow-hidden bg-muted/20 rounded-xl border border-border/50 flex flex-col">
          {isLoading ? (
            <div className="flex flex-col items-center justify-center h-60 gap-4">
              <Loader2 className="w-10 h-10 animate-spin text-primary" />
              <p className="text-sm text-muted-foreground">AI 正在阅读聊天记录并生成总结...</p>
            </div>
          ) : (
            <div className="flex-1 overflow-y-auto custom-scrollbar">
              <div className="text-sm leading-relaxed whitespace-pre-wrap text-foreground/90 p-5">
                {summary || "点击上方按钮开始生成总结"}
              </div>
            </div>
          )}
        </div>

        <style>{`
          .custom-scrollbar::-webkit-scrollbar {
            width: 6px;
          }
          .custom-scrollbar::-webkit-scrollbar-track {
            background: transparent;
          }
          .custom-scrollbar::-webkit-scrollbar-thumb {
            background: rgba(0, 0, 0, 0.1);
            border-radius: 10px;
          }
          .custom-scrollbar::-webkit-scrollbar-thumb:hover {
            background: rgba(0, 0, 0, 0.2);
          }
          .dark .custom-scrollbar::-webkit-scrollbar-thumb {
            background: rgba(255, 255, 255, 0.1);
          }
          .dark .custom-scrollbar::-webkit-scrollbar-thumb:hover {
            background: rgba(255, 255, 255, 0.2);
          }
        `}</style>

        <div className="flex justify-end mt-4 border-t pt-4 shrink-0">
          <Button variant="ghost" onClick={onClose}>
            关闭
          </Button>
        </div>
      </div>
      <div className="absolute inset-0 -z-10" onClick={onClose} />
    </div>,
    document.body
  ))}