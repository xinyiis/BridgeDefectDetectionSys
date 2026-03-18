// Package service 定义领域服务层
// 包含核心业务逻辑和领域规则
package service

import (
	"errors"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"gorm.io/gorm"
)

// DroneService 无人机领域服务
// 处理无人机相关的核心业务逻辑
type DroneService struct {
	db        *gorm.DB                    // GORM数据库连接
	droneRepo repository.DroneRepository  // 无人机仓储
}

// NewDroneService 创建无人机服务实例
// 参数：
//   - db: GORM数据库连接
//   - droneRepo: 无人机Repository接口
//
// 返回：
//   - *DroneService: 无人机服务实例
func NewDroneService(db *gorm.DB, droneRepo repository.DroneRepository) *DroneService {
	return &DroneService{
		db:        db,
		droneRepo: droneRepo,
	}
}

// CreateDrone 创建无人机
// 参数：
//   - drone: 无人机实体
//
// 返回：
//   - error: 操作错误
func (s *DroneService) CreateDrone(drone *model.Drone) error {
	// 1. 参数验证
	if drone.Name == "" {
		return errors.New("无人机名称不能为空")
	}
	if len(drone.Name) > 100 {
		return errors.New("无人机名称不能超过100字符")
	}

	// 2. 直接插入数据库（无需检查唯一性）
	return s.droneRepo.Create(drone)
}

// GetByID 根据ID获取无人机
// 参数：
//   - id: 无人机ID
//
// 返回：
//   - *model.Drone: 无人机实体
//   - error: 操作错误
func (s *DroneService) GetByID(id uint) (*model.Drone, error) {
	return s.droneRepo.FindByID(id)
}

// ListDrones 分页获取无人机列表（带权限过滤）
// 参数：
//   - currentUser: 当前用户
//   - page: 页码
//   - pageSize: 每页数量
//
// 返回：
//   - []model.Drone: 无人机列表
//   - int64: 总数量
//   - error: 操作错误
func (s *DroneService) ListDrones(currentUser *model.User, page, pageSize int) ([]model.Drone, int64, error) {
	var drones []model.Drone
	var total int64

	// 参数验证
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := s.db.Model(&model.Drone{})

	// 🔑 关键：根据用户角色自动过滤
	if !currentUser.IsAdmin() {
		// 普通用户：只查询自己的无人机
		query = query.Where("user_id = ?", currentUser.ID)
	}
	// 管理员：不添加user_id条件，查询所有无人机

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&drones).Error; err != nil {
		return nil, 0, err
	}

	return drones, total, nil
}

// UpdateDrone 更新无人机信息
// 参数：
//   - drone: 无人机实体（包含更新后的数据）
//
// 返回：
//   - error: 操作错误
func (s *DroneService) UpdateDrone(drone *model.Drone) error {
	// 参数验证
	if drone.Name != "" && len(drone.Name) > 100 {
		return errors.New("无人机名称不能超过100字符")
	}

	return s.droneRepo.Update(drone)
}

// DeleteDrone 删除无人机（物理删除）
// 参数：
//   - droneID: 无人机ID
//   - currentUser: 当前用户
//
// 返回：
//   - error: 操作错误
func (s *DroneService) DeleteDrone(droneID uint, currentUser *model.User) error {
	// 1. 查询无人机（验证存在性）
	drone, err := s.droneRepo.FindByID(droneID)
	if err != nil {
		return err
	}
	if drone == nil {
		return errors.New("无人机不存在")
	}

	// 2. 权限验证
	if !currentUser.IsAdmin() && !drone.IsOwnedBy(currentUser.ID) {
		return errors.New("权限不足")
	}

	// 3. 物理删除（无需事务、无级联删除、无文件删除）
	return s.db.Unscoped().Delete(&model.Drone{}, droneID).Error
}
