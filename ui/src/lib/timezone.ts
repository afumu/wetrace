/**
 * 时区转换工具函数
 * 处理时间格式转换，统一使用东八区（Asia/Shanghai, UTC+8）
 */

/**
 * 将 Date 对象转换为东八区的 ISO 8601 格式字符串
 * 格式：YYYY-MM-DDTHH:mm:ss.sss+08:00
 *
 * @param date Date 对象
 * @returns 东八区格式的 ISO 8601 字符串
 *
 * @example
 * const date = new Date('2024-01-15T10:30:00Z')
 * toCST(date) // '2024-01-15T18:30:00.000+08:00'
 */
export function toCST(date: Date): string {
  // 获取 UTC 时间戳
  const utcTime = date.getTime()

  // 转换为东八区（UTC+8，即 +8 小时）
  const cstTime = utcTime + (8 * 60 * 60 * 1000)
  const cstDate = new Date(cstTime)

  // 手动构建 ISO 8601 格式字符串
  const year = cstDate.getUTCFullYear()
  const month = String(cstDate.getUTCMonth() + 1).padStart(2, '0')
  const day = String(cstDate.getUTCDate()).padStart(2, '0')
  const hours = String(cstDate.getUTCHours()).padStart(2, '0')
  const minutes = String(cstDate.getUTCMinutes()).padStart(2, '0')
  const seconds = String(cstDate.getUTCSeconds()).padStart(2, '0')
  const milliseconds = String(cstDate.getUTCMilliseconds()).padStart(3, '0')

  return `${year}-${month}-${day}T${hours}:${minutes}:${seconds}.${milliseconds}+08:00`
}

