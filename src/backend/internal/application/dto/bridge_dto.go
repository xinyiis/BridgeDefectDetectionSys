// Package dto 定义数据传输对象（Data Transfer Object）
// 用于API层与业务逻辑层之间的数据传递
package dto

import "time"

// CreateBridgeRequest 创建桥梁请求
type CreateBridgeRequest struct {
	BridgeName  string  `form:"bridge_name" binding:"required"`
	BridgeCode  string  `form:"bridge_code" binding:"required"`
	Address     string  `form:"address" binding:"required"`
	Longitude   float64 `form:"longitude" binding:"required"`
	Latitude    float64 `form:"latitude" binding:"required"`
	BridgeType  string  `form:"bridge_type" binding:"required"`
	BuildYear   int     `form:"build_year" binding:"required"`
	Length      float64 `form:"length" binding:"required"`
	Width       float64 `form:"width" binding:"required"`
	Remark      string  `form:"remark"`
	Model3DPath string  `form:"-"` // 由Handler设置
	UserID      uint    `form:"-"` // 由Handler设置
}

// UpdateBridgeRequest 更新桥梁请求
type UpdateBridgeRequest struct {
	BridgeName  string  `form:"bridge_name"`
	Address     string  `form:"address"`
	Longitude   float64 `form:"longitude"`
	Latitude    float64 `form:"latitude"`
	BridgeType  string  `form:"bridge_type"`
	BuildYear   int     `form:"build_year"`
	Length      float64 `form:"length"`
	Width       float64 `form:"width"`
	Status      string  `form:"status"`
	Remark      string  `form:"remark"`
	Model3DPath string  `form:"-"`
}

// BridgeResponse 桥梁响应
type BridgeResponse struct {
	ID          uint      `json:"id"`
	BridgeName  string    `json:"bridge_name"`
	BridgeCode  string    `json:"bridge_code"`
	Address     string    `json:"address"`
	Longitude   float64   `json:"longitude"`
	Latitude    float64   `json:"latitude"`
	BridgeType  string    `json:"bridge_type"`
	BuildYear   int       `json:"build_year"`
	Length      float64   `json:"length"`
	Width       float64   `json:"width"`
	Status      string    `json:"status"`
	Model3DPath string    `json:"model_3d_path"`
	Remark      string    `json:"remark"`
	UserID      uint      `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BridgeListResponse 桥梁列表响应
type BridgeListResponse struct {
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	List     []BridgeResponse `json:"list"`
}
