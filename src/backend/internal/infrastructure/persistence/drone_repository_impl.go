// Package persistence 实现数据访问层
// 包含所有Repository接口的具体实现
package persistence

import (
	"errors"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"gorm.io/gorm"
)

// DroneRepositoryImpl 无人机Repository实现
// 实现 repository.DroneRepository 接口
type DroneRepositoryImpl struct {
	db *gorm.DB // GORM数据库连接
}

// NewDroneRepository 创建无人机Repository实例
// 参数：
//   - db: GORM数据库连接
//
// 返回：
//   - repository.DroneRepository: 无人机Repository接口
func NewDroneRepository(db *gorm.DB) repository.DroneRepository {
	return &DroneRepositoryImpl{db: db}
}

// Create 创建新无人机
func (r *DroneRepositoryImpl) Create(drone *model.Drone) error {
	return r.db.Create(drone).Error
}

// FindByID 根据ID查询无人机
func (r *DroneRepositoryImpl) FindByID(id uint) (*model.Drone, error) {
	var drone model.Drone
	err := r.db.First(&drone, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到返回nil而不是错误
		}
		return nil, err
	}
	return &drone, nil
}

// Update 更新无人机信息
func (r *DroneRepositoryImpl) Update(drone *model.Drone) error {
	return r.db.Save(drone).Error
}

// Delete 删除无人机（物理删除）
// 注意：这是物理删除，不使用软删除
// 无需事务、无需修改编号、无需级联删除
func (r *DroneRepositoryImpl) Delete(id uint) error {
	return r.db.Unscoped().Delete(&model.Drone{}, id).Error
}

// List 分页查询无人机列表
func (r *DroneRepositoryImpl) List(page, pageSize int) ([]model.Drone, int64, error) {
	var drones []model.Drone
	var total int64

	// 计算总数
	if err := r.db.Model(&model.Drone{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&drones).Error; err != nil {
		return nil, 0, err
	}

	return drones, total, nil
}

// ListByUserID 根据用户ID分页查询无人机列表
func (r *DroneRepositoryImpl) ListByUserID(userID uint, page, pageSize int) ([]model.Drone, int64, error) {
	var drones []model.Drone
	var total int64

	// 计算总数
	if err := r.db.Model(&model.Drone{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&drones).Error; err != nil {
		return nil, 0, err
	}

	return drones, total, nil
}
