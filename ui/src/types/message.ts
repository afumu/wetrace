/**
 * 消息类型枚举
 */
export enum MessageType {
  Text = 1,
  Image = 3,
  Voice = 34,
  ContactCard = 42,
  Video = 43,
  Emoji = 47,
  Location = 48,
  File = 49,
  VoiceCall = 50,
  System = 10000,
  Revoke = 10002,
  Gap = 99999,
  EmptyRange = 99998,
  QQMail = 35,
}

/**
 * 富文本消息子类型枚举
 */
export enum RichMessageSubType {
  QQMusic = 3,
  VideoLink = 4,
  Link = 5,
  File = 6,
  CardPackage = 16,
  Forwarded = 19,
  MiniProgram = 33,
  ShoppingMiniProgram = 36,
  ShortVideo = 51,
  Jielong = 53,
  Refer = 57,
  Pat = 62,
  Live = 63,
  FileDownloading = 74,
  Transfer = 2000,
  RedPacket = 2001,
}

/**
 * 后端返回的消息数据结构
 */
export interface MessageResponse {
  seq: number
  time: string
  talker: string
  talkerName: string
  isChatRoom: boolean
  sender: string
  senderName: string
  isSelf: boolean
  type: number
  subType: number
  content: string
  contents?: {
    md5?: string
    title?: string
    desc?: string
    url?: string
    recordInfo?: RecordInfo
    [key: string]: any
  }
  smallHeadURL?: string
  bigHeadURL?: string
}

export interface RecordItem {
  DataType: string
  DataID: string
  DataDesc: string
  SourceName: string
  SourceTime: string
  SourceHeadURL: string
  FullMD5?: string
  ThumbFullMD5?: string
  CDNThumbURL?: string
  CDNDataURL?: string
  [key: string]: any
}

export interface RecordInfo {
  Title: string
  Desc: string
  DataList: {
    Count: string
    DataItems: RecordItem[]
  }
  [key: string]: any
}

/**
 * 前端使用的消息接口
 */
export interface Message {
  id: number
  seq: number
  time: string
  createTime: number
  talker: string
  talkerName: string
  talkerAvatar?: string
  sender: string
  senderName: string
  isSelf: boolean
  isSend: number
  isChatRoom: boolean
  type: MessageType
  subType: number
  content: string
  contents?: {
    md5?: string
    title?: string
    desc?: string
    url?: string
    recordInfo?: RecordInfo
    [key: string]: any
  }
  imageUrl?: string
  videoUrl?: string
  voiceUrl?: string
  fileUrl?: string
  fileName?: string
  fileSize?: number
  duration?: number
  smallHeadURL?: string
  bigHeadURL?: string
  // Gap 消息标识
  isGap?: boolean
  gapData?: {
    timeRange: string
    beforeTime: number
  }
  // EmptyRange 消息标识
  isEmptyRange?: boolean
  emptyRangeData?: {
    timeRange: string
    triedTimes: number
    suggestedBeforeTime: number
  }
}

/**
 * 消息分组（按日期）
 */
export interface MessageGroup {
  date: string
  messages: Message[]
}

/**
 * 消息内容类型
 */
export interface MessageContent {
  text?: string
  url?: string
  fileName?: string
  fileSize?: number
  duration?: number
  width?: number
  height?: number
}

/**
 * 消息类型显示名称映射
 */
export const MessageTypeNames: Record<MessageType, string> = {
  [MessageType.Text]: '文本',
  [MessageType.Image]: '图片',
  [MessageType.Voice]: '语音',
  [MessageType.ContactCard]: '个人名片',
  [MessageType.Video]: '视频',
  [MessageType.Emoji]: '表情',
  [MessageType.Location]: '位置',
  [MessageType.File]: '文件',
  [MessageType.VoiceCall]: '语音通话',
  [MessageType.System]: '系统消息',
  [MessageType.Revoke]: '撤回消息',
  [MessageType.Gap]: '虚拟间隙消息',
  [MessageType.EmptyRange]: '虚拟空范围消息',
  [MessageType.QQMail]: 'QQ邮箱消息',
}

/**
 * 消息图标映射（字符串键版本）
 */
export const MessageIconMap: Record<string, string> = {
  '1': 'ChatLineSquare',
  '2': 'Picture',
  '3': 'Headset',
  '4': 'VideoPlay',
  '8': 'Document',
  '16': 'Tickets',
  '33': 'Grid',
  '34': 'Microphone',
  '35': 'Message',
  '36': 'ShoppingCart',
  '42': 'User',
  '43': 'VideoPlay',
  '48': 'Location',
  '50': 'Phone',
  '51': 'VideoCameraFilled',
  '53': 'List',
  '62': 'Pointer',
  '63': 'VideoCamera',
  '2000': 'Wallet',
  '2001': 'Present'
}

/**
 * 文件大小单位
 */
export const FileSizeUnits = ['B', 'KB', 'MB', 'GB'] as const
export const FileSizeBase = 1024