package model

import "time"

// VideoTaskStatus 视频分析任务状态。
type VideoTaskStatus string

const (
	VideoTaskUploaded       VideoTaskStatus = "uploaded"
	VideoTaskReadyToProcess VideoTaskStatus = "ready_to_process"
	VideoTaskProcessing     VideoTaskStatus = "processing"
	VideoTaskCompleted      VideoTaskStatus = "completed"
	VideoTaskFailed         VideoTaskStatus = "failed"
	VideoTaskCanceled       VideoTaskStatus = "canceled"
)

// VideoAnalysisTask 视频分析任务。
type VideoAnalysisTask struct {
	ID               uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID           string          `gorm:"type:varchar(64);uniqueIndex;not null" json:"task_id"`
	SessionID        string          `gorm:"type:varchar(64);index" json:"session_id"`
	UserID           uint            `gorm:"not null;index" json:"user_id"`
	BridgeID         uint            `gorm:"not null;index" json:"bridge_id"`
	VideoPath        string          `gorm:"type:varchar(255);not null" json:"video_path"`
	FPS              float64         `gorm:"type:decimal(8,2);default:1" json:"fps"`
	ModelName        string          `gorm:"type:varchar(50);default:'baseline'" json:"model_name"`
	PixelRatio       float64         `gorm:"type:decimal(12,6);not null" json:"pixel_ratio"`
	EnableSegment    bool            `gorm:"default:false" json:"enable_segment"`
	Status           VideoTaskStatus `gorm:"type:varchar(32);index;not null" json:"status"`
	TotalFrames      int             `gorm:"default:0" json:"total_frames"`
	ProcessedFrames  int             `gorm:"default:0" json:"processed_frames"`
	ConfirmedDefects int             `gorm:"default:0" json:"confirmed_defects"`
	ErrorMessage     string          `gorm:"type:text" json:"error_message"`
	CreatedAt        time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	StartedAt        *time.Time      `json:"started_at,omitempty"`
	CompletedAt      *time.Time      `json:"completed_at,omitempty"`
	CanceledAt       *time.Time      `json:"canceled_at,omitempty"`
}

// TableName 指定表名。
func (VideoAnalysisTask) TableName() string {
	return "video_analysis_tasks"
}

// DefectObservation 帧级观测记录。
type DefectObservation struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID      string    `gorm:"type:varchar(64);index;not null" json:"task_id"`
	BridgeID    uint      `gorm:"not null;index" json:"bridge_id"`
	FrameNo     int       `gorm:"not null;index" json:"frame_no"`
	TimestampMS int       `gorm:"not null" json:"timestamp_ms"`
	DefectType  string    `gorm:"type:varchar(50);not null" json:"defect_type"`
	BBox        string    `gorm:"type:text" json:"bbox"`
	Confidence  float64   `gorm:"type:decimal(5,4)" json:"confidence"`
	FramePath   string    `gorm:"type:varchar(255)" json:"frame_path"`
	ResultPath  string    `gorm:"type:varchar(255)" json:"result_path"`
	TrackID     string    `gorm:"type:varchar(64);index" json:"track_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名。
func (DefectObservation) TableName() string {
	return "defect_observations"
}
