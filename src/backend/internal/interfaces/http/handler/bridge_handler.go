// Package handler 定义HTTP请求处理器
// 负责接收HTTP请求，调用UseCase，返回HTTP响应
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/response"
)

// BridgeHandler 桥梁HTTP处理器
type BridgeHandler struct {
	bridgeUseCase *usecase.BridgeUseCase // 桥梁用例
	fileService   service.FileService    // 文件服务
}

// NewBridgeHandler 创建桥梁Handler实例
// 参数：
//   - bridgeUseCase: 桥梁用例
//   - fileService: 文件服务
// 返回：
//   - *BridgeHandler: 桥梁Handler实例
func NewBridgeHandler(bridgeUseCase *usecase.BridgeUseCase, fileService service.FileService) *BridgeHandler {
	return &BridgeHandler{
		bridgeUseCase: bridgeUseCase,
		fileService:   fileService,
	}
}

// CreateBridge 创建桥梁
// @Summary 创建桥梁
// @Tags 桥梁管理
// @Accept multipart/form-data
// @Produce json
// @Param bridge_name formData string true "桥梁名称"
// @Param bridge_code formData string true "桥梁编号"
// @Param model_3d_file formData file false "3D模型文件"
// @Success 200 {object} response.Response
// @Router /api/bridges [post]
func (h *BridgeHandler) CreateBridge(c *gin.Context) {
	// 1. 绑定表单数据
	var req dto.CreateBridgeRequest
	if err := c.ShouldBind(&req); err != nil {
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

	// 3. 处理文件上传（可选）
	file, err := c.FormFile("model_3d_file")
	var model3DPath string

	if err == nil && file != nil {
		// 3.1 验证文件格式
		allowedFormats := []string{".obj", ".fbx", ".gltf", ".glb"}
		if err := h.fileService.ValidateFileFormat(file, allowedFormats); err != nil {
			response.BadRequest(c, err.Error())
			return
		}

		// 3.2 验证文件大小（50MB）
		if err := h.fileService.ValidateFileSize(file, 50*1024*1024); err != nil {
			response.BadRequest(c, err.Error())
			return
		}

		// 3.3 保存文件
		model3DPath, err = h.fileService.SaveUploadedFile(file, "models")
		if err != nil {
			response.InternalErrorWithDetail(c, "文件上传失败")
			return
		}
	}

	// 4. 创建桥梁
	req.Model3DPath = model3DPath
	req.UserID = user.ID
	bridge, err := h.bridgeUseCase.CreateBridge(&req)

	if err != nil {
		// 🔑 关键：失败时删除已上传的文件
		if model3DPath != "" {
			h.fileService.DeleteFile(model3DPath)
		}
		response.BadRequest(c, err.Error())
		return
	}

	// 5. 返回结果
	response.Success(c, gin.H{
		"bridge_id":   bridge.ID,
		"bridge_name": bridge.BridgeName,
		"bridge_code": bridge.BridgeCode,
	})
}

// GetBridge 获取桥梁详情
// @Summary 获取桥梁详情
// @Tags 桥梁管理
// @Produce json
// @Param id path int true "桥梁ID"
// @Success 200 {object} response.Response
// @Router /api/bridges/{id} [get]
func (h *BridgeHandler) GetBridge(c *gin.Context) {
	// 1. 从上下文获取桥梁（中间件已查询并验证权限）
	bridgeInterface, exists := c.Get("bridge")
	if exists {
		bridge := bridgeInterface.(*model.Bridge)
		// 转换为响应DTO
		bridgeResp := dto.BridgeResponse{
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
		response.Success(c, bridgeResp)
		return
	}

	// 2. 如果上下文中没有，则重新查询
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的桥梁ID")
		return
	}

	bridge, err := h.bridgeUseCase.GetBridge(uint(id))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, bridge)
}

// ListBridges 获取桥梁列表
// @Summary 获取桥梁列表
// @Tags 桥梁管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "状态过滤"
// @Success 200 {object} response.Response
// @Router /api/bridges [get]
func (h *BridgeHandler) ListBridges(c *gin.Context) {
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
	status := c.Query("status")

	// 3. 查询桥梁列表
	bridges, err := h.bridgeUseCase.ListBridges(user, page, pageSize, status)
	if err != nil {
		response.InternalErrorWithDetail(c, "查询失败")
		return
	}

	// 4. 返回结果
	response.Success(c, bridges)
}

// UpdateBridge 更新桥梁信息
// @Summary 更新桥梁信息
// @Tags 桥梁管理
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "桥梁ID"
// @Success 200 {object} response.Response
// @Router /api/bridges/{id} [put]
func (h *BridgeHandler) UpdateBridge(c *gin.Context) {
	// 1. 获取桥梁ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的桥梁ID")
		return
	}

	// 2. 绑定表单数据
	var req dto.UpdateBridgeRequest
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 3. 处理文件上传（可选）
	file, err := c.FormFile("model_3d_file")
	if err == nil && file != nil {
		// 验证文件
		allowedFormats := []string{".obj", ".fbx", ".gltf", ".glb"}
		if err := h.fileService.ValidateFileFormat(file, allowedFormats); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		if err := h.fileService.ValidateFileSize(file, 50*1024*1024); err != nil {
			response.BadRequest(c, err.Error())
			return
		}

		// 保存新文件
		model3DPath, err := h.fileService.SaveUploadedFile(file, "models")
		if err != nil {
			response.InternalErrorWithDetail(c, "文件上传失败")
			return
		}

		// TODO: 删除旧文件（需要先查询桥梁获取旧文件路径）
		req.Model3DPath = model3DPath
	}

	// 4. 更新桥梁
	bridge, err := h.bridgeUseCase.UpdateBridge(uint(id), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 5. 返回结果
	response.Success(c, bridge)
}

// DeleteBridge 删除桥梁
// @Summary 删除桥梁
// @Tags 桥梁管理
// @Produce json
// @Param id path int true "桥梁ID"
// @Success 200 {object} response.Response
// @Router /api/bridges/{id} [delete]
func (h *BridgeHandler) DeleteBridge(c *gin.Context) {
	// 1. 获取桥梁ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的桥梁ID")
		return
	}

	// 2. 获取当前用户
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	// 3. 删除桥梁
	if err := h.bridgeUseCase.DeleteBridge(uint(id), user); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 4. 返回结果
	response.Success(c, gin.H{
		"message": "删除成功",
	})
}
