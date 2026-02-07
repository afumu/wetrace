import type { Message } from "@/types/message"
import { ArrowRightLeft } from "lucide-react"

interface TransferMessageProps {
  message: Message
}

export function TransferMessage({ message }: TransferMessageProps) {
  // Parse content like "[转账|发送 ￥0.10]"
  let amount = ""
  let note = ""

  // Simple parsing logic
  // If content matches the pattern, extract amount.
  // The provided example content is just the summary.
  // Real WeChat transfer usually carries more info in `contents` or xml.
  // For now, we display the raw content or try to clean it up.
  
  if (message.content.startsWith("[转账]")) {
      // Legacy format?
      note = message.content
  } else if (message.content.includes("|")) {
      // "[转账|发送 ￥0.10]"
      const parts = message.content.replace("[", "").replace("]", "").split("|")
      if (parts.length > 1) {
          // "发送 ￥0.10"
          const info = parts[1].trim() // "发送 ￥0.10" or "接收 ￥0.10"
          // Keep the full info to show direction as requested
          amount = info
          note = parts[0] // "转账"
      } else {
          note = message.content
      }
  } else {
      note = message.content
  }
  
  // If we couldn't parse a clean amount, just show the whole content as title
  const title = amount || message.content.replace(/^\[|\]$/g, '')
  const subTitle = note || "转账" // Use note if available, else default text

  return (
    <div className="w-[240px] overflow-hidden rounded-lg cursor-pointer">
      {/* Orange Top Part */}
      <div className="bg-[#fa9d3b] p-3 flex items-center gap-3">
        {/* Transfer Icon */}
        <div className="w-10 h-10 rounded-full border-2 border-white/90 flex items-center justify-center shrink-0">
          <ArrowRightLeft className="w-5 h-5 text-white" />
        </div>
        
        <div className="flex flex-col text-white">
          <span className="text-[15px] font-medium leading-tight truncate">{title}</span>
          {subTitle && <span className="text-xs opacity-80 mt-0.5">{subTitle}</span>}
        </div>
      </div>
      
      {/* Bottom Part */}
      <div className="bg-white dark:bg-card px-3 py-1.5 border-t border-transparent">
        <span className="text-[11px] text-muted-foreground">微信转账</span>
      </div>
    </div>
  )
}
