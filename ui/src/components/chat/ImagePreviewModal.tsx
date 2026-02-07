import { useEffect, useCallback } from "react"
import { createPortal } from "react-dom"
import { useImagePreviewStore } from "@/stores/image-preview"
import { mediaApi } from "@/api/media"
import { X, ChevronLeft, ChevronRight, Download, ExternalLink } from "lucide-react"

export function ImagePreviewModal() {
  const { isOpen, currentIndex, images, closePreview, nextImage, prevImage } = useImagePreviewStore()

  const currentImage = images[currentIndex]

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (!isOpen) return
    switch (e.key) {
      case "Escape":
        closePreview()
        break
      case "ArrowRight":
        nextImage()
        break
      case "ArrowLeft":
        prevImage()
        break
    }
  }, [isOpen, closePreview, nextImage, prevImage])

  useEffect(() => {
    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [handleKeyDown])

  if (!isOpen || !currentImage) return null

  const imageUrl = currentImage.content || mediaApi.getImageUrl(currentImage.md5, currentImage.path)

  return createPortal(
    <div 
      className="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-sm animate-in fade-in duration-200 select-none"
      onClick={closePreview}
    >
      <div 
        className="relative w-full h-full flex items-center justify-center" 
        onClick={e => e.stopPropagation()}
      >
        {/* Main Content Area: Left Button - Image Container (Image + Controls) - Right Button */}
        <div className="flex items-center justify-center gap-6 w-full h-full px-4">
          
          {/* Left Button */}
          <div className="flex-shrink-0 w-14 flex justify-end">
            {currentIndex > 0 && (
              <button
                className="p-3 bg-white/10 text-white/70 hover:text-white hover:bg-white/20 rounded-full transition-all backdrop-blur-md hover:scale-110 active:scale-95"
                onClick={(e) => {
                  e.stopPropagation()
                  prevImage()
                }}
                title="上一张 (←)"
              >
                <ChevronLeft className="w-8 h-8" />
              </button>
            )}
          </div>

          {/* Image Container */}
          <div className="relative flex flex-col items-center">
            {/* Controls - Positioned relative to image top right */}
            <div className="absolute -top-12 right-0 flex gap-3 animate-in slide-in-from-bottom-2 duration-300">
               <a 
                href={imageUrl} 
                download={`image-${currentImage.md5 || 'download'}.jpg`}
                className="p-2 bg-black/40 text-white/80 hover:text-white hover:bg-black/60 rounded-full transition-all backdrop-blur-md ring-1 ring-white/10"
                title="下载图片"
                onClick={(e) => e.stopPropagation()}
              >
                <Download className="w-5 h-5" />
              </a>
               <a 
                href={imageUrl} 
                target="_blank"
                rel="noreferrer"
                className="p-2 bg-black/40 text-white/80 hover:text-white hover:bg-black/60 rounded-full transition-all backdrop-blur-md ring-1 ring-white/10"
                title="在新标签页打开"
                onClick={(e) => e.stopPropagation()}
              >
                <ExternalLink className="w-5 h-5" />
              </a>
              <button 
                className="p-2 bg-red-500/40 text-white/80 hover:text-white hover:bg-red-500/60 rounded-full transition-all backdrop-blur-md ring-1 ring-white/10"
                onClick={closePreview}
                title="关闭 (Esc)"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            
            {/* Image */}
            <img
              src={imageUrl}
              alt="Preview"
              className="max-w-[calc(100vw-200px)] max-h-[85vh] object-contain shadow-[0_0_50px_rgba(0,0,0,0.6)] ring-1 ring-white/10 transition-all duration-300 rounded-sm bg-black/20"
            />
          </div>

          {/* Right Button */}
          <div className="flex-shrink-0 w-14 flex justify-start">
            {currentIndex < images.length - 1 && (
              <button
                className="p-3 bg-white/10 text-white/70 hover:text-white hover:bg-white/20 rounded-full transition-all backdrop-blur-md hover:scale-110 active:scale-95"
                onClick={(e) => {
                  e.stopPropagation()
                  nextImage()
                }}
                title="下一张 (→)"
              >
                <ChevronRight className="w-8 h-8" />
              </button>
            )}
          </div>
        </div>
        
        {/* Counter */}
        <div className="absolute bottom-10 left-1/2 -translate-x-1/2 px-4 py-1.5 bg-black/60 backdrop-blur-md text-white/90 rounded-full text-xs font-medium ring-1 ring-white/10">
          {currentIndex + 1} / {images.length}
        </div>
      </div>
    </div>,
    document.body
  )
}
