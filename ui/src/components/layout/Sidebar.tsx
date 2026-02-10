import { useAppStore } from "@/stores/app"
import { cn } from "@/lib/utils"
import { MessageSquare, RefreshCw, Moon, Sun, Monitor, Search, Key, ImageIcon, CalendarDays, Heart, Cloud, BrainCircuit, Users, Settings, Shield, PlayCircle } from "lucide-react"
import { useNavigate } from "react-router-dom"
import { useState } from "react"
import { systemApi, mediaApi } from "@/api"
import { toast } from "sonner"
import { KeyManagerModal } from "./KeyManagerModal"
import { ImageCacheManager } from "../chat/ImageCacheManager"

export function Sidebar() {
  const { activeNav, setActiveNav, toggleTheme, settings } = useAppStore()
  const navigate = useNavigate()
  const [isSyncing, setIsSyncing] = useState(false)
  const [showKeyManager, setShowKeyManager] = useState(false)

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

  const navItems = [
    { key: 'chat', icon: MessageSquare, label: '聊天', path: '/chat' },
    { key: 'contacts', icon: Users, label: '联系人', path: '/contacts' },
    { key: 'search', icon: Search, label: '搜索', path: '/search' },
    { key: 'gallery', icon: ImageIcon, label: '图片', path: '/gallery' },
    { key: 'report', icon: CalendarDays, label: '年度报告', path: '/report' },
    { key: 'sentiment', icon: Heart, label: '情感分析', path: '/sentiment' },
    { key: 'wordcloud', icon: Cloud, label: '词云', path: '/wordcloud' },
    { key: 'ai-tools', icon: BrainCircuit, label: 'AI工具箱', path: '/ai-tools' },
    { key: 'monitor', icon: Shield, label: '监控配置', path: '/monitor' },
    { key: 'replay', icon: PlayCircle, label: '对话回放', path: '/replay' },
    { key: 'settings', icon: Settings, label: '设置', path: '/settings' },
  ]

  const handleNavClick = (key: string, path: string) => {
    setActiveNav(key)
    navigate(path)
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

  const ThemeIcon = settings.theme === 'dark' ? Moon : settings.theme === 'light' ? Sun : Monitor

  return (
    <div className="w-[60px] h-full bg-background border-r border-border flex flex-col items-center py-4 z-50">
      <div className="mb-8">
        <div 
          className="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center text-primary cursor-pointer hover:bg-primary/20 transition-colors"
          onClick={() => {
            setActiveNav('chat')
            navigate('/chat')
          }}
        >
          <MessageSquare className="w-6 h-6" />
        </div>
      </div>

      <div className="flex-1 w-full flex flex-col gap-2 px-2">
        {navItems.map((item) => (
          <button
            key={item.key}
            onClick={() => handleNavClick(item.key, item.path)}
            className={cn(
              "w-full aspect-square flex items-center justify-center rounded-lg transition-colors",
              activeNav === item.key 
                ? "bg-primary/10 text-primary" 
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            )}
            title={item.label}
          >
            <item.icon className="w-5 h-5" />
          </button>
        ))}
      </div>

      <div className="mt-auto w-full flex flex-col gap-2 px-2">
        <button
          onClick={() => setShowKeyManager(true)}
          className={cn(
            "w-full aspect-square flex items-center justify-center rounded-lg transition-colors",
            "text-muted-foreground hover:bg-muted hover:text-foreground"
          )}
          title="获取微信密钥"
        >
          <Key className="w-5 h-5" />
        </button>

        <button
          onClick={handleFullCache}
          className={cn(
            "w-full aspect-square flex items-center justify-center rounded-lg transition-colors",
            "text-muted-foreground hover:bg-muted hover:text-foreground"
          )}
          title="预加载全量图片"
        >
          <ImageIcon className="w-5 h-5" />
        </button>

        <button
          onClick={handleSync}
          disabled={isSyncing}
          className={cn(
            "w-full aspect-square flex items-center justify-center rounded-lg transition-colors",
            "text-muted-foreground hover:bg-muted hover:text-foreground"
          )}
          title="重新同步数据"
        >
          <RefreshCw className={cn("w-5 h-5", isSyncing && "animate-spin")} />
        </button>
        
        <button
          onClick={toggleTheme}
          className="w-full aspect-square flex items-center justify-center rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
          title="切换主题"
        >
          <ThemeIcon className="w-5 h-5" />
        </button>
      </div>

      {showKeyManager && (
        <KeyManagerModal onClose={() => setShowKeyManager(false)} />
      )}

      <ImageCacheManager />
    </div>
  )
}
