// Package repository 定义数据访问层接口
// 采用Repository模式，实现业务逻辑与数据访问的解耦
package repository

import (
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
)

// BridgeRepository 桥梁数据访问接口
// 定义所有桥梁相关的数据库操作
type BridgeRepository interface {
	// Create 创建新桥梁
	// 参数：
	//   - bridge: 桥梁实体
	// 返回：
	//   - error: 操作错误（如编号重复）
	Create(bridge *model.Bridge) error

	// FindByID 根据ID查询桥梁
	// 参数：
	//   - id: 桥梁ID
	// 返回：
	//   - *model.Bridge: 桥梁实体（未找到返回nil）
	//   - error: 操作错误
	FindByID(id uint) (*model.Bridge, error)

	// FindByCode 根据桥梁编号查询
	// 参数：
	//   - code: 桥梁编号
	// 返回：
	//   - *model.Bridge: 桥梁实体（未找到返回nil）
	//   - error: 操作错误
	FindByCode(code string) (*model.Bridge, error)

	// Update 更新桥梁信息
	// 参数：
	//   - bridge: 桥梁实体（包含更新后的数据）
	// 返回：
	//   - error: 操作错误
	Update(bridge *model.Bridge) error

	// Delete 删除桥梁（软删除）
	// 参数：
	//   - id: 桥梁ID
	// 返回：
	//   - error: 操作错误
	Delete(id uint) error

	// List 分页查询桥梁列表
	// 参数：
	//   - page: 页码（从1开始）
	//   - pageSize: 每页数量
	// 返回：
	//   - []model.Bridge: 桥梁列表
	//   - int64: 总数量
	//   - error: 操作错误
	List(page, pageSize int) ([]model.Bridge, int64, error)

	// ListByUserID 根据用户ID分页查询桥梁列表
	// 参数：
	//   - userID: 用户ID
	//   - page: 页码（从1开始）
	//   - pageSize: 每页数量
	// 返回：
	//   - []model.Bridge: 桥梁列表
	//   - int64: 总数量
	//   - error: 操作错误
	ListByUserID(userID uint, page, pageSize int) ([]model.Bridge, int64, error)
}
