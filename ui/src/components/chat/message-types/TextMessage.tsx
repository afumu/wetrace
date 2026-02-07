import { cn } from "@/lib/utils"
import { EmojiText } from "../EmojiText"

interface TextMessageProps {
  content: string
  isSelf: boolean
}

export function TextMessage({ content, isSelf }: TextMessageProps) {
  return (
    <div className={cn(
      "whitespace-pre-wrap break-words text-sm",
      isSelf ? "text-[hsl(var(--chat-self-text))]" : "text-foreground"
    )}>
      <EmojiText text={content} />
    </div>
  )
}
