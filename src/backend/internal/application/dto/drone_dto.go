// Package dto 定义数据传输对象（Data Transfer Object）
// 用于API层与业务逻辑层之间的数据传递
package dto

import "time"

// CreateDroneRequest 创建无人机请求
// 使用JSON binding，不涉及文件上传
type CreateDroneRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=100"` // 无人机名称（必填，1-100字符）
	Model     string `json:"model" binding:"max=100"`                // 设备型号（可选，最长100字符）
	StreamURL string `json:"stream_url" binding:"max=255"`           // 视频流地址（可选，最长255字符）
	UserID    uint   `json:"-"`                                      // 所属用户ID（由Handler设置）
}

// UpdateDroneRequest 更新无人机请求
// 所有字段都是可选的
type UpdateDroneRequest struct {
	Name      string `json:"name" binding:"omitempty,max=100"`      // 无人机名称（可选）
	Model     string `json:"model" binding:"omitempty,max=100"`     // 设备型号（可选）
	StreamURL string `json:"stream_url" binding:"omitempty,max=255"` // 视频流地址（可选）
}

// DroneResponse 无人机响应
type DroneResponse struct {
	ID        uint      `json:"id"`         // 无人机ID
	Name      string    `json:"name"`       // 无人机名称
	Model     string    `json:"model"`      // 设备型号
	StreamURL string    `json:"stream_url"` // 视频流地址
	UserID    uint      `json:"user_id"`    // 所属用户ID
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// DroneListResponse 无人机列表响应
type DroneListResponse struct {
	Total    int64           `json:"total"`     // 总记录数
	Page     int             `json:"page"`      // 当前页码
	PageSize int             `json:"page_size"` // 每页数量
	List     []DroneResponse `json:"list"`      // 无人机列表
}
