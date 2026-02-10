import { useState, useRef, useEffect } from "react"
import { aiApi } from "@/api/ai"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Send, Loader2, Bot, X } from "lucide-react"
import { cn } from "@/lib/utils"
import { EmojiText } from "../chat/EmojiText"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"

interface Message {
  role: 'user' | 'ai'
  content: string
}

interface AISimulateChatProps {
  talker: string
  displayName: string
  contactAvatar?: string
  selfAvatar?: string
  onClose: () => void
}

export function AISimulateChat({ talker, displayName, contactAvatar, selfAvatar, onClose }: AISimulateChatProps) {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [messages])

  const handleSend = async () => {
    if (!input.trim() || isLoading) return

    const userMsg = input.trim()
    setInput("")
    setMessages(prev => [...prev, { role: 'user', content: userMsg }])
    setIsLoading(true)

    try {
      const res = await aiApi.simulate({ talker, message: userMsg })
      
      // 模拟“正在输入”的真实感
      // 基础延迟 800ms + 每字 50-100ms 的随机打字时间，最高不超过 4 秒
      const typingSpeed = 50 + Math.random() * 50
      const delay = Math.min(800 + (res.length * typingSpeed), 4000)
      
      await new Promise(resolve => setTimeout(resolve, delay))
      
      setMessages(prev => [...prev, { role: 'ai', content: res }])
    } catch (err) {
      console.error("AI simulation failed:", err)
      setMessages(prev => [...prev, { role: 'ai', content: "抱歉，我现在无法模拟回复。请检查 AI 配置。" }])
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="absolute inset-0 bg-background flex flex-col z-50 animate-in slide-in-from-right duration-300">
      <div className="h-14 border-b flex items-center justify-between px-4 bg-muted/20">
        <div className="flex items-center gap-2">
          <Bot className="w-5 h-5 text-primary" />
          <h3 className="font-medium text-sm">与 {displayName} (AI 模拟) 对话</h3>
        </div>
        <Button variant="ghost" size="icon" onClick={onClose}>
          <X className="w-5 h-5" />
        </Button>
      </div>

      <ScrollArea className="flex-1 p-4" ref={scrollRef}>
        <div className="flex flex-col gap-4">
          <div className="bg-muted/50 p-3 rounded-lg text-xs text-muted-foreground italic">
            AI 已经学习了与 {displayName} 的历史聊天记录，现在你可以尝试与“他/她”直接对话。
          </div>
          {messages.map((msg, i) => (
            <div key={i} className={cn(
              "flex gap-3 max-w-[85%]",
              msg.role === 'user' ? "ml-auto flex-row-reverse" : "mr-auto"
            )}>
              <Avatar className="w-8 h-8 shrink-0 rounded-md">
                <AvatarImage src={msg.role === 'user' ? selfAvatar : contactAvatar} />
                <AvatarFallback className="rounded-md bg-primary text-primary-foreground text-[10px]">
                  {msg.role === 'user' ? "我" : displayName.slice(0, 1)}
                </AvatarFallback>
              </Avatar>
              <div className={cn(
                "p-3 rounded-2xl text-sm leading-relaxed",
                msg.role === 'user' ? "bg-primary text-primary-foreground rounded-tr-none" : "bg-muted rounded-tl-none"
              )}>
                <EmojiText text={msg.content} />
              </div>
            </div>
          ))}
          {isLoading && (
            <div className="flex gap-3 max-w-[85%] mr-auto">
              <Avatar className="w-8 h-8 shrink-0 rounded-md">
                <AvatarImage src={contactAvatar} />
                <AvatarFallback className="rounded-md bg-muted text-[10px]">
                  {displayName.slice(0, 1)}
                </AvatarFallback>
              </Avatar>
              <div className="bg-muted p-3 rounded-2xl rounded-tl-none flex items-center">
                <Loader2 className="w-4 h-4 animate-spin" />
              </div>
            </div>
          )}
        </div>
      </ScrollArea>

      <div className="p-4 border-t bg-background">
        <form 
          className="flex gap-2" 
          onSubmit={(e) => {
            e.preventDefault()
            handleSend()
          }}
        >
          <Input 
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="输入消息，由 AI 模拟对方回复..."
            disabled={isLoading}
            className="flex-1"
          />
          <Button type="submit" size="icon" disabled={isLoading || !input.trim()}>
            <Send className="w-4 h-4" />
          </Button>
        </form>
      </div>
    </div>
  )
}
