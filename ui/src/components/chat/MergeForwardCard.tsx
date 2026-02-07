import { useState } from 'react'
import type { Message, RecordInfo } from '@/types/message'
import { MergeForwardModal } from './MergeForwardModal'
import { EmojiText } from './EmojiText'

interface Props {
  message: Message
}

export function MergeForwardCard({ message }: Props) {
  const [showModal, setShowModal] = useState(false)
  
  const recordInfo = message.contents?.recordInfo || message.contents
  
  if (!recordInfo) {
    return <div className="text-red-500 text-xs">[合并转发数据错误]</div>
  }

  // Use title from recordInfo, or contents.title, or fallback
  const title = recordInfo.Title || message.contents?.title || '聊天记录'
  
  // Use desc from recordInfo, or contents.desc, or construct from items
  const desc = recordInfo.Desc || message.contents?.desc || ''

  return (
    <>
      <div 
        className="w-64 bg-card rounded-md overflow-hidden border border-border cursor-pointer hover:bg-accent/50 transition-colors shadow-sm"
        onClick={() => setShowModal(true)}
      >
        <div className="p-3">
          <h4 className="text-sm font-medium text-card-foreground mb-1 truncate">
            {title}
          </h4>
          <div className="text-xs text-muted-foreground line-clamp-4 whitespace-pre-wrap leading-relaxed">
            <EmojiText text={desc} />
          </div>
        </div>
        
        <div className="px-3 py-1.5 border-t border-border/50 bg-muted/30">
          <span className="text-[10px] text-muted-foreground/70">
            聊天记录
          </span>
        </div>
      </div>

      <MergeForwardModal 
        isOpen={showModal} 
        onClose={() => setShowModal(false)} 
        recordInfo={recordInfo as RecordInfo}
      />
    </>
  )
}
