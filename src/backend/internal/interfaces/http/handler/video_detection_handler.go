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

// VideoDetectionHandler 视频检测 HTTP 处理器。
type VideoDetectionHandler struct {
	videoUseCase *usecase.VideoDetectionUseCase
}

// NewVideoDetectionHandler 创建视频检测处理器。
func NewVideoDetectionHandler(videoUseCase *usecase.VideoDetectionUseCase) *VideoDetectionHandler {
	return &VideoDetectionHandler{videoUseCase: videoUseCase}
}

// UploadVideo 上传视频并创建任务。
func (h *VideoDetectionHandler) UploadVideo(c *gin.Context) {
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	var req dto.VideoUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.Video == nil {
		response.BadRequest(c, "请上传视频文件")
		return
	}
	if req.EnableSegment {
		response.BadRequest(c, "视频检测链路仅支持 detect-only，暂不支持 enable_segment")
		return
	}

	ext := strings.ToLower(filepath.Ext(req.Video.Filename))
	switch ext {
	case ".mp4", ".avi", ".mov", ".mkv":
	default:
		response.BadRequest(c, "不支持的视频格式，仅支持 mp4/avi/mov/mkv")
		return
	}

	maxSize := int64(200 * 1024 * 1024)
	if req.Video.Size > maxSize {
		response.BadRequest(c, "视频大小不能超过200MB")
		return
	}

	result, err := h.videoUseCase.UploadVideo(&req, user)
	if err != nil {
		h.handleVideoError(c, err)
		return
	}

	response.SuccessWithMessage(c, "视频上传成功", result)
}

// StartTask 创建视频分析会话。
func (h *VideoDetectionHandler) StartTask(c *gin.Context) {
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	result, err := h.videoUseCase.StartTask(c.Param("task_id"), user)
	if err != nil {
		h.handleVideoError(c, err)
		return
	}

	response.SuccessWithMessage(c, "分析会话创建成功", result)
}

// GetTask 获取视频任务状态。
func (h *VideoDetectionHandler) GetTask(c *gin.Context) {
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	result, err := h.videoUseCase.GetTask(c.Param("task_id"), user)
	if err != nil {
		h.handleVideoError(c, err)
		return
	}

	response.Success(c, result)
}

// CancelTask 取消视频分析任务。
func (h *VideoDetectionHandler) CancelTask(c *gin.Context) {
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	result, err := h.videoUseCase.CancelTask(c.Param("task_id"), user)
	if err != nil {
		h.handleVideoError(c, err)
		return
	}

	response.SuccessWithMessage(c, "任务已取消", result)
}

func (h *VideoDetectionHandler) handleVideoError(c *gin.Context, err error) {
	message := err.Error()
	switch {
	case strings.Contains(message, "不存在"):
		response.NotFound(c, strings.TrimSuffix(message, "不存在"))
	case strings.Contains(message, "无权"):
		response.ForbiddenWithMessage(c, message)
	case strings.Contains(message, "参数"), strings.Contains(message, "状态"), strings.Contains(message, "会话"), strings.Contains(message, "取消"):
		response.BadRequest(c, message)
	default:
		response.InternalErrorWithDetail(c, message)
	}
}
