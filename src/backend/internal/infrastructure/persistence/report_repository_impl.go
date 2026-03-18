// Package persistence 实现数据持久化（Repository接口的GORM实现）
package persistence

import (
	"errors"

	"gorm.io/gorm"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
)

// ReportRepositoryImpl 报表仓储实现
type ReportRepositoryImpl struct {
	db *gorm.DB
}

// NewReportRepository 创建报表仓储实例
// 参数：
//   - db: GORM数据库实例
// 返回：
//   - repository.ReportRepository: 报表仓储接口
func NewReportRepository(db *gorm.DB) repository.ReportRepository {
	return &ReportRepositoryImpl{db: db}
}

// Create 创建报表记录
func (r *ReportRepositoryImpl) Create(report *model.Report) error {
	return r.db.Create(report).Error
}

// FindByID 根据ID查询报表
func (r *ReportRepositoryImpl) FindByID(id uint) (*model.Report, error) {
	var report model.Report
	err := r.db.Preload("User").Preload("Bridge").First(&report, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("报表不存在")
		}
		return nil, err
	}
	return &report, nil
}

// Update 更新报表
func (r *ReportRepositoryImpl) Update(report *model.Report) error {
	return r.db.Save(report).Error
}

// Delete 删除报表（软删除）
func (r *ReportRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&model.Report{}, id).Error
}

// List 查询报表列表（分页）
func (r *ReportRepositoryImpl) List(page, pageSize int, reportType *model.ReportType) ([]model.Report, int64, error) {
	var reports []model.Report
	var total int64

	query := r.db.Model(&model.Report{}).Preload("Bridge")

	// 类型过滤
	if reportType != nil && *reportType != "" {
		query = query.Where("report_type = ?", *reportType)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// ListByUserID 查询指定用户的报表列表
func (r *ReportRepositoryImpl) ListByUserID(userID uint, page, pageSize int, reportType *model.ReportType) ([]model.Report, int64, error) {
	var reports []model.Report
	var total int64

	query := r.db.Model(&model.Report{}).Where("user_id = ?", userID).Preload("Bridge")

	// 类型过滤
	if reportType != nil && *reportType != "" {
		query = query.Where("report_type = ?", *reportType)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// ListByBridgeID 查询指定桥梁的报表列表
func (r *ReportRepositoryImpl) ListByBridgeID(bridgeID uint, page, pageSize int) ([]model.Report, int64, error) {
	var reports []model.Report
	var total int64

	query := r.db.Model(&model.Report{}).Where("bridge_id = ?", bridgeID).Preload("Bridge")

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// CountByStatus 统计指定状态的报表数量
func (r *ReportRepositoryImpl) CountByStatus(status model.ReportStatus) (int64, error) {
	var count int64
	err := r.db.Model(&model.Report{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
