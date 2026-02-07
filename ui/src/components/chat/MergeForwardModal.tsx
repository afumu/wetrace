import { X } from 'lucide-react'
import type { RecordInfo, RecordItem } from '@/types/message'
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar'
import { EmojiText } from './EmojiText'

interface Props {
  isOpen: boolean
  onClose: () => void
  recordInfo: RecordInfo
}

export function MergeForwardModal({ isOpen, onClose, recordInfo }: Props) {
  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 animate-in fade-in duration-200">
      <div className="bg-background w-full max-w-2xl max-h-[80vh] rounded-lg shadow-xl flex flex-col overflow-hidden animate-in zoom-in-95 duration-200 border border-border">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-border bg-muted/40">
          <h3 className="font-medium text-foreground truncate pr-4">
            {recordInfo.Title || '聊天记录'}
          </h3>
          <button 
            onClick={onClose}
            className="p-1 rounded-full hover:bg-muted transition-colors text-muted-foreground hover:text-foreground"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content List */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {recordInfo.DataList?.DataItems?.map((item, index) => (
            <MergeForwardItem key={item.DataID || index} item={item} />
          ))}
        </div>
      </div>
    </div>
  )
}

function MergeForwardItem({ item }: { item: RecordItem }) {
  const isImage = item.DataType === '2'
  
  return (
    <div className="flex gap-3 items-start">
      <Avatar className="w-10 h-10 mt-0.5 border border-border">
        <AvatarImage src={item.SourceHeadURL} />
        <AvatarFallback className="text-muted-foreground bg-muted">{item.SourceName?.slice(0, 1)}</AvatarFallback>
      </Avatar>
      
      <div className="flex-1 min-w-0">
        <div className="flex items-baseline gap-2 mb-1">
          <span className="text-sm font-medium text-muted-foreground">
            {item.SourceName}
          </span>
          <span className="text-xs text-muted-foreground/70">
            {item.SourceTime}
          </span>
        </div>
        
        <div className="text-foreground text-sm leading-relaxed break-words">
          {isImage ? (
            <MergeForwardImage item={item} />
          ) : (
            <div className="whitespace-pre-wrap"><EmojiText text={item.DataDesc} /></div>
          )}
        </div>
      </div>
    </div>
  )
}

function MergeForwardImage({ item }: { item: RecordItem }) {
  const imageUrl = item.FullMD5 ? `/api/v1/media/image_merge/${item.FullMD5}` : ''
  const thumbUrl = item.ThumbFullMD5 ? `/api/v1/media/image_merge/${item.ThumbFullMD5}` : ''
  
  if (!imageUrl && !thumbUrl) {
    return (
      <div className="bg-muted p-4 rounded text-xs text-muted-foreground border border-dashed border-border w-fit">
        [图片无法加载: 缺少 MD5]
      </div>
    )
  }

  return (
    <div className="w-fit max-w-sm rounded-lg overflow-hidden border border-border bg-muted/30">
      <img 
        src={imageUrl || thumbUrl} 
        alt="聊天图片" 
        className="max-w-full h-auto max-h-[300px] object-contain cursor-zoom-in block"
        loading="lazy"
        onClick={() => window.open(imageUrl || thumbUrl, '_blank')}
        onError={(e) => {
          const target = e.target as HTMLImageElement
          if (thumbUrl && target.src !== thumbUrl && !target.src.endsWith(thumbUrl)) {
             target.src = thumbUrl
          } else {
             target.style.display = 'none'
             target.parentElement!.innerHTML = '<div class="p-4 text-xs text-muted-foreground">[图片加载失败]</div>'
          }
        }}
      />
    </div>
  )
}