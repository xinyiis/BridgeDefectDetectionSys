// Package handler 定义HTTP请求处理器
// 负责接收HTTP请求，调用UseCase，返回HTTP响应
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/response"
)

// DroneHandler 无人机HTTP处理器
type DroneHandler struct {
	droneUseCase *usecase.DroneUseCase // 无人机用例
}

// NewDroneHandler 创建无人机Handler实例
// 参数：
//   - droneUseCase: 无人机用例
//
// 返回：
//   - *DroneHandler: 无人机Handler实例
func NewDroneHandler(droneUseCase *usecase.DroneUseCase) *DroneHandler {
	return &DroneHandler{
		droneUseCase: droneUseCase,
	}
}

// CreateDrone 创建无人机
// @Summary 创建无人机
// @Tags 无人机管理
// @Accept json
// @Produce json
// @Param drone body dto.CreateDroneRequest true "无人机信息"
// @Success 200 {object} response.Response
// @Router /api/v1/drones [post]
func (h *DroneHandler) CreateDrone(c *gin.Context) {
	// 1. 绑定JSON数据
	var req dto.CreateDroneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 2. 获取当前用户
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)
	req.UserID = user.ID

	// 3. 创建无人机
	drone, err := h.droneUseCase.CreateDrone(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 4. 返回成功响应
	response.Success(c, gin.H{
		"id":    drone.ID,
		"name":  drone.Name,
		"model": drone.Model,
	})
}

// GetDrone 获取无人机详情
// @Summary 获取无人机详情
// @Tags 无人机管理
// @Produce json
// @Param id path int true "无人机ID"
// @Success 200 {object} response.Response
// @Router /api/v1/drones/{id} [get]
func (h *DroneHandler) GetDrone(c *gin.Context) {
	// 1. 从上下文获取无人机（中间件已查询并验证权限）
	droneInterface, exists := c.Get("drone")
	if exists {
		drone := droneInterface.(*model.Drone)
		// 转换为响应DTO
		droneResp := dto.DroneResponse{
			ID:        drone.ID,
			Name:      drone.Name,
			Model:     drone.Model,
			StreamURL: drone.StreamURL,
			UserID:    drone.UserID,
			CreatedAt: drone.CreatedAt,
			UpdatedAt: drone.UpdatedAt,
		}
		response.Success(c, droneResp)
		return
	}

	// 2. 如果上下文中没有，则重新查询
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的无人机ID")
		return
	}

	drone, err := h.droneUseCase.GetDrone(uint(id))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, drone)
}

// ListDrones 获取无人机列表
// @Summary 获取无人机列表
// @Tags 无人机管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/drones [get]
func (h *DroneHandler) ListDrones(c *gin.Context) {
	// 1. 获取当前用户
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	// 2. 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 3. 查询无人机列表
	drones, err := h.droneUseCase.ListDrones(user, page, pageSize)
	if err != nil {
		response.InternalErrorWithDetail(c, "查询失败")
		return
	}

	// 4. 返回响应
	response.Success(c, drones)
}

// UpdateDrone 更新无人机信息
// @Summary 更新无人机信息
// @Tags 无人机管理
// @Accept json
// @Produce json
// @Param id path int true "无人机ID"
// @Param drone body dto.UpdateDroneRequest true "无人机信息"
// @Success 200 {object} response.Response
// @Router /api/v1/drones/{id} [put]
func (h *DroneHandler) UpdateDrone(c *gin.Context) {
	// 1. 解析ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的无人机ID")
		return
	}

	// 2. 绑定JSON数据
	var req dto.UpdateDroneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 3. 更新无人机
	drone, err := h.droneUseCase.UpdateDrone(uint(id), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 4. 返回成功响应
	response.Success(c, gin.H{
		"id":   drone.ID,
		"name": drone.Name,
	})
}

// DeleteDrone 删除无人机
// @Summary 删除无人机
// @Tags 无人机管理
// @Produce json
// @Param id path int true "无人机ID"
// @Success 200 {object} response.Response
// @Router /api/v1/drones/{id} [delete]
func (h *DroneHandler) DeleteDrone(c *gin.Context) {
	// 1. 解析ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的无人机ID")
		return
	}

	// 2. 获取当前用户
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	// 3. 删除无人机
	if err := h.droneUseCase.DeleteDrone(uint(id), user); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 4. 返回成功响应
	response.Success(c, gin.H{
		"message": "删除成功",
	})
}
