package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetHourlyActivity 获取每小时活跃度
func (a *API) GetHourlyActivity(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session id"})
		return
	}

	stats, err := a.Store.GetHourlyActivity(c.Request.Context(), sessionID)
	if err != nil {
		log.Error().Err(err).Msg("获取时段活跃度失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetDailyActivity 获取每日活跃度
func (a *API) GetDailyActivity(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session id"})
		return
	}

	stats, err := a.Store.GetDailyActivity(c.Request.Context(), sessionID)
	if err != nil {
		log.Error().Err(err).Msg("获取每日活跃度失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetWeekdayActivity 获取星期活跃度
func (a *API) GetWeekdayActivity(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session id"})
		return
	}

	stats, err := a.Store.GetWeekdayActivity(c.Request.Context(), sessionID)
	if err != nil {
		log.Error().Err(err).Msg("获取星期活跃度失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetMonthlyActivity 获取月份活跃度
func (a *API) GetMonthlyActivity(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session id"})
		return
	}

	stats, err := a.Store.GetMonthlyActivity(c.Request.Context(), sessionID)
	if err != nil {
		log.Error().Err(err).Msg("获取月份活跃度失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetMessageTypeDistribution 获取消息类型分布
func (a *API) GetMessageTypeDistribution(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session id"})
		return
	}

	stats, err := a.Store.GetMessageTypeDistribution(c.Request.Context(), sessionID)
	if err != nil {
		log.Error().Err(err).Msg("获取消息类型分布失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetMemberActivity 获取成员活跃度
func (a *API) GetMemberActivity(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session id"})
		return
	}

	stats, err := a.Store.GetMemberActivity(c.Request.Context(), sessionID)
	if err != nil {
		log.Error().Err(err).Msg("获取成员活跃度失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRepeatAnalysis 获取复读分析
func (a *API) GetRepeatAnalysis(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session id"})
		return
	}

	stats, err := a.Store.GetRepeatAnalysis(c.Request.Context(), sessionID)
	if err != nil {
		log.Error().Err(err).Msg("获取复读分析失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetPersonalTopContacts 获取个人社交排行榜
func (a *API) GetPersonalTopContacts(c *gin.Context) {
	stats, err := a.Store.GetPersonalTopContacts(c.Request.Context(), 100)
	if err != nil {
		log.Error().Err(err).Msg("获取个人社交排行榜失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
