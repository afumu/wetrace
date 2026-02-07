import { format, isToday, isYesterday, isThisWeek, isThisYear } from 'date-fns'
import { zhCN } from 'date-fns/locale'

export function formatSessionTime(timestamp: number): string {
  if (!timestamp) return ''
  
  const date = new Date(timestamp < 10000000000 ? timestamp * 1000 : timestamp)
  
  if (isToday(date)) {
    return format(date, 'HH:mm')
  }
  
  if (isYesterday(date)) {
    return '昨天'
  }
  
  if (isThisWeek(date)) {
    return format(date, 'eee', { locale: zhCN })
  }
  
  if (isThisYear(date)) {
    return format(date, 'MM/dd')
  }
  
  return format(date, 'yyyy/MM/dd')
}

export function formatMessageTime(timestamp: number | string): string {
  if (!timestamp) return ''
  const date = typeof timestamp === 'string' ? new Date(timestamp) : new Date(timestamp < 10000000000 ? timestamp * 1000 : timestamp)
  
  if (isToday(date)) {
    return format(date, 'HH:mm')
  }
  
  if (isYesterday(date)) {
    return `昨天 ${format(date, 'HH:mm')}`
  }
  
  if (isThisWeek(date)) {
    return `${format(date, 'eee', { locale: zhCN })} ${format(date, 'HH:mm')}`
  }
  
  if (isThisYear(date)) {
    return format(date, 'M月d日 HH:mm')
  }
  
  return format(date, 'yyyy年M月d日 HH:mm')
}
