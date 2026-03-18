// Package dto 定义数据传输对象（Data Transfer Object）
package dto

import "time"

// DefectDTO 缺陷信息DTO
type DefectDTO struct {
	ID         uint      `json:"id"`          // 缺陷ID
	BridgeID   uint      `json:"bridge_id"`   // 所属桥梁ID
	BridgeName string    `json:"bridge_name,omitempty"` // 桥梁名称（列表查询时包含）
	DefectType string    `json:"defect_type"` // 缺陷类型
	ImagePath  string    `json:"image_path"`  // 原始图片路径
	ResultPath string    `json:"result_path"` // 结果图片路径
	BBox       string    `json:"bbox"`        // 边界框坐标（JSON字符串）
	Length     float64   `json:"length"`      // 缺陷长度（米）
	Width      float64   `json:"width"`       // 缺陷宽度（米）
	Area       float64   `json:"area"`        // 缺陷面积（平方米）
	Confidence float64   `json:"confidence"`  // 置信度（0-1）
	DetectedAt time.Time `json:"detected_at"` // 检测时间
	CreatedAt  time.Time `json:"created_at,omitempty"` // 创建时间
}

// DefectListRequest 缺陷列表查询请求
type DefectListRequest struct {
	Page       int    `form:"page,default=1"`         // 页码
	PageSize   int    `form:"page_size,default=10"`   // 每页数量
	BridgeID   *uint  `form:"bridge_id"`              // 按桥梁ID过滤（可选）
	DefectType string `form:"defect_type"`            // 按缺陷类型过滤（可选）
	StartDate  string `form:"start_date"`             // 开始时间（YYYY-MM-DD）
	EndDate    string `form:"end_date"`               // 结束时间（YYYY-MM-DD）
}

// DefectListResponse 缺陷列表响应
type DefectListResponse struct {
	Total    int64       `json:"total"`     // 总数量
	Page     int         `json:"page"`      // 当前页码
	PageSize int         `json:"page_size"` // 每页数量
	List     []DefectDTO `json:"list"`      // 缺陷列表
}

// DefectDetailResponse 缺陷详情响应（包含桥梁信息）
type DefectDetailResponse struct {
	DefectDTO
	Bridge *BridgeSimpleInfo `json:"bridge,omitempty"` // 关联桥梁信息
}

// BridgeSimpleInfo 桥梁简要信息（用于缺陷详情）
type BridgeSimpleInfo struct {
	ID         uint   `json:"id"`
	BridgeName string `json:"bridge_name"`
	BridgeCode string `json:"bridge_code"`
}
