import { useState, useMemo } from "react"
import { useQuery } from "@tanstack/react-query"
import { contactApi } from "@/api"
import type { Contact, Session } from "@/types"
import { ContactType } from "@/types"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card } from "@/components/ui/card"
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  Users,
  Search,
  Download,
} from "lucide-react"
import { useSessions } from "@/hooks/useSession"

type ContactFilter = "all" | "friend" | "chatroom"

function ExportDropdown({
  onExport,
  onClose,
}: {
  onExport: (format: "csv" | "xlsx") => void
  onClose: () => void
}) {
  return (
    <>
      <div className="fixed inset-0 z-40" onClick={onClose} />
      <div className="absolute right-0 top-full mt-1 z-50 bg-card border rounded-lg shadow-md py-1 w-36">
        <button
          className="w-full px-3 py-2 text-sm text-left hover:bg-muted/50 transition-colors"
          onClick={() => onExport("csv")}
        >
          导出 CSV
        </button>
        <button
          className="w-full px-3 py-2 text-sm text-left hover:bg-muted/50 transition-colors"
          onClick={() => onExport("xlsx")}
        >
          导出 XLSX
        </button>
      </div>
    </>
  )
}

export default function ContactsView() {
  const [keyword, setKeyword] = useState("")
  const [searchKeyword, setSearchKeyword] = useState("")
  const [showExportMenu, setShowExportMenu] = useState(false)
  const [contactFilter, setContactFilter] = useState<ContactFilter>("all")

  const { data: contacts, isLoading } = useQuery({
    queryKey: ["contacts-list", searchKeyword],
    queryFn: () =>
      contactApi.getContacts(
        searchKeyword ? { keyword: searchKeyword } : undefined
      ),
  })

  const { data: sessions = [] } = useSessions()

  const filteredContacts = useMemo(() => {
    if (!contacts) return undefined
    if (contactFilter === "all") return contacts
    if (contactFilter === "friend") return contacts.filter(c => c.type === ContactType.Friend)
    return contacts.filter(c => c.type === ContactType.Chatroom)
  }, [contacts, contactFilter])

  const handleSearch = () => {
    setSearchKeyword(keyword.trim())
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") handleSearch()
  }

  const handleExport = (format: "csv" | "xlsx") => {
    const url = contactApi.exportContacts(format, searchKeyword || undefined)
    window.open(url, "_blank")
    setShowExportMenu(false)
  }

  return (
    <div className="flex flex-col h-full bg-background">
      {/* Header */}
      <div className="p-6 pb-0 max-w-5xl mx-auto w-full">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-2xl font-bold tracking-tight">联系人管理</h2>
            <p className="text-sm text-muted-foreground mt-1">
              查看和管理所有联系人
            </p>
          </div>
          <div className="relative">
            <Button
              variant="outline"
              size="sm"
              className="gap-2"
              onClick={() => setShowExportMenu(!showExportMenu)}
            >
              <Download className="w-4 h-4" />
              导出
            </Button>
            {showExportMenu && (
              <ExportDropdown
                onExport={handleExport}
                onClose={() => setShowExportMenu(false)}
              />
            )}
          </div>
        </div>

        {/* Search bar */}
        <div className="flex gap-3 mb-4">
          <div className="relative flex-1">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
            <Input
              placeholder="搜索昵称、备注、微信号..."
              className="pl-12 h-12 text-lg shadow-sm rounded-xl border-muted-foreground/20"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onKeyDown={handleKeyDown}
            />
          </div>
          <Button onClick={handleSearch} className="h-12 px-6 rounded-xl">
            搜索
          </Button>
        </div>

        {/* Filter tabs */}
        <div className="flex gap-1 mb-4">
          {([
            { value: "all", label: "全部" },
            { value: "friend", label: "私人好友" },
            { value: "chatroom", label: "群聊" },
          ] as const).map((tab) => (
            <Button
              key={tab.value}
              variant={contactFilter === tab.value ? "default" : "outline"}
              size="sm"
              className="text-xs"
              onClick={() => setContactFilter(tab.value)}
            >
              {tab.label}
              {contacts && (
                <span className="ml-1 opacity-70">
                  ({tab.value === "all"
                    ? contacts.length
                    : tab.value === "friend"
                      ? contacts.filter(c => c.type === ContactType.Friend).length
                      : contacts.filter(c => c.type === ContactType.Chatroom).length})
                </span>
              )}
            </Button>
          ))}
        </div>

        {/* Result count */}
        {filteredContacts && (
          <div className="text-sm text-muted-foreground mb-4">
            共 <span className="font-bold text-foreground">{filteredContacts.length}</span> 个联系人
          </div>
        )}
      </div>

      {/* Contact list */}
      <ScrollArea className="flex-1 px-6">
        <div className="max-w-5xl mx-auto w-full pb-20">
          <ContactList contacts={filteredContacts} isLoading={isLoading} sessions={sessions} />
        </div>
      </ScrollArea>
    </div>
  )
}

function ContactList({
  contacts,
  isLoading,
  sessions,
}: {
  contacts: Contact[] | undefined
  isLoading: boolean
  sessions: Session[]
}) {
  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-4">
        <div className="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin" />
        <p className="text-muted-foreground animate-pulse text-sm">加载中...</p>
      </div>
    )
  }

  if (!contacts || contacts.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-4">
        <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center">
          <Users className="w-8 h-8 text-muted-foreground/30" />
        </div>
        <p className="text-muted-foreground text-sm font-medium">暂无联系人</p>
      </div>
    )
  }

  return (
    <div className="grid gap-3">
      {contacts.map((c) => (
        <ContactCard key={c.wxid} contact={c} sessions={sessions} />
      ))}
    </div>
  )
}

function ContactCard({ contact, sessions }: { contact: Contact; sessions: Session[] }) {
  const displayName = contact.remark || contact.nickname || contact.wxid
  const firstChar = (contact.nickname || contact.remark || contact.wxid || "?").charAt(0)
  const avatarUrl = useMemo(() => {
    const session = sessions.find(s => s.talker === contact.wxid)
    return session?.avatar || ""
  }, [contact.wxid, sessions])
  const typeConfig =
    contact.type === "chatroom"
      ? { label: "群聊", className: "bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300" }
      : contact.type === "official"
        ? { label: "公众号", className: "bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300" }
        : { label: "好友", className: "bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300" }

  return (
    <Card className="overflow-hidden border-none shadow-sm bg-card hover:shadow-md hover:ring-1 hover:ring-primary/20 transition-all">
      <div className="p-4 flex items-center gap-3">
        <Avatar className="w-10 h-10 flex-shrink-0">
              <AvatarImage src={avatarUrl} alt={displayName} />
              <AvatarFallback className="text-sm font-medium">{firstChar}</AvatarFallback>
            </Avatar>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium truncate">{displayName}</span>
            <span className={`text-[10px] px-1.5 py-0.5 rounded flex-shrink-0 font-medium ${typeConfig.className}`}>
              {typeConfig.label}
            </span>
          </div>
          <div className="flex items-center gap-3 mt-0.5">
            {contact.alias && (
              <span className="text-xs text-muted-foreground truncate">
                微信号: {contact.alias}
              </span>
            )}
            {contact.remark && contact.nickname && (
              <span className="text-xs text-muted-foreground truncate">
                昵称: {contact.nickname}
              </span>
            )}
          </div>
        </div>
        <span className="text-[10px] text-muted-foreground font-mono flex-shrink-0">
          {contact.wxid.length > 20
            ? contact.wxid.slice(0, 20) + "..."
            : contact.wxid}
        </span>
      </div>
    </Card>
  )
}