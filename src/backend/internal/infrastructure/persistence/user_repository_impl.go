// Package persistence 实现数据访问层
// 包含所有Repository接口的具体实现
package persistence

import (
	"errors"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"gorm.io/gorm"
)

// UserRepositoryImpl 用户Repository实现
// 实现 repository.UserRepository 接口
type UserRepositoryImpl struct {
	db *gorm.DB // GORM数据库连接
}

// NewUserRepository 创建用户Repository实例
// 参数：
//   - db: GORM数据库连接
// 返回：
//   - repository.UserRepository: 用户Repository接口
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &UserRepositoryImpl{db: db}
}

// Create 创建新用户
func (r *UserRepositoryImpl) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// FindByID 根据ID查询用户
func (r *UserRepositoryImpl) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到返回nil而不是错误
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查询用户
func (r *UserRepositoryImpl) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到返回nil而不是错误
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查询用户
func (r *UserRepositoryImpl) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到返回nil而不是错误
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *UserRepositoryImpl) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete 删除用户
func (r *UserRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

// List 分页查询用户列表
func (r *UserRepositoryImpl) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// 计算总数
	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
