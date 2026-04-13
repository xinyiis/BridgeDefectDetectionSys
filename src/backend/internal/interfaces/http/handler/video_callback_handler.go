package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/response"
)

// VideoCallbackHandler 处理算法端视频异步回调。
type VideoCallbackHandler struct {
	videoUseCase *usecase.VideoDetectionUseCase
}

// NewVideoCallbackHandler 创建视频回调处理器。
func NewVideoCallbackHandler(videoUseCase *usecase.VideoDetectionUseCase) *VideoCallbackHandler {
	return &VideoCallbackHandler{videoUseCase: videoUseCase}
}

// TaskQueued 处理任务入队回调。
func (h *VideoCallbackHandler) TaskQueued(c *gin.Context) {
	var req dto.AlgoVideoTaskQueuedCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.videoUseCase.HandleAlgoTaskQueued(&req)
	if err != nil {
		h.handleCallbackError(c, err)
		return
	}

	response.Success(c, result)
}

// TaskStarted 处理任务开始回调。
func (h *VideoCallbackHandler) TaskStarted(c *gin.Context) {
	var req dto.AlgoVideoTaskStartedCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.videoUseCase.HandleAlgoTaskStarted(&req)
	if err != nil {
		h.handleCallbackError(c, err)
		return
	}

	response.Success(c, result)
}

// FrameResult 处理帧结果回调。
func (h *VideoCallbackHandler) FrameResult(c *gin.Context) {
	var req dto.AlgoVideoFrameResultCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.videoUseCase.HandleAlgoFrameResult(&req)
	if err != nil {
		h.handleCallbackError(c, err)
		return
	}

	response.Success(c, result)
}

// TaskCompleted 处理任务完成回调。
func (h *VideoCallbackHandler) TaskCompleted(c *gin.Context) {
	var req dto.AlgoVideoTaskCompletedCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.videoUseCase.HandleAlgoTaskCompleted(&req)
	if err != nil {
		h.handleCallbackError(c, err)
		return
	}

	response.Success(c, result)
}

// TaskFailed 处理任务失败回调。
func (h *VideoCallbackHandler) TaskFailed(c *gin.Context) {
	var req dto.AlgoVideoTaskFailedCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.videoUseCase.HandleAlgoTaskFailed(&req)
	if err != nil {
		h.handleCallbackError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *VideoCallbackHandler) handleCallbackError(c *gin.Context, err error) {
	message := err.Error()
	switch {
	case strings.Contains(message, "不存在"):
		response.NotFound(c, message)
	case strings.Contains(message, "参数"):
		response.BadRequest(c, message)
	default:
		response.InternalErrorWithDetail(c, message)
	}
}
