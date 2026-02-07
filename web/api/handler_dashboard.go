package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetDashboard 获取总览数据
func (a *API) GetDashboard(c *gin.Context) {
	data, err := a.Store.GetDashboardData(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("获取总览数据失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}
