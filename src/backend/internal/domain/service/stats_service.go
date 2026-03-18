// Package service 定义领域服务
package service

import (
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
)

// StatsService 统计服务接口
type StatsService interface {
	// GetOverview 获取用户概览统计
	GetOverview(currentUser *model.User) (*dto.StatsOverview, error)

	// GetDefectTypeDistribution 获取缺陷类型分布
	GetDefectTypeDistribution(currentUser *model.User, days int) (*dto.DefectTypeDistributionResponse, error)

	// GetDefectTrend 获取缺陷趋势统计
	GetDefectTrend(currentUser *model.User, days int, granularity string) (*dto.DefectTrendResponse, error)

	// GetBridgeRanking 获取桥梁健康度排名
	GetBridgeRanking(currentUser *model.User, limit int, order string) (*dto.BridgeRankingResponse, error)

	// GetRecentDetections 获取最近检测记录
	GetRecentDetections(currentUser *model.User, limit int) (*dto.RecentDetectionsResponse, error)

	// GetHighRiskAlerts 获取高危缺陷告警
	GetHighRiskAlerts(currentUser *model.User, severity string, limit int) (*dto.HighRiskAlertsResponse, error)
}
