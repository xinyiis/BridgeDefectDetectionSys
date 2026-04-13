package dto

import (
	"encoding/json"
	"mime/multipart"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
)

// VideoUploadRequest 视频上传请求。
type VideoUploadRequest struct {
	Video         *multipart.FileHeader `form:"video" binding:"required"`
	BridgeID      uint                  `form:"bridge_id" binding:"required"`
	ModelName     string                `form:"model_name"`
	PixelRatio    float64               `form:"pixel_ratio" binding:"required,gt=0"`
	FPS           float64               `form:"fps" binding:"omitempty,gt=0,lte=25"`
	EnableSegment bool                  `form:"enable_segment"`
}

// VideoUploadResponse 视频上传响应。
type VideoUploadResponse struct {
	TaskID    string                `json:"task_id"`
	Status    model.VideoTaskStatus `json:"status"`
	VideoPath string                `json:"video_path"`
}

// VideoStartResponse 开始分析响应。
type VideoStartResponse struct {
	TaskID    string                `json:"task_id"`
	SessionID string                `json:"session_id"`
	Status    model.VideoTaskStatus `json:"status"`
	WSURL     string                `json:"ws_url"`
}

// VideoTaskResponse 视频任务状态响应。
type VideoTaskResponse struct {
	TaskID           string                `json:"task_id"`
	SessionID        string                `json:"session_id,omitempty"`
	BridgeID         uint                  `json:"bridge_id"`
	Status           model.VideoTaskStatus `json:"status"`
	ModelName        string                `json:"model_name"`
	PixelRatio       float64               `json:"pixel_ratio"`
	FPS              float64               `json:"fps"`
	EnableSegment    bool                  `json:"enable_segment"`
	VideoPath        string                `json:"video_path"`
	TotalFrames      int                   `json:"total_frames"`
	ProcessedFrames  int                   `json:"processed_frames"`
	ConfirmedDefects int                   `json:"confirmed_defects"`
	ErrorMessage     string                `json:"error_message,omitempty"`
	CanCancel        bool                  `json:"can_cancel"`
	CanRetry         bool                  `json:"can_retry"`
	CreatedAt        time.Time             `json:"created_at"`
	StartedAt        *time.Time            `json:"started_at,omitempty"`
	CompletedAt      *time.Time            `json:"completed_at,omitempty"`
	CanceledAt       *time.Time            `json:"canceled_at,omitempty"`
}

// VideoBBox 视频检测框。
type VideoBBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// JSON 返回 JSON 字符串。
func (b VideoBBox) JSON() string {
	data, _ := json.Marshal(b)
	return string(data)
}

// VideoFrameDefect 单帧缺陷。
type VideoFrameDefect struct {
	DefectType        string    `json:"defect_type"`
	BBox              VideoBBox `json:"bbox"`
	Length            float64   `json:"length"`
	Width             float64   `json:"width"`
	Area              float64   `json:"area"`
	MeasurementSource string    `json:"measurement_source"`
	MeasurementStatus string    `json:"measurement_status"`
}

