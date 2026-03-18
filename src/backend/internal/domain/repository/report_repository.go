// Package repository 定义数据访问层接口
// 采用Repository模式，实现业务逻辑与数据访问的解耦
package repository

import (
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
)

// ReportRepository 报表数据访问接口
// 定义所有报表相关的数据库操作
type ReportRepository interface {
	// Create 创建报表记录
	// 参数：
	//   - report: 报表实体
	// 返回：
	//   - error: 操作错误
	Create(report *model.Report) error

	// FindByID 根据ID查询报表
	// 参数：
	//   - id: 报表ID
	// 返回：
	//   - *model.Report: 报表实体（未找到返回nil）
	//   - error: 操作错误
	FindByID(id uint) (*model.Report, error)

	// Update 更新报表
	// 参数：
	//   - report: 报表实体（包含更新后的数据）
	// 返回：
	//   - error: 操作错误
	Update(report *model.Report) error

	// Delete 删除报表（软删除）
	// 参数：
	//   - id: 报表ID
	// 返回：
	//   - error: 操作错误
	Delete(id uint) error

	// List 查询报表列表（分页）
	// 参数：
	//   - page: 页码（从1开始）
	//   - pageSize: 每页数量
	//   - reportType: 报表类型过滤（可选）
	// 返回：
	//   - []model.Report: 报表列表
	//   - int64: 总数量
	//   - error: 操作错误
	List(page, pageSize int, reportType *model.ReportType) ([]model.Report, int64, error)

	// ListByUserID 查询指定用户的报表列表
	// 参数：
	//   - userID: 用户ID
	//   - page: 页码（从1开始）
	//   - pageSize: 每页数量
	//   - reportType: 报表类型过滤（可选）
	// 返回：
	//   - []model.Report: 报表列表
	//   - int64: 总数量
	//   - error: 操作错误
	ListByUserID(userID uint, page, pageSize int, reportType *model.ReportType) ([]model.Report, int64, error)

	// ListByBridgeID 查询指定桥梁的报表列表
	// 参数：
	//   - bridgeID: 桥梁ID
	//   - page: 页码（从1开始）
	//   - pageSize: 每页数量
	// 返回：
	//   - []model.Report: 报表列表
	//   - int64: 总数量
	//   - error: 操作错误
	ListByBridgeID(bridgeID uint, page, pageSize int) ([]model.Report, int64, error)

	// CountByStatus 统计指定状态的报表数量
	// 参数：
	//   - status: 报表状态
	// 返回：
	//   - int64: 报表数量
	//   - error: 操作错误
	CountByStatus(status model.ReportStatus) (int64, error)
}
