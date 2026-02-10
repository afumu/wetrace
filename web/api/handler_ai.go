package api

import (
	"context"
	"fmt"
	"net/http"
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

	// 提取对方的名字和聊天样本
	var history strings.Builder
	var targetName string
	var targetSamples []string

	// 如果消息太多，取最近的 150 条作为上下文
	if len(msgs) > 150 {
		msgs = msgs[len(msgs)-150:]
	}

	for _, m := range msgs {
		if m.Sender == req.Talker {
			targetName = m.SenderName
			if m.Type == 1 {
				targetSamples = append(targetSamples, m.Content)
			}
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
- 直接输出回复内容，不要附带任何解释或前缀。`, targetName, targetName, targetName, history.String(), targetName)

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
