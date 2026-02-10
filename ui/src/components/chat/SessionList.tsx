import { useState } from "react"
import { useSessions } from "@/hooks/useSession"
import { SessionItem } from "./SessionItem"
import { Input } from "@/components/ui/input"
import { Search } from "lucide-react"
import { Virtuoso } from "react-virtuoso"
import { useChat } from "@/hooks/useChat"
import { useQueryClient } from "@tanstack/react-query"
import { sessionApi } from "@/api/session"
import { toast } from "sonner"

export function SessionList() {
  const [search, setSearch] = useState("")
  const { activeTalker, setActiveTalker } = useChat()
  const queryClient = useQueryClient()
  
  // TODO: Add debounce for search
  const { data: sessions = [], isLoading } = useSessions()

  const handleDelete = async (talker: string) => {
    try {
      await sessionApi.deleteSession(talker)
      if (activeTalker === talker) {
        setActiveTalker(null)
      }
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      toast.success("会话已删除")
    } catch (error) {
      console.error("Failed to delete session:", error)
      toast.error("删除会话失败")
    }
  }

  // Client-side filter for demo if API search isn't fully integrated yet
  const filteredSessions = sessions.filter(s => {
    // Exclude system and special placeholders
    if (s.talker === '@placeholder_foldgroup' || s.talker === 'brandsessionholder' || s.talker === 'brandservicesessionholder') {
      return false
    }
    
    const isOfficial = s.type === 'official'
    if (isOfficial) return false

    return s.name?.toLowerCase().includes(search.toLowerCase()) || 
      s.talkerName?.toLowerCase().includes(search.toLowerCase())
  })

  return (
    <div className="flex flex-col h-full bg-background/50 backdrop-blur-xl">
      <div className="h-14 flex-shrink-0 px-3 flex items-center border-b border-border/30">
        <div className="relative flex items-center w-full">
          <Search className="absolute left-2.5 w-3.5 h-3.5 text-muted-foreground pointer-events-none" />
          <Input 
            placeholder="搜索" 
            className="pl-8 bg-muted/50 border-none focus-visible:ring-1 h-8 w-full text-xs"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
      </div>
      
      <div className="flex-1 overflow-hidden pt-2">
        {isLoading ? (
          <div className="p-4 text-center text-muted-foreground text-sm">加载中...</div>
        ) : (
          <Virtuoso
            style={{ height: '100%' }}
            data={filteredSessions}
            itemContent={(_, session) => (
              <div className="py-0.5 px-0">
                <SessionItem 
                  session={session} 
                  isActive={activeTalker === session.talker}
                  onClick={() => setActiveTalker(session.talker)}
                  onDelete={() => handleDelete(session.talker)}
                />
              </div>
            )}
          />
        )}
      </div>
    </div>
  )
}