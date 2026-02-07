import { useState, useMemo } from "react"
import { useContacts, groupContacts } from "@/hooks/useContacts"
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar"
import { Input } from "@/components/ui/input"
import { Search } from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"

export function ContactList() {
  const [search, setSearch] = useState("")
  const { data: contacts, isLoading } = useContacts()

  const filteredContacts = useMemo(() => {
    if (!contacts) return []
    if (!search) return contacts
    
    return contacts.filter(c => 
      (c.remark || c.nickname || c.wxid).toLowerCase().includes(search.toLowerCase())
    )
  }, [contacts, search])

  const groupedContacts = useMemo(() => groupContacts(filteredContacts), [filteredContacts])

  return (
    <div className="flex flex-col h-full w-full bg-background">
      <div className="p-4 border-b border-border">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input 
            placeholder="搜索联系人" 
            className="pl-9 bg-muted/50 border-none focus-visible:ring-1"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
      </div>

      <ScrollArea className="flex-1">
        {isLoading ? (
          <div className="p-8 text-center text-muted-foreground">加载中...</div>
        ) : (
          <div className="pb-4">
            {Object.entries(groupedContacts).map(([initial, group]) => (
              group && group.length > 0 && (
                <div key={initial} className="mb-2">
                  <div className="sticky top-0 z-10 bg-background/95 backdrop-blur px-6 py-1 text-xs font-semibold text-muted-foreground border-b border-border/50">
                    {initial}
                  </div>
                  <div>
                    {group.map(contact => (
                      <div 
                        key={contact.wxid} 
                        className="flex items-center gap-3 px-6 py-3 hover:bg-accent/50 cursor-pointer transition-colors"
                      >
                        <Avatar className="rounded-md">
                          <AvatarImage src={contact.avatar} alt={contact.nickname} />
                          <AvatarFallback className="rounded-md">{(contact.remark || contact.nickname)?.slice(0, 1)}</AvatarFallback>
                        </Avatar>
                        <div className="flex-1 min-w-0">
                          <div className="font-medium text-sm truncate">
                            {contact.remark || contact.nickname}
                          </div>
                          {contact.remark && (
                            <div className="text-xs text-muted-foreground truncate">
                              昵称: {contact.nickname}
                            </div>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )
            ))}
            
            {filteredContacts.length === 0 && !isLoading && (
              <div className="p-8 text-center text-muted-foreground text-sm">
                暂无联系人
              </div>
            )}
          </div>
        )}
      </ScrollArea>
    </div>
  )
}
