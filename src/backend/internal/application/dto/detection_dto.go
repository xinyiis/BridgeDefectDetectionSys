// Package dto 定义数据传输对象（Data Transfer Object）
package dto

import (
	"mime/multipart"
)

// DetectionUploadRequest 图片上传检测请求
type DetectionUploadRequest struct {
	Image      *multipart.FileHeader `form:"image" binding:"required"`            // 图片文件
	BridgeID   uint                  `form:"bridge_id" binding:"required"`        // 关联桥梁ID
	ModelName  string                `form:"model_name" binding:"required"`       // 模型名称/版本
	PixelRatio float64               `form:"pixel_ratio" binding:"required,gt=0"` // 像素实际系数
}

// DetectionResponse 检测响应（支持多缺陷）
type DetectionResponse struct {
	TotalDefects       int            `json:"total_defects"`                 // 检测到的缺陷总数
	ImagePath          string         `json:"image_path"`                    // 原始图片路径
	ResultPath         string         `json:"result_path"`                   // 结果图片路径
	ProcessingTime     float64        `json:"processing_time"`               // 处理时间（秒）
	PersistenceStatus  string         `json:"persistence_status"`            // 持久化状态：pending/completed/failed
	PersistenceTaskID  string         `json:"persistence_task_id,omitempty"` // 持久化任务ID
	Defects            []DefectDTO    `json:"defects"`                       // 缺陷列表
	DefectSummary      map[string]int `json:"defect_summary"`                // 缺陷统计（类型 -> 数量）
	PersistenceMessage string         `json:"persistence_message,omitempty"` // 持久化失败或状态说明
}

// DetectionPersistenceStatusResponse 检测持久化状态查询响应。
type DetectionPersistenceStatusResponse struct {
	TaskID         string `json:"task_id"`
	Status         string `json:"status"`
	ImagePath      string `json:"image_path"`
	ResultPath     string `json:"result_path"`
	SavedDefectIDs []uint `json:"saved_defect_ids"`
	ErrorMessage   string `json:"error_message,omitempty"`
}
