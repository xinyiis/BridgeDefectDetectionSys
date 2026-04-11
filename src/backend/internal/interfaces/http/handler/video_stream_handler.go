package handler

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/response"
	"golang.org/x/net/websocket"
)

// VideoStreamHandler 视频分析流处理器。
type VideoStreamHandler struct {
	videoUseCase *usecase.VideoDetectionUseCase
}

// NewVideoStreamHandler 创建视频流处理器。
func NewVideoStreamHandler(videoUseCase *usecase.VideoDetectionUseCase) *VideoStreamHandler {
	return &VideoStreamHandler{videoUseCase: videoUseCase}
}

// Stream 建立 WebSocket 连接并触发分析任务。
func (h *VideoStreamHandler) Stream(c *gin.Context) {
	currentUser, exists := c.Get("current_user")
	if !exists {
		response.Unauthorized(c)
		return
	}
	user := currentUser.(*model.User)

	taskID := c.Query("task_id")
	sessionID := c.Query("session_id")
	if taskID == "" || sessionID == "" {
		response.BadRequest(c, "task_id 和 session_id 不能为空")
		return
	}

	websocket.Handler(func(conn *websocket.Conn) {
		defer conn.Close()

		var writeMu sync.Mutex
		send := func(message any) error {
			writeMu.Lock()
			defer writeMu.Unlock()
			return websocket.JSON.Send(conn, message)
		}

		if err := h.videoUseCase.StreamTask(taskID, sessionID, user, send); err != nil {
			_ = send(dto.VideoErrorMessage{
				Type:      "error",
				TaskID:    taskID,
				SessionID: sessionID,
				Message:   err.Error(),
			})
		}
	}).ServeHTTP(c.Writer, c.Request)
}
