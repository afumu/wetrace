/**
 * API 响应基础接口
 */
export interface ApiResponse<T = any> {
  success: boolean
  data: T
  message?: string
}

/**
 * 分页响应接口
 */
export interface PaginatedResponse<T> {
  total: number
  items: T[]
  page?: number
  pageSize?: number
  hasMore?: boolean
}

/**
 * 聊天记录查询参数
 * @property time 必填，格式：YYYY-MM-DD 或 YYYY-MM-DD~YYYY-MM-DD
 */
export interface ChatlogParams {
  talker_id?: string
  time: string  // 必填：YYYY-MM-DD 或 YYYY-MM-DD~YYYY-MM-DD
  sender?: string
  limit?: number
  offset?: number
  bottom?: number
  format?: 'json' | 'csv' | 'text'
}

/**
 * 搜索参数
 */
export interface SearchParams {
  keyword: string
  talker_id?: string
  time?: string
  type?: number
  sender?: string
  limit?: number
  offset?: number
}

/**
 * 会话列表查询参数
 */
export interface SessionParams {
  limit?: number
  offset?: number
  type?: string
}

/**
 * 联系人查询参数
 */
export interface ContactParams {
  type?: string
  keyword?: string
  limit?: number
  offset?: number
}

/** QA
*
 * API 错误接口
 */
export interface ApiError {
  code: number
  message: string
  timestamp?: string
  path?: string
  details?: any
}

/**
 * API 请求配置
 */
export interface ApiConfig {
  baseURL: string
  timeout: number
  headers?: Record<string, string>
}

/**
 * HTTP 方法枚举
 */
export enum HttpMethod {
  GET = 'GET',
  POST = 'POST',
  PUT = 'PUT',
  DELETE = 'DELETE',
  PATCH = 'PATCH',
}

/**
 * 请求状态枚举
 */
export enum RequestStatus {
  Idle = 'idle',
  Loading = 'loading',
  Success = 'success',
  Error = 'error',
}

/**
 * 时间范围接口
 */
export interface TimeRange {
  start?: string
  end?: string
}

/**
 * 搜索结果接口
 */
export interface SearchResult<T = any> {
  item: T
  highlight?: string
  score?: number
  context?: {
    before?: string
    after?: string
  }
}
