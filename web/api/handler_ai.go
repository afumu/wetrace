package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/afumu/wetrace/internal/ai"
	"github.com/afumu/wetrace/pkg/util"
	"github.com/afumu/wetrace/store/types"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
)

// AISummarizeRequest AI 总结请求
type AISummarizeRequest struct {
	Talker    string `json:"talker" binding:"required"`
	TimeRange string `json:"time_range"`
}

// AISimulateRequest AI 模拟对话请求
type AISimulateRequest struct {
	Talker  string `json:"talker" binding:"required"`
	Message string `json:"message" binding:"required"`
}

// AISentimentRequest AI 情感分析请求
type AISentimentRequest struct {
	Talker    string `json:"talker" binding:"required"`
	TimeRange string `json:"time_range"`
}

// AISentimentResponse AI 情感分析响应
type AISentimentResponse struct {
	OverallScore          float64                    `json:"overall_score"`
	OverallLabel          string                     `json:"overall_label"`
	RelationshipHealth    string                     `json:"relationship_health"`
	Summary               string                     `json:"summary"`
	EmotionTimeline       []EmotionTimelineItem      `json:"emotion_timeline"`
	SentimentDistribution SentimentDistribution       `json:"sentiment_distribution"`
	RelationshipIndicators RelationshipIndicators     `json:"relationship_indicators"`
}

// EmotionTimelineItem 情绪时间线项
type EmotionTimelineItem struct {
	Period   string   `json:"period"`
	Score    float64  `json:"score"`
	Label    string   `json:"label"`
	Keywords []string `json:"keywords"`
}

// SentimentDistribution 情感分布
type SentimentDistribution struct {
	Positive float64 `json:"positive"`
	Neutral  float64 `json:"neutral"`
	Negative float64 `json:"negative"`
}

// RelationshipIndicators 关系指标
type RelationshipIndicators struct {
	InitiativeRatio float64 `json:"initiative_ratio"`
	ResponseSpeed   string  `json:"response_speed"`
	IntimacyTrend   string  `json:"intimacy_trend"`
}

