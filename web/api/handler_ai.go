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

	prompt := GetAIPrompt("summarize") + sb.String()

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

	// 精细化 Prompt - 使用可配置提示词
	promptTpl := GetAIPrompt("simulate")
	replacer := strings.NewReplacer(
		"{{target_name}}", targetName,
		"{{history}}", history.String(),
	)
	systemPrompt := replacer.Replace(promptTpl)

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

// buildSentimentPrompt 构建情感分析的 prompt（使用可配置提示词）
func buildSentimentPrompt(monthlyTexts map[string]string) string {
	// 按月份排序
	months := make([]string, 0, len(monthlyTexts))
	for k := range monthlyTexts {
		months = append(months, k)
	}
	sort.Strings(months)

	var monthlyTextsSB strings.Builder
	for _, month := range months {
		monthlyTextsSB.WriteString(fmt.Sprintf("=== %s ===\n%s\n", month, monthlyTexts[month]))
	}

	promptTpl := GetAIPrompt("sentiment")
	return strings.Replace(promptTpl, "{{monthly_texts}}", monthlyTextsSB.String(), 1)
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

	prompt := GetAIPrompt("summary") + sb.String()

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

	prompt := GetAIPrompt("extract_todos") + sb.String()

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

	promptTpl := GetAIPrompt("extract_info")
	prompt := strings.Replace(promptTpl, "{{types_hint}}", typesHint, 1) + sb.String()

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
