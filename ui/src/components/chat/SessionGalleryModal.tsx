import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { mediaApi } from "@/api"
import type { ImageListItem } from "@/api/media"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import {
  X,
  ChevronLeft,
  ChevronRight,
  ImageIcon,
} from "lucide-react"

interface SessionGalleryModalProps {
  talker: string
  displayName: string
  isOpen: boolean
  onClose: () => void
}

export function SessionGalleryModal({
  talker,
  displayName,
  isOpen,
  onClose,
}: SessionGalleryModalProps) {
  const [offset, setOffset] = useState(0)
  const [previewItem, setPreviewItem] = useState<ImageListItem | null>(null)
  const limit = 50

  const { data, isLoading } = useQuery({
    queryKey: ["session-gallery", talker, offset],
    queryFn: () =>
      mediaApi.getImageList({
        talker,
        limit,
        offset,
      }),
    enabled: isOpen && !!talker,
  })

  const totalPages = data ? Math.ceil(data.total / limit) : 0
  const currentPage = Math.floor(offset / limit) + 1

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 bg-black/60 flex items-center justify-center p-4">
      <div
        className="bg-background rounded-xl shadow-2xl w-full max-w-4xl max-h-[85vh] flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b">
          <div>
            <h3 className="text-lg font-semibold">会话图片</h3>
            <p className="text-sm text-muted-foreground">
              {displayName}
              {data && ` - 共 ${data.total} 张图片`}
            </p>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="w-5 h-5" />
          </Button>
        </div>

        {/* Content */}
        <ScrollArea className="flex-1 p-6">
          <SessionImageGrid
            items={data?.items}
            isLoading={isLoading}
            onPreview={setPreviewItem}
          />
        </ScrollArea>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-4 px-6 py-3 border-t">
            <Button
              variant="outline"
              size="sm"
              disabled={currentPage <= 1}
              onClick={() => setOffset(Math.max(0, offset - limit))}
            >
              上一页
            </Button>
            <span className="text-sm text-muted-foreground">
              第 {currentPage} / {totalPages} 页
            </span>
            <Button
              variant="outline"
              size="sm"
              disabled={currentPage >= totalPages}
              onClick={() => setOffset(offset + limit)}
            >
              下一页
            </Button>
          </div>
        )}
      </div>

      {/* Lightbox */}
      {previewItem && (
        <SessionImageLightbox
          item={previewItem}
          items={data?.items || []}
          onClose={() => setPreviewItem(null)}
          onNavigate={setPreviewItem}
        />
      )}
    </div>
  )
}

function SessionImageGrid({
  items,
  isLoading,
  onPreview,
}: {
  items: ImageListItem[] | undefined
  isLoading: boolean
  onPreview: (item: ImageListItem) => void
}) {
  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-4">
        <div className="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin" />
        <p className="text-muted-foreground animate-pulse text-sm">加载中...</p>
      </div>
    )
  }

  if (!items || items.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-4">
        <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center">
          <ImageIcon className="w-8 h-8 text-muted-foreground/30" />
        </div>
        <p className="text-muted-foreground text-sm font-medium">暂无图片</p>
      </div>
    )
  }

  return (
    <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 gap-3">
      {items.map((item) => (
        <div
          key={`${item.key}-${item.seq}`}
          className="group relative aspect-square rounded-lg overflow-hidden bg-muted cursor-pointer hover:ring-2 hover:ring-primary/40 transition-all"
          onClick={() => onPreview(item)}
        >
          <img
            src={item.thumbnailUrl}
            alt=""
            className="w-full h-full object-cover"
            loading="lazy"
          />
          <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/60 to-transparent p-2 opacity-0 group-hover:opacity-100 transition-opacity">
            <p className="text-white/70 text-[10px]">
              {new Date(item.time).toLocaleDateString()}
            </p>
          </div>
        </div>
      ))}
    </div>
  )
}

function SessionImageLightbox({
  item,
  items,
  onClose,
  onNavigate,
}: {
  item: ImageListItem
  items: ImageListItem[]
  onClose: () => void
  onNavigate: (item: ImageListItem) => void
}) {
  const currentIndex = items.findIndex(
    (i) => i.key === item.key && i.seq === item.seq
  )
  const hasPrev = currentIndex > 0
  const hasNext = currentIndex < items.length - 1

  const handlePrev = () => {
    if (hasPrev) onNavigate(items[currentIndex - 1])
  }
  const handleNext = () => {
    if (hasNext) onNavigate(items[currentIndex + 1])
  }

  return (
    <div
      className="fixed inset-0 z-[60] bg-black/80 flex items-center justify-center"
      onClick={onClose}
      tabIndex={0}
      role="dialog"
      onKeyDown={(e) => {
        if (e.key === "Escape") onClose()
        if (e.key === "ArrowLeft") handlePrev()
        if (e.key === "ArrowRight") handleNext()
      }}
    >
      <button
        className="absolute top-4 right-4 text-white/70 hover:text-white z-10"
        onClick={onClose}
      >
        <X className="w-6 h-6" />
      </button>

      {hasPrev && (
        <button
          className="absolute left-4 top-1/2 -translate-y-1/2 text-white/70 hover:text-white z-10"
          onClick={(e) => { e.stopPropagation(); handlePrev() }}
        >
          <ChevronLeft className="w-8 h-8" />
        </button>
      )}

      <img
        src={item.fullUrl || item.thumbnailUrl}
        alt=""
        className="max-w-[90vw] max-h-[85vh] object-contain rounded-lg"
        onClick={(e) => e.stopPropagation()}
        onError={(e) => {
          const target = e.target as HTMLImageElement
          if (item.thumbnailUrl && target.src !== item.thumbnailUrl && !target.src.endsWith(item.thumbnailUrl)) {
            target.src = item.thumbnailUrl
          }
        }}
      />

      {hasNext && (
        <button
          className="absolute right-4 top-1/2 -translate-y-1/2 text-white/70 hover:text-white z-10"
          onClick={(e) => { e.stopPropagation(); handleNext() }}
        >
          <ChevronRight className="w-8 h-8" />
        </button>
      )}

      <div className="absolute bottom-4 left-1/2 -translate-x-1/2 bg-black/60 text-white text-xs px-4 py-2 rounded-full">
        {new Date(item.time).toLocaleString()}
        {items.length > 1 && ` (${currentIndex + 1}/${items.length})`}
      </div>
    </div>
  )
}
