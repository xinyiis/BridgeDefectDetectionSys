package repository

import "github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"

// VideoTaskRepository 视频任务仓储接口。
type VideoTaskRepository interface {
	Create(task *model.VideoAnalysisTask) error
	FindByTaskID(taskID string) (*model.VideoAnalysisTask, error)
	Update(task *model.VideoAnalysisTask) error
}

// VideoFrameTaskRepository 视频帧级请求仓储接口。
type VideoFrameTaskRepository interface {
	Create(frameTask *model.VideoFrameTaskRequest) error
	FindByRequestID(requestID string) (*model.VideoFrameTaskRequest, error)
	Update(frameTask *model.VideoFrameTaskRequest) error
}
