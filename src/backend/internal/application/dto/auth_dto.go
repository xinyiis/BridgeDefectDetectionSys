// Package dto 定义数据传输对象（Data Transfer Object）
// 用于API请求和响应的数据结构
package dto

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`      // 用户名（必填，3-50字符）
	Password string `json:"password" binding:"required,min=6,max=50"`      // 密码（必填，6-50字符）
	RealName string `json:"real_name" binding:"required,min=2,max=50"`     // 真实姓名（必填，2-50字符）
	Phone    string `json:"phone" binding:"omitempty,len=11,numeric"`      // 手机号（可选，11位数字）
	Email    string `json:"email" binding:"required,email,max=100"`        // 邮箱（必填，需符合邮箱格式）
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名（必填）
	Password string `json:"password" binding:"required"` // 密码（必填）
}

// LoginResponse 登录响应
type LoginResponse struct {
	User    *UserResponse `json:"user"`    // 用户信息
	Message string        `json:"message"` // 提示信息
}
