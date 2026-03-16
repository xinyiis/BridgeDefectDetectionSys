// Package middleware 提供资源所有权验证中间件
// 负责验证用户是否有权访问特定资源
package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/response"
)

// BridgeOwnershipRequired 验证用户是否有权访问桥梁资源
// 功能：
//   1. 获取当前用户
//   2. 管理员直接放行（不查数据库）
//   3. 获取桥梁ID并查询桥梁
//   4. 验证所有权
//   5. 将桥梁对象存入上下文（避免Handler重复查询）
//
// 参数：
//   - bridgeRepo: 桥梁Repository接口
//
// 返回值：
//   - gin.HandlerFunc: Gin中间件函数
//
// 使用示例：
//   bridgeResource := bridges.Group("/:id")
//   bridgeResource.Use(middleware.BridgeOwnershipRequired(bridgeRepo))
//   {
//       bridgeResource.GET("", handler.GetBridge)
//       bridgeResource.PUT("", handler.UpdateBridge)
//       bridgeResource.DELETE("", handler.DeleteBridge)
//   }
func BridgeOwnershipRequired(bridgeRepo repository.BridgeRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取当前用户
		currentUser, exists := c.Get("current_user")
		if !exists {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		user := currentUser.(*model.User)

		// 2. 管理员直接放行（不查数据库）
		if user.IsAdmin() {
			c.Next()
			return
		}

		// 3. 获取桥梁ID
		bridgeIDStr := c.Param("id")
		bridgeID, err := strconv.ParseUint(bridgeIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效的桥梁ID")
			c.Abort()
			return
		}

		// 4. 查询桥梁（验证存在性和所有权）
		bridge, err := bridgeRepo.FindByID(uint(bridgeID))
		if err != nil {
			response.InternalErrorWithDetail(c, "查询桥梁失败")
			c.Abort()
			return
		}
		if bridge == nil {
			response.NotFound(c, "桥梁不存在")
			c.Abort()
			return
		}

		// 5. 验证所有权
		if !bridge.IsOwnedBy(user.ID) {
			response.Forbidden(c)
			c.Abort()
			return
		}

		// 6. 将桥梁信息存入上下文（避免Handler重复查询）
		c.Set("bridge", bridge)
		c.Next()
	}
}
