// Package repository 定义数据访问层接口
package repository

import (
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
)

// DefectRepository 缺陷数据访问接口
// 定义所有缺陷相关的数据库操作
type DefectRepository interface {
	// Create 创建新缺陷记录
	// 参数：
	//   - defect: 缺陷实体
	// 返回：
	//   - error: 操作错误
	Create(defect *model.Defect) error

	// FindByID 根据ID查询缺陷
	// 参数：
	//   - id: 缺陷ID
	// 返回：
	//   - *model.Defect: 缺陷实体（未找到返回nil）
	//   - error: 操作错误
	FindByID(id uint) (*model.Defect, error)

	// Delete 删除缺陷（软删除）
	// 参数：
	//   - id: 缺陷ID
	// 返回：
	//   - error: 操作错误
	Delete(id uint) error

	// List 查询缺陷列表（支持复杂过滤和权限控制）
	// 参数：
	//   - filters: 过滤条件
	// 返回：
	//   - []model.Defect: 缺陷列表
	//   - int64: 总数量
	//   - error: 操作错误
	List(filters DefectListFilters) ([]model.Defect, int64, error)

	// ListByBridgeID 根据桥梁ID查询缺陷列表
	// 参数：
	//   - bridgeID: 桥梁ID
	//   - page: 页码（从1开始）
	//   - pageSize: 每页数量
	// 返回：
	//   - []model.Defect: 缺陷列表
	//   - int64: 总数量
	//   - error: 操作错误
	ListByBridgeID(bridgeID uint, page, pageSize int) ([]model.Defect, int64, error)
}

// DefectListFilters 缺陷列表过滤条件
type DefectListFilters struct {
	Page        int          // 页码（从1开始）
	PageSize    int          // 每页数量
	BridgeID    *uint        // 按桥梁ID过滤（可选）
	DefectType  string       // 按缺陷类型过滤（可选）
	StartTime   *time.Time   // 检测开始时间（可选）
	EndTime     *time.Time   // 检测结束时间（可选）
	CurrentUser *model.User  // 当前用户（用于权限过滤）
}
