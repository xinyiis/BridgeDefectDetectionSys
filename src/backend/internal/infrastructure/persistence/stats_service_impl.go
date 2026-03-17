// Package persistence 实现数据访问层
package persistence

import (
	"fmt"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/cache"
	"gorm.io/gorm"
)

// StatsServiceImpl 统计服务实现
type StatsServiceImpl struct {
	db *gorm.DB
}

// NewStatsService 创建统计服务实例
func NewStatsService(db *gorm.DB) service.StatsService {
	return &StatsServiceImpl{db: db}
}

// GetOverview 获取用户概览统计
func (s *StatsServiceImpl) GetOverview(currentUser *model.User) (*dto.StatsOverview, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf(cache.CacheKeyOverview, currentUser.ID)
	var overview dto.StatsOverview
	if err := cache.StatsCache.Get(cacheKey, &overview); err == nil {
		return &overview, nil
	}

	// 2. 缓存未命中，查询数据库
	overview = dto.StatsOverview{}

	// 基础统计（子查询）
	if currentUser.IsAdmin() {
		s.db.Raw(`
			SELECT
				(SELECT COUNT(*) FROM bridges WHERE deleted_at IS NULL) AS bridge_count,
				(SELECT COUNT(*) FROM drones WHERE deleted_at IS NULL) AS drone_count,
				(SELECT COUNT(*) FROM defects WHERE deleted_at IS NULL) AS defect_count,
				(SELECT COUNT(DISTINCT image_path) FROM defects WHERE deleted_at IS NULL) AS detection_count
		`).Scan(&overview)
	} else {
		s.db.Raw(`
			SELECT
				(SELECT COUNT(*) FROM bridges WHERE user_id = ? AND deleted_at IS NULL) AS bridge_count,
				(SELECT COUNT(*) FROM drones WHERE user_id = ? AND deleted_at IS NULL) AS drone_count,
				(SELECT COUNT(d.id) FROM defects d
				 JOIN bridges b ON b.id = d.bridge_id
				 WHERE b.user_id = ? AND d.deleted_at IS NULL) AS defect_count,
				(SELECT COUNT(DISTINCT d.image_path) FROM defects d
				 JOIN bridges b ON b.id = d.bridge_id
				 WHERE b.user_id = ? AND d.deleted_at IS NULL) AS detection_count
		`, currentUser.ID, currentUser.ID, currentUser.ID, currentUser.ID).Scan(&overview)
	}

	// 增量统计（今日、本周、昨日）
	baseQuery := s.buildAuthorizedQuery(currentUser)

	var todayDefects, weekDefects, yesterdayDefects int64
	baseQuery.Where("DATE(defects.detected_at) = CURDATE()").Count(&todayDefects)
	baseQuery.Where("defects.detected_at >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)").Count(&weekDefects)
	baseQuery.Where("DATE(defects.detected_at) = DATE_SUB(CURDATE(), INTERVAL 1 DAY)").Count(&yesterdayDefects)

	overview.TodayDefects = int(todayDefects)
	overview.WeekDefects = int(weekDefects)

	// 计算趋势
	if yesterdayDefects > 0 {
		overview.DefectTrend = float64(todayDefects-yesterdayDefects) / float64(yesterdayDefects) * 100
	}

	if todayDefects > yesterdayDefects {
		overview.TrendDirection = "up"
	} else if todayDefects < yesterdayDefects {
		overview.TrendDirection = "down"
	} else {
		overview.TrendDirection = "stable"
	}

	// 3. 写入缓存
	cache.StatsCache.Set(cacheKey, overview, cache.CacheTTLOverview)

	return &overview, nil
}

// GetDefectTypeDistribution 获取缺陷类型分布
func (s *StatsServiceImpl) GetDefectTypeDistribution(currentUser *model.User, days int) (*dto.DefectTypeDistributionResponse, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf(cache.CacheKeyDefectTypes, currentUser.ID, days)
	var response dto.DefectTypeDistributionResponse
	if err := cache.StatsCache.Get(cacheKey, &response); err == nil {
		return &response, nil
	}

	// 2. 构建查询
	query := s.buildAuthorizedQuery(currentUser)

	if days > 0 {
		query = query.Where("defects.detected_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)", days)
	}

	// 3. 聚合查询
	var distribution []dto.DefectTypeDistribution
	query.Select("defect_type, COUNT(*) AS count, ROUND(AVG(confidence), 2) AS avg_confidence").
		Group("defect_type").
		Order("count DESC").
		Scan(&distribution)

	// 4. 计算百分比
	total := 0
	for _, d := range distribution {
		total += d.Count
	}
	for i := range distribution {
		if total > 0 {
			distribution[i].Percentage = float64(distribution[i].Count) / float64(total) * 100
		}
	}

	response = dto.DefectTypeDistributionResponse{
		Total:        total,
		Distribution: distribution,
	}

	// 5. 写入缓存
	cache.StatsCache.Set(cacheKey, response, cache.CacheTTLDefectType)

	return &response, nil
}

