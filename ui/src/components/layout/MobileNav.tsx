import { useAppStore } from "@/stores/app"
import { cn } from "@/lib/utils"
import {
  MessageSquare, Search, Users, ImageIcon, MoreHorizontal,
  Shield, Settings, X, Clock,
  CalendarDays, Heart, Cloud, BrainCircuit, PlayCircle,
} from "lucide-react"
import { useNavigate } from "react-router-dom"
import { useState, useRef, useEffect } from "react"

type MoreItem = {
  key: string
  icon: React.ComponentType<{ className?: string }>
  label: string
  path: string
}

type MoreGroup = {
  label: string
  items: MoreItem[]
}

export function MobileNav() {
  const { activeNav, setActiveNav } = useAppStore()
  const navigate = useNavigate()
  const [showMore, setShowMore] = useState(false)
  const panelRef = useRef<HTMLDivElement>(null)

  const primaryItems = [
    { key: 'chat', icon: MessageSquare, label: '聊天', path: '/chat' },
    { key: 'contacts', icon: Users, label: '联系人', path: '/contacts' },
    { key: 'gallery', icon: ImageIcon, label: '图片', path: '/gallery' },
    { key: 'search', icon: Search, label: '搜索', path: '/search' },
  ]

  const moreGroups: MoreGroup[] = [
    {
      label: '分析',
      items: [
        { key: 'report', icon: CalendarDays, label: '年度报告', path: '/report' },
        { key: 'sentiment', icon: Heart, label: '情感分析', path: '/sentiment' },
        { key: 'wordcloud', icon: Cloud, label: '词云', path: '/wordcloud' },
      ],
    },
    {
      label: 'AI工具',
      items: [
        { key: 'ai-tools', icon: BrainCircuit, label: 'AI工具箱', path: '/ai-tools' },
        { key: 'replay', icon: PlayCircle, label: '对话回放', path: '/replay' },
      ],
    },
    {
      label: '其他',
      items: [
        { key: 'contact-reminder', icon: Clock, label: '联系提醒', path: '/contact-reminder' },
        { key: 'monitor', icon: Shield, label: '监控', path: '/monitor' },
        { key: 'settings', icon: Settings, label: '设置', path: '/settings' },
      ],
    },
  ]

  // All "more" keys for highlighting the more button
  const moreKeys = moreGroups.flatMap((g) => g.items.map((i) => i.key))
  const isMoreActive = moreKeys.includes(activeNav)

  // Close panel on outside click
  useEffect(() => {
    if (!showMore) return
    const handler = (e: MouseEvent) => {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
        setShowMore(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [showMore])

  const handleNavClick = (key: string, path: string) => {
    setActiveNav(key)
    navigate(path)
    setShowMore(false)
  }

  return (
    <>
      {/* More panel overlay */}
      {showMore && (
        <div className="fixed inset-0 bg-black/30 z-40" />
      )}

      {/* More panel */}
      {showMore && (
        <div
          ref={panelRef}
          className="fixed bottom-[50px] left-0 right-0 bg-background border-t border-border rounded-t-xl z-50 px-4 pt-4 pb-2 animate-in slide-in-from-bottom-4 duration-200"
        >
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm font-medium text-foreground">更多功能</span>
            <button
              onClick={() => setShowMore(false)}
              className="p-1 rounded-lg text-muted-foreground hover:bg-muted"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          {moreGroups.map((group) => (
            <div key={group.label} className="mb-3">
              <div className="text-xs text-muted-foreground mb-1.5 px-1">{group.label}</div>
              <div className="grid grid-cols-4 gap-2">
                {group.items.map((item) => (
                  <button
                    key={item.key}
                    onClick={() => handleNavClick(item.key, item.path)}
                    className={cn(
                      "flex flex-col items-center gap-1 py-2 rounded-lg transition-colors",
                      activeNav === item.key
                        ? "bg-primary/10 text-primary"
                        : "text-muted-foreground hover:bg-muted"
                    )}
                  >
                    <item.icon className="w-5 h-5" />
                    <span className="text-[10px] font-medium">{item.label}</span>
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Bottom tab bar */}
      <div className="h-[50px] bg-background border-t border-border flex items-center justify-around px-4 pb-safe z-50">
        {primaryItems.map((item) => (
          <button
            key={item.key}
            onClick={() => handleNavClick(item.key, item.path)}
            className={cn(
              "flex flex-col items-center justify-center w-full h-full gap-0.5",
              activeNav === item.key
                ? "text-primary"
                : "text-muted-foreground"
            )}
          >
            <item.icon className="w-5 h-5" />
            <span className="text-[10px] font-medium">{item.label}</span>
          </button>
        ))}
        <button
          onClick={() => setShowMore(!showMore)}
          className={cn(
            "flex flex-col items-center justify-center w-full h-full gap-0.5",
            isMoreActive || showMore
              ? "text-primary"
              : "text-muted-foreground"
          )}
        >
          <MoreHorizontal className="w-5 h-5" />
          <span className="text-[10px] font-medium">更多</span>
        </button>
      </div>
    </>
  )
}
