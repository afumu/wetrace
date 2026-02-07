import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse, type AxiosError, type InternalAxiosRequestConfig } from 'axios'
import type { ApiResponse, ApiError } from '@/types/api'

// Extend axios config
declare module 'axios' {
  export interface InternalAxiosRequestConfig {
    metadata?: {
      startTime?: number
      retryCount?: number
    }
  }
}

// Get settings from localStorage
const getSettings = () => {
  try {
    const settings = localStorage.getItem('chatlog-settings')
    return settings ? JSON.parse(settings) : {}
  } catch {
    return {}
  }
}

// Get API Base URL
export const getApiBaseUrl = (): string => {
  const directUrl = localStorage.getItem('apiBaseUrl')
  if (directUrl) return directUrl
  
  const settings = getSettings()
  if (settings.apiBaseUrl) return settings.apiBaseUrl
  
  return import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:5200'
}

const getDynamicConfig = (): AxiosRequestConfig => {
  const settings = getSettings()
  return {
    baseURL: getApiBaseUrl(),
    timeout: settings.apiTimeout || Number(import.meta.env.VITE_API_TIMEOUT) || 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  }
}

const service: AxiosInstance = axios.create(getDynamicConfig())

service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    if (!config.metadata) config.metadata = {}
    config.metadata.startTime = Date.now()
    if (config.metadata.retryCount === undefined) config.metadata.retryCount = 0
    
    // Dynamic config update
    const apiBaseUrl = getApiBaseUrl()
    if (apiBaseUrl) config.baseURL = apiBaseUrl
    
    const settings = getSettings()
    if (settings.apiTimeout) config.timeout = settings.apiTimeout
    
    // Add default params for GET
    if (config.method?.toLowerCase() === 'get') {
      const userParams = config.params || {}
      config.params = {
        limit: 200,
        offset: 0,
        ...userParams,
        _t: Date.now(),
      }
      
      // Only add format=json if not a blob/arraybuffer request
      if (config.responseType !== 'blob' && config.responseType !== 'arraybuffer') {
        config.params.format = 'json'
      }
    } else {
      config.params = { ...config.params, format: 'json' }
    }

    if (import.meta.env.DEV) {
      console.log('📤 API Request:', config.method?.toUpperCase(), (config.baseURL || '') + (config.url || ''))
    }

    return config
  },
  (error: AxiosError) => {
    console.error('❌ Request Error:', error)
    return Promise.reject(error)
  }
)

service.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data } = response
    
    // Handle backend response format: { success: true, data: ... }
    if (data && typeof data === 'object' && 'success' in data) {
      if (data.success) {
        return data.data
      }
      const errorMessage = (data as any).error?.message || (data as any).message || 'Request failed'
      console.error(errorMessage)
      return Promise.reject(new Error(errorMessage))
    }

    // Handle legacy/other formats
    if (Array.isArray(data)) return data as any
    if ('items' in data) return data as any
    
    return data as any
  },
  async (error: AxiosError<ApiError>) => {
    let errorMessage = error.message
    if (error.response?.data) {
      const data = error.response.data as any
      errorMessage = data.error?.message || data.message || errorMessage
    }
    
    // Assign the server error message to the error object so downstream catch blocks can display it
    error.message = errorMessage

    const settings = getSettings()
    const config = error.config as InternalAxiosRequestConfig
    
    const retryCount = settings.apiRetryCount ?? 3
    const retryDelay = settings.apiRetryDelay ?? 1000
    
    const shouldRetry = config && 
                       config.metadata &&
                       config.metadata.retryCount !== undefined &&
                       config.metadata.retryCount < retryCount &&
                       (!error.response || error.response.status >= 500 || error.code === 'ECONNABORTED')
    
    if (shouldRetry && config.metadata) {
      config.metadata.retryCount = (config.metadata.retryCount || 0) + 1
      await new Promise(resolve => setTimeout(resolve, retryDelay))
      return service(config)
    }
    
    console.error('❌ API Error:', error.message)
    return Promise.reject(error)
  }
)

class Request {
  get<T = any>(url: string, params?: any, config?: AxiosRequestConfig): Promise<T> {
    return service.get(url, { params, ...config })
  }
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return service.post(url, data, config)
  }
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return service.put(url, data, config)
  }
  delete<T = any>(url: string, params?: any, config?: AxiosRequestConfig): Promise<T> {
    return service.delete(url, { params, ...config })
  }
  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return service.patch(url, data, config)
  }
}

export const request = new Request()
export default service