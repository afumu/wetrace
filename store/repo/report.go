package repo

import (
	"context"
	"time"

	"github.com/afumu/wetrace/internal/model"
	"github.com/rs/zerolog/log"
)

// GetAnnualReport 获取年度报告数据
func (r *Repository) GetAnnualReport(ctx context.Context, year int) (*model.AnnualReport, error) {
	startTime := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endTime := time.Date(year, 12, 31, 23, 59, 59, 0, time.Local)

	report := &model.AnnualReport{
		Year:         year,
		MessageTypes: make(map[string]int),
	}

	// 1. 获取概览数据
	overview, err := r.getAnnualOverview(ctx, startTime, endTime)
	if err != nil {
		log.Warn().Err(err).Msg("获取年度概览失败")
	}
	report.Overview = overview

	// 2. 获取亲密度排行
	topContacts, err := r.getAnnualTopContacts(ctx, startTime, endTime, 20)
	if err != nil {
		log.Warn().Err(err).Msg("获取年度亲密度排行失败")
	}
	report.TopContacts = topContacts

	// 3. 获取月度趋势
	report.MonthlyTrend = r.getAnnualMonthlyTrend(ctx, startTime, endTime)

	// 4. 获取星期分布
	report.WeekdayDist = r.getAnnualWeekdayDist(ctx, startTime, endTime)

	// 5. 获取小时分布
	report.HourlyDist = r.getAnnualHourlyDist(ctx, startTime, endTime)

	// 6. 获取消息类型分布
	report.MessageTypes = r.getAnnualMessageTypes(ctx, startTime, endTime)

	// 7. 获取亮点数据
	report.Highlights = r.getAnnualHighlights(ctx, startTime, endTime)

	return report, nil
}
