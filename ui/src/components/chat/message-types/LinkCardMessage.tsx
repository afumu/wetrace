import type { Message } from "@/types/message"
import { cn } from "@/lib/utils"

interface LinkCardMessageProps {
  message: Message
}

export function LinkCardMessage({ message }: LinkCardMessageProps) {
  const { title, desc, url, displayname, thumburl } = message.contents || {}

  // Fix URL encoding
  const targetUrl = url?.replace(/\\u0026/g, "&")

  const handleClick = () => {
    if (targetUrl) {
      window.open(targetUrl, '_blank')
    }
  }

  return (
    <div 
      className={cn(
        "flex flex-col gap-2 p-3 w-[280px] cursor-pointer transition-colors rounded-lg",
        "bg-card border border-border/50 shadow-sm",
        "hover:bg-muted/50"
      )}
      onClick={handleClick}
    > 
      <div className="font-medium text-[15px] leading-snug line-clamp-2 text-foreground text-left">
        {title || "无标题"}
      </div>
      
      <div className="flex gap-2 items-start mt-1">
        <div className="flex-1 text-xs text-muted-foreground line-clamp-2 text-left h-[34px]">
          {desc}
        </div>
        {thumburl && (
          <img 
            src={thumburl} 
            alt="缩略图" 
            className="w-[40px] h-[40px] object-cover rounded shrink-0 bg-muted"
          />
        )}
      </div>

      <div className="flex items-center gap-1.5 pt-2 border-t border-border/50 mt-1">
        {/* If we had an icon for the source, we could show it. Using a default text for now if displayname exists */}
        {displayname && (
          <span className="text-[10px] text-muted-foreground flex items-center gap-1">
             <span className="w-3 h-3 rounded-full bg-indigo-500/20 text-indigo-500 flex items-center justify-center text-[8px]">
                L
             </span>
            {displayname}
          </span>
        )}
      </div>
    </div>
  )
}
