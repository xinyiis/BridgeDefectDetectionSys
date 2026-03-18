// Package dto 定义数据传输对象（Data Transfer Object）
package dto

import (
	"time"
)

// StatsOverview 概览统计
type StatsOverview struct {
	BridgeCount    int     `json:"bridge_count"`    // 桥梁总数
	DroneCount     int     `json:"drone_count"`     // 无人机总数
	DefectCount    int     `json:"defect_count"`    // 缺陷总数
	DetectionCount int     `json:"detection_count"` // 检测任务数
	TodayDefects   int     `json:"today_defects"`   // 今日新增缺陷
	WeekDefects    int     `json:"week_defects"`    // 本周新增缺陷
	DefectTrend    float64 `json:"defect_trend"`    // 缺陷增长率 (%)
	TrendDirection string  `json:"trend_direction"` // "up" / "down" / "stable"
}

// DefectTypeDistribution 缺陷类型分布
type DefectTypeDistribution struct {
	DefectType    string  `json:"defect_type"`    // 缺陷类型
	Count         int     `json:"count"`          // 数量
	Percentage    float64 `json:"percentage"`     // 占比 (%)
	AvgConfidence float64 `json:"avg_confidence"` // 平均置信度
}

// DefectTypeDistributionResponse 类型分布响应
type DefectTypeDistributionResponse struct {
	Total        int                      `json:"total"`        // 总数
	Distribution []DefectTypeDistribution `json:"distribution"` // 分布列表
}

// DefectTrend 缺陷趋势
type DefectTrend struct {
	Date            string `json:"date"`              // 日期 (YYYY-MM-DD)
	Count           int    `json:"count"`             // 当天缺陷数
	CumulativeCount int    `json:"cumulative_count"`  // 累计缺陷数
}

// DefectTrendResponse 趋势响应
type DefectTrendResponse struct {
	Period      string        `json:"period"`       // 统计周期（7days/30days/90days）
	Granularity string        `json:"granularity"`  // 时间粒度（day/week/month）
	Trend       []DefectTrend `json:"trend"`        // 趋势数据
	Total       int           `json:"total"`        // 周期内总缺陷数
	AvgPerDay   float64       `json:"avg_per_day"`  // 日均缺陷数
	PeakDate    string        `json:"peak_date"`    // 峰值日期
	PeakCount   int           `json:"peak_count"`   // 峰值数量
}

// BridgeHealthRanking 桥梁健康度排名
type BridgeHealthRanking struct {
	BridgeID      uint       `json:"bridge_id"`       // 桥梁ID
	BridgeName    string     `json:"bridge_name"`     // 桥梁名称
	DefectCount   int        `json:"defect_count"`    // 缺陷总数
	HighRiskCount int        `json:"high_risk_count"` // 高危缺陷数 (置信度>0.9)
	LastDetection *time.Time `json:"last_detection"`  // 最近检测时间（可空）
	HealthScore   float64    `json:"health_score"`    // 健康评分 (0-100)
	HealthLevel   string     `json:"health_level"`    // "优秀" / "良好" / "一般" / "较差" / "危险"
}

// BridgeRankingResponse 排名响应
type BridgeRankingResponse struct {
	Ranking []BridgeHealthRanking `json:"ranking"` // 排名列表
}

// RecentDetection 最近检测记录
type RecentDetection struct {
	DetectionID    uint      `json:"detection_id"`    // 缺陷ID（用作检测记录标识）
	BridgeID       uint      `json:"bridge_id"`       // 桥梁ID
	BridgeName     string    `json:"bridge_name"`     // 桥梁名称
	ImagePath      string    `json:"image_path"`      // 检测图片
	DefectCount    int       `json:"defect_count"`    // 本次检测到的缺陷数
	DetectedAt     time.Time `json:"detected_at"`     // 检测时间
	ProcessingTime float64   `json:"processing_time"` // 处理耗时（秒）
}

// RecentDetectionsResponse 最近检测响应
type RecentDetectionsResponse struct {
	Detections []RecentDetection `json:"detections"` // 检测记录列表
}

// HighRiskAlert 高危缺陷告警
type HighRiskAlert struct {
	DefectID   uint      `json:"defect_id"`   // 缺陷ID
	BridgeID   uint      `json:"bridge_id"`   // 桥梁ID
	BridgeName string    `json:"bridge_name"` // 桥梁名称
	DefectType string    `json:"defect_type"` // 缺陷类型
	Confidence float64   `json:"confidence"`  // 置信度
	Area       float64   `json:"area"`        // 面积（平方米）
	Severity   string    `json:"severity"`    // "高危" / "严重" / "紧急"
	DetectedAt time.Time `json:"detected_at"` // 检测时间
	ImagePath  string    `json:"image_path"`  // 缺陷图片
}

// HighRiskAlertsResponse 高危告警响应
type HighRiskAlertsResponse struct {
	Total  int             `json:"total"`  // 总数
	Alerts []HighRiskAlert `json:"alerts"` // 告警列表
}
