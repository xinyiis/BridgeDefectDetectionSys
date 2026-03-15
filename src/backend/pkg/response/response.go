// Package response 提供统一的 HTTP 响应工具
// 标准化所有 API 接口的响应格式
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`              // 状态码（200表示成功，其他表示错误）
	Message string      `json:"message"`           // 响应消息
	Data    interface{} `json:"data,omitempty"`    // 响应数据（成功时返回）
	Error   string      `json:"error,omitempty"`   // 错误详情（失败时返回）
}

// Success 返回成功响应
// 用于 API 请求成功时返回数据
// 参数：
//   - c: Gin 上下文
//   - data: 要返回的数据（会被放入 response.data 字段）
//
// 响应格式：
//
//	{
//	  "code": 200,
//	  "message": "success",
//	  "data": { ... }
//	}
//
// 使用示例：
//
//	response.Success(c, gin.H{"user": user})
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 返回带自定义消息的成功响应
// 参数：
//   - c: Gin 上下文
//   - message: 自定义成功消息
//   - data: 要返回的数据
//
// 使用示例：
//
//	response.SuccessWithMessage(c, "创建成功", gin.H{"id": 1})
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	})
}

// Error 返回错误响应
// 用于通用错误处理
// 参数：
//   - c: Gin 上下文
//   - code: HTTP 状态码
//   - message: 错误消息
//
// 响应格式：
//
//	{
//	  "code": 400,
//	  "message": "参数错误"
//	}
//
// 使用示例：
//
//	response.Error(c, 400, "参数错误")
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// ErrorWithDetail 返回带详细错误信息的响应
// 参数：
//   - c: Gin 上下文
//   - code: HTTP 状态码
//   - message: 错误消息
//   - detail: 详细错误信息
//
// 使用示例：
//
//	response.ErrorWithDetail(c, 500, "数据库错误", err.Error())
func ErrorWithDetail(c *gin.Context, code int, message string, detail string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
		Error:   detail,
	})
}

// BadRequest 返回 400 错误（请求参数错误）
// 参数：
//   - c: Gin 上下文
//   - message: 错误消息
//
// 使用示例：
//
//	response.BadRequest(c, "用户名不能为空")
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized 返回 401 错误（未授权/未登录）
// 参数：
//   - c: Gin 上下文
//
// 使用示例：
//
//	response.Unauthorized(c)
func Unauthorized(c *gin.Context) {
	Error(c, http.StatusUnauthorized, "未登录或登录已过期")
}

// UnauthorizedWithMessage 返回带自定义消息的 401 错误
// 参数：
//   - c: Gin 上下文
//   - message: 自定义错误消息
func UnauthorizedWithMessage(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden 返回 403 错误（权限不足）
// 参数：
//   - c: Gin 上下文
//
// 使用示例：
//
//	response.Forbidden(c)
func Forbidden(c *gin.Context) {
	Error(c, http.StatusForbidden, "权限不足")
}

// ForbiddenWithMessage 返回带自定义消息的 403 错误
// 参数：
//   - c: Gin 上下文
//   - message: 自定义错误消息
func ForbiddenWithMessage(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound 返回 404 错误（资源不存在）
// 参数：
//   - c: Gin 上下文
//   - resource: 资源名称
//
// 使用示例：
//
//	response.NotFound(c, "用户")
func NotFound(c *gin.Context, resource string) {
	Error(c, http.StatusNotFound, resource+"不存在")
}

// InternalError 返回 500 错误（服务器内部错误）
// 参数：
//   - c: Gin 上下文
//
// 使用示例：
//
//	response.InternalError(c)
func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "服务器内部错误")
}

// InternalErrorWithDetail 返回带详细信息的 500 错误
// 参数：
//   - c: Gin 上下文
//   - detail: 详细错误信息（开发环境可显示，生产环境建议隐藏）
func InternalErrorWithDetail(c *gin.Context, detail string) {
	ErrorWithDetail(c, http.StatusInternalServerError, "服务器内部错误", detail)
}
