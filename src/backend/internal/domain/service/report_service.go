// Package service 定义领域服务层
// 包含核心业务逻辑和领域规则
package service

import (
	"errors"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"gorm.io/gorm"
)

// ReportService 报表领域服务
// 处理报表相关的核心业务逻辑
type ReportService struct {
	db         *gorm.DB                    // GORM数据库连接
	reportRepo repository.ReportRepository // 报表仓储
}

// NewReportService 创建报表服务实例
// 参数：
//   - db: GORM数据库连接
//   - reportRepo: 报表Repository接口
// 返回：
//   - *ReportService: 报表服务实例
func NewReportService(db *gorm.DB, reportRepo repository.ReportRepository) *ReportService {
	return &ReportService{
		db:         db,
		reportRepo: reportRepo,
	}
}

// CreateReport 创建报表记录
// 参数：
//   - report: 报表实体
// 返回：
//   - error: 操作错误
func (s *ReportService) CreateReport(report *model.Report) error {
	// 参数验证
	if report.ReportName == "" {
		return errors.New("报表名称不能为空")
	}
	if report.UserID == 0 {
		return errors.New("用户ID不能为空")
	}

	// 创建报表
	return s.reportRepo.Create(report)
}

// GetReport 获取报表详情（带权限验证）
// 参数：
//   - reportID: 报表ID
//   - currentUser: 当前用户
// 返回：
//   - *model.Report: 报表实体
//   - error: 操作错误
func (s *ReportService) GetReport(reportID uint, currentUser *model.User) (*model.Report, error) {
	// 查询报表
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, errors.New("报表不存在")
	}

	// 权限验证
	if !currentUser.IsAdmin() && !report.IsOwnedBy(currentUser.ID) {
		return nil, errors.New("无权访问此报表")
	}

	return report, nil
}

// UpdateReport 更新报表信息
// 参数：
//   - report: 报表实体（包含更新后的数据）
// 返回：
//   - error: 操作错误
func (s *ReportService) UpdateReport(report *model.Report) error {
	if report.ID == 0 {
		return errors.New("报表ID不能为空")
	}

	return s.reportRepo.Update(report)
}

// DeleteReport 删除报表（软删除）
// 参数：
//   - reportID: 报表ID
//   - currentUser: 当前用户
// 返回：
//   - error: 操作错误
func (s *ReportService) DeleteReport(reportID uint, currentUser *model.User) error {
	// 1. 查询报表（验证存在性）
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return err
	}
	if report == nil {
		return errors.New("报表不存在")
	}

	// 2. 权限验证
	if !currentUser.IsAdmin() && !report.IsOwnedBy(currentUser.ID) {
		return errors.New("无权删除此报表")
	}

	// 3. 软删除报表
	return s.reportRepo.Delete(reportID)
}

// ListReports 分页获取报表列表（带权限过滤）
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
//   - reportType: 报表类型过滤（可选）
//   - bridgeID: 桥梁ID过滤（可选）
//   - currentUser: 当前用户
// 返回：
//   - []model.Report: 报表列表
//   - int64: 总数量
//   - error: 操作错误
func (s *ReportService) ListReports(
	page, pageSize int,
	reportType *model.ReportType,
	bridgeID *uint,
	currentUser *model.User,
) ([]model.Report, int64, error) {
	// 参数验证
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 🔑 关键：根据用户角色选择查询方法
	if currentUser.IsAdmin() {
		// 管理员：查询所有报表（带过滤条件）
		if bridgeID != nil {
			// 按桥梁ID过滤
			return s.reportRepo.ListByBridgeID(*bridgeID, page, pageSize)
		}
		// 全局查询
		return s.reportRepo.List(page, pageSize, reportType)
	}

	// 普通用户：只查询自己的报表
	return s.reportRepo.ListByUserID(currentUser.ID, page, pageSize, reportType)
}
