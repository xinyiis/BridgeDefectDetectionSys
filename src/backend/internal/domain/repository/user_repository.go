// Package repository 定义数据访问层接口
// 采用Repository模式，实现业务逻辑与数据访问的解耦
package repository

import (
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
)

// UserRepository 用户数据访问接口
// 定义所有用户相关的数据库操作
type UserRepository interface {
	// Create 创建新用户
	// 参数：
	//   - user: 用户实体
	// 返回：
	//   - error: 操作错误（如用户名/邮箱重复）
	Create(user *model.User) error

	// FindByID 根据ID查询用户
	// 参数：
	//   - id: 用户ID
	// 返回：
	//   - *model.User: 用户实体（未找到返回nil）
	//   - error: 操作错误
	FindByID(id uint) (*model.User, error)

	// FindByUsername 根据用户名查询用户
	// 参数：
	//   - username: 用户名
	// 返回：
	//   - *model.User: 用户实体（未找到返回nil）
	//   - error: 操作错误
	FindByUsername(username string) (*model.User, error)

	// FindByEmail 根据邮箱查询用户
	// 参数：
	//   - email: 邮箱地址
	// 返回：
	//   - *model.User: 用户实体（未找到返回nil）
	//   - error: 操作错误
	FindByEmail(email string) (*model.User, error)

	// Update 更新用户信息
	// 参数：
	//   - user: 用户实体（包含更新后的数据）
	// 返回：
	//   - error: 操作错误
	Update(user *model.User) error

	// Delete 删除用户（物理删除）
	// 参数：
	//   - id: 用户ID
	// 返回：
	//   - error: 操作错误
	Delete(id uint) error

	// List 分页查询用户列表
	// 参数：
	//   - page: 页码（从1开始）
	//   - pageSize: 每页数量
	// 返回：
	//   - []model.User: 用户列表
	//   - int64: 总数量
	//   - error: 操作错误
	List(page, pageSize int) ([]model.User, int64, error)
}
