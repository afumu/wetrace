import { request, getApiBaseUrl } from "@/lib/request"

export const mediaApi = {
  getImageUrl: (id: string, path?: string): string => {
    const baseURL = getApiBaseUrl()
    let url = `${baseURL}/api/v1/media/image/${encodeURIComponent(id)}`
    if (path) {
      url += `?path=${encodeURIComponent(path)}`
    }
    return url
  },

  getThumbnailUrl: (id: string, path?: string): string => {
    const baseURL = getApiBaseUrl()
    let url = `${baseURL}/api/v1/media/image/${encodeURIComponent(id)}?thumb=1`
    if (path) {
      url += `&path=${encodeURIComponent(path)}`
    }
    return url
  },

  getVideoUrl: (id: string): string => {
    const baseURL = getApiBaseUrl()
    return `${baseURL}/api/v1/media/video/${encodeURIComponent(id)}`
  },

  getVoiceUrl: (id: string): string => {
    const baseURL = getApiBaseUrl()
    return `${baseURL}/api/v1/media/voice/${encodeURIComponent(id)}`
  },

  getFileUrl: (id: string): string => {
    const baseURL = getApiBaseUrl()
    return `${baseURL}/api/v1/media/file/${encodeURIComponent(id)}`
  },

  getAvatarUrl: (avatarPath: string): string => {
    const baseURL = getApiBaseUrl()
    if (!avatarPath) return ''
    if (avatarPath.startsWith('http://') || avatarPath.startsWith('https://')) return avatarPath
    // Avatar logic might depend on how it's served. 
    // If it's a relative path to a static asset or another API:
    return `${baseURL}${avatarPath.startsWith('/') ? '' : '/'}${avatarPath}`
  },

  isMediaMessage: (type: number): boolean => {
    return [3, 34, 43, 47, 49].includes(type)
  },

  startCache: (scope: 'all' | 'session', talker?: string) => {
    return request.post('/api/v1/media/cache/start', { scope, talker })
  },

  getCacheStatus: () => {
    return request.get('/api/v1/media/cache/status')
  },

  getImageList: (params?: {
    talker?: string;
    time_range?: string;
    limit?: number;
    offset?: number;
  }) => {
    return request.get<{
      total: number;
      items: ImageListItem[];
    }>('/api/v1/media/images', params)
  },

  transcribeVoice: (id: string) => {
    return request.post<{ text: string }>('/api/v1/media/voice/transcribe', { id })
  },

  getExportVoicesUrl: (talker: string, name?: string): string => {
    const baseURL = getApiBaseUrl()
    let url = `${baseURL}/api/v1/export/voices?talker=${encodeURIComponent(talker)}`
    if (name) {
      url += `&name=${encodeURIComponent(name)}`
    }
    return url
  },
}

export interface ImageListItem {
  key: string;
  talker: string;
  talkerName: string;
  time: string;
  thumbnailUrl: string;
  fullUrl: string;
  seq: number;
}
