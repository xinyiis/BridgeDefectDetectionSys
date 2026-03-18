// Package persistence 实现数据访问层
// 包含所有Repository接口的具体实现
package persistence

import (
	"errors"
	"fmt"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"gorm.io/gorm"
)

// BridgeRepositoryImpl 桥梁Repository实现
// 实现 repository.BridgeRepository 接口
type BridgeRepositoryImpl struct {
	db *gorm.DB // GORM数据库连接
}

// NewBridgeRepository 创建桥梁Repository实例
// 参数：
//   - db: GORM数据库连接
// 返回：
//   - repository.BridgeRepository: 桥梁Repository接口
func NewBridgeRepository(db *gorm.DB) repository.BridgeRepository {
	return &BridgeRepositoryImpl{db: db}
}

// Create 创建新桥梁
func (r *BridgeRepositoryImpl) Create(bridge *model.Bridge) error {
	return r.db.Create(bridge).Error
}

// FindByID 根据ID查询桥梁
func (r *BridgeRepositoryImpl) FindByID(id uint) (*model.Bridge, error) {
	var bridge model.Bridge
	err := r.db.First(&bridge, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到返回nil而不是错误
		}
		return nil, err
	}
	return &bridge, nil
}

// FindByCode 根据桥梁编号查询
func (r *BridgeRepositoryImpl) FindByCode(code string) (*model.Bridge, error) {
	var bridge model.Bridge
	err := r.db.Where("bridge_code = ?", code).First(&bridge).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到返回nil而不是错误
		}
		return nil, err
	}
	return &bridge, nil
}

// Update 更新桥梁信息
func (r *BridgeRepositoryImpl) Update(bridge *model.Bridge) error {
	return r.db.Save(bridge).Error
}

// Delete 删除桥梁（软删除 + 修改编号释放唯一索引）
// 解决软删除与唯一索引冲突的问题
func (r *BridgeRepositoryImpl) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 查询桥梁
		var bridge model.Bridge
		if err := tx.First(&bridge, id).Error; err != nil {
			return err
		}

		// 2. 修改编号，释放唯一索引
		// 在编号后加时间戳，避免与新建桥梁冲突
		timestamp := time.Now().Unix()
		bridge.BridgeCode = fmt.Sprintf("%s_deleted_%d", bridge.BridgeCode, timestamp)

		if err := tx.Save(&bridge).Error; err != nil {
			return err
		}

		// 3. 软删除桥梁
		return tx.Delete(&bridge).Error
	})
}

// List 分页查询桥梁列表
func (r *BridgeRepositoryImpl) List(page, pageSize int) ([]model.Bridge, int64, error) {
	var bridges []model.Bridge
	var total int64

	// 计算总数
	if err := r.db.Model(&model.Bridge{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&bridges).Error; err != nil {
		return nil, 0, err
	}

	return bridges, total, nil
}

// ListByUserID 根据用户ID分页查询桥梁列表
func (r *BridgeRepositoryImpl) ListByUserID(userID uint, page, pageSize int) ([]model.Bridge, int64, error) {
	var bridges []model.Bridge
	var total int64

	// 计算总数
	if err := r.db.Model(&model.Bridge{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&bridges).Error; err != nil {
		return nil, 0, err
	}

	return bridges, total, nil
}
