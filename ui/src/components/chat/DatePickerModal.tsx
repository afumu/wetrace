import { useMemo, useState } from "react"
import { createPortal } from "react-dom"
import { DayPicker } from "react-day-picker"
import { zhCN } from "date-fns/locale"
import { format } from "date-fns"
import { X } from "lucide-react"
import { Button } from "../ui/button"

// Import default styles (we will override some with Tailwind)
import "react-day-picker/dist/style.css"

interface DatePickerModalProps {
  isOpen: boolean
  onClose: () => void
  dateMap: Record<string, number>
  onSelect: (index: number) => void
}

export function DatePickerModal({ isOpen, onClose, dateMap, onSelect }: DatePickerModalProps) {
  const [month, setMonth] = useState<Date>(new Date())

  // Create a Set of dates that have messages for fast lookup
  const messageDates = useMemo(() => {
    return new Set(Object.keys(dateMap))
  }, [dateMap])

  if (!isOpen) return null

  // Custom Day component or Modifiers to highlight dates with messages
  const modifiers = {
    hasMessages: (date: Date) => messageDates.has(format(date, 'yyyy-MM-dd'))
  }

  const handleDayClick = (day: Date) => {
    const dateKey = format(day, 'yyyy-MM-dd')
    if (dateMap[dateKey] !== undefined) {
      onSelect(dateMap[dateKey])
    }
  }

  return createPortal(
    <div className="fixed inset-0 z-[150] flex items-center justify-center bg-black/50 backdrop-blur-sm animate-in fade-in duration-200">
      <div 
        className="bg-background border shadow-2xl rounded-xl p-6 w-auto animate-in zoom-in-95 duration-200 relative"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between mb-4 border-b pb-4">
          <div>
            <h3 className="text-lg font-bold">按日期查找</h3>
            <p className="text-xs text-muted-foreground">点击高亮日期查看当天记录</p>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose} className="rounded-full">
            <X className="h-5 w-5" />
          </Button>
        </div>

        <div className="flex flex-col gap-4">
          <style>{`
            .rdp {
              --rdp-cell-size: 45px;
              --rdp-accent-color: hsl(var(--primary));
              --rdp-background-color: hsl(var(--accent));
              margin: 0;
            }
            .rdp-day_hasMessages {
              position: relative;
              background-color: hsl(var(--primary) / 0.1) !important;
              color: hsl(var(--primary)) !important;
              font-weight: 900 !important;
              border: 1px solid hsl(var(--primary) / 0.2);
            }
            .rdp-day_hasMessages:hover {
              background-color: hsl(var(--primary)) !important;
              color: white !important;
            }
            .rdp-day_hasMessages::after {
              content: '';
              position: absolute;
              bottom: 4px;
              left: 50%;
              transform: translateX(-50%);
              width: 4px;
              height: 4px;
              background-color: currentColor;
              border-radius: 50%;
            }
          `}</style>
          
          <DayPicker
            mode="single"
            locale={zhCN}
            month={month}
            onMonthChange={setMonth}
            modifiers={modifiers}
            modifiersClassNames={{
              hasMessages: "rdp-day_hasMessages"
            }}
            onDayClick={handleDayClick}
            footer={
              <div className="mt-4 pt-4 border-t text-xs text-center text-muted-foreground">
                提示：带有蓝色背景的日期表示该日有聊天记录
              </div>
            }
          />
        </div>
      </div>
      <div className="absolute inset-0 -z-10" onClick={onClose} />
    </div>,
    document.body
  )
}