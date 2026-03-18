// Package handler 定义HTTP请求处理器
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/response"
)

// DefectHandler 缺陷HTTP处理器
type DefectHandler struct {
	defectUseCase *usecase.DefectUseCase // 缺陷用例
}

// NewDefectHandler 创建缺陷Handler实例
// 参数：
//   - defectUseCase: 缺陷用例
// 返回：
//   - *DefectHandler: 缺陷Handler实例
func NewDefectHandler(defectUseCase *usecase.DefectUseCase) *DefectHandler {
	return &DefectHandler{
		defectUseCase: defectUseCase,
	}
}

// ListDefects 缺陷列表
// @Summary 获取缺陷列表
// @Tags 缺陷管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param bridge_id query int false "桥梁ID过滤"
// @Param defect_type query string false "缺陷类型过滤"
// @Param start_date query string false "开始时间(YYYY-MM-DD)"
// @Param end_date query string false "结束时间(YYYY-MM-DD)"
// @Success 200 {object} response.Response
// @Router /api/v1/defects [get]
func (h *DefectHandler) ListDefects(c *gin.Context) {
	// 1. 获取当前用户
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	// 2. 绑定查询参数
	var req dto.DefectListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 3. 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	// 4. 查询列表（UseCase会自动进行权限过滤）
	result, err := h.defectUseCase.ListDefects(&req, user)
	if err != nil {
		response.InternalErrorWithDetail(c, err.Error())
		return
	}

	// 5. 返回结果
	response.Success(c, result)
}

// GetDefect 缺陷详情
// @Summary 获取缺陷详情
// @Tags 缺陷管理
// @Produce json
// @Param id path int true "缺陷ID"
// @Success 200 {object} response.Response
// @Router /api/v1/defects/{id} [get]
func (h *DefectHandler) GetDefect(c *gin.Context) {
	// 1. 优先从Context获取（中间件已查询并验证权限）
	defectInterface, exists := c.Get("defect")
	if exists {
		defect := defectInterface.(*model.Defect)
		// 构建详情响应
		detailResp := &dto.DefectDetailResponse{
			DefectDTO: dto.DefectDTO{
				ID:         defect.ID,
				BridgeID:   defect.BridgeID,
				DefectType: defect.DefectType,
				ImagePath:  defect.ImagePath,
				ResultPath: defect.ResultPath,
				BBox:       defect.BBox,
				Length:     defect.Length,
				Width:      defect.Width,
				Area:       defect.Area,
				Confidence: defect.Confidence,
				DetectedAt: defect.DetectedAt,
				CreatedAt:  defect.CreatedAt,
			},
		}
		if defect.Bridge != nil {
			detailResp.Bridge = &dto.BridgeSimpleInfo{
				ID:         defect.Bridge.ID,
				BridgeName: defect.Bridge.BridgeName,
				BridgeCode: defect.Bridge.BridgeCode,
			}
		}
		response.Success(c, detailResp)
		return
	}

	// 2. Fallback：解析ID并查询
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的缺陷ID")
		return
	}

	// 3. 查询缺陷
	result, err := h.defectUseCase.GetDefect(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	// 4. 返回结果
	response.Success(c, result)
}

// DeleteDefect 删除缺陷
// @Summary 删除缺陷
// @Tags 缺陷管理
// @Produce json
// @Param id path int true "缺陷ID"
// @Success 200 {object} response.Response
// @Router /api/v1/defects/{id} [delete]
func (h *DefectHandler) DeleteDefect(c *gin.Context) {
	// 1. 解析ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的缺陷ID")
		return
	}

	// 2. 删除缺陷（中间件已验证权限）
	if err := h.defectUseCase.DeleteDefect(uint(id)); err != nil {
		response.InternalErrorWithDetail(c, err.Error())
		return
	}

	// 3. 返回成功
	response.Success(c, gin.H{
		"message": "删除成功",
	})
}
