// API 相关类型
export * from './api';

// 消息相关类型
export * from './message';

// 会话相关类型
export * from './session';

// 联系人相关类型
export * from './contact';

// 应用相关类型
export interface AppConfig {
  title: string
  version: string
  apiBaseUrl: string
  apiTimeout: number
  pageSize: number
  maxPageSize: number
  enableDebug: boolean
  enableMock: boolean
}

export interface UserSettings {
  theme: 'light' | 'dark' | 'auto'
  language: 'zh-CN' | 'en-US'
  fontSize: 'small' | 'medium' | 'large'
  messageDensity: 'compact' | 'comfortable' | 'spacious'
  enterToSend: boolean
  autoPlayVoice: boolean
  showMessagePreview: boolean
  showTimestamp: boolean
  showAvatar: boolean
  timeFormat: '12h' | '24h'
  showMediaResources: boolean
  disableServerPinning: boolean
}

export interface LoadingState {
  sessions: boolean
  messages: boolean
  contacts: boolean
  search: boolean
}

export interface ErrorState {
  sessions: Error | null
  messages: Error | null
  contacts: Error | null
  search: Error | null
}

export type SessionFilterType = 'chat' | 'private' | 'group' | 'official' | 'all'
export type ContactFilterType = 'friend' | 'chatroom' | 'official' | 'all'
