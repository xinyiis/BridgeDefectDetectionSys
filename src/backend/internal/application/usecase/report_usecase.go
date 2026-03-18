// Package usecase 实现应用层用例
// 负责业务流程编排和DTO转换
package usecase

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/pdf"
)

// ReportUseCase 报表用例
type ReportUseCase struct {
	reportService *service.ReportService
	bridgeService *service.BridgeService
	defectService *service.DefectService
	pdfGenerator  *pdf.ReportGenerator
	reportBaseDir string // 报表存储根目录
}

// NewReportUseCase 创建报表用例
func NewReportUseCase(
	reportService *service.ReportService,
	bridgeService *service.BridgeService,
	defectService *service.DefectService,
	pdfGenerator *pdf.ReportGenerator,
	reportBaseDir string,
) *ReportUseCase {
	return &ReportUseCase{
		reportService: reportService,
		bridgeService: bridgeService,
		defectService: defectService,
		pdfGenerator:  pdfGenerator,
		reportBaseDir: reportBaseDir,
	}
}

// CreateReport 创建报表
func (uc *ReportUseCase) CreateReport(req *dto.CreateReportRequest, currentUser *model.User) (*dto.ReportResponse, error) {
	// 1. 参数验证
	if err := uc.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// 2. 解析时间
	startTime, err := time.Parse("2006-01-02", req.StartTime)
	if err != nil {
		return nil, errors.New("开始时间格式错误，应为：YYYY-MM-DD")
	}
	endTime, err := time.Parse("2006-01-02", req.EndTime)
	if err != nil {
		return nil, errors.New("结束时间格式错误，应为：YYYY-MM-DD")
	}

	// 验证时间范围
	if endTime.Before(startTime) {
		return nil, errors.New("结束时间不能早于开始时间")
	}

	// 3. 验证桥梁所有权（针对bridge_inspection类型）
	if req.BridgeID != nil {
		bridge, err := uc.bridgeService.GetByID(*req.BridgeID)
		if err != nil {
			return nil, errors.New("桥梁不存在")
		}
		if !currentUser.IsAdmin() && !bridge.IsOwnedBy(currentUser.ID) {
			return nil, errors.New("无权生成该桥梁的报表")
		}
	}

	// 4. 创建报表记录（状态为generating）
	report := &model.Report{
		ReportName: req.ReportName,
		ReportType: req.ReportType,
		UserID:     req.UserID,
		BridgeID:   req.BridgeID,
		BridgeIDs:  req.BridgeIDs,
		StartTime:  startTime,
		EndTime:    endTime,
		Status:     model.ReportStatusGenerating,
	}

	if err := uc.reportService.CreateReport(report); err != nil {
		return nil, err
	}

	// 5. 异步生成PDF
	go uc.generatePDFAsync(report, currentUser)

	// 6. 返回报表信息
	return uc.toReportResponse(report), nil
}

// generatePDFAsync 异步生成PDF
func (uc *ReportUseCase) generatePDFAsync(report *model.Report, currentUser *model.User) {
	// 1. 确保reports目录存在
	if err := os.MkdirAll(uc.reportBaseDir, 0755); err != nil {
		report.MarkAsFailed(fmt.Sprintf("创建报表目录失败：%s", err.Error()))
		uc.reportService.UpdateReport(report)
		return
	}

	// 2. 获取桥梁信息
	if report.BridgeID == nil {
		report.MarkAsFailed("报表缺少桥梁ID")
		uc.reportService.UpdateReport(report)
		return
	}

	bridge, err := uc.bridgeService.GetByID(*report.BridgeID)
	if err != nil {
		report.MarkAsFailed(fmt.Sprintf("获取桥梁信息失败：%s", err.Error()))
		uc.reportService.UpdateReport(report)
		return
	}

	// 3. 查询缺陷列表
	defects, err := uc.defectService.ListDefectsByBridgeAndTime(
		*report.BridgeID,
		report.StartTime,
		report.EndTime,
		currentUser,
	)
	if err != nil {
		report.MarkAsFailed(fmt.Sprintf("查询缺陷数据失败：%s", err.Error()))
		uc.reportService.UpdateReport(report)
		return
	}

	// 4. 计算统计数据
	highRiskCount := 0
	for _, defect := range defects {
		if defect.Confidence >= 0.85 || defect.Area >= 0.02 {
			highRiskCount++
		}
	}

	// 计算健康度评分
	healthScore := 100.0 - float64(len(defects)) - float64(highRiskCount)*5.0
	if healthScore < 0 {
		healthScore = 0
	}

	// 更新报表统计字段
	report.DefectCount = len(defects)
	report.HighRiskCount = highRiskCount
	report.HealthScore = healthScore

	// 5. 生成PDF
	fileName := fmt.Sprintf("report_%d_%s.pdf", report.ID, time.Now().Format("20060102150405"))
	filePath := filepath.Join(uc.reportBaseDir, fileName)

	err = uc.pdfGenerator.GenerateBridgeInspectionReport(report, bridge, defects, filePath)
	if err != nil {
		report.MarkAsFailed(fmt.Sprintf("PDF生成失败：%s", err.Error()))
		uc.reportService.UpdateReport(report)
		return
	}

	// 6. 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		report.MarkAsFailed(fmt.Sprintf("获取文件信息失败：%s", err.Error()))
		uc.reportService.UpdateReport(report)
		return
	}

	// 7. 粗略估算页数（每4KB约1页）
	totalPages := int(fileInfo.Size() / 4096)
	if totalPages < 1 {
		totalPages = 1
	}

	// 8. 更新报表状态为completed
	report.MarkAsCompleted(filePath, fileInfo.Size(), totalPages)
	uc.reportService.UpdateReport(report)
}

