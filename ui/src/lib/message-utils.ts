import { MessageType } from "@/types/message";
import type { Message } from "@/types/message";
import { formatCSTTime, toCST } from "./timezone";

/**
 * 解析时间范围的起始时间
 */
export function parseTimeRangeStart(timeRange: string): number {
  const parts = timeRange.split('~')
  if (parts.length !== 2) {
    return Date.now()
  }
  return new Date(parts[0].trim()).getTime()
}

/**
 * 解析时间范围的结束时间
 */
export function parseTimeRangeEnd(timeRange: string): number {
  const parts = timeRange.split('~')
  if (parts.length !== 2) {
    return Date.now()
  }
  return new Date(parts[1].trim()).getTime()
}

/**
 * 创建 EmptyRange 消息
 */
export function createEmptyRangeMessage(
  talker: string,
  timeRange: string,
  newestMsgTime: string | undefined,
  triedTimes: number,
  suggestedBeforeTime: number
): Message {
  const startTime = parseTimeRangeStart(timeRange)
  const endTime = parseTimeRangeEnd(timeRange)
  const startDate = new Date(startTime)
  let endDate = new Date(endTime)
  if( newestMsgTime){
    endDate = new Date(newestMsgTime)
  }

  const startStr = formatCSTTime(startDate)
  const endStr = formatCSTTime(endDate)

  return {
    id: -Date.now() - 1000,
    seq: -1,
    time: toCST(startDate),
    createTime: startDate.getTime(),
    talker,
    talkerName: '',
    sender: '',
    senderName: '',
    isSelf: false,
    isSend: 0,
    isChatRoom: false,
    type: MessageType.EmptyRange,
    subType: 0,
    content: `${startStr} ~ ${endStr} 没有消息`,
    isEmptyRange: true,
    emptyRangeData: {
      timeRange,
      triedTimes,
      suggestedBeforeTime,
    },
  }
}

/**
 * 创建 Gap 消息
 * Gap 消息标记指定时间范围内有数据需要加载
 * 
 * @param talker 会话 ID
 * @param gapStartTime Gap 起始时间（时间戳或 ISO 字符串）
 * @param gapEndTime Gap 结束时间（时间戳或 ISO 字符串）
 * @param estimatedCount 预估消息数量
 * @returns Gap 消息对象
 */
export function createGapMessage(
  talker: string,
  gapStartTime: string | number,
  gapEndTime: string | number,
  estimatedCount?: number
): Message {
  const startDate = typeof gapStartTime === 'string' 
    ? new Date(gapStartTime) 
    : new Date(gapStartTime)
  const endDate = typeof gapEndTime === 'string' 
    ? new Date(gapEndTime) 
    : new Date(gapEndTime)

  const startStr = formatCSTTime(startDate)
  const endStr = formatCSTTime(endDate)
  const timeRange = `${toCST(startDate)}~${toCST(endDate)}`

  const content = estimatedCount && estimatedCount > 0
    ? `${startStr} ~ ${endStr} 还有约 ${estimatedCount} 条消息`
    : `${startStr} ~ ${endStr} 还有更多消息`

  return {
    id: -Date.now(),
    seq: -1,
    time: toCST(startDate),
    createTime: startDate.getTime(),
    talker,
    talkerName: '',
    sender: '',
    senderName: '',
    isSelf: false,
    isSend: 0,
    isChatRoom: false,
    type: MessageType.Gap,
    subType: 0,
    content,
    isGap: true,
    gapData: {
      timeRange,
      beforeTime: endDate.getTime(),
    },
  }
}