// AISummarize 总结聊天内容
func (a *API) AISummarize(c *gin.Context) {
	if a.AI == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI 功能未启用"})
		return
	}

	var req AISummarizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 解析时间范围
	var start, end time.Time
	var ok bool
	if req.TimeRange != "" {
		start, end, ok = util.TimeRangeOf(req.TimeRange)
	}

	if !ok {
		// 如果未指定或无效，则默认为过去 20 年
		end = time.Now()
		start = end.AddDate(-20, 0, 0)
	}

	// 获取消息进行总结
	msgs, err := a.Store.GetMessages(context.Background(), types.MessageQuery{
		Talker:    req.Talker,
		StartTime: start,
		EndTime:   end,
		Limit:     500, // 时间范围总结可能需要更多上下文，增加到 500 条
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(msgs) == 0 {
		transport.SendSuccess(c, "暂无聊天记录可总结")
		return
	}

	// 我们希望总结的是该范围内的前 500 条（或全部）
	if len(msgs) > 500 {
		msgs = msgs[:500]
	}

	var sb strings.Builder
	for _, m := range msgs {
		if m.Type == 1 { // 文本消息
			sb.WriteString(fmt.Sprintf("%s: %s\n", m.SenderName, m.Content))
		}
	}

	prompt := "以下是一段微信聊天记录，请简要总结对话的核心内容和主要结论：\n\n" + sb.String()

	summary, err := a.AI.Chat([]ai.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	transport.SendSuccess(c, summary)
}

// AISimulate 模拟对方回复
func (a *API) AISimulate(c *gin.Context) {
	if a.AI == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI 功能未启用"})
		return
	}

	var req AISimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取更多历史记录以进行深度学习
	end := time.Now()
	start := end.AddDate(-20, 0, 0)

	msgs, err := a.Store.GetMessages(context.Background(), types.MessageQuery{
		Talker:    req.Talker,
		StartTime: start,
		EndTime:   end,
		Limit:     300, // 增加采样量
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 提取对方的名字和聊天记录
	var history strings.Builder
	var targetName string

	// 如果消息太多，取最近的 150 条作为上下文
	if len(msgs) > 150 {
		msgs = msgs[len(msgs)-150:]
	}

	for _, m := range msgs {
		if m.Sender == req.Talker {
			targetName = m.SenderName
		}
		if m.Type == 1 {
			role := "用户"
			if m.Sender == req.Talker {
				role = targetName
			}
			history.WriteString(fmt.Sprintf("[%s]: %s\n", role, m.Content))
		}
	}

	if targetName == "" {
		targetName = "对方"
	}

	// 精细化 Prompt
	systemPrompt := fmt.Sprintf(`你现在是一个高级人工智能，你的任务是精准模拟一个名为 "%s" 的人的微信聊天风格。

你需要通过分析以下提供的聊天记录，学习并模仿 %s 的以下特征：
1. 语气与口吻：是热情、冷淡、幽默还是严肃？
2. 常用词汇：是否有特定的口头禅、简称或习惯性用语？
3. 表情习惯：是否经常使用表情符号（如 [微笑]、[呲牙]）或 Emoji？使用的频率如何？
4. 回复长度：习惯发长句子还是短句？
5. 标点符号：是否经常使用标点，还是习惯直接空格？

历史聊天记录（参考上下文）：
%s

模仿要点：
- 你现在就是 %s。
- 严禁以 AI 助手的身份说话。
- 回复内容必须简洁自然，符合微信聊天的即时性。
- 直接输出回复内容，不要附带任何解释或前缀。`, targetName, targetName, history.String(), targetName)

	reply, err := a.AI.Chat([]ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: req.Message},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	transport.SendSuccess(c, reply)
}

// AISentiment 分析对话情感倾向与关系变化趋势
func (a *API) AISentiment(c *gin.Context) {
	if a.AI == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI 功能未启用"})
		return
	}

	var req AISentimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 解析时间范围
	var start, end time.Time
	var ok bool
	if req.TimeRange != "" {
		start, end, ok = util.TimeRangeOf(req.TimeRange)
	}
	if !ok {
		end = time.Now()
		start = end.AddDate(-20, 0, 0)
	}

	// 按月分段采样消息
	monthlyTexts := a.sampleMessagesByMonth(start, end, req.Talker)
	if len(monthlyTexts) == 0 {
		transport.SendSuccess(c, "暂无聊天记录可分析")
		return
	}

	// 构建 prompt
	prompt := buildSentimentPrompt(monthlyTexts)

	result, err := a.AI.Chat([]ai.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 解析 AI 返回的 JSON
	resp, err := parseSentimentResponse(result)
	if err != nil {
		// 如果解析失败，返回原始文本
		transport.SendSuccess(c, result)
		return
	}

	transport.SendSuccess(c, resp)
}

// sampleMessagesByMonth 按月分段采样文本消息，每月最多 100 条
func (a *API) sampleMessagesByMonth(start, end time.Time, talker string) map[string]string {
	monthlyTexts := make(map[string]string)

	current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	for current.Before(end) {
		monthStart := current
		monthEnd := current.AddDate(0, 1, 0).Add(-time.Second)
		if monthEnd.After(end) {
			monthEnd = end
		}

		msgs, err := a.Store.GetMessages(context.Background(), types.MessageQuery{
			Talker:    talker,
			StartTime: monthStart,
			EndTime:   monthEnd,
			Limit:     100,
		})
		if err != nil {
			current = current.AddDate(0, 1, 0)
			continue
		}

		var sb strings.Builder
		for _, m := range msgs {
			if m.Type == 1 {
				sb.WriteString(fmt.Sprintf("%s: %s\n", m.SenderName, m.Content))
			}
		}

		text := sb.String()
		if text != "" {
			key := monthStart.Format("2006-01")
			monthlyTexts[key] = text
		}

		current = current.AddDate(0, 1, 0)
	}

	return monthlyTexts
}

// buildSentimentPrompt 构建情感分析的 prompt
func buildSentimentPrompt(monthlyTexts map[string]string) string {
	// 按月份排序
	months := make([]string, 0, len(monthlyTexts))
	for k := range monthlyTexts {
		months = append(months, k)
	}
	sort.Strings(months)

	var sb strings.Builder
	sb.WriteString("以下是按月份整理的微信聊天记录，请进行情感分析。\n\n")

	for _, month := range months {
		sb.WriteString(fmt.Sprintf("=== %s ===\n%s\n", month, monthlyTexts[month]))
	}

	sb.WriteString(`
请严格按照以下 JSON 格式返回分析结果，不要包含任何其他文字：
{
  "overall_score": 0.72,
  "overall_label": "积极/消极/中立",
  "relationship_health": "良好/一般/需关注",
  "summary": "整体分析总结...",
  "emotion_timeline": [
    {
      "period": "2025-01",
      "score": 0.8,
      "label": "积极/消极/中立",
      "keywords": ["关键词1", "关键词2"]
    }
  ],
  "sentiment_distribution": {
    "positive": 0.58,
    "neutral": 0.30,
    "negative": 0.12
  },
  "relationship_indicators": {
    "initiative_ratio": 0.52,
    "response_speed": "快/中/慢",
    "intimacy_trend": "上升/稳定/下降"
  }
}

说明：
- overall_score: 0-1之间的情感评分，越高越积极
- emotion_timeline: 按月份的情绪变化，每个月一条记录
- initiative_ratio: 主动发起对话的比例（0-1）
- 所有数值保留两位小数`)

	return sb.String()
}

// parseSentimentResponse 解析 AI 返回的情感分析 JSON
func parseSentimentResponse(raw string) (*AISentimentResponse, error) {
	// 尝试提取 JSON 内容（AI 可能返回 markdown 代码块包裹的 JSON）
	jsonStr := raw
	if idx := strings.Index(raw, "{"); idx >= 0 {
		if endIdx := strings.LastIndex(raw, "}"); endIdx >= 0 {
			jsonStr = raw[idx : endIdx+1]
		}
	}

	var resp AISentimentResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- AI 高级功能：待办提取、关键信息抽取、长对话摘要、语音转文字 ---

// AITodosRequest 待办事项提取请求
type AITodosRequest struct {
	Talker    string `json:"talker" binding:"required"`
	TimeRange string `json:"time_range"`
}

// AITodoItem 单条待办事项
type AITodoItem struct {
	Content    string `json:"content"`
	Deadline   string `json:"deadline"`
	Priority   string `json:"priority"`
	SourceMsg  string `json:"source_msg"`
	SourceTime string `json:"source_time"`
}

// AITodosResponse 待办事项提取响应
type AITodosResponse struct {
	Todos []AITodoItem `json:"todos"`
}

// AIExtractRequest 关键信息抽取请求
type AIExtractRequest struct {
	Talker    string   `json:"talker" binding:"required"`
	TimeRange string   `json:"time_range"`
	Types     []string `json:"types"`
}

// AIExtractItem 单条抽取信息
type AIExtractItem struct {
	Type    string `json:"type"`
	Value   string `json:"value"`
	Context string `json:"context"`
	Time    string `json:"time"`
}

// AIExtractResponse 关键信息抽取响应
type AIExtractResponse struct {
	Extractions []AIExtractItem `json:"extractions"`
}

// AISummaryRequest 长对话/群聊摘要请求
type AISummaryRequest struct {
	Talker    string `json:"talker" binding:"required"`
	TimeRange string `json:"time_range"`
}

// AIVoice2TextRequest 语音转文字请求
type AIVoice2TextRequest struct {
	Talker string `json:"talker" binding:"required"`
	Seq    int64  `json:"seq" binding:"required"`
}

// AISummary 长对话/群聊自动摘要（增强版）
func (a *API) AISummary(c *gin.Context) {
	if a.AI == nil {
		transport.BadRequest(c, "AI 功能未启用")
		return
	}

	var req AISummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, err.Error())
		return
	}

	var start, end time.Time
	var ok bool
	if req.TimeRange != "" {
		start, end, ok = util.TimeRangeOf(req.TimeRange)
	}
	if !ok {
		end = time.Now()
		start = end.AddDate(-20, 0, 0)
	}

	msgs, err := a.Store.GetMessages(context.Background(), types.MessageQuery{
		Talker:    req.Talker,
		StartTime: start,
		EndTime:   end,
		Limit:     500,
	})
	if err != nil {
		transport.InternalServerError(c, err.Error())
		return
	}

	if len(msgs) == 0 {
		transport.SendSuccess(c, gin.H{"summary": "暂无聊天记录可摘要"})
		return
	}

	if len(msgs) > 500 {
		msgs = msgs[:500]
	}

	var sb strings.Builder
	for _, m := range msgs {
		if m.Type == 1 {
			sb.WriteString(fmt.Sprintf("[%s] %s: %s\n",
				m.Time.Format("01-02 15:04"), m.SenderName, m.Content))
		}
	}

	prompt := `以下是一段微信聊天记录，请生成结构化摘要。要求：
1. 核心话题：列出讨论的主要话题（不超过5个）
2. 关键结论：列出达成的共识或结论
3. 待跟进事项：列出需要后续跟进的事项
4. 一句话总结：用一句话概括整段对话

请严格按照以下 JSON 格式返回，不要包含其他文字：
{
  "topics": ["话题1", "话题2"],
  "conclusions": ["结论1", "结论2"],
  "follow_ups": ["跟进事项1", "跟进事项2"],
  "one_line_summary": "一句话总结"
}

聊天记录：
` + sb.String()

	result, err := a.AI.Chat([]ai.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		transport.InternalServerError(c, err.Error())
		return
	}

	parsed := parseJSONFromAI(result)
	if parsed != nil {
		transport.SendSuccess(c, parsed)
		return
	}
	transport.SendSuccess(c, gin.H{"summary": result})
}

// AIExtractTodos 从聊天记录中提取待办事项
func (a *API) AIExtractTodos(c *gin.Context) {
	if a.AI == nil {
		transport.BadRequest(c, "AI 功能未启用")
		return
	}

	var req AITodosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, err.Error())
		return
	}

	var start, end time.Time
	var ok bool
	if req.TimeRange != "" {
		start, end, ok = util.TimeRangeOf(req.TimeRange)
	}
	if !ok {
		// 默认最近一周
		end = time.Now()
		start = end.AddDate(0, 0, -7)
	}

	msgs, err := a.Store.GetMessages(context.Background(), types.MessageQuery{
		Talker:    req.Talker,
		StartTime: start,
		EndTime:   end,
		Limit:     300,
	})
	if err != nil {
		transport.InternalServerError(c, err.Error())
		return
	}

	if len(msgs) == 0 {
		transport.SendSuccess(c, AITodosResponse{Todos: []AITodoItem{}})
		return
	}

	var sb strings.Builder
	for _, m := range msgs {
		if m.Type == 1 {
			sb.WriteString(fmt.Sprintf("[%s] %s: %s\n",
				m.Time.Format("2006-01-02 15:04"), m.SenderName, m.Content))
		}
	}

	prompt := `以下是一段微信聊天记录，请从中提取所有待办事项、任务、提醒和承诺。
包括但不限于：
- 明确的任务分配（"帮我..."、"你去..."）
- 时间约定（"明天..."、"下周..."）
- 承诺和提醒（"记得..."、"别忘了..."）
- 需要回复或跟进的事项

请严格按照以下 JSON 格式返回，不要包含其他文字：
{
  "todos": [
    {
      "content": "待办事项描述",
      "deadline": "截止时间（如有，ISO 8601格式，无则为空字符串）",
      "priority": "high/medium/low",
      "source_msg": "原始消息内容",
      "source_time": "消息时间"
    }
  ]
}

如果没有找到任何待办事项，返回 {"todos": []}

聊天记录：
` + sb.String()

	result, err := a.AI.Chat([]ai.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		transport.InternalServerError(c, err.Error())
		return
	}

	var resp AITodosResponse
	if err := parseAIJSON(result, &resp); err != nil {
		transport.SendSuccess(c, gin.H{"todos": []interface{}{}, "raw": result})
		return
	}
	transport.SendSuccess(c, resp)
}