// validateCreateRequest 验证创建请求
func (uc *ReportUseCase) validateCreateRequest(req *dto.CreateReportRequest) error {
	switch req.ReportType {
	case model.ReportTypeBridgeInspection:
		if req.BridgeID == nil {
			return errors.New("桥梁检测报告必须指定bridge_id")
		}
	case model.ReportTypeDefectAnalysis:
		if req.BridgeID == nil {
			return errors.New("缺陷分析报告必须指定bridge_id")
		}
	case model.ReportTypeHealthComparison:
		if len(req.BridgeIDs) < 2 {
			return errors.New("健康对比报告至少需要2座桥梁")
		}
	default:
		return errors.New("不支持的报表类型")
	}
	return nil
}

// GetReport 获取报表详情
func (uc *ReportUseCase) GetReport(reportID uint, currentUser *model.User) (*dto.ReportResponse, error) {
	report, err := uc.reportService.GetReport(reportID, currentUser)
	if err != nil {
		return nil, err
	}
	return uc.toReportResponse(report), nil
}

// ListReports 查询报表列表
func (uc *ReportUseCase) ListReports(params *dto.ReportQueryParams, currentUser *model.User) (*dto.ReportListResponse, error) {
	// 设置默认值
	if params.Page == 0 {
		params.Page = 1
	}
	if params.PageSize == 0 {
		params.PageSize = 10
	}

	reports, total, err := uc.reportService.ListReports(
		params.Page,
		params.PageSize,
		params.ReportType,
		params.BridgeID,
		currentUser,
	)
	if err != nil {
		return nil, err
	}

	list := make([]dto.ReportResponse, 0, len(reports))
	for _, report := range reports {
		list = append(list, *uc.toReportResponse(&report))
	}

	return &dto.ReportListResponse{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		List:     list,
	}, nil
}

// DeleteReport 删除报表
func (uc *ReportUseCase) DeleteReport(reportID uint, currentUser *model.User) error {
	// 1. 获取报表
	report, err := uc.reportService.GetReport(reportID, currentUser)
	if err != nil {
		return err
	}

	// 2. 删除PDF文件
	if report.FilePath != "" {
		if err := os.Remove(report.FilePath); err != nil {
			// 文件删除失败只记录日志，不影响数据库删除
			fmt.Printf("警告：删除PDF文件失败: %s, error: %v\n", report.FilePath, err)
		}
	}

	// 3. 软删除数据库记录
	return uc.reportService.DeleteReport(reportID, currentUser)
}

// toReportResponse DTO转换
func (uc *ReportUseCase) toReportResponse(report *model.Report) *dto.ReportResponse {
	resp := &dto.ReportResponse{
		ID:            report.ID,
		ReportName:    report.ReportName,
		ReportType:    report.ReportType,
		UserID:        report.UserID,
		BridgeID:      report.BridgeID,
		StartTime:     report.StartTime,
		EndTime:       report.EndTime,
		FilePath:      report.FilePath,
		FileSize:      report.FileSize,
		Status:        report.Status,
		ErrorMessage:  report.ErrorMessage,
		TotalPages:    report.TotalPages,
		DefectCount:   report.DefectCount,
		HighRiskCount: report.HighRiskCount,
		HealthScore:   report.HealthScore,
		CreatedAt:     report.CreatedAt,
	}

	if report.Bridge != nil {
		resp.BridgeName = report.Bridge.BridgeName
	}

	return resp
}