export function formatCSTTime (date: Date) : string {
  const cstHours = (date.getUTCHours() + 8) % 24
  const minutes = date.getUTCMinutes()
  const seconds = date.getUTCSeconds()
  return `${String(cstHours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
}

/**
 * 将东八区 ISO 8601 字符串转换为 Date 对象
 *
 * @param cstString 东八区格式的 ISO 8601 字符串
 * @returns Date 对象
 *
 * @example
 * fromCST('2024-01-15T18:30:00.000+08:00') // Date object
 */
export function fromCST(cstString: string): Date {
  return new Date(cstString)
}

/**
 * 获取当前东八区时间的 ISO 8601 格式字符串
 *
 * @returns 当前东八区时间
 *
 * @example
 * nowCST() // '2024-01-15T18:30:00.000+08:00'
 */
export function nowCST(): string {
  return toCST(new Date())
}

/**
 * 格式化东八区时间范围
 * 格式：startTime~endTime
 *
 * @param startDate 开始时间
 * @param endDate 结束时间
 * @returns 时间范围字符串
 *
 * @example
 * const start = new Date('2024-01-15T00:00:00Z')
 * const end = new Date('2024-01-16T00:00:00Z')
 * formatCSTRange(start, end) // '2024-01-15T08:00:00.000+08:00~2024-01-16T08:00:00.000+08:00'
 */
export function formatCSTRange(startDate: Date, endDate: Date): string {
  return `${toCST(startDate)}~${toCST(endDate)}`
}

/**
 * 将 Unix 时间戳（秒）转换为东八区 ISO 8601 格式
 *
 * @param timestamp Unix 时间戳（秒）
 * @returns 东八区格式的 ISO 8601 字符串
 *
 * @example
 * timestampToCST(1705318245) // '2024-01-15T18:30:45.000+08:00'
 */
export function timestampToCST(timestamp: number): string {
  return toCST(new Date(timestamp * 1000))
}

/**
 * 格式化东八区日期为 YYYY-MM-DD 格式
 *
 * @param date Date 对象
 * @returns YYYY-MM-DD 格式的日期字符串
 *
 * @example
 * formatCSTDate(new Date('2024-01-15T10:30:00Z')) // '2024-01-15'
 */
export function formatCSTDate(date: Date): string {
  const cstString = toCST(date)
  return cstString.split('T')[0]
}

/**
 * 获取指定日期在东八区的开始时间（00:00:00.000）
 *
 * @param date Date 对象
 * @returns 该日期在东八区的开始时间
 *
 * @example
 * const date = new Date('2024-01-15T10:30:00Z')
 * getCSTDayStart(date) // '2024-01-15T00:00:00.000+08:00'
 */
export function getCSTDayStart(date: Date): string {
  const cstDateStr = formatCSTDate(date)
  return `${cstDateStr}T00:00:00.000+08:00`
}

/**
 * 获取指定日期在东八区的结束时间（23:59:59.999）
 *
 * @param date Date 对象
 * @returns 该日期在东八区的结束时间
 *
 * @example
 * const date = new Date('2024-01-15T10:30:00Z')
 * getCSTDayEnd(date) // '2024-01-15T23:59:59.999+08:00'
 */
export function getCSTDayEnd(date: Date): string {
  const cstDateStr = formatCSTDate(date)
  return `${cstDateStr}T23:59:59.999+08:00`
}

/**
 * 计算东八区时间下两个日期之间的天数差
 *
 * @param date1 第一个日期
 * @param date2 第二个日期
 * @returns 天数差（date2 - date1）
 *
 * @example
 * const d1 = new Date('2024-01-15T10:00:00Z')
 * const d2 = new Date('2024-01-16T10:00:00Z')
 * getCSTDaysDiff(d1, d2) // 1
 */
export function getCSTDaysDiff(date1: Date, date2: Date): number {
  const ms1 = date1.getTime()
  const ms2 = date2.getTime()
  const diff = ms2 - ms1
  return Math.floor(diff / (1000 * 60 * 60 * 24))
}

/**
 * 判断两个时间是否在东八区的同一天
 *
 * @param date1 第一个日期
 * @param date2 第二个日期
 * @returns 是否同一天
 *
 * @example
 * const d1 = new Date('2024-01-15T16:00:00Z') // CST 2024-01-16 00:00
 * const d2 = new Date('2024-01-15T20:00:00Z') // CST 2024-01-16 04:00
 * isSameCSTDay(d1, d2) // true
 */
export function isSameCSTDay(date1: Date, date2: Date): boolean {
  return formatCSTDate(date1) === formatCSTDate(date2)
}

/**
 * 获取 N 天前的东八区时间
 *
 * @param days 天数
 * @param baseDate 基准日期（默认为当前时间）
 * @returns N 天前的东八区时间
 *
 * @example
 * getCSTDaysAgo(7) // 7天前的当前时间（东八区）
 */
export function getCSTDaysAgo(days: number, baseDate: Date = new Date()): string {
  const targetDate = new Date(baseDate.getTime() - days * 24 * 60 * 60 * 1000)
  return toCST(targetDate)
}

/**
 * 调整日期到指定天数前（保留时分秒）
 *
 * @param date 基准日期
 * @param days 向前的天数
 * @returns 调整后的 Date 对象
 *
 * @example
 * const date = new Date('2024-01-15T18:30:00+08:00')
 * subtractDays(date, 3) // Date for 2024-01-12T18:30:00+08:00
 */
export function subtractDays(date: Date, days: number): Date {
  const result = new Date(date)
  result.setDate(result.getDate() - days)
  return result
}

/**
 * 解析东八区时间字符串为毫秒时间戳
 *
 * @param cstString 东八区格式的时间字符串
 * @returns 毫秒时间戳
 *
 * @example
 * parseCSTToTimestamp('2024-01-15T18:30:00.000+08:00') // 1705318200000
 */
export function parseCSTToTimestamp(cstString: string): number {
  return new Date(cstString).getTime()
}

/**
 * 验证是否为有效的东八区时间格式
 *
 * @param str 待验证的字符串
 * @returns 是否为有效格式
 *
 * @example
 * isValidCSTString('2024-01-15T18:30:00.000+08:00') // true
 * isValidCSTString('2024-01-15T18:30:00Z') // false
 */
export function isValidCSTString(str: string): boolean {
  const cstPattern = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}\+08:00$/
  return cstPattern.test(str)
}

/**
 * 将任意时间字符串或时间戳转换为东八区格式
 *
 * @param time 时间字符串或 Unix 时间戳（秒）
 * @returns 东八区格式的 ISO 8601 字符串
 *
 * @example
 * toSafeCST('2024-01-15T10:30:00Z') // '2024-01-15T18:30:00.000+08:00'
 * toSafeCST(1705318245) // '2024-01-15T18:30:45.000+08:00'
 */
export function toSafeCST(time: string | number): string {
  if (typeof time === 'string') {
    return toCST(new Date(time))
  } else {
    return timestampToCST(time)
  }
}