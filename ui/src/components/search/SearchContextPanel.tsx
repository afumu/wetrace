import { useQuery } from "@tanstack/react-query"
import { searchApi, type SearchItem } from "@/api/search"
import { createPortal } from "react-dom"
import { X, ExternalLink } from "lucide-react"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { cn } from "@/lib/utils"

interface Props {
  item: SearchItem
  keyword: string
  onClose: () => void
  onJumpToChat: (item: SearchItem) => void
}

export function SearchContextPanel({ item, keyword, onClose, onJumpToChat }: Props) {
  const { data, isLoading } = useQuery({
    queryKey: ["search-context", item.talker, item.seq],
    queryFn: () => searchApi.getContext(item.talker, item.seq, 10, 10),
  })

  const highlightText = (text: string) => {
    if (!keyword) return text
    const regex = new RegExp(`(${keyword.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")})`, "gi")
    return text.replace(regex, "<em class='bg-yellow-200 dark:bg-yellow-800 not-italic px-0.5 rounded'>$1</em>")
  }

  return createPortal(
    <div className="fixed inset-0 z-[150] flex items-center justify-center bg-black/50 backdrop-blur-sm animate-in fade-in duration-200 p-4">
      <div
        className="bg-background border shadow-2xl rounded-xl w-full max-w-[600px] max-h-[80vh] flex flex-col animate-in zoom-in-95 duration-200"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b shrink-0">
          <div>
            <h3 className="font-bold">消息上下文</h3>
            <p className="text-xs text-muted-foreground">
              {item.talkerName || item.talker}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              className="h-7 text-xs gap-1"
              onClick={() => onJumpToChat(item)}
            >
              <ExternalLink className="w-3 h-3" />
              跳转到会话
            </Button>
            <Button variant="ghost" size="icon" onClick={onClose} className="rounded-full h-8 w-8">
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* Messages */}
        <ScrollArea className="flex-1 p-4">
          {isLoading ? (
            <div className="flex flex-col items-center justify-center py-12 gap-3">
              <div className="w-8 h-8 border-3 border-primary border-t-transparent rounded-full animate-spin" />
              <p className="text-sm text-muted-foreground">加载上下文...</p>
            </div>
          ) : (
            <div className="space-y-2">
              {data?.messages.map((msg, idx) => (
                <div
                  key={`${msg.talker}-${msg.seq}-${idx}`}
                  className={cn(
                    "p-3 rounded-lg text-sm",
                    idx === data.anchor_index
                      ? "bg-primary/10 border border-primary/30 ring-1 ring-primary/20"
                      : "bg-muted/30"
                  )}
                >
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-medium text-xs text-primary">
                      {msg.senderName || msg.sender}
                    </span>
                    <span className="text-[10px] text-muted-foreground">
                      {new Date(msg.time).toLocaleString()}
                    </span>
                  </div>
                  <div
                    className="text-foreground/80 break-words"
                    dangerouslySetInnerHTML={{
                      __html: idx === data.anchor_index
                        ? highlightText(msg.content)
                        : msg.content,
                    }}
                  />
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
      </div>
      <div className="absolute inset-0 -z-10" onClick={onClose} />
    </div>,
    document.body
  )
}
