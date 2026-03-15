// Package dto 定义数据传输对象（Data Transfer Object）
package dto

import "time"

// UserResponse 用户信息响应（脱敏）
type UserResponse struct {
	ID        uint      `json:"id"`         // 用户ID
	Username  string    `json:"username"`   // 用户名
	RealName  string    `json:"real_name"`  // 真实姓名
	Phone     string    `json:"phone"`      // 手机号
	Email     string    `json:"email"`      // 邮箱
	Role      string    `json:"role"`       // 角色: user/admin
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	RealName string `json:"real_name" binding:"omitempty,min=2,max=50"` // 真实姓名（可选，2-50字符）
	Phone    string `json:"phone" binding:"omitempty,len=11,numeric"`   // 手机号（可选，11位数字）
	Email    string `json:"email" binding:"omitempty,email,max=100"`    // 邮箱（可选，需符合邮箱格式）
	Password string `json:"password" binding:"omitempty,min=6,max=50"`  // 密码（可选，6-50字符）
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Total    int64           `json:"total"`     // 总数
	Page     int             `json:"page"`      // 当前页码
	PageSize int             `json:"page_size"` // 每页数量
	Users    []UserResponse  `json:"users"`     // 用户列表
}

// PromoteUserRequest 提升用户为管理员请求
type PromoteUserRequest struct {
	UserID uint `json:"user_id" binding:"required"` // 用户ID（必填）
}
