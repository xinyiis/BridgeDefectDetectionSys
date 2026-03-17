// Package handler 定义HTTP处理器
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
)

// UserHandler 用户管理处理器
// 处理用户信息查询、更新、删除等HTTP请求
type UserHandler struct {
	userUseCase *usecase.UserUseCase // 用户管理用例
}

// NewUserHandler 创建用户管理处理器实例
// 参数：
//   - userUseCase: 用户管理用例
// 返回：
//   - *UserHandler: 用户管理处理器实例
func NewUserHandler(userUseCase *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

// GetUserInfo 获取当前用户信息接口
// GET /api/user/info
// 响应: UserResponse
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	// 1. 从Session获取当前用户ID
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	// 2. 调用UseCase获取用户信息
	user, err := h.userUseCase.GetUserInfo(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	// 3. 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    user,
	})
}

// UpdateUserInfo 更新当前用户信息接口
// PUT /api/user/info
// 请求体: UpdateUserRequest
// 响应: UserResponse
func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	// 1. 从Session获取当前用户ID
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	// 2. 参数绑定和验证
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 3. 调用UseCase更新用户信息
	user, err := h.userUseCase.UpdateUserInfo(userID.(uint), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	// 4. 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    user,
	})
}

// ListUsers 获取用户列表接口（管理员）
// GET /api/users?page=1&page_size=10
// 响应: UserListResponse
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 1. 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 2. 调用UseCase获取用户列表
	userList, err := h.userUseCase.ListUsers(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	// 3. 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    userList,
	})
}

// GetUserByID 获取指定用户信息接口（管理员）
// GET /api/users/:id
// 响应: UserResponse
func (h *UserHandler) GetUserByID(c *gin.Context) {
	// 1. 获取用户ID参数
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID格式错误",
		})
		return
	}

	// 2. 调用UseCase获取用户信息
	user, err := h.userUseCase.GetUserInfo(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}

	// 3. 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    user,
	})
}

// DeleteUser 删除用户接口（管理员）
// DELETE /api/users/:id
// 响应: 成功消息
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 1. 获取用户ID参数
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID格式错误",
		})
		return
	}

	// 2. 调用UseCase删除用户
	if err := h.userUseCase.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	// 3. 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// PromoteToAdmin 提升用户为管理员接口（管理员）
// POST /api/users/promote
// 请求体: PromoteUserRequest
// 响应: UserResponse
func (h *UserHandler) PromoteToAdmin(c *gin.Context) {
	// 1. 参数绑定和验证
	var req dto.PromoteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 2. 调用UseCase提升为管理员
	user, err := h.userUseCase.PromoteToAdmin(&req)
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
		"message": "提升成功",
		"data":    user,
	})
}
