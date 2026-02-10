import { useAppStore } from "@/stores/app"
import { cn } from "@/lib/utils"
import { MessageSquare, Search, CalendarDays, Heart, MoreHorizontal } from "lucide-react"
import { useNavigate } from "react-router-dom"

export function MobileNav() {
  const { activeNav, setActiveNav } = useAppStore()
  const navigate = useNavigate()

  const navItems = [
    { key: 'chat', icon: MessageSquare, label: '聊天', path: '/chat' },
    { key: 'search', icon: Search, label: '搜索', path: '/search' },
    { key: 'report', icon: CalendarDays, label: '报告', path: '/report' },
    { key: 'sentiment', icon: Heart, label: '情感', path: '/sentiment' },
    { key: 'more', icon: MoreHorizontal, label: '更多', path: '/wordcloud' },
  ]

  const handleNavClick = (key: string, path: string) => {
    setActiveNav(key)
    navigate(path)
  }

  return (
    <div className="h-[50px] bg-background border-t border-border flex items-center justify-around px-4 pb-safe z-50">
      {navItems.map((item) => (
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
    </div>
  )
}