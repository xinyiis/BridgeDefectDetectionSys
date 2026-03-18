// Package middleware 提供资源所有权验证中间件
// 负责验证用户是否有权访问特定资源
package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
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

// DroneOwnershipRequired 验证用户是否有权访问无人机资源
// 功能：
//   1. 获取当前用户
//   2. 管理员直接放行（不查数据库）
//   3. 获取无人机ID并查询无人机
//   4. 验证所有权
//   5. 将无人机对象存入上下文（避免Handler重复查询）
//
// 参数：
//   - droneRepo: 无人机Repository接口
//
// 返回值：
//   - gin.HandlerFunc: Gin中间件函数
//
// 使用示例：
//   droneResource := drones.Group("/:id")
//   droneResource.Use(middleware.DroneOwnershipRequired(droneRepo))
//   {
//       droneResource.GET("", handler.GetDrone)
//       droneResource.PUT("", handler.UpdateDrone)
//       droneResource.DELETE("", handler.DeleteDrone)
//   }
func DroneOwnershipRequired(droneRepo repository.DroneRepository) gin.HandlerFunc {
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

		// 3. 获取无人机ID
		droneIDStr := c.Param("id")
		droneID, err := strconv.ParseUint(droneIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效的无人机ID")
			c.Abort()
			return
		}

		// 4. 查询无人机（验证存在性和所有权）
		drone, err := droneRepo.FindByID(uint(droneID))
		if err != nil {
			response.InternalErrorWithDetail(c, "查询无人机失败")
			c.Abort()
			return
		}
		if drone == nil {
			response.NotFound(c, "无人机不存在")
			c.Abort()
			return
		}

		// 5. 验证所有权
		if !drone.IsOwnedBy(user.ID) {
			response.Forbidden(c)
			c.Abort()
			return
		}

		// 6. 将无人机信息存入上下文（避免Handler重复查询）
		c.Set("drone", drone)
		c.Next()
	}
}

// DefectOwnershipRequired 验证用户是否有权访问缺陷资源
// 功能：
//   1. 获取当前用户
//   2. 管理员直接放行（不查数据库）
//   3. 获取缺陷ID并查询缺陷
//   4. 查询关联的桥梁
//   5. 验证桥梁所有权
//   6. 将缺陷对象存入上下文（避免Handler重复查询）
//
// 参数：
//   - defectService: 缺陷领域服务（用于验证所有权）
//
// 返回值：
//   - gin.HandlerFunc: Gin中间件函数
//
// 使用示例：
//   defectResource := defects.Group("/:id")
//   defectResource.Use(middleware.DefectOwnershipRequired(defectService))
//   {
//       defectResource.GET("", handler.GetDefect)
//       defectResource.DELETE("", handler.DeleteDefect)
//   }
func DefectOwnershipRequired(defectService *service.DefectService) gin.HandlerFunc {
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

		// 3. 获取缺陷ID
		defectIDStr := c.Param("id")
		defectID, err := strconv.ParseUint(defectIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效的缺陷ID")
			c.Abort()
			return
		}

		// 4. 验证所有权（跨表查询：defect -> bridge -> user_id）
		defect, err := defectService.VerifyDefectOwnership(uint(defectID), user.ID, user.IsAdmin())
		if err != nil {
			// 根据错误类型返回不同响应
			if err.Error() == "缺陷不存在" || err.Error() == "关联桥梁不存在" {
				response.NotFound(c, err.Error())
			} else {
				response.Forbidden(c)
			}
			c.Abort()
			return
		}

		// 5. 将缺陷信息存入上下文（避免Handler重复查询）
		c.Set("defect", defect)
		c.Next()
	}
}

// ReportOwnershipRequired 验证用户是否有权访问报表资源
// 功能：
//   1. 获取当前用户
//   2. 管理员直接放行（不查数据库）
//   3. 获取报表ID并查询报表
//   4. 验证所有权
//   5. 将报表对象存入上下文（避免Handler重复查询）
//
// 参数：
//   - reportRepo: 报表Repository接口
//
// 返回值：
//   - gin.HandlerFunc: Gin中间件函数
//
// 使用示例：
//   reportResource := reports.Group("/:id")
//   reportResource.Use(middleware.ReportOwnershipRequired(reportRepo))
//   {
//       reportResource.GET("", handler.GetReport)
//       reportResource.GET("/download", handler.DownloadReport)
//       reportResource.DELETE("", handler.DeleteReport)
//   }
func ReportOwnershipRequired(reportRepo repository.ReportRepository) gin.HandlerFunc {
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

		// 3. 获取报表ID
		reportIDStr := c.Param("id")
		reportID, err := strconv.ParseUint(reportIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效的报表ID")
			c.Abort()
			return
		}

		// 4. 查询报表（验证存在性和所有权）
		report, err := reportRepo.FindByID(uint(reportID))
		if err != nil {
			response.InternalErrorWithDetail(c, "查询报表失败")
			c.Abort()
			return
		}
		if report == nil {
			response.NotFound(c, "报表")
			c.Abort()
			return
		}

		// 5. 验证所有权
		if !report.IsOwnedBy(user.ID) {
			response.ForbiddenWithMessage(c, "无权访问此报表")
			c.Abort()
			return
		}

		// 6. 将报表信息存入上下文（避免Handler重复查询）
		c.Set("report", report)
		c.Next()
	}
}
