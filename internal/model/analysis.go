package model

// HourlyStat 每小时活跃度统计
type HourlyStat struct {
	Hour  int `json:"hour"`  // 0-23
	Count int `json:"count"` // 消息数量
}

// DailyStat 每日活跃度统计
type DailyStat struct {
	Date  string `json:"date"`  // YYYY-MM-DD
	Count int    `json:"count"` // 消息数量
}

// WeekdayStat 星期活跃度统计
type WeekdayStat struct {
	Weekday int `json:"weekday"` // 1-7 (周一到周日)
	Count   int `json:"count"`   // 消息数量
}

// MonthlyStat 月份活跃度统计
type MonthlyStat struct {
	Month int `json:"month"` // 1-12
	Count int `json:"count"` // 消息数量
}

// MessageTypeStat 消息类型统计
type MessageTypeStat struct {
	Type  int `json:"type"`  // 消息类型
	Count int `json:"count"` // 消息数量
}

// RepeatStat 复读统计项
type RepeatStat struct {
	Content    string `json:"content"`    // 复读的内容
	Count      int    `json:"count"`      // 该内容复读总次数
	MemberName string `json:"memberName"` // 最近一次发起的成员
}

// PersonalTopContact 个人分析：亲密度排行榜项
type PersonalTopContact struct {
	Talker       string `json:"talker"`       // 联系人ID
	Name         string `json:"name"`         // 联系人姓名
	Avatar       string `json:"avatar"`       // 头像
	MessageCount int    `json:"messageCount"` // 总消息数
	SentCount    int    `json:"sentCount"`    // 我发送的消息数
	RecvCount    int    `json:"recvCount"`    // 对方发送的消息数
	LastTime     int64  `json:"lastTime"`     // 最后互动时间
}

// RelationshipDetail 深度关系分析
type RelationshipDetail struct {
	Talker          string       `json:"talker"`
	InitiatorCount  int          `json:"initiatorCount"`  // 我发起的对话数
	TerminatorCount int          `json:"terminatorCount"` // 我终结的对话数
	LateNightCount  int          `json:"lateNightCount"`  // 深夜(0-5点)互动数
	WordCountRatio  float64      `json:"wordCountRatio"`  // 字数比例 (我/总)
	TopKeywords     []string     `json:"topKeywords"`     // 关键词
	HourlyActivity  []HourlyStat `json:"hourlyActivity"`  // 24小时分布
	FirstMessage    string       `json:"firstMessage"`    // 第一句话
	LastMessage     string       `json:"lastMessage"`     // 最近一句话
}

// DashboardData 总览数据
type DashboardData struct {
	Overview DashboardOverview `json:"overview"`
}

type DashboardOverview struct {
	User     string            `json:"user"`
	DBStats  DBStats           `json:"dbStats"`
	MsgStats MsgStats          `json:"msgStats"`
	MsgTypes map[string]int    `json:"msgTypes"`
	Groups   []DashboardGroup  `json:"groups"`
	Timeline DashboardTimeline `json:"timeline"`
}

type DBStats struct {
	DBSizeMB  float64 `json:"db_size_mb"`
	DirSizeMB float64 `json:"dir_size_mb"`
}

type MsgStats struct {
	TotalMsgs      int `json:"total_msgs"`
	SentMsgs       int `json:"sent_msgs"`
	ReceivedMsgs   int `json:"received_msgs"`
	UniqueMsgTypes int `json:"unique_msg_types"`
}

type DashboardGroup struct {
	ChatRoomName string `json:"ChatRoomName"`
	NickName     string `json:"NickName"`
	MemberCount  int    `json:"member_count"`
	MessageCount int    `json:"message_count"`
}

type DashboardTimeline struct {
	EarliestMsgTime int64 `json:"earliest_msg_time"`
	LatestMsgTime   int64 `json:"latest_msg_time"`
	DurationDays    int   `json:"duration_days"`
}

// AnnualReport 年度报告
type AnnualReport struct {
	Year         int                   `json:"year"`
	Overview     AnnualOverview        `json:"overview"`
	TopContacts  []*PersonalTopContact `json:"top_contacts"`
	MonthlyTrend []*MonthlyStat        `json:"monthly_trend"`
	WeekdayDist  []*WeekdayStat        `json:"weekday_distribution"`
	HourlyDist   []*HourlyStat         `json:"hourly_distribution"`
	MessageTypes map[string]int        `json:"message_types"`
	Highlights   AnnualHighlights      `json:"highlights"`
}

// AnnualOverview 年度概览
type AnnualOverview struct {
	TotalMessages    int    `json:"total_messages"`
	SentMessages     int    `json:"sent_messages"`
	ReceivedMessages int    `json:"received_messages"`
	TotalContacts    int    `json:"total_contacts"`
	ActiveContacts   int    `json:"active_contacts"`
	TotalChatrooms   int    `json:"total_chatrooms"`
	ActiveChatrooms  int    `json:"active_chatrooms"`
	FirstMessageDate string `json:"first_message_date"`
	LastMessageDate  string `json:"last_message_date"`
	ActiveDays       int    `json:"active_days"`
}

// AnnualHighlights 年度亮点
type AnnualHighlights struct {
	BusiestDay          DayCount `json:"busiest_day"`
	QuietestDay         DayCount `json:"quietest_day"`
	LongestStreak       int      `json:"longest_streak"`
	LateNightCount      int      `json:"late_night_count"`
	EarliestMessageTime string   `json:"earliest_message_time"`
	LatestMessageTime   string   `json:"latest_message_time"`
}

// DayCount 日期消息计数
type DayCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// NeedContactItem 需要联系的客户项
type NeedContactItem struct {
	UserName         string `json:"userName"`
	NickName         string `json:"nickName"`
	Remark           string `json:"remark"`
	SmallHeadURL     string `json:"smallHeadURL"`
	BigHeadURL       string `json:"bigHeadURL"`
	LastContactTime  int64  `json:"lastContactTime"`
	DaysSinceContact int    `json:"daysSinceContact"`
}
