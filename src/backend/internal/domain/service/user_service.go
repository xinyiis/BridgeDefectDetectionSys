// Package service 定义领域服务层
// 包含核心业务逻辑和领域规则
package service

import (
	"errors"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

// UserService 用户领域服务
// 处理用户相关的核心业务逻辑
type UserService struct {
	userRepo repository.UserRepository // 用户数据仓库
}

// NewUserService 创建用户服务实例
// 参数：
//   - userRepo: 用户Repository接口
// 返回：
//   - *UserService: 用户服务实例
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// HashPassword 使用bcrypt加密密码
// 参数：
//   - password: 明文密码
// 返回：
//   - string: 加密后的密码
//   - error: 加密错误
func (s *UserService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword 验证密码是否正确
// 参数：
//   - hashedPassword: 加密后的密码（数据库存储）
//   - password: 用户输入的明文密码
// 返回：
//   - bool: true表示密码正确，false表示密码错误
func (s *UserService) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// CreateUser 创建新用户（包含密码加密）
// 参数：
//   - user: 用户实体（密码为明文）
// 返回：
//   - error: 操作错误
func (s *UserService) CreateUser(user *model.User) error {
	// 1. 检查用户名是否已存在
	existingUser, err := s.userRepo.FindByUsername(user.Username)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("用户名已存在")
	}

	// 2. 检查邮箱是否已存在
	existingUser, err = s.userRepo.FindByEmail(user.Email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("邮箱已被注册")
	}

	// 3. 加密密码
	hashedPassword, err := s.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	// 4. 设置默认角色
	if user.Role == "" {
		user.Role = "user"
	}

	// 5. 创建用户
	return s.userRepo.Create(user)
}

// AuthenticateUser 验证用户登录
// 参数：
//   - username: 用户名
//   - password: 明文密码
// 返回：
//   - *model.User: 用户实体（验证成功）
//   - error: 验证失败错误
func (s *UserService) AuthenticateUser(username, password string) (*model.User, error) {
	// 1. 查找用户
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 2. 验证密码
	if !s.VerifyPassword(user.Password, password) {
		return nil, errors.New("用户名或密码错误")
	}

	return user, nil
}

// UpdateUser 更新用户信息
// 参数：
//   - user: 用户实体（包含更新后的数据）
//   - newPassword: 新密码（如果为空则不更新密码）
// 返回：
//   - error: 操作错误
func (s *UserService) UpdateUser(user *model.User, newPassword string) error {
	// 1. 如果提供了新密码，则加密
	if newPassword != "" {
		hashedPassword, err := s.HashPassword(newPassword)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	}

	// 2. 更新用户
	return s.userRepo.Update(user)
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	return s.userRepo.Delete(id)
}

// ListUsers 分页获取用户列表
func (s *UserService) ListUsers(page, pageSize int) ([]model.User, int64, error) {
	return s.userRepo.List(page, pageSize)
}

// PromoteToAdmin 提升用户为管理员
// 参数：
//   - userID: 用户ID
// 返回：
//   - error: 操作错误
func (s *UserService) PromoteToAdmin(userID uint) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("用户不存在")
	}

	user.Role = "admin"
	return s.userRepo.Update(user)
}
