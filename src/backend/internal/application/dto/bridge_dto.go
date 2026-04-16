// Package dto 定义数据传输对象（Data Transfer Object）
// 用于API层与业务逻辑层之间的数据传递
package dto

import "time"

// CreateBridgeRequest 创建桥梁请求
type CreateBridgeRequest struct {
	BridgeName  string  `json:"bridge_name" form:"bridge_name" binding:"required"`
	BridgeCode  string  `json:"bridge_code" form:"bridge_code" binding:"required"`
	Address     string  `json:"address" form:"address" binding:"required"`
	Longitude   float64 `json:"longitude" form:"longitude" binding:"required"`
	Latitude    float64 `json:"latitude" form:"latitude" binding:"required"`
	BridgeType  string  `json:"bridge_type" form:"bridge_type" binding:"required"`
	BuildYear   int     `json:"build_year" form:"build_year" binding:"required"`
	Length      float64 `json:"length" form:"length" binding:"required"`
	Width       float64 `json:"width" form:"width" binding:"required"`
	Status      string  `json:"status" form:"status"`
	Remark      string  `json:"remark" form:"remark"`
	Model3DPath string  `json:"model_3d_path" form:"-"` // 由Handler设置
	UserID      uint    `json:"user_id" form:"-"`       // 由Handler设置
}

// UpdateBridgeRequest 更新桥梁请求
type UpdateBridgeRequest struct {
	BridgeName string  `json:"bridge_name" form:"bridge_name"`
	Address    string  `json:"address" form:"address"`
	Longitude  float64 `json:"longitude" form:"longitude"`
	Latitude   float64 `json:"latitude" form:"latitude"`
	BridgeType string  `json:"bridge_type" form:"bridge_type"`
	BuildYear  int     `json:"build_year" form:"build_year"`
	Length     float64 `json:"length" form:"length"`
	Width      float64 `json:"width" form:"width"`
	Status     string  `json:"status" form:"status"`
	Remark     string  `json:"remark" form:"remark"`
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
