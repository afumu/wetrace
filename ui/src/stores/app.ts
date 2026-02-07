import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { UserSettings, AppConfig } from '@/types'

interface AppState {
  config: AppConfig
  settings: UserSettings
  isMobile: boolean
  sidebarCollapsed: boolean
  activeNav: string
  
  // Actions
  setMobile: (isMobile: boolean) => void
  toggleSidebar: () => void
  setActiveNav: (nav: string) => void
  updateSettings: (settings: Partial<UserSettings>) => void
  toggleTheme: () => void
}

const defaultConfig: AppConfig = {
  title: 'Chatlog Session',
  version: '1.0.0',
  apiBaseUrl: 'http://127.0.0.1:5030',
  apiTimeout: 30000,
  pageSize: 500,
  maxPageSize: 5000,
  enableDebug: false,
  enableMock: false,
}

const defaultSettings: UserSettings = {
  theme: 'light',
  language: 'zh-CN',
  fontSize: 'medium',
  messageDensity: 'comfortable',
  enterToSend: true,
  autoPlayVoice: false,
  showMessagePreview: true,
  showTimestamp: true,
  showAvatar: true,
  timeFormat: '24h',
  showMediaResources: true,
  disableServerPinning: false,
}

export const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      config: defaultConfig,
      settings: defaultSettings,
      isMobile: false,
      sidebarCollapsed: false,
      activeNav: 'chat',

      setMobile: (isMobile) => set({ isMobile }),
      
      toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
      
      setActiveNav: (nav) => set({ activeNav: nav }),
      
      updateSettings: (newSettings) => {
        set((state) => {
          const settings = { ...state.settings, ...newSettings }
          // Apply theme side-effect
          if (newSettings.theme) {
            const html = document.documentElement
            if (newSettings.theme === 'dark' || (newSettings.theme === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
              html.classList.add('dark')
            } else {
              html.classList.remove('dark')
            }
          }
          return { settings }
        })
      },

      toggleTheme: () => {
        const { settings, updateSettings } = get()
        const themes: Array<'light' | 'dark' | 'auto'> = ['light', 'dark', 'auto']
        const currentIndex = themes.indexOf(settings.theme)
        const nextIndex = (currentIndex + 1) % themes.length
        updateSettings({ theme: themes[nextIndex] })
      },
    }),
    {
      name: 'app-storage',
      partialize: (state) => ({ settings: state.settings }), // Only persist settings
    }
  )
)
