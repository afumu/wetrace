import { useAppStore } from "@/stores/app"
import { cn } from "@/lib/utils"
import {
  MessageSquare, RefreshCw, Moon, Sun, Monitor, Search, Key,
  ImageIcon, BarChart3, Sparkles, Users, Settings, Shield,
  ChevronDown, CalendarDays, Heart, Cloud, BrainCircuit, PlayCircle,
} from "lucide-react"
import { useNavigate, useLocation } from "react-router-dom"
import { useState, useEffect } from "react"
import { systemApi, mediaApi } from "@/api"
import { toast } from "sonner"
import { KeyManagerModal } from "./KeyManagerModal"
import { ImageCacheManager } from "../chat/ImageCacheManager"

type NavItem = {
  key: string
  icon: React.ComponentType<{ className?: string }>
  label: string
  path: string
}

type NavGroup = {
  key: string
  icon: React.ComponentType<{ className?: string }>
  label: string
  children: NavItem[]
}

type NavEntry = NavItem | NavGroup

function isGroup(entry: NavEntry): entry is NavGroup {
  return 'children' in entry
}

export function Sidebar() {
  const { activeNav, setActiveNav, toggleTheme, settings } = useAppStore()
  const navigate = useNavigate()
  const location = useLocation()
  const [isSyncing, setIsSyncing] = useState(false)
  const [showKeyManager, setShowKeyManager] = useState(false)
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set())

  const navEntries: NavEntry[] = [
    { key: 'chat', icon: MessageSquare, label: '聊天', path: '/chat' },
    { key: 'contacts', icon: Users, label: '联系人', path: '/contacts' },
    { key: 'gallery', icon: ImageIcon, label: '图片', path: '/gallery' },
    {
      key: 'analysis',
      icon: BarChart3,
      label: '分析',
      children: [
        { key: 'report', icon: CalendarDays, label: '年度报告', path: '/report' },
        { key: 'sentiment', icon: Heart, label: '情感分析', path: '/sentiment' },
        { key: 'wordcloud', icon: Cloud, label: '词云', path: '/wordcloud' },
      ],
    },
    {
      key: 'ai',
      icon: Sparkles,
      label: 'AI工具',
      children: [
        { key: 'ai-tools', icon: BrainCircuit, label: 'AI工具箱', path: '/ai-tools' },
        { key: 'replay', icon: PlayCircle, label: '对话回放', path: '/replay' },
      ],
    },
    { key: 'search', icon: Search, label: '搜索', path: '/search' },
    { key: 'monitor', icon: Shield, label: '监控', path: '/monitor' },
  ]

  const settingsItem: NavItem = { key: 'settings', icon: Settings, label: '设置', path: '/settings' }

  // Auto-expand groups that contain the active route
  useEffect(() => {
    const path = location.pathname
    for (const entry of navEntries) {
      if (isGroup(entry)) {
        const match = entry.children.some((child) => path.startsWith(child.path))
        if (match) {
          setExpandedGroups((prev) => {
            const next = new Set(prev)
            next.add(entry.key)
            return next
          })
        }
      }
    }
  }, [location.pathname])

  const toggleGroup = (key: string) => {
    setExpandedGroups((prev) => {
      const next = new Set(prev)
      if (next.has(key)) {
        next.delete(key)
      } else {
        next.add(key)
      }
      return next
    })
  }

  const handleNavClick = (key: string, path: string) => {
    setActiveNav(key)
    navigate(path)
  }

  const handleFullCache = async () => {
    try {
      await mediaApi.startCache('all')
      window.dispatchEvent(new CustomEvent('image-cache-start'))
      toast.success("全量图片预加载任务已启动，你可以在右下角查看进度。")
    } catch (err) {
      console.error("Failed to start full cache:", err)
      toast.error("启动缓存任务失败")
    }
  }

  const handleSync = async () => {
    try {
      setIsSyncing(true)
      await systemApi.decrypt()
      toast.success("数据同步成功！")
      window.location.reload()
    } catch (error: any) {
      console.error("Sync failed:", error)
      const message = error.message || "同步失败，请检查日志。"
      toast.error(message)
    } finally {
      setIsSyncing(false)
    }
  }

  const isActive = (key: string) => activeNav === key
  const isGroupActive = (group: NavGroup) => group.children.some((c) => activeNav === c.key)

  const ThemeIcon = settings.theme === 'dark' ? Moon : settings.theme === 'light' ? Sun : Monitor

  const renderNavItem = (item: NavItem, indent = false) => (
    <button
      key={item.key}
      onClick={() => handleNavClick(item.key, item.path)}
      className={cn(
        "w-full h-9 flex items-center gap-3 rounded-lg px-3 text-sm transition-colors",
        indent && "pl-9",
        isActive(item.key)
          ? "bg-primary/10 text-primary font-medium"
          : "text-muted-foreground hover:bg-muted hover:text-foreground"
      )}
    >
      <item.icon className="w-4 h-4 shrink-0" />
      <span className="truncate">{item.label}</span>
    </button>
  )

  const renderNavGroup = (group: NavGroup) => {
    const expanded = expandedGroups.has(group.key)
    const groupActive = isGroupActive(group)

    return (
      <div key={group.key}>
        <button
          onClick={() => toggleGroup(group.key)}
          className={cn(
            "w-full h-9 flex items-center gap-3 rounded-lg px-3 text-sm transition-colors",
            groupActive
              ? "text-primary font-medium"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          )}
        >
          <group.icon className="w-4 h-4 shrink-0" />
          <span className="truncate flex-1 text-left">{group.label}</span>
          <ChevronDown
            className={cn(
              "w-3.5 h-3.5 shrink-0 transition-transform duration-200",
              expanded && "rotate-180"
            )}
          />
        </button>
        <div
          className={cn(
            "overflow-hidden transition-all duration-200",
            expanded ? "max-h-40 opacity-100 mt-0.5" : "max-h-0 opacity-0"
          )}
        >
          <div className="flex flex-col gap-0.5">
            {group.children.map((child) => renderNavItem(child, true))}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="w-[180px] h-full bg-background border-r border-border flex flex-col py-3 z-50">
      {/* Logo / Brand */}
      <div className="px-4 mb-4">
        <div
          className="h-9 flex items-center gap-2 cursor-pointer text-primary"
          onClick={() => handleNavClick('chat', '/chat')}
        >
          <MessageSquare className="w-5 h-5" />
          <span className="font-semibold text-sm">WeTrace</span>
        </div>
      </div>

      {/* Main nav */}
      <nav className="flex-1 overflow-y-auto px-2 flex flex-col gap-0.5">
        {navEntries.map((entry) =>
          isGroup(entry) ? renderNavGroup(entry) : renderNavItem(entry)
        )}
      </nav>

      {/* Bottom section: settings + utility buttons */}
      <div className="mt-auto px-2 flex flex-col gap-0.5 pt-2 border-t border-border mx-2">
        {renderNavItem(settingsItem)}

        <div className="flex items-center gap-1 mt-1 px-1">
          <button
            onClick={() => setShowKeyManager(true)}
            className="flex-1 h-8 flex items-center justify-center rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            title="获取微信密钥"
          >
            <Key className="w-4 h-4" />
          </button>
          <button
            onClick={handleFullCache}
            className="flex-1 h-8 flex items-center justify-center rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            title="预加载全量图片"
          >
            <ImageIcon className="w-4 h-4" />
          </button>
          <button
            onClick={handleSync}
            disabled={isSyncing}
            className="flex-1 h-8 flex items-center justify-center rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            title="重新同步数据"
          >
            <RefreshCw className={cn("w-4 h-4", isSyncing && "animate-spin")} />
          </button>
          <button
            onClick={toggleTheme}
            className="flex-1 h-8 flex items-center justify-center rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            title="切换主题"
          >
            <ThemeIcon className="w-4 h-4" />
          </button>
        </div>
      </div>

      {showKeyManager && (
        <KeyManagerModal onClose={() => setShowKeyManager(false)} />
      )}

      <ImageCacheManager />
    </div>
  )
}
