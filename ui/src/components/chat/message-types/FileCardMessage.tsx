import type { Message } from "@/types/message"
import { mediaApi } from "@/api/media"
import { FileText, Music, Video, Image as ImageIcon } from "lucide-react"

interface FileCardMessageProps {
  message: Message
}

export function FileCardMessage({ message }: FileCardMessageProps) {
  const { md5, title } = message.contents || {}
  const fileName = title || "未知文件"
  
  // Determine icon based on file extension
  const getFileIcon = (name: string) => {
    const ext = name.split('.').pop()?.toLowerCase()
    if (['mp3', 'wav', 'm4a', 'flac', 'aac'].includes(ext || '')) {
      return <Music className="w-8 h-8 text-blue-500" />
    }
    if (['mp4', 'mov', 'avi', 'mkv'].includes(ext || '')) {
      return <Video className="w-8 h-8 text-purple-500" />
    }
    if (['jpg', 'jpeg', 'png', 'gif', 'webp'].includes(ext || '')) {
      return <ImageIcon className="w-8 h-8 text-green-500" />
    }
    return <FileText className="w-8 h-8 text-gray-500" />
  }

  const handleOpen = () => {
    if (md5) {
      const url = mediaApi.getFileUrl(md5)
      window.open(url, '_blank')
    }
  }

  return (
    <div 
      className="flex items-center gap-3 p-3 w-[240px] bg-card border border-border/50 shadow-sm rounded-lg cursor-pointer hover:bg-muted/50 transition-colors"
      onClick={handleOpen}
      title="点击打开文件"
    >
      <div className="w-12 h-12 bg-muted/50 rounded flex items-center justify-center shrink-0">
        {getFileIcon(fileName)}
      </div>
      
      <div className="flex-1 min-w-0 flex flex-col justify-center">
        <span className="text-sm font-medium text-foreground truncate break-all line-clamp-2 leading-tight mb-1">
          {fileName}
        </span>
        <div className="flex items-center gap-1 text-[10px] text-muted-foreground">
          <span className="truncate">点击预览或查看</span>
        </div>
      </div>
    </div>
  )
}