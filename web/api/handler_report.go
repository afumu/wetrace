package api

import (
	"strconv"
	"time"

	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetAnnualReport 获取年度报告
func (a *API) GetAnnualReport(c *gin.Context) {
	yearStr := c.Query("year")
	year := time.Now().Year()

	if yearStr != "" {
		y, err := strconv.Atoi(yearStr)
		if err != nil || y < 2000 || y > 2100 {
			transport.BadRequest(c, "无效的年份参数")
			return
		}
		year = y
	}

	report, err := a.Store.GetAnnualReport(c.Request.Context(), year)
	if err != nil {
		log.Error().Err(err).Int("year", year).Msg("获取年度报告失败")
		transport.InternalServerError(c, "获取年度报告失败")
		return
	}

	transport.SendSuccess(c, report)
}
