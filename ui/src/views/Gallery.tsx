import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { mediaApi } from "@/api"
import type { ImageListItem } from "@/api/media"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  ImageIcon,
  X,
  ChevronLeft,
  ChevronRight,
} from "lucide-react"

export default function GalleryView() {
  const [talkerFilter, setTalkerFilter] = useState("")
  const [timeRange, setTimeRange] = useState("all")
  const [offset, setOffset] = useState(0)
  const [previewItem, setPreviewItem] = useState<ImageListItem | null>(null)
  const limit = 50

  const { data, isLoading } = useQuery({
    queryKey: ["gallery-images", talkerFilter, timeRange, offset],
    queryFn: () =>
      mediaApi.getImageList({
        talker: talkerFilter || undefined,
        time_range: timeRange !== "all" ? timeRange : undefined,
        limit,
        offset,
      }),
  })

  const totalPages = data ? Math.ceil(data.total / limit) : 0
  const currentPage = Math.floor(offset / limit) + 1

  return (
    <div className="flex flex-col h-full bg-background">
      {/* Header */}
      <div className="p-6 pb-0 max-w-5xl mx-auto w-full">
        <div className="mb-6">
          <h2 className="text-2xl font-bold tracking-tight">图片画廊</h2>
          <p className="text-sm text-muted-foreground mt-1">
            浏览所有聊天中的图片
          </p>
        </div>

        <GalleryFilters
          talkerFilter={talkerFilter}
          setTalkerFilter={setTalkerFilter}
          timeRange={timeRange}
          setTimeRange={setTimeRange}
          onReset={() => {
            setTalkerFilter("")
            setTimeRange("all")
            setOffset(0)
          }}
        />

        {data && (
          <div className="text-sm text-muted-foreground mb-4">
            共 <span className="font-bold text-foreground">{data.total}</span> 张图片
          </div>
        )}
      </div>

      {/* Image grid */}
      <ScrollArea className="flex-1 px-6">
        <div className="max-w-5xl mx-auto w-full pb-20">
          <ImageGrid
            items={data?.items}
            isLoading={isLoading}
            onPreview={setPreviewItem}
          />

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-4 mt-6">
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
      </ScrollArea>

      {/* Lightbox */}
      {previewItem && (
        <ImageLightbox
          item={previewItem}
          items={data?.items || []}
          onClose={() => setPreviewItem(null)}
          onNavigate={setPreviewItem}
        />
      )}
    </div>
  )
}

/* ============================================================
 * Gallery Filters
 * ============================================================ */
function GalleryFilters({
  talkerFilter,
  setTalkerFilter,
  timeRange,
  setTimeRange,
  onReset,
}: {
  talkerFilter: string
  setTalkerFilter: (v: string) => void
  timeRange: string
  setTimeRange: (v: string) => void
  onReset: () => void
}) {
  const timeOptions = [
    { value: "all", label: "全部" },
    { value: "last_week", label: "最近一周" },
    { value: "last_month", label: "最近一月" },
    { value: "last_year", label: "最近一年" },
  ]

  return (
    <div className="flex flex-wrap items-center gap-3 mb-4">
      <div className="flex-1 min-w-[200px]">
        <Input
          value={talkerFilter}
          onChange={(e) => setTalkerFilter(e.target.value)}
          placeholder="按会话ID筛选..."
          className="h-9"
        />
      </div>
      <div className="flex gap-1">
        {timeOptions.map((opt) => (
          <Button
            key={opt.value}
            variant={timeRange === opt.value ? "default" : "outline"}
            size="sm"
            className="text-xs"
            onClick={() => setTimeRange(opt.value)}
          >
            {opt.label}
          </Button>
        ))}
      </div>
      {(talkerFilter || timeRange !== "all") && (
        <Button variant="ghost" size="sm" className="text-xs gap-1" onClick={onReset}>
          <X className="w-3 h-3" />
          清除
        </Button>
      )}
    </div>
  )
}

/* ============================================================
 * Image Grid
 * ============================================================ */
function ImageGrid({
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
    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
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
            <p className="text-white text-[10px] truncate">{item.talkerName}</p>
            <p className="text-white/70 text-[10px]">
              {new Date(item.time).toLocaleDateString()}
            </p>
          </div>
        </div>
      ))}
    </div>
  )
}

/* ============================================================
 * Image Lightbox
 * ============================================================ */
function ImageLightbox({
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
      className="fixed inset-0 z-50 bg-black/80 flex items-center justify-center"
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
        src={item.thumbnailUrl}
        alt=""
        className="max-w-[90vw] max-h-[85vh] object-contain rounded-lg"
        onClick={(e) => e.stopPropagation()}
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
        {item.talkerName} - {new Date(item.time).toLocaleString()}
        {items.length > 1 && ` (${currentIndex + 1}/${items.length})`}
      </div>
    </div>
  )
}