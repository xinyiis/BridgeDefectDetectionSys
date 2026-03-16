// Package repository 定义仓储层接口
// 仓储层负责数据持久化的抽象定义
package repository

import "github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"

// DroneRepository 无人机仓储接口
// 定义无人机数据访问的标准方法
type DroneRepository interface {
	// Create 创建无人机
	// 参数：
	//   - drone: 要创建的无人机实体
	// 返回值：
	//   - error: 创建失败时返回错误信息
	Create(drone *model.Drone) error

	// FindByID 根据ID查询无人机
	// 参数：
	//   - id: 无人机ID
	// 返回值：
	//   - *model.Drone: 查询到的无人机实体，未找到返回nil
	//   - error: 查询失败时返回错误信息
	FindByID(id uint) (*model.Drone, error)

	// Update 更新无人机信息
	// 参数：
	//   - drone: 要更新的无人机实体（必须包含ID）
	// 返回值：
	//   - error: 更新失败时返回错误信息
	Update(drone *model.Drone) error

	// Delete 删除无人机（物理删除）
	// 参数：
	//   - id: 要删除的无人机ID
	// 返回值：
	//   - error: 删除失败时返回错误信息
	// 注意：这是物理删除，记录将从数据库中永久删除
	Delete(id uint) error

	// List 分页查询所有无人机
	// 参数：
	//   - page: 页码（从1开始）
	//   - pageSize: 每页数量
	// 返回值：
	//   - []model.Drone: 无人机列表
	//   - int64: 总记录数
	//   - error: 查询失败时返回错误信息
	List(page, pageSize int) ([]model.Drone, int64, error)

	// ListByUserID 分页查询指定用户的无人机
	// 参数：
	//   - userID: 用户ID
	//   - page: 页码（从1开始）
	//   - pageSize: 每页数量
	// 返回值：
	//   - []model.Drone: 无人机列表
	//   - int64: 总记录数
	//   - error: 查询失败时返回错误信息
	ListByUserID(userID uint, page, pageSize int) ([]model.Drone, int64, error)
}
