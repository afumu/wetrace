import { MessageType, RichMessageSubType, type Message } from "@/types/message"
import { cn } from "@/lib/utils"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { formatMessageTime } from "@/lib/date"
import { TextMessage } from "./message-types/TextMessage"
import { ImageMessage } from "./message-types/ImageMessage"
import { VoiceMessage } from "./message-types/VoiceMessage"
import { VideoMessage } from "./message-types/VideoMessage"
import { MergeForwardCard } from "./MergeForwardCard"
import { EmojiMessage } from "./message-types/EmojiMessage"
import { ReferMessage } from "./message-types/ReferMessage"
import { LinkCardMessage } from "./message-types/LinkCardMessage"
import { RedPacketMessage } from "./message-types/RedPacketMessage"
import { TransferMessage } from "./message-types/TransferMessage"
import { FileCardMessage } from "./message-types/FileCardMessage"
import { mediaApi } from "@/api/media"

interface MessageBubbleProps {
  message: Message
  showAvatar?: boolean
  showTime?: boolean
  showName?: boolean
}

export function MessageBubble({ message, showAvatar = true, showTime = false, showName = true }: MessageBubbleProps) {
  const isSelf = message.isSelf

  // 系统消息或拍一拍特殊处理：居中显示，无头像和气泡
  if (message.type === MessageType.System || (message.type === MessageType.File && message.subType === RichMessageSubType.Pat)) {
    return (
      <div className="flex flex-col items-center my-4 px-10">
        <span className="text-[11px] text-muted-foreground bg-muted/30 px-3 py-0.5 rounded-full text-center leading-relaxed">
          {message.content}
        </span>
      </div>
    )
  }

  // Get avatar URL
  // Logic from Vue component:
  // If self -> sender
  // If other -> 
  //    Group chat -> sender
  //    Private chat -> talker
  let username = ''
  if (isSelf) {
    username = message.sender
  } else {
    username = message.isChatRoom ? message.sender : message.talker
  }
  
  const avatarUrl = message.smallHeadURL || mediaApi.getAvatarUrl(`avatar/${username}`)
  const displayName = message.senderName || message.sender

  const renderContent = () => {
    switch (message.type) {
      case MessageType.Text:
        return <TextMessage content={message.content} isSelf={isSelf} />
      case MessageType.Image:
        return <ImageMessage id={message.id || message.seq} md5={message.contents?.md5} path={message.contents?.path} content={message.content} />
      case MessageType.Voice:
        return <VoiceMessage id={message.contents?.voice} isSelf={isSelf} duration={message.duration} />
      case MessageType.Video:
        return <VideoMessage md5={message.contents?.md5} />
      case MessageType.Emoji:
        return <EmojiMessage contents={message.contents} />
      case MessageType.File:
        if (message.subType === RichMessageSubType.Forwarded) {
          return <MergeForwardCard message={message} />
        }
        if (message.subType === RichMessageSubType.Refer) {
          return <ReferMessage message={message} isSelf={isSelf} />
        }
        if (message.subType === RichMessageSubType.Link || message.subType === RichMessageSubType.VideoLink) {
          return <LinkCardMessage message={message} />
        }
        if (message.subType === RichMessageSubType.RedPacket) {
          return <RedPacketMessage message={message} />
        }
        if (message.subType === RichMessageSubType.Transfer) {
          return <TransferMessage message={message} />
        }
        if (message.subType === RichMessageSubType.File) {
          return <FileCardMessage message={message} />
        }
        // Fallback for other file types
        return (
          <div className="text-xs italic">
            [文件/链接消息: {message.subType}] {message.content}
          </div>
        )
      default:
        // Fallback for unknown types
        return (
          <div className="text-xs text-muted-foreground italic">
            [不支持的消息类型: {message.type}] {message.content}
          </div>
        )
    }
  }

  // Check if we should remove bubble styling for custom cards like MergeForward or Emoji
  // Refer message keeps the bubble styling
  const isCustomCard = (message.type === MessageType.File && (
    message.subType === RichMessageSubType.Forwarded || 
    message.subType === RichMessageSubType.Refer ||
    message.subType === RichMessageSubType.Link ||
    message.subType === RichMessageSubType.VideoLink ||
    message.subType === RichMessageSubType.RedPacket ||
    message.subType === RichMessageSubType.Transfer ||
    message.subType === RichMessageSubType.File
  )) || message.type === MessageType.Emoji

  return (
    <div className={cn("flex flex-col mb-4 px-4", isSelf ? "items-end" : "items-start")}>
      {showTime && (
        <div className="w-full text-center my-4">
          <span className="text-xs text-muted-foreground bg-muted/30 px-2 py-1 rounded-full">
            {formatMessageTime(message.createTime)}
          </span>
        </div>
      )}

      <div className={cn("flex max-w-[80%] gap-2", isSelf ? "flex-row-reverse" : "flex-row")}>
        {showAvatar && (
          <Avatar className="w-9 h-9 mt-1 shrink-0 rounded-md">
            <AvatarImage src={avatarUrl} alt={displayName} />
            <AvatarFallback className="rounded-md">{displayName?.slice(0, 1)}</AvatarFallback>
          </Avatar>
        )}

        <div className={cn("flex flex-col", isSelf ? "items-end" : "items-start")}>
          {showName && !isSelf && message.isChatRoom && (
            <span className="text-[10px] text-muted-foreground mb-1 ml-1">
              {displayName}
            </span>
          )}
          
          <div className={cn(
            "rounded-lg shadow-sm relative group max-w-full break-all",
            !isCustomCard && "px-3 py-2",
            !isCustomCard && "before:content-[''] before:absolute before:top-3 before:border-[6px] before:border-transparent",
            !isCustomCard && isSelf 
              ? "bg-[hsl(var(--chat-self-bg))] text-[hsl(var(--chat-self-text))] before:right-[-12px] before:border-l-[hsl(var(--chat-self-bg))]" // Wechat Green
              : !isCustomCard && "bg-card before:left-[-12px] before:border-r-card"
          )}>
            {renderContent()}
          </div>
        </div>
      </div>
    </div>
  )
}