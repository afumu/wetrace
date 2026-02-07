import { useEffect, useMemo, useRef, useState } from "react"
import { Virtuoso, type VirtuosoHandle } from "react-virtuoso"
import { useMessages } from "@/hooks/useChatLog"
import { useChat } from "@/hooks/useChat"
import { MessageBubble } from "./MessageBubble"
import { isSameDay, format } from "date-fns"
import { formatMessageTime } from "@/lib/date"
import { MessageType } from "@/types/message"
import { useImagePreviewStore } from "@/stores/image-preview"
import { ImagePreviewModal } from "./ImagePreviewModal"
import { ChevronUp, Calendar as CalendarIcon, Search, X as CloseIcon } from "lucide-react"
import { Button } from "../ui/button"
import { Input } from "../ui/input"
import { DatePickerModal } from "./DatePickerModal"
import { useSearchParams } from "react-router-dom"

export function MessageList() {
  const { activeTalker } = useChat()
  const [searchParams] = useSearchParams()
  const targetSeq = searchParams.get('seq')
  
  const virtuosoRef = useRef<VirtuosoHandle>(null)
  const setImages = useImagePreviewStore(state => state.setImages)
  const [showScrollTop, setShowScrollTop] = useState(false)
  const [isDatePickerOpen, setIsDatePickerOpen] = useState(false)
  const [searchQuery, setSearchQuery] = useState("")
  const [isSearching, setIsSearching] = useState(false)
  
  const { 
    data: allMessages = [], 
    isLoading 
  } = useMessages(activeTalker)

  // Sync images to store for preview navigation
  const imageMessages = useMemo(() => {
    return allMessages
      .filter(msg => msg.type === MessageType.Image && (msg.contents?.md5 || msg.content))
      .map(msg => ({
        id: msg.id || msg.seq,
        md5: msg.contents?.md5 || '',
        path: msg.contents?.path,
        content: msg.content
      }))
  }, [allMessages])

  useEffect(() => {
    setImages(imageMessages)
  }, [imageMessages, setImages])

  // Process messages to add date headers and determine avatar visibility
  // Also collect date indexes and message indexes for jumping
  const { items, dateMap, msgIndexMap } = useMemo(() => {
    const res: Array<{ type: 'date' | 'message'; data: any }> = []
    const dMap: Record<string, number> = {}
    const mIndexMap: Record<number, number> = {}
    
    for (let i = 0; i < allMessages.length; i++) {
      const msg = allMessages[i]
      const prevMsg = allMessages[i - 1]
      
      // Date header logic: Show time if first message, or > 5 mins gap, or different day
      const SHOW_TIME_THRESHOLD = 300 // 5 minutes in seconds
      const msgDate = new Date(msg.createTime * 1000)
      const prevDate = prevMsg ? new Date(prevMsg.createTime * 1000) : null
      
      let shouldShowTime = false
      if (!prevDate) {
        shouldShowTime = true
      } else {
        const timeDiff = msg.createTime - prevMsg.createTime
        if (timeDiff > SHOW_TIME_THRESHOLD || !isSameDay(msgDate, prevDate)) {
          shouldShowTime = true
        }
      }

      if (shouldShowTime) {
        const dateKey = format(msgDate, 'yyyy-MM-dd')
        // Store the index of the date header if it's the first one of that day
        if (!(dateKey in dMap)) {
          dMap[dateKey] = res.length
        }

        res.push({
          type: 'date',
          data: formatMessageTime(msg.createTime * 1000)
        })
      }
      
      // Determine props
      const isNewGroup = !prevMsg || prevMsg.sender !== msg.sender || (msg.createTime - prevMsg.createTime > 300)
      const showAvatar = true
      const showName = isNewGroup
      const showTime = false // Handled by date header usually, or bubble timestamp
      
      const itemIndex = res.length
      mIndexMap[msg.seq] = itemIndex

      res.push({
        type: 'message',
        data: {
          message: msg,
          showAvatar,
          showName,
          showTime
        }
      })
    }
    
    return { items: res, dateMap: dMap, msgIndexMap: mIndexMap }
  }, [allMessages])

  const searchResults = useMemo(() => {
    if (!searchQuery.trim()) return []
    const q = searchQuery.toLowerCase()
    return allMessages.filter(msg => 
      msg.content && msg.content.toLowerCase().includes(q)
    ).slice(0, 100) // Limit results for performance
  }, [searchQuery, allMessages])

  // Scroll to bottom on initial load or channel change, or to targetSeq if provided
  useEffect(() => {
    if (!isLoading && items.length > 0) {
      if (targetSeq) {
        const idx = msgIndexMap[Number(targetSeq)]
        if (idx !== undefined) {
          // Small delay to ensure virtuoso is ready
          setTimeout(() => {
            virtuosoRef.current?.scrollToIndex({ index: idx, align: 'center' })
          }, 200)
          return
        }
      }

      // Default: scroll to bottom
      setTimeout(() => {
        virtuosoRef.current?.scrollToIndex({ index: items.length - 1, align: 'end' })
      }, 100)
    }
  }, [activeTalker, isLoading, targetSeq, msgIndexMap])

  if (!activeTalker) return null

  if (isLoading) {
    return <div className="h-full flex items-center justify-center text-muted-foreground">加载中...</div>
  }

  if (items.length === 0) {
    return <div className="h-full flex items-center justify-center text-muted-foreground">暂无消息</div>
  }

  return (
    <div className="relative h-full flex flex-col">
      {/* Search Header */}
      <div className="px-4 py-2 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 flex items-center gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="搜索聊天记录..."
            className="pl-9 h-9"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onFocus={() => setIsSearching(true)}
          />
          {searchQuery && (
            <button 
              className="absolute right-2.5 top-1/2 -translate-y-1/2 hover:bg-muted p-0.5 rounded-full"
              onClick={() => setSearchQuery("")}
            >
              <CloseIcon className="h-4 w-4 text-muted-foreground" />
            </button>
          )}
        </div>
      </div>

      <div className="relative flex-1 overflow-hidden">
        <Virtuoso
          ref={virtuosoRef}
          style={{ height: '100%' }}
          data={items}
          atTopStateChange={(atTop) => {
            setShowScrollTop(!atTop)
          }}
          itemContent={(_, item) => {
            if (item.type === 'date') {
              return (
                <div className="flex justify-center py-4">
                  <span className="text-xs text-muted-foreground bg-muted/30 px-2 py-1 rounded-full">
                    {item.data}
                  </span>
                </div>
              )
            }
            
            return (
              <MessageBubble 
                message={item.data.message}
                showAvatar={item.data.showAvatar}
                showName={item.data.showName}
                showTime={item.data.showTime}
              />
            )
          }}
        />

        {/* Search Results Dropdown */}
        {isSearching && searchQuery.trim() && (
          <div className="absolute top-0 left-0 right-0 max-h-[60%] bg-background border-b shadow-xl overflow-y-auto z-50 animate-in slide-in-from-top-2 duration-200">
            <div className="p-2 sticky top-0 bg-muted/50 backdrop-blur text-xs font-bold flex justify-between items-center">
              <span>搜索结果 ({searchResults.length})</span>
              <button onClick={() => setIsSearching(false)} className="hover:text-primary">关闭</button>
            </div>
            {searchResults.length === 0 ? (
              <div className="p-8 text-center text-muted-foreground">未找到匹配消息</div>
            ) : (
              <div className="divide-y">
                {searchResults.map((msg) => (
                  <button
                    key={msg.seq}
                    className="w-full p-3 text-left hover:bg-muted flex flex-col gap-1 transition-colors"
                    onClick={() => {
                      const targetIdx = msgIndexMap[msg.seq]
                      if (targetIdx !== undefined) {
                        virtuosoRef.current?.scrollToIndex({ index: targetIdx, align: 'center' })
                        setIsSearching(false)
                      }
                    }}
                  >
                    <div className="flex justify-between items-center">
                      <span className="text-xs font-bold text-primary">{msg.senderName}</span>
                      <span className="text-[10px] text-muted-foreground">{formatMessageTime(msg.createTime * 1000)}</span>
                    </div>
                    <div className="text-sm line-clamp-2 text-foreground/80 break-all">{msg.content}</div>
                  </button>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Floating Actions */}
        <div className="absolute top-4 right-4 flex flex-col gap-2 z-10">
          <Button
            variant="secondary"
            size="icon"
            className="rounded-full shadow-md opacity-80 hover:opacity-100 transition-opacity"
            onClick={() => setIsDatePickerOpen(true)}
          >
            <CalendarIcon className="h-4 w-4" />
          </Button>
          {showScrollTop && (
            <Button
              variant="secondary"
              size="icon"
              className="rounded-full shadow-md opacity-80 hover:opacity-100 transition-opacity"
              onClick={() => {
                virtuosoRef.current?.scrollToIndex({ index: 0, align: 'start', behavior: 'smooth' })
              }}
            >
              <ChevronUp className="h-4 w-4" />
            </Button>
          )}
        </div>
      </div>

      <DatePickerModal 
        isOpen={isDatePickerOpen}
        onClose={() => setIsDatePickerOpen(false)}
        dateMap={dateMap}
        onSelect={(index) => {
          virtuosoRef.current?.scrollToIndex({ index, align: 'start' })
          setIsDatePickerOpen(false)
        }}
      />

      <ImagePreviewModal />
    </div>
  )
}
