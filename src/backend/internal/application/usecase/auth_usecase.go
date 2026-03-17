// Package usecase 定义应用层用例
// 编排业务流程，协调领域服务完成具体业务功能
package usecase

import (
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
)

// AuthUseCase 认证用例
// 处理用户注册、登录等认证相关业务流程
type AuthUseCase struct {
	userService *service.UserService // 用户领域服务
}

// NewAuthUseCase 创建认证用例实例
// 参数：
//   - userService: 用户领域服务
// 返回：
//   - *AuthUseCase: 认证用例实例
func NewAuthUseCase(userService *service.UserService) *AuthUseCase {
	return &AuthUseCase{
		userService: userService,
	}
}

// Register 用户注册
// 参数：
//   - req: 注册请求DTO
// 返回：
//   - *dto.UserResponse: 用户信息响应（脱敏）
//   - error: 操作错误
func (uc *AuthUseCase) Register(req *dto.RegisterRequest) (*dto.UserResponse, error) {
	// 1. 构建用户实体
	user := &model.User{
		Username: req.Username,
		Password: req.Password, // 明文密码，由Service层加密
		RealName: req.RealName,
		Phone:    req.Phone,
		Email:    req.Email,
		Role:     "user", // 默认角色
	}

	// 2. 调用Service创建用户（包含密码加密和验证）
	if err := uc.userService.CreateUser(user); err != nil {
		return nil, err
	}

	// 3. 转换为响应DTO（脱敏）
	return uc.toUserResponse(user), nil
}

// Login 用户登录
// 参数：
//   - req: 登录请求DTO
// 返回：
//   - *dto.LoginResponse: 登录响应（包含用户信息）
//   - error: 操作错误
func (uc *AuthUseCase) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 1. 验证用户名和密码
	user, err := uc.userService.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	// 2. 构建登录响应
	return &dto.LoginResponse{
		User:    uc.toUserResponse(user),
		Message: "登录成功",
	}, nil
}

// toUserResponse 将User实体转换为UserResponse DTO（脱敏）
// 参数：
//   - user: 用户实体
// 返回：
//   - *dto.UserResponse: 用户响应DTO
func (uc *AuthUseCase) toUserResponse(user *model.User) *dto.UserResponse {
	return &dto.UserResponse{
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
