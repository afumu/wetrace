import { getEmojiPath } from 'wechat-emojis'

interface Props {
  text: string
  className?: string
}

export function EmojiText({ text, className }: Props) {
  if (!text) return null

  // Split text by [emoji] pattern
  const parts = text.split(/(\[[^\]]+\])/g)

  return (
    <span className={className}>
      {parts.map((part, index) => {
        if (part.startsWith('[') && part.endsWith(']')) {
          // Extract name without brackets
          const name = part.slice(1, -1)
          // Get path from library
          const path = getEmojiPath(name as any)
          
          if (path) {
            // Ensure path starts with / for absolute path from public root
            const fullPath = path.startsWith('/') ? path : `/${path}`
            
            return (
              <img
                key={index}
                src={fullPath}
                alt={name}
                title={part}
                className="inline-block w-[1.4em] h-[1.4em] align-text-bottom mx-px"
              />
            )
          }
        }
        return <span key={index}>{part}</span>
      })}
    </span>
  )
}