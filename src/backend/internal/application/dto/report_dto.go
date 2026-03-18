// Package dto 定义数据传输对象（Data Transfer Object）
// 用于API层与业务逻辑层之间的数据传递
package dto

import (
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
)

// CreateReportRequest 创建报表请求
type CreateReportRequest struct {
	ReportName string            `json:"report_name" binding:"required,max=200"`                                                      // 报表名称
	ReportType model.ReportType  `json:"report_type" binding:"required,oneof=bridge_inspection defect_analysis health_comparison"` // 报表类型
	BridgeID   *uint             `json:"bridge_id"`                                                                                  // 关联桥梁ID（单桥梁报表）
	BridgeIDs  []uint            `json:"bridge_ids"`                                                                                 // 桥梁ID列表（多桥梁报表）
	StartTime  string            `json:"start_time" binding:"required"`                                                              // 报表开始时间（格式：2006-01-02）
	EndTime    string            `json:"end_time" binding:"required"`                                                                // 报表结束时间（格式：2006-01-02）
	UserID     uint              `json:"-"`                                                                                          // 创建用户ID（从Session获取，不从请求体绑定）
}

// UpdateReportRequest 更新报表请求
type UpdateReportRequest struct {
	ReportName *string `json:"report_name" binding:"omitempty,max=200"` // 报表名称（可选）
}

// ReportResponse 报表响应
type ReportResponse struct {
	ID            uint               `json:"id"`                       // 报表ID
	ReportName    string             `json:"report_name"`              // 报表名称
	ReportType    model.ReportType   `json:"report_type"`              // 报表类型
	UserID        uint               `json:"user_id"`                  // 创建用户ID
	BridgeID      *uint              `json:"bridge_id,omitempty"`      // 关联桥梁ID
	BridgeName    string             `json:"bridge_name,omitempty"`    // 桥梁名称
	StartTime     time.Time          `json:"start_time"`               // 报表开始时间
	EndTime       time.Time          `json:"end_time"`                 // 报表结束时间
	FilePath      string             `json:"file_path,omitempty"`      // PDF文件路径
	FileSize      int64              `json:"file_size"`                // 文件大小（字节）
	Status        model.ReportStatus `json:"status"`                   // 生成状态
	ErrorMessage  string             `json:"error_message,omitempty"`  // 错误信息
	TotalPages    int                `json:"total_pages"`              // 总页数
	DefectCount   int                `json:"defect_count"`             // 缺陷数量
	HighRiskCount int                `json:"high_risk_count"`          // 高危缺陷数量
	HealthScore   float64            `json:"health_score"`             // 健康度评分
	CreatedAt     time.Time          `json:"created_at"`               // 创建时间
}

// ReportListResponse 报表列表响应
type ReportListResponse struct {
	Total    int64            `json:"total"`     // 总数量
	Page     int              `json:"page"`      // 当前页码
	PageSize int              `json:"page_size"` // 每页数量
	List     []ReportResponse `json:"list"`      // 报表列表
}

// ReportQueryParams 报表查询参数
type ReportQueryParams struct {
	Page       int               `form:"page" binding:"omitempty,min=1"`                                                             // 页码
	PageSize   int               `form:"page_size" binding:"omitempty,min=1,max=100"`                                                // 每页数量
	ReportType *model.ReportType `form:"report_type" binding:"omitempty,oneof=bridge_inspection defect_analysis health_comparison"` // 报表类型过滤
	BridgeID   *uint             `form:"bridge_id"`                                                                                  // 桥梁ID过滤
}
