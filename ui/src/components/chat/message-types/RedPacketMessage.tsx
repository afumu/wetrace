import type { Message } from "@/types/message"

interface RedPacketMessageProps {
  message: Message
}

export function RedPacketMessage({ message }: RedPacketMessageProps) {
  // Try to find the wish text. Usually in content or a specific field.
  // If content is empty, default to "恭喜发财，大吉大利"
  // For the provided example where content is empty, we use a default.
  // If content has text, use it.
  const text = message.content || "恭喜发财，大吉大利"

  return (
    <div className="w-[240px] overflow-hidden rounded-lg cursor-pointer">
      {/* Orange/Red Top Part */}
      <div className="bg-[#fa9d3b] p-3 flex items-center gap-3">
        {/* Red Envelope Icon/Image */}
        <div className="w-10 h-12 bg-[#e75e58] rounded flex items-center justify-center shrink-0">
          <div className="w-4 h-4 rounded-full bg-[#f8d757] flex items-center justify-center text-[#e75e58] text-[10px] font-bold">
            ￥
          </div>
        </div>
        
        <div className="flex flex-col text-white">
          <span className="text-[15px] font-medium leading-tight line-clamp-1">{text}</span>
          {/* <span className="text-xs opacity-80">查看详情</span> */}
        </div>
      </div>
      
      {/* Bottom Part */}
      <div className="bg-white dark:bg-card px-3 py-1.5 border-t border-transparent">
        <span className="text-[11px] text-muted-foreground">微信红包</span>
      </div>
    </div>
  )
}
