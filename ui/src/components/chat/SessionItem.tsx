import type { Session } from "@/types"
import { cn } from "@/lib/utils"
import { formatSessionTime } from "@/lib/date"
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar"
import { EmojiText } from "./EmojiText"

interface SessionItemProps {
  session: Session
  isActive?: boolean
  onClick?: () => void
}

export function SessionItem({ session, isActive, onClick }: SessionItemProps) {
  return (
    <div
      onClick={onClick}
      className={cn(
        "flex items-center gap-3 p-3 cursor-pointer hover:bg-accent/50 transition-colors rounded-lg mx-2",
        isActive && "bg-accent"
      )}
    >
      <div className="relative">
        <Avatar className="w-12 h-12 rounded-lg">
          <AvatarImage src={session.smallHeadURL || session.avatar} alt={session.name} className="object-cover" />
          <AvatarFallback className="rounded-lg">{session.name?.slice(0, 1)}</AvatarFallback>
        </Avatar>
        {session.unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 bg-red-500 text-white text-[10px] px-1.5 py-0.5 rounded-full min-w-[18px] text-center border-2 border-background font-medium shadow-sm">
            {session.unreadCount > 99 ? '99+' : session.unreadCount}
          </span>
        )}
      </div>
      
      <div className="flex-1 min-w-0">
        <div className="flex justify-between items-start mb-1">
          <EmojiText 
            text={session.name || session.talkerName || ""} 
            className="font-medium truncate text-sm text-foreground block" 
          />
          <span className="text-[10px] text-muted-foreground whitespace-nowrap ml-2">
            {session.lastMessage ? formatSessionTime(session.lastMessage.createTime) : ''}
          </span>
        </div>
        <div className="text-xs text-muted-foreground truncate">
          <EmojiText 
            text={session.lastMessage?.content || "[图片]"}
            className="truncate block"
          />
        </div>
      </div>
    </div>
  )
}
