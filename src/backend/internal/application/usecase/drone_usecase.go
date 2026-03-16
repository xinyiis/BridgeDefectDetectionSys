// Package usecase 定义应用层用例
// 编排业务流程，协调领域服务完成具体业务功能
package usecase

import (
	"errors"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
)

// DroneUseCase 无人机用例
// 处理无人机CRUD相关业务流程
type DroneUseCase struct {
	droneService *service.DroneService // 无人机领域服务
}

// NewDroneUseCase 创建无人机用例实例
// 参数：
//   - droneService: 无人机领域服务
//
// 返回：
//   - *DroneUseCase: 无人机用例实例
func NewDroneUseCase(droneService *service.DroneService) *DroneUseCase {
	return &DroneUseCase{
		droneService: droneService,
	}
}

// CreateDrone 创建无人机
// 参数：
//   - req: 创建无人机请求DTO
//
// 返回：
//   - *dto.DroneResponse: 无人机信息响应
//   - error: 操作错误
func (uc *DroneUseCase) CreateDrone(req *dto.CreateDroneRequest) (*dto.DroneResponse, error) {
	// 1. 构建无人机实体
	drone := &model.Drone{
		Name:      req.Name,
		Model:     req.Model,
		StreamURL: req.StreamURL,
		UserID:    req.UserID,
	}

	// 2. 调用Service创建无人机
	if err := uc.droneService.CreateDrone(drone); err != nil {
		return nil, err
	}

	// 3. 转换为响应DTO
	return uc.toDroneResponse(drone), nil
}

// GetDrone 获取无人机详情
// 参数：
//   - id: 无人机ID
//
// 返回：
//   - *dto.DroneResponse: 无人机信息响应
//   - error: 操作错误
func (uc *DroneUseCase) GetDrone(id uint) (*dto.DroneResponse, error) {
	drone, err := uc.droneService.GetByID(id)
	if err != nil {
		return nil, err
	}
	if drone == nil {
		return nil, errors.New("无人机不存在")
	}

	return uc.toDroneResponse(drone), nil
}

// ListDrones 获取无人机列表
// 参数：
//   - currentUser: 当前用户
//   - page: 页码
//   - pageSize: 每页数量
//
// 返回：
//   - *dto.DroneListResponse: 无人机列表响应
//   - error: 操作错误
func (uc *DroneUseCase) ListDrones(currentUser *model.User, page, pageSize int) (*dto.DroneListResponse, error) {
	// 1. 查询无人机列表
	drones, total, err := uc.droneService.ListDrones(currentUser, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 2. 转换为响应DTO
	droneResponses := make([]dto.DroneResponse, len(drones))
	for i, drone := range drones {
		droneResponses[i] = *uc.toDroneResponse(&drone)
	}

	return &dto.DroneListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     droneResponses,
	}, nil
}

// UpdateDrone 更新无人机信息
// 参数：
//   - id: 无人机ID
//   - req: 更新无人机请求DTO
//
// 返回：
//   - *dto.DroneResponse: 更新后的无人机信息
//   - error: 操作错误
func (uc *DroneUseCase) UpdateDrone(id uint, req *dto.UpdateDroneRequest) (*dto.DroneResponse, error) {
	// 1. 获取无人机
	drone, err := uc.droneService.GetByID(id)
	if err != nil {
		return nil, err
	}
	if drone == nil {
		return nil, errors.New("无人机不存在")
	}

	// 2. 更新字段（只更新非空字段）
	if req.Name != "" {
		drone.Name = req.Name
	}
	if req.Model != "" {
		drone.Model = req.Model
	}
	if req.StreamURL != "" {
		drone.StreamURL = req.StreamURL
	}

	// 3. 调用Service更新
	if err := uc.droneService.UpdateDrone(drone); err != nil {
		return nil, err
	}

	// 4. 返回更新后的无人机信息
	return uc.toDroneResponse(drone), nil
}

// DeleteDrone 删除无人机
// 参数：
//   - id: 无人机ID
//   - currentUser: 当前用户
//
// 返回：
//   - error: 操作错误
func (uc *DroneUseCase) DeleteDrone(id uint, currentUser *model.User) error {
	return uc.droneService.DeleteDrone(id, currentUser)
}

// toDroneResponse 转换为响应DTO
func (uc *DroneUseCase) toDroneResponse(drone *model.Drone) *dto.DroneResponse {
	return &dto.DroneResponse{
		ID:        drone.ID,
		Name:      drone.Name,
		Model:     drone.Model,
		StreamURL: drone.StreamURL,
		UserID:    drone.UserID,
		CreatedAt: drone.CreatedAt,
		UpdatedAt: drone.UpdatedAt,
	}
}
