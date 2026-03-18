// Package usecase 定义应用层用例
package usecase

import (
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
)

// StatsUseCase 统计用例
type StatsUseCase struct {
	statsService service.StatsService
}

// NewStatsUseCase 创建统计用例实例
func NewStatsUseCase(statsService service.StatsService) *StatsUseCase {
	return &StatsUseCase{statsService: statsService}
}

// GetOverview 获取概览统计
func (uc *StatsUseCase) GetOverview(currentUser *model.User) (*dto.StatsOverview, error) {
	return uc.statsService.GetOverview(currentUser)
}

// GetDefectTypeDistribution 获取缺陷类型分布
func (uc *StatsUseCase) GetDefectTypeDistribution(currentUser *model.User, days int) (*dto.DefectTypeDistributionResponse, error) {
	// 参数校验
	if days < 0 {
		days = 30 // 默认30天
	}
	if days > 365 {
		days = 365 // 最大1年
	}

	return uc.statsService.GetDefectTypeDistribution(currentUser, days)
}

// GetDefectTrend 获取缺陷趋势统计
func (uc *StatsUseCase) GetDefectTrend(currentUser *model.User, days int, granularity string) (*dto.DefectTrendResponse, error) {
	// 参数校验
	if days <= 0 {
		days = 7 // 默认7天
	}
	if granularity != "day" && granularity != "week" && granularity != "month" {
		granularity = "day" // 默认按天
	}

	return uc.statsService.GetDefectTrend(currentUser, days, granularity)
}

// GetBridgeRanking 获取桥梁健康度排名
func (uc *StatsUseCase) GetBridgeRanking(currentUser *model.User, limit int, order string) (*dto.BridgeRankingResponse, error) {
	// 参数校验
	if limit <= 0 {
		limit = 10 // 默认10条
	}
	if limit > 50 {
		limit = 50 // 最大50条
	}
	if order != "worst" && order != "best" {
		order = "worst" // 默认最差优先
	}

	return uc.statsService.GetBridgeRanking(currentUser, limit, order)
}

// GetRecentDetections 获取最近检测记录
func (uc *StatsUseCase) GetRecentDetections(currentUser *model.User, limit int) (*dto.RecentDetectionsResponse, error) {
	// 参数校验
	if limit <= 0 {
		limit = 10 // 默认10条
	}
	if limit > 50 {
		limit = 50 // 最大50条
	}

	return uc.statsService.GetRecentDetections(currentUser, limit)
}

// GetHighRiskAlerts 获取高危缺陷告警
func (uc *StatsUseCase) GetHighRiskAlerts(currentUser *model.User, severity string, limit int) (*dto.HighRiskAlertsResponse, error) {
	// 参数校验
	if limit <= 0 {
		limit = 20 // 默认20条
	}
	if limit > 100 {
		limit = 100 // 最大100条
	}

	return uc.statsService.GetHighRiskAlerts(currentUser, severity, limit)
}
