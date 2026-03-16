// Package persistence 实现数据访问层
package persistence

import (
	"errors"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"gorm.io/gorm"
)

// DefectRepositoryImpl 缺陷Repository实现
// 实现 repository.DefectRepository 接口
type DefectRepositoryImpl struct {
	db *gorm.DB // GORM数据库连接
}

// NewDefectRepository 创建缺陷Repository实例
// 参数：
//   - db: GORM数据库连接
// 返回：
//   - repository.DefectRepository: 缺陷Repository接口
func NewDefectRepository(db *gorm.DB) repository.DefectRepository {
	return &DefectRepositoryImpl{db: db}
}

// Create 创建新缺陷记录
func (r *DefectRepositoryImpl) Create(defect *model.Defect) error {
	return r.db.Create(defect).Error
}

// FindByID 根据ID查询缺陷
func (r *DefectRepositoryImpl) FindByID(id uint) (*model.Defect, error) {
	var defect model.Defect
	err := r.db.Preload("Bridge").First(&defect, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到返回nil而不是错误
		}
		return nil, err
	}
	return &defect, nil
}

// Delete 删除缺陷（软删除）
func (r *DefectRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&model.Defect{}, id).Error
}

// List 查询缺陷列表（支持复杂过滤和权限控制）
func (r *DefectRepositoryImpl) List(filters repository.DefectListFilters) ([]model.Defect, int64, error) {
	// 1. 构建基础查询
	query := r.db.Model(&model.Defect{})

	// 2. 权限过滤（JOIN bridges表）
	if filters.CurrentUser != nil && !filters.CurrentUser.IsAdmin() {
		// 普通用户只能看自己桥梁的缺陷
		query = query.Joins("JOIN bridges ON bridges.id = defects.bridge_id").
			Where("bridges.user_id = ?", filters.CurrentUser.ID)
	}

	// 3. 其他过滤条件
	if filters.BridgeID != nil {
		query = query.Where("defects.bridge_id = ?", *filters.BridgeID)
	}
	if filters.DefectType != "" {
		query = query.Where("defects.defect_type = ?", filters.DefectType)
	}
	if filters.StartTime != nil {
		query = query.Where("defects.detected_at >= ?", *filters.StartTime)
	}
	if filters.EndTime != nil {
		query = query.Where("defects.detected_at <= ?", *filters.EndTime)
	}

	// 4. 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 5. 分页查询（需要SELECT defects.*避免字段冲突）
	var defects []model.Defect
	offset := (filters.Page - 1) * filters.PageSize
	err := query.Select("defects.*").
		Preload("Bridge").
		Offset(offset).
		Limit(filters.PageSize).
		Order("defects.detected_at DESC"). // 按检测时间倒序
		Find(&defects).Error

	if err != nil {
		return nil, 0, err
	}

	return defects, total, nil
}

// ListByBridgeID 根据桥梁ID查询缺陷列表
func (r *DefectRepositoryImpl) ListByBridgeID(bridgeID uint, page, pageSize int) ([]model.Defect, int64, error) {
	var defects []model.Defect
	var total int64

	// 计算总数
	if err := r.db.Model(&model.Defect{}).Where("bridge_id = ?", bridgeID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := r.db.Where("bridge_id = ?", bridgeID).
		Preload("Bridge").
		Offset(offset).
		Limit(pageSize).
		Order("detected_at DESC").
		Find(&defects).Error

	if err != nil {
		return nil, 0, err
	}

	return defects, total, nil
}