// VideoConnectedMessage WebSocket 建连消息。
type VideoConnectedMessage struct {
	Type      string `json:"type"`
	TaskID    string `json:"task_id"`
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// VideoFrameMessage 单帧推送消息。
type VideoFrameMessage struct {
	Type                string             `json:"type"`
	TaskID              string             `json:"task_id"`
	SessionID           string             `json:"session_id"`
	FrameNo             int                `json:"frame_no"`
	ProcessedFrames     int                `json:"processed_frames"`
	TotalFrames         int                `json:"total_frames"`
	Progress            int                `json:"progress"`
	TimestampMS         int                `json:"timestamp_ms"`
	ImageBase64         string             `json:"image_base64"`
	PersistedResultPath string             `json:"persisted_result_path,omitempty"`
	Defects             []VideoFrameDefect `json:"defects"`
	DefectCount         int                `json:"defect_count"`
	QueueLatencyMS      int                `json:"queue_latency_ms,omitempty"`
	DetectLatencyMS     int                `json:"detect_latency_ms,omitempty"`
}

// VideoProgressMessage 进度消息。
type VideoProgressMessage struct {
	Type            string `json:"type"`
	TaskID          string `json:"task_id"`
	SessionID       string `json:"session_id"`
	ProcessedFrames int    `json:"processed_frames"`
	TotalFrames     int    `json:"total_frames"`
	Progress        int    `json:"progress"`
	TotalDefects    int    `json:"total_defects"`
}

// VideoCompletedMessage 完成消息。
type VideoCompletedMessage struct {
	Type            string         `json:"type"`
	TaskID          string         `json:"task_id"`
	SessionID       string         `json:"session_id"`
	ProcessedFrames int            `json:"processed_frames"`
	TotalFrames     int            `json:"total_frames"`
	TotalDefects    int            `json:"total_defects"`
	DefectSummary   map[string]int `json:"defect_summary"`
}

// VideoErrorMessage 错误消息。
type VideoErrorMessage struct {
	Type      string `json:"type"`
	TaskID    string `json:"task_id"`
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// AlgoVideoTaskQueuedCallbackRequest 算法端任务入队回调。
type AlgoVideoTaskQueuedCallbackRequest struct {
	TaskID     string `json:"task_id" binding:"required"`
	AlgoTaskID string `json:"algo_task_id"`
	Status     string `json:"status" binding:"required"`
	QueuedAt   string `json:"queued_at"`
}

// AlgoVideoTaskStartedCallbackRequest 算法端任务开始回调。
type AlgoVideoTaskStartedCallbackRequest struct {
	TaskID     string `json:"task_id" binding:"required"`
	AlgoTaskID string `json:"algo_task_id"`
	Status     string `json:"status" binding:"required"`
	StartedAt  string `json:"started_at"`
}

// AlgoVideoBBox 算法端返回的 bbox。
type AlgoVideoBBox struct {
	BoxID      int        `json:"box_id"`
	ClassIdx   int        `json:"class_idx"`
	ClassName  string     `json:"class_name"`
	YOLOCoords [4]float64 `json:"yolo_coords"`
	Confidence float64    `json:"confidence"`
}

// AlgoVideoFrameResultCallbackRequest 算法端单帧结果回调。
type AlgoVideoFrameResultCallbackRequest struct {
	RequestID      string         `json:"request_id" binding:"required"`
	TaskID         string         `json:"task_id" binding:"required"`
	AlgoTaskID     string         `json:"algo_task_id"`
	FrameNo        int            `json:"frame_no" binding:"required"`
	TimestampMS    int            `json:"timestamp_ms"`
	Status         string         `json:"status" binding:"required"`
	QueueLatencyMS int            `json:"queue_latency_ms"`
	DetectTotalMS  int            `json:"detect_total_ms"`
	DecodeMS       int            `json:"decode_ms"`
	InferMS        int            `json:"infer_ms"`
	PostprocessMS  int            `json:"postprocess_ms"`
	FrameRef       string         `json:"frame_ref"`
	YOLOBBoxes     []AlgoVideoBBox `json:"yolo_bboxes"`
	ErrorMessage   string         `json:"error_message"`
}

// AlgoVideoTaskCompletedCallbackRequest 算法端任务完成回调。
type AlgoVideoTaskCompletedCallbackRequest struct {
	TaskID          string `json:"task_id" binding:"required"`
	AlgoTaskID      string `json:"algo_task_id"`
	Status          string `json:"status" binding:"required"`
	TotalFrames     int    `json:"total_frames"`
	ProcessedFrames int    `json:"processed_frames"`
	FailedFrames    int    `json:"failed_frames"`
	CompletedAt     string `json:"completed_at"`
}

// AlgoVideoTaskFailedCallbackRequest 算法端任务失败回调。
type AlgoVideoTaskFailedCallbackRequest struct {
	TaskID       string `json:"task_id" binding:"required"`
	AlgoTaskID   string `json:"algo_task_id"`
	Status       string `json:"status" binding:"required"`
	ErrorMessage string `json:"error_message"`
	FailedAt     string `json:"failed_at"`
}

// VideoCallbackAckResponse 视频回调确认响应。
type VideoCallbackAckResponse struct {
	TaskID    string `json:"task_id"`
	RequestID string `json:"request_id,omitempty"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
}
