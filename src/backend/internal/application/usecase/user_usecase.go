// Package usecase 定义应用层用例
package usecase

import (
	"errors"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
)

// UserUseCase 用户管理用例
// 处理用户信息查询、更新、删除等业务流程
type UserUseCase struct {
	userService *service.UserService // 用户领域服务
}

// NewUserUseCase 创建用户管理用例实例
// 参数：
//   - userService: 用户领域服务
// 返回：
//   - *UserUseCase: 用户管理用例实例
func NewUserUseCase(userService *service.UserService) *UserUseCase {
	return &UserUseCase{
		userService: userService,
	}
}

// GetUserInfo 获取用户信息
// 参数：
//   - userID: 用户ID
// 返回：
//   - *dto.UserResponse: 用户信息响应
//   - error: 操作错误
func (uc *UserUseCase) GetUserInfo(userID uint) (*dto.UserResponse, error) {
	user, err := uc.userService.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		RealName:  user.RealName,
		Phone:     user.Phone,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// UpdateUserInfo 更新用户信息
// 参数：
//   - userID: 用户ID
//   - req: 更新请求DTO
// 返回：
//   - *dto.UserResponse: 更新后的用户信息
//   - error: 操作错误
func (uc *UserUseCase) UpdateUserInfo(userID uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// 1. 获取用户
	user, err := uc.userService.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 2. 更新字段（只更新非空字段）
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	// 3. 调用Service更新（包含密码加密）
	if err := uc.userService.UpdateUser(user, req.Password); err != nil {
		return nil, err
	}

	// 4. 返回更新后的用户信息
	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		RealName:  user.RealName,
		Phone:     user.Phone,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// DeleteUser 删除用户
// 参数：
//   - userID: 用户ID
// 返回：
//   - error: 操作错误
func (uc *UserUseCase) DeleteUser(userID uint) error {
	return uc.userService.DeleteUser(userID)
}

// ListUsers 获取用户列表
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - *dto.UserListResponse: 用户列表响应
//   - error: 操作错误
func (uc *UserUseCase) ListUsers(page, pageSize int) (*dto.UserListResponse, error) {
	// 1. 参数验证
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 2. 查询用户列表
	users, total, err := uc.userService.ListUsers(page, pageSize)
	if err != nil {
		return nil, err
	}

	// 3. 转换为响应DTO
	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = dto.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			RealName:  user.RealName,
			Phone:     user.Phone,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	return &dto.UserListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Users:    userResponses,
	}, nil
}

// PromoteToAdmin 提升用户为管理员
// 参数：
//   - req: 提升请求DTO
// 返回：
//   - *dto.UserResponse: 更新后的用户信息
//   - error: 操作错误
func (uc *UserUseCase) PromoteToAdmin(req *dto.PromoteUserRequest) (*dto.UserResponse, error) {
	// 1. 提升为管理员
	if err := uc.userService.PromoteToAdmin(req.UserID); err != nil {
		return nil, err
	}

	// 2. 获取更新后的用户信息
	return uc.GetUserInfo(req.UserID)
}
