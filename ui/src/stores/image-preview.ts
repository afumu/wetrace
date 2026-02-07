import { create } from 'zustand'

export interface PreviewImage {
  id: string | number // Message ID or unique key
  md5: string
  path?: string
  content?: string
}

interface ImagePreviewState {
  isOpen: boolean
  currentIndex: number
  images: PreviewImage[]
  
  // Actions
  openPreview: (initialImageId: string | number) => void
  closePreview: () => void
  nextImage: () => void
  prevImage: () => void
  setImages: (images: PreviewImage[]) => void
}

export const useImagePreviewStore = create<ImagePreviewState>((set, get) => ({
  isOpen: false,
  currentIndex: 0,
  images: [],

  setImages: (images) => set({ images }),

  openPreview: (initialImageId) => {
    const { images } = get()
    const index = images.findIndex(img => img.id === initialImageId)
    if (index !== -1) {
      set({ isOpen: true, currentIndex: index })
    }
  },

  closePreview: () => set({ isOpen: false }),

  nextImage: () => {
    const { currentIndex, images } = get()
    if (currentIndex < images.length - 1) {
      set({ currentIndex: currentIndex + 1 })
    }
  },

  prevImage: () => {
    const { currentIndex } = get()
    if (currentIndex > 0) {
      set({ currentIndex: currentIndex - 1 })
    }
  }
}))
