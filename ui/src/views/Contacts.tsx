import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { contactApi } from "@/api"
import type { Contact } from "@/types"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  Users,
  Search,
  Download,
} from "lucide-react"

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

  const { data: contacts, isLoading } = useQuery({
    queryKey: ["contacts-list", searchKeyword],
    queryFn: () =>
      contactApi.getContacts(
        searchKeyword ? { keyword: searchKeyword } : undefined
      ),
  })

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

        {/* Result count */}
        {contacts && (
          <div className="text-sm text-muted-foreground mb-4">
            共 <span className="font-bold text-foreground">{contacts.length}</span> 个联系人
          </div>
        )}
      </div>

      {/* Contact list */}
      <ScrollArea className="flex-1 px-6">
        <div className="max-w-5xl mx-auto w-full pb-20">
          <ContactList contacts={contacts} isLoading={isLoading} />
        </div>
      </ScrollArea>
    </div>
  )
}

function ContactList({
  contacts,
  isLoading,
}: {
  contacts: Contact[] | undefined
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
        <ContactCard key={c.wxid} contact={c} />
      ))}
    </div>
  )
}

function ContactCard({ contact }: { contact: Contact }) {
  const displayName = contact.remark || contact.nickname || contact.wxid
  const typeLabel =
    contact.type === "chatroom"
      ? "群聊"
      : contact.type === "official"
        ? "公众号"
        : "好友"

  return (
    <Card className="overflow-hidden border-none shadow-sm bg-card hover:shadow-md hover:ring-1 hover:ring-primary/20 transition-all">
      <div className="p-4 flex items-center gap-3">
        <div className="w-10 h-10 bg-muted rounded-full flex items-center justify-center flex-shrink-0">
          <Users className="w-5 h-5 text-muted-foreground/50" />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium truncate">{displayName}</span>
            <span className="text-[10px] bg-muted px-1.5 py-0.5 rounded text-muted-foreground flex-shrink-0">
              {typeLabel}
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