import { useState, useMemo } from "react"
import { useQuery } from "@tanstack/react-query"
import { contactApi, mediaApi } from "@/api"
import type { NeedContactItem } from "@/api/contact"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card } from "@/components/ui/card"
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Clock, Search, UserX } from "lucide-react"

const DAY_OPTIONS = [
  { value: 3, label: "3天" },
  { value: 7, label: "7天" },
  { value: 15, label: "15天" },
  { value: 30, label: "30天" },
  { value: 60, label: "60天" },
]

function Header({
  days,
  onDaysChange,
  keyword,
  onKeywordChange,
  total,
}: {
  days: number
  onDaysChange: (d: number) => void
  keyword: string
  onKeywordChange: (k: string) => void
  total: number | undefined
}) {
  return (
    <div className="p-6 pb-0 max-w-5xl mx-auto w-full">
      <div className="mb-6">
        <h2 className="text-2xl font-bold tracking-tight">客户联系提醒</h2>
        <p className="text-sm text-muted-foreground mt-1">
          查看超过指定天数未联系的客户，及时跟进维护关系
        </p>
      </div>

      {/* Day selector */}
      <div className="flex flex-wrap gap-2 mb-4">
        {DAY_OPTIONS.map((opt) => (
          <Button
            key={opt.value}
            variant={days === opt.value ? "default" : "outline"}
            size="sm"
            onClick={() => onDaysChange(opt.value)}
          >
            {opt.label}未联系
          </Button>
        ))}
      </div>

      {/* Search */}
      <div className="relative mb-4">
        <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
        <Input
          placeholder="搜索昵称、备注、微信号..."
          className="pl-10 h-10 rounded-xl border-muted-foreground/20"
          value={keyword}
          onChange={(e) => onKeywordChange(e.target.value)}
        />
      </div>

      {/* Count */}
      {total !== undefined && (
        <div className="text-sm text-muted-foreground mb-4">
          共 <span className="font-bold text-foreground">{total}</span> 位客户需要联系
        </div>
      )}
    </div>
  )
}

export default function ContactReminder() {
  const [days, setDays] = useState(7)
  const [keyword, setKeyword] = useState("")

  const { data: items, isLoading } = useQuery({
    queryKey: ["need-contact", days],
    queryFn: () => contactApi.getNeedContactList(days),
  })

  const filtered = useMemo(() => {
    if (!items) return undefined
    if (!keyword.trim()) return items
    const kw = keyword.trim().toLowerCase()
    return items.filter(
      (item) =>
        item.nickName.toLowerCase().includes(kw) ||
        item.remark.toLowerCase().includes(kw) ||
        item.userName.toLowerCase().includes(kw)
    )
  }, [items, keyword])

  return (
    <div className="flex flex-col h-full bg-background">
      <Header
        days={days}
        onDaysChange={setDays}
        keyword={keyword}
        onKeywordChange={setKeyword}
        total={filtered?.length}
      />
      <ScrollArea className="flex-1 px-6">
        <div className="max-w-5xl mx-auto w-full pb-20">
          <ContactList items={filtered} isLoading={isLoading} />
        </div>
      </ScrollArea>
    </div>
  )
}

function ContactList({
  items,
  isLoading,
}: {
  items: NeedContactItem[] | undefined
  isLoading: boolean
}) {
  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-4">
        <div className="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin" />
        <p className="text-muted-foreground animate-pulse text-sm">加载中...</p>
      </div>
    )
  }

  if (!items || items.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-4">
        <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center">
          <UserX className="w-8 h-8 text-muted-foreground/30" />
        </div>
        <p className="text-muted-foreground text-sm font-medium">
          暂无需要联系的客户
        </p>
      </div>
    )
  }

  return (
    <div className="grid gap-3">
      {items.map((item) => (
        <ContactCard key={item.userName} item={item} />
      ))}
    </div>
  )
}

function formatLastContactTime(ts: number): string {
  if (!ts) return "未知"
  const d = new Date(ts * 1000)
  return d.toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  })
}

function getDaysColor(days: number): string {
  if (days >= 60) return "text-red-600 dark:text-red-400"
  if (days >= 30) return "text-orange-600 dark:text-orange-400"
  if (days >= 15) return "text-yellow-600 dark:text-yellow-400"
  return "text-muted-foreground"
}

function ContactCard({ item }: { item: NeedContactItem }) {
  const displayName = item.remark || item.nickName || item.userName
  const firstChar = displayName.charAt(0) || "?"
  const avatarUrl = item.smallHeadURL || mediaApi.getAvatarUrl(`avatar/${item.userName}`)

  return (
    <Card className="overflow-hidden border-none shadow-sm bg-card hover:shadow-md hover:ring-1 hover:ring-primary/20 transition-all">
      <div className="p-4 flex items-center gap-3">
        <Avatar className="w-10 h-10 flex-shrink-0">
          <AvatarImage src={avatarUrl} alt={displayName} />
          <AvatarFallback className="text-sm font-medium">
            {firstChar}
          </AvatarFallback>
        </Avatar>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium truncate">{displayName}</span>
            {item.remark && item.nickName && item.remark !== item.nickName && (
              <span className="text-xs text-muted-foreground truncate">
                ({item.nickName})
              </span>
            )}
          </div>
          <div className="flex items-center gap-1 mt-0.5 text-xs text-muted-foreground">
            <Clock className="w-3 h-3" />
            <span>最后联系: {formatLastContactTime(item.lastContactTime)}</span>
          </div>
        </div>

        <div className="flex-shrink-0 text-right">
          <span className={`text-sm font-bold ${getDaysColor(item.daysSinceContact)}`}>
            {item.daysSinceContact}天
          </span>
          <div className="text-[10px] text-muted-foreground">未联系</div>
        </div>
      </div>
    </Card>
  )
}
