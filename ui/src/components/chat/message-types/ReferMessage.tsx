import { MessageType, type Message } from '@/types/message'
import { cn } from '@/lib/utils'
import { EmojiText } from '../EmojiText'
import { useState } from 'react'

interface ReferMessageProps {
  message: Message
  isSelf?: boolean
}

export function ReferMessage({ message, isSelf }: ReferMessageProps) {
  const refer = message.contents?.refer
  const mainContent = message.content
  const [imageError, setImageError] = useState(false)

  // 引用部分的内容
  const referName = refer?.senderName || refer?.sender || '未知用户'
  
  const renderReferContent = () => {
    if (!refer) return '[引用内容缺失]'

    // 图片类型
    if (refer.type === MessageType.Image && refer.contents?.md5 && !imageError) {
      return (
        <img 
          src={`/api/v1/media/image/${refer.contents.md5}`} 
          alt="图片" 
          className="h-8 w-auto rounded border border-border/50 cursor-pointer hover:opacity-90"
          loading="lazy"
          onError={() => setImageError(true)}
          onClick={(e) => {
            e.stopPropagation()
            window.open(`/api/v1/media/image/${refer.contents?.md5}`, '_blank')
          }}
        />
      )
    }

    if (refer.type === MessageType.Image) {
      return <span>[图片]</span>
    }

    // 默认文本 (使用 EmojiText 解析)
    return <EmojiText text={refer.content || '[不支持的消息类型]'} />
  }

  return (
    <div className={cn("flex flex-col max-w-full", isSelf ? "items-end" : "items-start")}>
      {/* 1. 主消息气泡 (复用 MessageBubble 的样式逻辑) */}
      <div className={cn(
        "rounded-lg px-3 py-2 shadow-sm relative group break-all mb-1",
        "before:content-[''] before:absolute before:top-3 before:border-[6px] before:border-transparent",
        isSelf 
          ? "bg-[hsl(var(--chat-self-bg))] text-[hsl(var(--chat-self-text))] before:right-[-12px] before:border-l-[hsl(var(--chat-self-bg))]" 
          : "bg-card before:left-[-12px] before:border-r-card"
      )}>
        <span className="whitespace-pre-wrap"><EmojiText text={mainContent} /></span>
      </div>
      
      {/* 2. 引用内容区域 (显示在下方) */}
      <div className={cn(
        "text-xs px-3 py-2 rounded-md max-w-full w-fit",
        "bg-muted/60 text-muted-foreground"
      )}>
        <div className="flex items-center gap-1 line-clamp-3">
          <span className="font-medium shrink-0">{referName}:</span>
          {renderReferContent()}
        </div>
      </div>
    </div>
  )
}