// GetDefectTrend 获取缺陷趋势统计
func (s *StatsServiceImpl) GetDefectTrend(currentUser *model.User, days int, granularity string) (*dto.DefectTrendResponse, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf(cache.CacheKeyDefectTrend, currentUser.ID, days, granularity)
	var response dto.DefectTrendResponse
	if err := cache.StatsCache.Get(cacheKey, &response); err == nil {
		return &response, nil
	}

	// 2. 构建查询
	query := s.buildAuthorizedQuery(currentUser)
	query = query.Where("defects.detected_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)", days)

	// 3. 时间序列查询
	var trend []dto.DefectTrend
	query.Select("DATE(defects.detected_at) AS date, COUNT(*) AS count").
		Group("DATE(defects.detected_at)").
		Order("date ASC").
		Scan(&trend)

	// 4. 计算累计数量
	cumulative := 0
	for i := range trend {
		cumulative += trend[i].Count
		trend[i].CumulativeCount = cumulative
	}

	// 5. 统计分析
	total := cumulative
	avgPerDay := 0.0
	if days > 0 {
		avgPerDay = float64(total) / float64(days)
	}

	peakDate := ""
	peakCount := 0
	for _, t := range trend {
		if t.Count > peakCount {
			peakCount = t.Count
			peakDate = t.Date
		}
	}

	response = dto.DefectTrendResponse{
		Period:      fmt.Sprintf("%ddays", days),
		Granularity: granularity,
		Trend:       trend,
		Total:       total,
		AvgPerDay:   avgPerDay,
		PeakDate:    peakDate,
		PeakCount:   peakCount,
	}

	// 6. 写入缓存
	cache.StatsCache.Set(cacheKey, response, cache.CacheTTLTrend)

	return &response, nil
}

// GetBridgeRanking 获取桥梁健康度排名
func (s *StatsServiceImpl) GetBridgeRanking(currentUser *model.User, limit int, order string) (*dto.BridgeRankingResponse, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf(cache.CacheKeyBridgeRanking, currentUser.ID, limit)
	var response dto.BridgeRankingResponse
	if err := cache.StatsCache.Get(cacheKey, &response); err == nil {
		return &response, nil
	}

	// 2. 构建查询
	query := s.db.Table("bridges b").
		Select(`
			b.id AS bridge_id,
			b.bridge_name,
			COUNT(d.id) AS defect_count,
			SUM(CASE WHEN d.confidence >= 0.9 THEN 1 ELSE 0 END) AS high_risk_count,
			MAX(d.detected_at) AS last_detection
		`).
		Joins("LEFT JOIN defects d ON d.bridge_id = b.id AND d.deleted_at IS NULL").
		Where("b.deleted_at IS NULL")

	// 权限过滤
	if !currentUser.IsAdmin() {
		query = query.Where("b.user_id = ?", currentUser.ID)
	}

	query = query.Group("b.id, b.bridge_name")

	// 排序
	if order == "best" {
		query = query.Order("defect_count ASC, high_risk_count ASC")
	} else {
		query = query.Order("defect_count DESC, high_risk_count DESC")
	}

	query = query.Limit(limit)

	// 3. 查询
	var ranking []dto.BridgeHealthRanking
	query.Scan(&ranking)

	// 4. 计算健康评分和等级
	for i := range ranking {
		ranking[i].HealthScore = calculateHealthScore(ranking[i].DefectCount, ranking[i].HighRiskCount)
		ranking[i].HealthLevel = getHealthLevel(ranking[i].HealthScore)
	}

	response = dto.BridgeRankingResponse{
		Ranking: ranking,
	}

	// 5. 写入缓存
	cache.StatsCache.Set(cacheKey, response, cache.CacheTTLRanking)

	return &response, nil
}

