// Package usecase 定义应用层用例
// 编排业务流程，协调领域服务完成具体业务功能
package usecase

import (
	"errors"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
)

// BridgeUseCase 桥梁用例
// 处理桥梁CRUD相关业务流程
type BridgeUseCase struct {
	bridgeService *service.BridgeService // 桥梁领域服务
}

// NewBridgeUseCase 创建桥梁用例实例
// 参数：
//   - bridgeService: 桥梁领域服务
// 返回：
//   - *BridgeUseCase: 桥梁用例实例
func NewBridgeUseCase(bridgeService *service.BridgeService) *BridgeUseCase {
	return &BridgeUseCase{
		bridgeService: bridgeService,
	}
}

// CreateBridge 创建桥梁
// 参数：
//   - req: 创建桥梁请求DTO
// 返回：
//   - *dto.BridgeResponse: 桥梁信息响应
//   - error: 操作错误
func (uc *BridgeUseCase) CreateBridge(req *dto.CreateBridgeRequest) (*dto.BridgeResponse, error) {
	// 1. 构建桥梁实体
	bridge := &model.Bridge{
		BridgeName:  req.BridgeName,
		BridgeCode:  req.BridgeCode,
		Address:     req.Address,
		Longitude:   req.Longitude,
		Latitude:    req.Latitude,
		BridgeType:  req.BridgeType,
		BuildYear:   req.BuildYear,
		Length:      req.Length,
		Width:       req.Width,
		Model3DPath: req.Model3DPath,
		Remark:      req.Remark,
		UserID:      req.UserID,
		Status:      "正常", // 默认状态
	}

	// 2. 调用Service创建桥梁
	if err := uc.bridgeService.CreateBridge(bridge); err != nil {
		return nil, err
	}

	// 3. 转换为响应DTO
	return uc.toBridgeResponse(bridge), nil
}

// GetBridge 获取桥梁详情
// 参数：
//   - id: 桥梁ID
// 返回：
//   - *dto.BridgeResponse: 桥梁信息响应
//   - error: 操作错误
func (uc *BridgeUseCase) GetBridge(id uint) (*dto.BridgeResponse, error) {
	bridge, err := uc.bridgeService.GetByID(id)
	if err != nil {
		return nil, err
	}
	if bridge == nil {
		return nil, errors.New("桥梁不存在")
	}

	return uc.toBridgeResponse(bridge), nil
}

// ListBridges 获取桥梁列表
// 参数：
//   - currentUser: 当前用户
//   - page: 页码
//   - pageSize: 每页数量
//   - status: 状态过滤
// 返回：
//   - *dto.BridgeListResponse: 桥梁列表响应
//   - error: 操作错误
func (uc *BridgeUseCase) ListBridges(currentUser *model.User, page, pageSize int, status string) (*dto.BridgeListResponse, error) {
	// 1. 查询桥梁列表
	bridges, total, err := uc.bridgeService.ListBridges(currentUser, page, pageSize, status)
	if err != nil {
		return nil, err
	}

	// 2. 转换为响应DTO
	bridgeResponses := make([]dto.BridgeResponse, len(bridges))
	for i, bridge := range bridges {
		bridgeResponses[i] = *uc.toBridgeResponse(&bridge)
	}

	return &dto.BridgeListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     bridgeResponses,
	}, nil
}

// UpdateBridge 更新桥梁信息
// 参数：
//   - id: 桥梁ID
//   - req: 更新桥梁请求DTO
// 返回：
//   - *dto.BridgeResponse: 更新后的桥梁信息
//   - error: 操作错误
func (uc *BridgeUseCase) UpdateBridge(id uint, req *dto.UpdateBridgeRequest) (*dto.BridgeResponse, error) {
	// 1. 获取桥梁
	bridge, err := uc.bridgeService.GetByID(id)
	if err != nil {
		return nil, err
	}
	if bridge == nil {
		return nil, errors.New("桥梁不存在")
	}

	// 2. 更新字段（只更新非空字段）
	if req.BridgeName != "" {
		bridge.BridgeName = req.BridgeName
	}
	if req.Address != "" {
		bridge.Address = req.Address
	}
	if req.Longitude != 0 {
		bridge.Longitude = req.Longitude
	}
	if req.Latitude != 0 {
		bridge.Latitude = req.Latitude
	}
	if req.BridgeType != "" {
		bridge.BridgeType = req.BridgeType
	}
	if req.BuildYear != 0 {
		bridge.BuildYear = req.BuildYear
	}
	if req.Length != 0 {
		bridge.Length = req.Length
	}
	if req.Width != 0 {
		bridge.Width = req.Width
	}
	if req.Status != "" {
		bridge.Status = req.Status
	}
	if req.Remark != "" {
		bridge.Remark = req.Remark
	}
	if req.Model3DPath != "" {
		bridge.Model3DPath = req.Model3DPath
	}

	// 3. 调用Service更新
	if err := uc.bridgeService.UpdateBridge(bridge); err != nil {
		return nil, err
	}

	// 4. 返回更新后的桥梁信息
	return uc.toBridgeResponse(bridge), nil
}

// DeleteBridge 删除桥梁
// 参数：
//   - id: 桥梁ID
//   - currentUser: 当前用户
// 返回：
//   - error: 操作错误
func (uc *BridgeUseCase) DeleteBridge(id uint, currentUser *model.User) error {
	return uc.bridgeService.DeleteBridge(id, currentUser)
}

// toBridgeResponse 转换为响应DTO
func (uc *BridgeUseCase) toBridgeResponse(bridge *model.Bridge) *dto.BridgeResponse {
	return &dto.BridgeResponse{
		ID:          bridge.ID,
		BridgeName:  bridge.BridgeName,
		BridgeCode:  bridge.BridgeCode,
		Address:     bridge.Address,
		Longitude:   bridge.Longitude,
		Latitude:    bridge.Latitude,
		BridgeType:  bridge.BridgeType,
		BuildYear:   bridge.BuildYear,
		Length:      bridge.Length,
		Width:       bridge.Width,
		Status:      bridge.Status,
		Model3DPath: bridge.Model3DPath,
		Remark:      bridge.Remark,
		UserID:      bridge.UserID,
		CreatedAt:   bridge.CreatedAt,
		UpdatedAt:   bridge.UpdatedAt,
	}
}