// AIExtractInfo 从聊天记录中抽取关键信息（地址、时间、金额、电话等）
func (a *API) AIExtractInfo(c *gin.Context) {
	if a.AI == nil {
		transport.BadRequest(c, "AI 功能未启用")
		return
	}

	var req AIExtractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, err.Error())
		return
	}

	var start, end time.Time
	var ok bool
	if req.TimeRange != "" {
		start, end, ok = util.TimeRangeOf(req.TimeRange)
	}
	if !ok {
		end = time.Now()
		start = end.AddDate(0, -1, 0)
	}

	msgs, err := a.Store.GetMessages(context.Background(), types.MessageQuery{
		Talker:    req.Talker,
		StartTime: start,
		EndTime:   end,
		Limit:     300,
	})
	if err != nil {
		transport.InternalServerError(c, err.Error())
		return
	}

	if len(msgs) == 0 {
		transport.SendSuccess(c, AIExtractResponse{Extractions: []AIExtractItem{}})
		return
	}

	var sb strings.Builder
	for _, m := range msgs {
		if m.Type == 1 {
			sb.WriteString(fmt.Sprintf("[%s] %s: %s\n",
				m.Time.Format("2006-01-02 15:04"), m.SenderName, m.Content))
		}
	}

	typesHint := "address（地址）、time（时间约定）、amount（金额）、phone（电话号码）"
	if len(req.Types) > 0 {
		typesHint = strings.Join(req.Types, "、")
	}

	prompt := fmt.Sprintf(`以下是一段微信聊天记录，请从中提取以下类型的关键信息：%s

提取规则：
- address: 具体地址、地点、位置信息
- time: 时间约定、日期安排（非消息本身的时间戳）
- amount: 金额、价格、费用
- phone: 电话号码、手机号

请严格按照以下 JSON 格式返回，不要包含其他文字：
{
  "extractions": [
    {
      "type": "address/time/amount/phone",
      "value": "提取的具体值",
      "context": "包含该信息的原始消息",
      "time": "消息时间"
    }
  ]
}

如果没有找到任何信息，返回 {"extractions": []}

聊天记录：
%s`, typesHint, sb.String())

	result, err := a.AI.Chat([]ai.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		transport.InternalServerError(c, err.Error())
		return
	}

	var extractResp AIExtractResponse
	if err := parseAIJSON(result, &extractResp); err != nil {
		transport.SendSuccess(c, gin.H{"extractions": []interface{}{}, "raw": result})
		return
	}
	transport.SendSuccess(c, extractResp)
}

