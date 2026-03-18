// Package handler 定义HTTP请求处理器
package handler

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/response"
)

// DetectionHandler 检测HTTP处理器
type DetectionHandler struct {
	detectionUseCase *usecase.DetectionUseCase // 检测用例
}

// NewDetectionHandler 创建检测Handler实例
// 参数：
//   - detectionUseCase: 检测用例
// 返回：
//   - *DetectionHandler: 检测Handler实例
func NewDetectionHandler(detectionUseCase *usecase.DetectionUseCase) *DetectionHandler {
	return &DetectionHandler{
		detectionUseCase: detectionUseCase,
	}
}

// UploadAndDetect 图片上传检测
// @Summary 上传图片进行缺陷检测
// @Tags 缺陷检测
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "图片文件"
// @Param bridge_id formData int true "桥梁ID"
// @Param model_name formData string true "模型名称"
// @Param pixel_ratio formData number true "像素实际系数"
// @Success 200 {object} response.Response
// @Router /api/v1/detection/upload [post]
func (h *DetectionHandler) UploadAndDetect(c *gin.Context) {
	// 1. 获取当前用户
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	// 2. 绑定请求参数
	var req dto.DetectionUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 3. 验证图片文件
	if req.Image == nil {
		response.BadRequest(c, "请上传图片文件")
		return
	}

	// 3.1 验证图片格式
	ext := strings.ToLower(filepath.Ext(req.Image.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".bmp" {
		response.BadRequest(c, "不支持的图片格式，仅支持 jpg/jpeg/png/bmp")
		return
	}

	// 3.2 验证图片大小（10MB）
	maxSize := int64(10 * 1024 * 1024)
	if req.Image.Size > maxSize {
		response.BadRequest(c, "图片大小不能超过10MB")
		return
	}

	// 4. 调用UseCase进行检测
	result, err := h.detectionUseCase.UploadAndDetect(&req, user)
	if err != nil {
		// 根据错误类型返回不同的HTTP状态码
		errMsg := err.Error()
		if strings.Contains(errMsg, "不存在") {
			response.NotFound(c, errMsg)
		} else if strings.Contains(errMsg, "无权访问") {
			response.Forbidden(c)
		} else {
			response.InternalErrorWithDetail(c, errMsg)
		}
		return
	}

	// 5. 返回检测结果
	response.Success(c, result)
}
