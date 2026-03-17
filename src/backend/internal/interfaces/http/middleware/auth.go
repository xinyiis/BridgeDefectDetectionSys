// Package middleware 提供 HTTP 中间件
// 包括认证、授权、CORS 等中间件
package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/config"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/response"
	"gorm.io/gorm"
)

// CORSMiddleware 创建 CORS 跨域中间件
// 允许前端（不同端口）访问后端 API
// 参数：
//   - cfg: 配置对象，包含允许的源列表
//
// 返回值：
//   - gin.HandlerFunc: Gin 中间件函数
//
// 功能：
//   - 设置允许的源（Origins）
//   - 设置允许的 HTTP 方法（GET, POST, PUT, DELETE）
//   - 设置允许的请求头
//   - 允许携带 Cookie（Session 认证必需）
//
// 使用示例：
//
//	r.Use(middleware.CORSMiddleware(cfg))
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return cors.New(cors.Config{
		// 允许的源列表（前端地址）
		AllowOrigins: cfg.CORS.AllowOrigins,

		// 允许的 HTTP 方法
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},

		// 允许的请求头
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},

		// 暴露的响应头（前端可以访问）
		ExposeHeaders: []string{"Content-Length"},

		// 允许携带 Cookie（Session 认证必需）
		AllowCredentials: cfg.CORS.AllowCredentials,

		// 预检请求的有效期（秒）
		MaxAge: 12 * 3600, // 12 小时
	})
}

// AuthRequired 认证中间件
// 检查用户是否已登录，未登录则返回 401 错误
// 参数：
//   - db: GORM 数据库连接对象
//
// 返回值：
//   - gin.HandlerFunc: Gin 中间件函数
//
// 工作流程：
//   1. 从 Session 中获取用户ID
//   2. 如果没有用户ID，返回 401 错误
//   3. 从数据库加载用户信息
//   4. 将用户信息存入 Context，供后续 handler 使用
//
// 使用示例：
//
//	auth := r.Group("/api")
//	auth.Use(middleware.AuthRequired(db))
//	auth.GET("/user/info", handler.GetUserInfo)
//
// 在 Handler 中获取当前用户：
//
//	user := c.MustGet("current_user").(*model.User)
func AuthRequired(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Session
		session := sessions.Default(c)
		userID := session.Get("user_id")

		// 2. 检查是否已登录
		if userID == nil {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		// 3. 从数据库加载用户信息
		var user model.User
		if err := db.First(&user, userID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 用户不存在（可能已被删除）
				response.UnauthorizedWithMessage(c, "用户不存在")
			} else {
				// 数据库错误
				response.InternalErrorWithDetail(c, err.Error())
			}
			c.Abort()
			return
		}

		// 4. 将用户信息存入 Context
		c.Set("current_user", &user)

		// 5. 继续执行后续处理
		c.Next()
	}
}

// AdminRequired 管理员权限中间件
// 检查当前用户是否为管理员，非管理员则返回 403 错误
// 注意：此中间件必须在 AuthRequired 之后使用
// 返回值：
//   - gin.HandlerFunc: Gin 中间件函数
//
// 工作流程：
//   1. 从 Context 获取当前用户（由 AuthRequired 中间件设置）
//   2. 检查用户角色是否为 admin
//   3. 如果不是管理员，返回 403 错误
//
// 使用示例：
//
//	admin := r.Group("/api/admin")
//	admin.Use(middleware.AuthRequired(db))
//	admin.Use(middleware.AdminRequired())
//	admin.GET("/users", handler.GetAllUsers)
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从 Context 获取当前用户
		userInterface, exists := c.Get("current_user")
		if !exists {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		// 2. 类型断言
		user, ok := userInterface.(*model.User)
		if !ok {
			response.InternalError(c)
			c.Abort()
			return
		}

		// 3. 检查是否为管理员
		if !user.IsAdmin() {
			response.Forbidden(c)
			c.Abort()
			return
		}

		// 4. 继续执行后续处理
		c.Next()
	}
}

// CheckResourceOwnership 资源所有权检查中间件
// 检查用户是否有权访问某个资源（如桥梁、无人机）
// 管理员可以访问所有资源，普通用户只能访问自己的资源
// 参数：
//   - db: GORM 数据库连接对象
//   - resourceType: 资源类型（"bridge", "drone"）
//
// 返回值：
//   - gin.HandlerFunc: Gin 中间件函数
//
// 工作流程：
//   1. 获取当前用户
//   2. 如果是管理员，直接放行
//   3. 获取资源ID（从 URL 参数）
//   4. 查询资源，检查 user_id 是否匹配
//
// 使用示例：
//
//	r.GET("/bridges/:id", middleware.CheckResourceOwnership(db, "bridge"), handler.GetBridge)
func CheckResourceOwnership(db *gorm.DB, resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取当前用户
		user := c.MustGet("current_user").(*model.User)

		// 2. 管理员跳过检查
		if user.IsAdmin() {
			c.Next()
			return
		}

		// 3. 获取资源ID
		resourceID := c.Param("id")
		if resourceID == "" {
			response.BadRequest(c, "资源ID不能为空")
			c.Abort()
			return
		}

		// 4. 根据资源类型检查所有权
		var userID uint
		var err error

		switch resourceType {
		case "bridge":
			var bridge model.Bridge
			err = db.Select("user_id").First(&bridge, resourceID).Error
			if err == nil {
				userID = bridge.UserID
			}

		case "drone":
			var drone model.Drone
			err = db.Select("user_id").First(&drone, resourceID).Error
			if err == nil {
				userID = drone.UserID
			}

		default:
			response.InternalError(c)
			c.Abort()
			return
		}

		// 5. 处理查询结果
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				response.NotFound(c, resourceType)
			} else {
				response.InternalErrorWithDetail(c, err.Error())
			}
			c.Abort()
			return
		}

		// 6. 检查所有权
		if userID != user.ID {
			response.ForbiddenWithMessage(c, "无权访问此资源")
			c.Abort()
			return
		}

		// 7. 继续执行后续处理
		c.Next()
	}
}

// GetCurrentUser 从 Context 获取当前登录用户
// 这是一个辅助函数，简化 handler 中获取用户的代码
// 参数：
//   - c: Gin 上下文
//
// 返回值：
//   - *model.User: 当前用户对象
//
// 注意：
//   - 此函数假设请求已通过 AuthRequired 中间件
//   - 如果未通过认证，会 panic（应该在中间件层面保证不会发生）
//
// 使用示例：
//
//	func GetUserInfo(c *gin.Context) {
//	    user := middleware.GetCurrentUser(c)
//	    response.Success(c, user)
//	}
func GetCurrentUser(c *gin.Context) *model.User {
	return c.MustGet("current_user").(*model.User)
}
