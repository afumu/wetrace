import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { chatlogApi } from "@/api/chatlog"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card } from "@/components/ui/card"
import { Search } from "lucide-react"
import { MessageBubble } from "@/components/chat/MessageBubble"
import { useChat } from "@/hooks/useChat"
import { useNavigate } from "react-router-dom"

export default function SearchView() {
  const [keyword, setKeyword] = useState("")
  const { setActiveTalker } = useChat()
  const navigate = useNavigate()

  const { data: results, isLoading } = useQuery({
    queryKey: ['search', keyword],
    queryFn: async () => {
      if (!keyword) return []
      // We assume global search here. 
      // The API supports global search if talker is not provided?
      // Check API implementation in src/api/chatlog.ts: globalSearch uses searchMessages without talker.
      return chatlogApi.globalSearch(keyword)
    },
    enabled: keyword.length > 0,
  })

  const handleMessageClick = (talker: string, seq: number) => {
    setActiveTalker(talker)
    navigate(`/chat?talker=${talker}&seq=${seq}`)
  }

  return (
    <div className="flex flex-col h-full bg-background p-6 max-w-5xl mx-auto w-full animate-in fade-in duration-300">
      <div className="mb-8">
        <h2 className="text-3xl font-bold tracking-tight mb-2">全局搜索</h2>
        <p className="text-muted-foreground">检索所有历史会话中的消息内容</p>
      </div>

      <div className="relative mb-8">
        <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
        <Input 
          placeholder="输入关键词进行搜索..." 
          className="pl-12 h-14 text-xl shadow-sm rounded-xl border-muted-foreground/20 focus:ring-primary/20"
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          autoFocus
        />
      </div>

      <ScrollArea className="flex-1 -mx-4 px-4">
        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-20 gap-4">
            <div className="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin" />
            <p className="text-muted-foreground animate-pulse">正在穿梭于海量数据中...</p>
          </div>
        ) : keyword && results?.length === 0 ? (
          <div className="text-center py-20">
            <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
              <Search className="w-8 h-8 text-muted-foreground/30" />
            </div>
            <p className="text-muted-foreground font-medium">未找到包含 "{keyword}" 的消息</p>
          </div>
        ) : (
          <div className="grid gap-4 pb-10">
            {results?.map((msg) => (
              <Card 
                key={msg.id || msg.seq} 
                className="overflow-hidden cursor-pointer border-none shadow-sm bg-card hover:shadow-md hover:ring-1 hover:ring-primary/20 transition-all group"
                onClick={() => handleMessageClick(msg.talker, msg.seq)}
              >
                <div className="p-4 flex flex-col gap-3">
                  <div className="flex justify-between items-center border-b pb-2">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-bold text-primary">{msg.talkerName || msg.talker}</span>
                      {msg.senderName && msg.senderName !== msg.talkerName && (
                        <span className="text-xs text-muted-foreground">· {msg.senderName}</span>
                      )}
                    </div>
                    <span className="text-[11px] font-medium text-muted-foreground bg-muted px-2 py-0.5 rounded-full">
                      {new Date(msg.createTime * 1000).toLocaleString()}
                    </span>
                  </div>
                  
                  <div className="pointer-events-none opacity-90 group-hover:opacity-100 transition-opacity">
                    <MessageBubble message={msg} showAvatar={false} showTime={false} showName={false} />
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}
      </ScrollArea>
    </div>
  )
}
