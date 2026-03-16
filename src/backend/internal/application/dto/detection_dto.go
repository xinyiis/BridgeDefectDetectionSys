// Package dto 定义数据传输对象（Data Transfer Object）
package dto

import (
	"mime/multipart"
)

// DetectionUploadRequest 图片上传检测请求
type DetectionUploadRequest struct {
	Image      *multipart.FileHeader `form:"image" binding:"required"`       // 图片文件
	BridgeID   uint                  `form:"bridge_id" binding:"required"`   // 关联桥梁ID
	ModelName  string                `form:"model_name" binding:"required"`  // 模型名称/版本
	PixelRatio float64               `form:"pixel_ratio" binding:"required,gt=0"` // 像素实际系数
}

// DetectionResponse 检测响应（支持多缺陷）
type DetectionResponse struct {
	TotalDefects   int            `json:"total_defects"`   // 检测到的缺陷总数
	ImagePath      string         `json:"image_path"`      // 原始图片路径
	ResultPath     string         `json:"result_path"`     // 结果图片路径
	ProcessingTime float64        `json:"processing_time"` // 处理时间（秒）
	Defects        []DefectDTO    `json:"defects"`         // 缺陷列表
	DefectSummary  map[string]int `json:"defect_summary"`  // 缺陷统计（类型 -> 数量）
}
