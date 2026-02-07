import { useState } from "react"
import { createPortal } from "react-dom"
import { X, Calendar, FileJson, FileText, Globe } from "lucide-react"
import { Button } from "../ui/button"
import { cn } from "@/lib/utils"
import { Label } from "../ui/label"
import { Input } from "../ui/input"

interface ExportModalProps {
  isOpen: boolean
  onClose: () => void
  onExport: (type: 'html' | 'json' | 'txt', range: { type: 'all' | 'custom', start?: string, end?: string }) => void
}

export function ExportModal({ isOpen, onClose, onExport }: ExportModalProps) {
  const [exportType, setExportType] = useState<'html' | 'json' | 'txt'>('html')
  const [rangeType, setRangeType] = useState<'all' | 'custom'>('all')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')

  if (!isOpen) return null

  const handleExport = () => {
    onExport(exportType, { 
      type: rangeType, 
      start: startDate, 
      end: endDate 
    })
    onClose()
  }

  const exportOptions = [
    { id: 'html', label: '网页导出', icon: Globe, desc: '包含图片、视频的可视化网页' },
    { id: 'json', label: 'JSON数据', icon: FileJson, desc: '原始数据，适合开发者' },
    { id: 'txt', label: '纯文本', icon: FileText, desc: '仅包含文字内容，易于阅读' },
  ] as const

  return createPortal(
    <div className="fixed inset-0 z-[150] flex items-center justify-center bg-black/50 backdrop-blur-sm animate-in fade-in duration-200">
      <div 
        className="bg-background border shadow-2xl rounded-xl p-6 w-[480px] animate-in zoom-in-95 duration-200 relative"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between mb-6 border-b pb-4">
          <div>
            <h3 className="text-lg font-bold">导出聊天记录</h3>
            <p className="text-xs text-muted-foreground">选择导出格式和时间范围</p>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose} className="rounded-full">
            <X className="h-5 w-5" />
          </Button>
        </div>

        <div className="space-y-6">
          {/* 格式选择 */}
          <div className="space-y-3">
            <Label>导出格式</Label>
            <div className="grid grid-cols-1 gap-3">
              {exportOptions.map((option) => (
                <div 
                  key={option.id}
                  onClick={() => setExportType(option.id)}
                  className={cn(
                    "flex items-center gap-4 p-3 rounded-lg border cursor-pointer transition-all hover:bg-muted/50",
                    exportType === option.id 
                      ? "border-primary bg-primary/5 ring-1 ring-primary" 
                      : "border-border"
                  )}
                >
                  <div className={cn(
                    "w-10 h-10 rounded-full flex items-center justify-center shrink-0",
                    exportType === option.id ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground"
                  )}>
                    <option.icon className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="font-medium">{option.label}</div>
                    <div className="text-xs text-muted-foreground">{option.desc}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* 时间范围 */}
          <div className="space-y-3">
            <Label>时间范围</Label>
            <div className="flex gap-4">
              <Button 
                variant={rangeType === 'all' ? "default" : "outline"}
                className="flex-1"
                onClick={() => setRangeType('all')}
              >
                全部记录
              </Button>
              <Button 
                variant={rangeType === 'custom' ? "default" : "outline"}
                className="flex-1 gap-2"
                onClick={() => setRangeType('custom')}
              >
                <Calendar className="w-4 h-4" />
                自定义范围
              </Button>
            </div>

            {rangeType === 'custom' && (
              <div className="grid grid-cols-2 gap-4 pt-2 animate-in slide-in-from-top-2 duration-200">
                <div className="space-y-1.5">
                  <Label className="text-xs text-muted-foreground">开始日期</Label>
                  <Input 
                    type="date" 
                    value={startDate}
                    onChange={(e) => setStartDate(e.target.value)}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label className="text-xs text-muted-foreground">结束日期</Label>
                  <Input 
                    type="date" 
                    value={endDate}
                    onChange={(e) => setEndDate(e.target.value)}
                  />
                </div>
              </div>
            )}
          </div>
        </div>

        <div className="flex justify-end gap-3 mt-8 border-t pt-4">
          <Button variant="ghost" onClick={onClose}>取消</Button>
          <Button onClick={handleExport} disabled={rangeType === 'custom' && (!startDate || !endDate)}>
            开始导出
          </Button>
        </div>
      </div>
      <div className="absolute inset-0 -z-10" onClick={onClose} />
    </div>,
    document.body
  )
}