// AIVoice2Text 语音消息转文字
func (a *API) AIVoice2Text(c *gin.Context) {
	if a.AI == nil {
		transport.BadRequest(c, "AI 功能未启用")
		return
	}

	var req AIVoice2TextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, err.Error())
		return
	}

	// 语音转文字需要 STT 服务支持，当前通过 AI 模拟实现
	// 实际生产环境应集成 Whisper API 或其他 STT 服务
	// 这里先获取语音消息的元信息，返回提示
	transport.SendSuccess(c, gin.H{
		"text":     "",
		"duration": 0,
		"language": "",
		"error":    "语音转文字功能需要配置 STT（语音识别）服务，当前暂未集成",
	})
}

// parseJSONFromAI 从 AI 返回的文本中提取 JSON 并解析为 map
func parseJSONFromAI(raw string) map[string]interface{} {
	jsonStr := extractJSON(raw)
	if jsonStr == "" {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil
	}
	return result
}

// parseAIJSON 从 AI 返回的文本中提取 JSON 并解析到指定结构体
func parseAIJSON(raw string, v interface{}) error {
	jsonStr := extractJSON(raw)
	if jsonStr == "" {
		return fmt.Errorf("no JSON found in AI response")
	}
	return json.Unmarshal([]byte(jsonStr), v)
}

// extractJSON 从可能包含 markdown 代码块的文本中提取 JSON 字符串
func extractJSON(raw string) string {
	// 尝试提取 JSON 内容（AI 可能返回 markdown 代码块包裹的 JSON）
	if idx := strings.Index(raw, "{"); idx >= 0 {
		if endIdx := strings.LastIndex(raw, "}"); endIdx >= 0 {
			return raw[idx : endIdx+1]
		}
	}
	return ""
}
