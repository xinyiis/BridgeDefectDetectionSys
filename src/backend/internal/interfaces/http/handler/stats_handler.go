// Package handler 定义HTTP请求处理器
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/response"
)

// StatsHandler 统计接口处理器
type StatsHandler struct {
	statsUseCase *usecase.StatsUseCase
}

// NewStatsHandler 创建统计Handler实例
func NewStatsHandler(statsUseCase *usecase.StatsUseCase) *StatsHandler {
	return &StatsHandler{statsUseCase: statsUseCase}
}

// GetOverview 获取概览统计
// @Summary 获取用户概览统计
// @Tags 统计
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/stats/overview [get]
func (h *StatsHandler) GetOverview(c *gin.Context) {
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	data, err := h.statsUseCase.GetOverview(user)
	if err != nil {
		response.InternalErrorWithDetail(c, err.Error())
		return
	}

	response.Success(c, data)
}

// GetDefectTypeDistribution 获取缺陷类型分布
// @Summary 获取缺陷类型分布统计
// @Tags 统计
// @Produce json
// @Param days query int false "统计天数"
// @Success 200 {object} response.Response
// @Router /api/v1/stats/defect-types [get]
func (h *StatsHandler) GetDefectTypeDistribution(c *gin.Context) {
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	data, err := h.statsUseCase.GetDefectTypeDistribution(user, days)
	if err != nil {
		response.InternalErrorWithDetail(c, err.Error())
		return
	}

	response.Success(c, data)
}

// GetDefectTrend 获取缺陷趋势统计
// @Summary 获取缺陷趋势统计
// @Tags 统计
// @Produce json
// @Param days query int false "统计天数"
// @Param granularity query string false "时间粒度（day/week/month）"
// @Success 200 {object} response.Response
// @Router /api/v1/stats/defect-trend [get]
func (h *StatsHandler) GetDefectTrend(c *gin.Context) {
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	granularity := c.DefaultQuery("granularity", "day")

	data, err := h.statsUseCase.GetDefectTrend(user, days, granularity)
	if err != nil {
		response.InternalErrorWithDetail(c, err.Error())
		return
	}

	response.Success(c, data)
}

// GetBridgeRanking 获取桥梁健康度排名
// @Summary 获取桥梁健康度排名
// @Tags 统计
// @Produce json
// @Param limit query int false "返回数量"
// @Param order query string false "排序方式（worst/best）"
// @Success 200 {object} response.Response
// @Router /api/v1/stats/bridge-ranking [get]
func (h *StatsHandler) GetBridgeRanking(c *gin.Context) {
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	order := c.DefaultQuery("order", "worst")

	data, err := h.statsUseCase.GetBridgeRanking(user, limit, order)
	if err != nil {
		response.InternalErrorWithDetail(c, err.Error())
		return
	}

	response.Success(c, data)
}

// GetRecentDetections 获取最近检测记录
// @Summary 获取最近检测记录
// @Tags 统计
// @Produce json
// @Param limit query int false "返回数量"
// @Success 200 {object} response.Response
// @Router /api/v1/stats/recent-detections [get]
func (h *StatsHandler) GetRecentDetections(c *gin.Context) {
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, err := h.statsUseCase.GetRecentDetections(user, limit)
	if err != nil {
		response.InternalErrorWithDetail(c, err.Error())
		return
	}

	response.Success(c, data)
}

// GetHighRiskAlerts 获取高危缺陷告警
// @Summary 获取高危缺陷告警
// @Tags 统计
// @Produce json
// @Param severity query string false "严重程度（high/serious/urgent）"
// @Param limit query int false "返回数量"
// @Success 200 {object} response.Response
// @Router /api/v1/stats/high-risk-alerts [get]
func (h *StatsHandler) GetHighRiskAlerts(c *gin.Context) {
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	severity := c.DefaultQuery("severity", "")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	data, err := h.statsUseCase.GetHighRiskAlerts(user, severity, limit)
	if err != nil {
		response.InternalErrorWithDetail(c, err.Error())
		return
	}

	response.Success(c, data)
}