// GetRecentDetections 获取最近检测记录
func (s *StatsServiceImpl) GetRecentDetections(currentUser *model.User, limit int) (*dto.RecentDetectionsResponse, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf(cache.CacheKeyRecentDetection, currentUser.ID, limit)
	var response dto.RecentDetectionsResponse
	if err := cache.StatsCache.Get(cacheKey, &response); err == nil {
		return &response, nil
	}

	// 2. 构建查询（按image_path分组，获取最近检测）
	query := s.db.Table("defects d").
		Select(`
			MIN(d.id) AS detection_id,
			b.id AS bridge_id,
			b.bridge_name,
			d.image_path,
			COUNT(d.id) AS defect_count,
			MIN(d.detected_at) AS detected_at,
			0 AS processing_time
		`).
		Joins("JOIN bridges b ON b.id = d.bridge_id").
		Where("d.deleted_at IS NULL")

	// 权限过滤
	if !currentUser.IsAdmin() {
		query = query.Where("b.user_id = ?", currentUser.ID)
	}

	query = query.Group("d.image_path, b.id, b.bridge_name").
		Order("detected_at DESC").
		Limit(limit)

	// 3. 查询
	var detections []dto.RecentDetection
	query.Scan(&detections)

	response = dto.RecentDetectionsResponse{
		Detections: detections,
	}

	// 4. 写入缓存
	cache.StatsCache.Set(cacheKey, response, cache.CacheTTLRecent)

	return &response, nil
}

// GetHighRiskAlerts 获取高危缺陷告警
func (s *StatsServiceImpl) GetHighRiskAlerts(currentUser *model.User, severity string, limit int) (*dto.HighRiskAlertsResponse, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf(cache.CacheKeyHighRiskAlert, currentUser.ID, severity, limit)
	var response dto.HighRiskAlertsResponse
	if err := cache.StatsCache.Get(cacheKey, &response); err == nil {
		return &response, nil
	}

	// 2. 构建查询
	query := s.db.Table("defects d").
		Select(`
			d.id AS defect_id,
			b.id AS bridge_id,
			b.bridge_name,
			d.defect_type,
			d.confidence,
			d.area,
			d.detected_at,
			d.image_path
		`).
		Joins("JOIN bridges b ON b.id = d.bridge_id").
		Where("d.deleted_at IS NULL")

	// 权限过滤
	if !currentUser.IsAdmin() {
		query = query.Where("b.user_id = ?", currentUser.ID)
	}

	// 高危条件（置信度 >= 0.85 或 面积 >= 0.02）
	query = query.Where("d.confidence >= 0.85 OR d.area >= 0.02")

	// 严重程度过滤
	if severity == "urgent" {
		query = query.Where("d.confidence >= 0.95 AND d.area >= 0.1")
	} else if severity == "serious" {
		query = query.Where("d.confidence >= 0.90 OR d.area >= 0.05")
	}

	query = query.Order("d.confidence DESC, d.area DESC").Limit(limit)

	// 3. 查询
	var alerts []dto.HighRiskAlert
	query.Scan(&alerts)

	// 4. 计算严重程度
	for i := range alerts {
		alerts[i].Severity = determineSeverity(alerts[i].Confidence, alerts[i].Area)
	}

	response = dto.HighRiskAlertsResponse{
		Total:  len(alerts),
		Alerts: alerts,
	}

	// 5. 写入缓存
	cache.StatsCache.Set(cacheKey, response, cache.CacheTTLAlert)

	return &response, nil
}

// buildAuthorizedQuery 构建带权限的基础查询
func (s *StatsServiceImpl) buildAuthorizedQuery(currentUser *model.User) *gorm.DB {
	query := s.db.Model(&model.Defect{})

	if !currentUser.IsAdmin() {
		// 普通用户：JOIN桥梁表过滤
		query = query.Joins("JOIN bridges ON bridges.id = defects.bridge_id").
			Where("bridges.user_id = ?", currentUser.ID)
	}

	return query.Where("defects.deleted_at IS NULL")
}

// calculateHealthScore 计算健康评分
// HealthScore = 100 - (缺陷总数 * 1 + 高危缺陷数 * 5)
func calculateHealthScore(defectCount, highRiskCount int) float64 {
	score := 100.0 - float64(defectCount)*1.0 - float64(highRiskCount)*5.0
	if score < 0 {
		score = 0
	}
	return score
}

// getHealthLevel 获取健康等级
func getHealthLevel(score float64) string {
	switch {
	case score >= 90:
		return "优秀"
	case score >= 70:
		return "良好"
	case score >= 50:
		return "一般"
	case score >= 30:
		return "较差"
	default:
		return "危险"
	}
}

// determineSeverity 判定严重程度
func determineSeverity(confidence, area float64) string {
	// 规则1: 置信度极高 + 面积大 = 紧急
	if confidence >= 0.95 && area >= 0.1 {
		return "紧急"
	}

	// 规则2: 置信度高 或 面积较大 = 严重
	if confidence >= 0.90 || area >= 0.05 {
		return "严重"
	}

	// 规则3: 置信度较高 或 面积一般 = 高危
	if confidence >= 0.85 || area >= 0.02 {
		return "高危"
	}

	return "一般"
}
