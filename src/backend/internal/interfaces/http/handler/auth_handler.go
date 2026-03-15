// Package handler 定义HTTP处理器
// 负责处理HTTP请求、参数验证、调用UseCase、返回响应
package handler

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
)

// AuthHandler 认证处理器
// 处理用户注册、登录、登出相关的HTTP请求
type AuthHandler struct {
	authUseCase *usecase.AuthUseCase // 认证用例
}

// NewAuthHandler 创建认证处理器实例
// 参数：
//   - authUseCase: 认证用例
// 返回：
//   - *AuthHandler: 认证处理器实例
func NewAuthHandler(authUseCase *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// Register 用户注册接口
// POST /api/register
// 请求体: RegisterRequest
// 响应: UserResponse
func (h *AuthHandler) Register(c *gin.Context) {
	// 1. 参数绑定和验证
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 2. 调用UseCase执行业务逻辑
	user, err := h.authUseCase.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	// 3. 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "注册成功",
		"data":    user,
	})
}

// Login 用户登录接口
// POST /api/login
// 请求体: LoginRequest
// 响应: LoginResponse
func (h *AuthHandler) Login(c *gin.Context) {
	// 1. 参数绑定和验证
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 2. 调用UseCase验证用户
	loginResp, err := h.authUseCase.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": err.Error(),
		})
		return
	}

	// 3. 创建Session（使用gin-contrib/sessions）
	session := sessions.Default(c)
	session.Set("user_id", loginResp.User.ID)
	session.Set("username", loginResp.User.Username)
	session.Set("role", loginResp.User.Role)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "登录失败，请稍后重试",
		})
		return
	}

	// 4. 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data":    loginResp.User,
	})
}

// Logout 用户登出接口
// POST /api/logout
// 响应: 成功消息
func (h *AuthHandler) Logout(c *gin.Context) {
	// 1. 获取Session
	session := sessions.Default(c)

	// 2. 清除Session
	session.Clear()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "登出失败，请稍后重试",
		})
		return
	}

	// 3. 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登出成功",
	})
}
