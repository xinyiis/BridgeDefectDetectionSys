// Package handler 实现HTTP处理器
// 负责HTTP请求处理和响应
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/response"
)

// ReportHandler 报表处理器
type ReportHandler struct {
	reportUseCase *usecase.ReportUseCase
}

// NewReportHandler 创建报表处理器
func NewReportHandler(reportUseCase *usecase.ReportUseCase) *ReportHandler {
	return &ReportHandler{reportUseCase: reportUseCase}
}

// CreateReport 创建报表
// @Summary 生成检测报告
// @Tags 报表管理
// @Accept json
// @Produce json
// @Param body body dto.CreateReportRequest true "报表信息"
// @Success 200 {object} response.Response{data=dto.ReportResponse}
// @Router /api/v1/reports [post]
func (h *ReportHandler) CreateReport(c *gin.Context) {
	// 1. 获取当前用户
	currentUser := c.MustGet("current_user").(*model.User)

	// 2. 绑定请求参数
	var req dto.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	// 3. 设置用户ID
	req.UserID = currentUser.ID

	// 4. 调用UseCase
	report, err := h.reportUseCase.CreateReport(&req, currentUser)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "报表生成任务已创建，正在后台生成中", report)
}

// GetReport 获取报表详情
// @Summary 获取报表详情
// @Tags 报表管理
// @Produce json
// @Param id path int true "报表ID"
// @Success 200 {object} response.Response{data=dto.ReportResponse}
// @Router /api/v1/reports/{id} [get]
func (h *ReportHandler) GetReport(c *gin.Context) {
	// 1. 优先从Context获取（中间件预查询）
	if report, exists := c.Get("report"); exists {
		response.Success(c, report)
		return
	}

	// 2. Fallback：解析ID并查询
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "报表ID格式错误")
		return
	}

	currentUser := c.MustGet("current_user").(*model.User)

	report, err := h.reportUseCase.GetReport(uint(reportID), currentUser)
	if err != nil {
		response.NotFound(c, "报表")
		return
	}

	response.Success(c, report)
}

// ListReports 查询报表列表
// @Summary 获取报表列表
// @Tags 报表管理
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param report_type query string false "报表类型"
// @Param bridge_id query int false "桥梁ID"
// @Success 200 {object} response.Response{data=dto.ReportListResponse}
// @Router /api/v1/reports [get]
func (h *ReportHandler) ListReports(c *gin.Context) {
	// 1. 解析查询参数
	var params dto.ReportQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	// 2. 获取当前用户
	currentUser := c.MustGet("current_user").(*model.User)

	// 3. 调用UseCase
	result, err := h.reportUseCase.ListReports(&params, currentUser)
	if err != nil {
		response.InternalErrorWithDetail(c, err.Error())
		return
	}

	response.Success(c, result)
}

// DownloadReport 下载报表
// @Summary 下载报表PDF
// @Tags 报表管理
// @Produce application/pdf
// @Param id path int true "报表ID"
// @Success 200 {file} binary
// @Router /api/v1/reports/{id}/download [get]
func (h *ReportHandler) DownloadReport(c *gin.Context) {
	// 1. 解析ID
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "报表ID格式错误")
		return
	}

	currentUser := c.MustGet("current_user").(*model.User)

	// 2. 获取报表
	report, err := h.reportUseCase.GetReport(uint(reportID), currentUser)
	if err != nil {
		response.NotFound(c, "报表")
		return
	}

	// 3. 检查状态
	if report.Status != model.ReportStatusCompleted {
		response.BadRequest(c, "报表尚未生成完成，请稍后重试")
		return
	}

	// 4. 检查文件是否存在
	if report.FilePath == "" {
		response.InternalError(c)
		return
	}

	// 5. 返回文件
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+report.ReportName+".pdf")
	c.Header("Content-Type", "application/pdf")
	c.File(report.FilePath)
}

// DeleteReport 删除报表
// @Summary 删除报表
// @Tags 报表管理
// @Produce json
// @Param id path int true "报表ID"
// @Success 200 {object} response.Response
// @Router /api/v1/reports/{id} [delete]
func (h *ReportHandler) DeleteReport(c *gin.Context) {
	// 1. 解析ID
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "报表ID格式错误")
		return
	}

	currentUser := c.MustGet("current_user").(*model.User)

	// 2. 调用UseCase
	if err := h.reportUseCase.DeleteReport(uint(reportID), currentUser); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}
