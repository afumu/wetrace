import { useState } from 'react'

interface EmojiMessageProps {
  contents?: {
    cdnurl?: string
    aeskey?: string
    [key: string]: any
  }
}

export function EmojiMessage({ contents }: EmojiMessageProps) {
  const [error, setError] = useState(false)

  if (!contents?.cdnurl || !contents?.aeskey) {
    return <div className="text-xs text-red-500">[表情包参数缺失]</div>
  }

  const url = `/api/v1/media/emoji?url=${encodeURIComponent(contents.cdnurl)}&key=${contents.aeskey}`

  if (error) {
    return (
      <div 
        className="w-[120px] h-[120px] bg-gray-100 dark:bg-gray-800 flex items-center justify-center rounded text-xs text-gray-400"
        title="表情加载失败"
      >
        [表情]
      </div>
    )
  }

  return (
    <img 
      src={url} 
      alt="表情" 
      className="max-w-[150px] max-h-[150px] object-contain cursor-pointer hover:opacity-90 transition-opacity"
      loading="lazy"
      onError={() => setError(true)}
      onClick={() => window.open(url, '_blank')}
    />
  )
}
