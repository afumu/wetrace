import { useState, useCallback } from "react"
import { useQuery } from "@tanstack/react-query"
import { searchApi, type SearchParams, type SearchItem } from "@/api/search"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Search, Filter, X, MessageSquare, ChevronDown, ChevronUp } from "lucide-react"
import { useChat } from "@/hooks/useChat"
import { useNavigate } from "react-router-dom"
import { cn } from "@/lib/utils"
import { SearchContextPanel } from "@/components/search/SearchContextPanel"

export default function SearchView() {
  const [keyword, setKeyword] = useState("")
  const [searchKeyword, setSearchKeyword] = useState("")
  const [showFilters, setShowFilters] = useState(false)
  const [talkerFilter, setTalkerFilter] = useState("")
  const [senderFilter, setSenderFilter] = useState("")
  const [startDate, setStartDate] = useState("")
  const [endDate, setEndDate] = useState("")
  const [offset, setOffset] = useState(0)
  const [contextItem, setContextItem] = useState<SearchItem | null>(null)
  const limit = 20

  const { setActiveTalker } = useChat()
  const navigate = useNavigate()

  const buildParams = useCallback((): SearchParams | null => {
    if (!searchKeyword) return null
    const params: SearchParams = { keyword: searchKeyword, limit, offset }
    if (talkerFilter) params.talker = talkerFilter
    if (senderFilter) params.sender = senderFilter
    if (startDate && endDate) params.time_range = `${startDate}~${endDate}`
    return params
  }, [searchKeyword, talkerFilter, senderFilter, startDate, endDate, offset])

  const { data, isLoading } = useQuery({
    queryKey: ["search", searchKeyword, talkerFilter, senderFilter, startDate, endDate, offset],
    queryFn: () => {
      const params = buildParams()
      if (!params) return { total: 0, items: [] }
      return searchApi.search(params)
    },
    enabled: !!searchKeyword,
  })

  const handleSearch = () => {
    if (!keyword.trim()) return
    setSearchKeyword(keyword.trim())
    setOffset(0)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") handleSearch()
  }

  const handleJumpToChat = (item: SearchItem) => {
    setActiveTalker(item.talker)
    navigate(`/chat?talker=${item.talker}&seq=${item.seq}`)
  }

  const totalPages = data ? Math.ceil(data.total / limit) : 0
  const currentPage = Math.floor(offset / limit) + 1

  return (
    <div className="flex flex-col h-full bg-background">
      <div className="p-6 pb-0 max-w-5xl mx-auto w-full">
        <div className="mb-6">
          <h2 className="text-2xl font-bold tracking-tight mb-1">全文搜索</h2>
          <p className="text-sm text-muted-foreground">跨会话关键词搜索，支持高级筛选</p>
        </div>

        {/* Search bar */}
        <div className="flex gap-3 mb-4">
          <div className="relative flex-1">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
            <Input
              placeholder="输入关键词进行搜索..."
              className="pl-12 h-12 text-lg shadow-sm rounded-xl border-muted-foreground/20"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onKeyDown={handleKeyDown}
              autoFocus
            />
          </div>
          <Button onClick={handleSearch} className="h-12 px-6 rounded-xl">
            搜索
          </Button>
          <Button
            variant="outline"
            className={cn("h-12 px-4 rounded-xl gap-2", showFilters && "border-primary text-primary")}
            onClick={() => setShowFilters(!showFilters)}
          >
            <Filter className="w-4 h-4" />
            筛选
            {showFilters ? <ChevronUp className="w-3 h-3" /> : <ChevronDown className="w-3 h-3" />}
          </Button>
        </div>

        {/* Filters */}
        {showFilters && (
          <Card className="p-4 mb-4 animate-in slide-in-from-top-2 duration-200">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="space-y-1.5">
                <Label className="text-xs text-muted-foreground">会话ID</Label>
                <Input
                  placeholder="限定会话..."
                  value={talkerFilter}
                  onChange={(e) => setTalkerFilter(e.target.value)}
                  className="h-9"
                />
              </div>
              <div className="space-y-1.5">
                <Label className="text-xs text-muted-foreground">发送人ID</Label>
                <Input
                  placeholder="限定发送人..."
                  value={senderFilter}
                  onChange={(e) => setSenderFilter(e.target.value)}
                  className="h-9"
                />
              </div>
              <div className="space-y-1.5">
                <Label className="text-xs text-muted-foreground">开始日期</Label>
                <Input
                  type="date"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  className="h-9"
                />
              </div>
              <div className="space-y-1.5">
                <Label className="text-xs text-muted-foreground">结束日期</Label>
                <Input
                  type="date"
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                  className="h-9"
                />
              </div>
            </div>
            <div className="flex justify-end mt-3">
              <Button
                variant="ghost"
                size="sm"
                className="text-xs gap-1"
                onClick={() => {
                  setTalkerFilter("")
                  setSenderFilter("")
                  setStartDate("")
                  setEndDate("")
                }}
              >
                <X className="w-3 h-3" />
                清除筛选
              </Button>
            </div>
          </Card>
        )}

        {/* Result count */}
        {searchKeyword && data && (
          <div className="text-sm text-muted-foreground mb-4">
            共找到 <span className="font-bold text-foreground">{data.total}</span> 条结果
          </div>
        )}
      </div>

      {/* Results */}
      <ScrollArea className="flex-1 px-6">
        <div className="max-w-5xl mx-auto w-full pb-10">
          {isLoading ? (
            <div className="flex flex-col items-center justify-center py-20 gap-4">
              <div className="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin" />
              <p className="text-muted-foreground animate-pulse">正在搜索中...</p>
            </div>
          ) : searchKeyword && data?.items.length === 0 ? (
            <div className="text-center py-20">
              <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
                <Search className="w-8 h-8 text-muted-foreground/30" />
              </div>
              <p className="text-muted-foreground font-medium">未找到包含 "{searchKeyword}" 的消息</p>
            </div>
          ) : (
            <>
              <div className="grid gap-3">
                {data?.items.map((item) => (
                  <Card
                    key={`${item.talker}-${item.seq}`}
                    className="overflow-hidden border-none shadow-sm bg-card hover:shadow-md hover:ring-1 hover:ring-primary/20 transition-all group"
                  >
                    <div className="p-4 flex flex-col gap-2">
                      <div className="flex justify-between items-center">
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-bold text-primary">
                            {item.talkerName || item.talker}
                          </span>
                          {item.senderName && item.senderName !== item.talkerName && (
                            <span className="text-xs text-muted-foreground">
                              · {item.senderName}
                            </span>
                          )}
                          {item.isChatRoom && (
                            <span className="text-[10px] bg-muted px-1.5 py-0.5 rounded text-muted-foreground">
                              群聊
                            </span>
                          )}
                        </div>
                        <span className="text-[11px] font-medium text-muted-foreground bg-muted px-2 py-0.5 rounded-full">
                          {new Date(item.time).toLocaleString()}
                        </span>
                      </div>

                      <div
                        className="text-sm text-foreground/80 line-clamp-3"
                        dangerouslySetInnerHTML={{
                          __html: item.highlight || item.content,
                        }}
                      />

                      <div className="flex gap-2 mt-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 text-xs gap-1 text-muted-foreground hover:text-primary"
                          onClick={() => setContextItem(item)}
                        >
                          <MessageSquare className="w-3 h-3" />
                          查看上下文
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 text-xs gap-1 text-muted-foreground hover:text-primary"
                          onClick={() => handleJumpToChat(item)}
                        >
                          跳转到会话
                        </Button>
                      </div>
                    </div>
                  </Card>
                ))}
              </div>

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-center gap-4 mt-6">
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={currentPage <= 1}
                    onClick={() => setOffset(Math.max(0, offset - limit))}
                  >
                    上一页
                  </Button>
                  <span className="text-sm text-muted-foreground">
                    第 {currentPage} / {totalPages} 页
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={currentPage >= totalPages}
                    onClick={() => setOffset(offset + limit)}
                  >
                    下一页
                  </Button>
                </div>
              )}
            </>
          )}
        </div>
      </ScrollArea>

      {/* Context Panel */}
      {contextItem && (
        <SearchContextPanel
          item={contextItem}
          keyword={searchKeyword}
          onClose={() => setContextItem(null)}
          onJumpToChat={handleJumpToChat}
        />
      )}
    </div>
  )
}
