// Package router 提供路由管理功能
// 负责注册所有 HTTP 路由和中间件
package router

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/internal/interfaces/http/middleware"
	"github.com/xinyiis/BridgeDefectDetectionSys/pkg/config"
	"gorm.io/gorm"
)

// SetupRouter 设置并返回 Gin 路由引擎
// 配置所有中间件和路由
// 参数：
//   - db: GORM 数据库连接对象
//   - cfg: 配置对象
//
// 返回值：
//   - *gin.Engine: 配置完成的 Gin 引擎
//
// 路由结构：
//   - /api                公开路由（无需登录）
//   - /api/auth          认证路由（需要登录）
//   - /api/admin         管理员路由（需要管理员权限）
//   - /uploads           静态文件（检测结果图片）
//
// 使用示例：
//
//	r := router.SetupRouter(db, cfg)
//	r.Run(":8080")
func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	// 1. 设置 Gin 运行模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 2. 创建 Gin 引擎
	r := gin.Default()

	// 3. 配置全局中间件
	setupGlobalMiddleware(r, cfg)

	// 4. 配置静态文件服务
	setupStaticFiles(r, cfg)

	// 5. 配置 API 路由
	setupAPIRoutes(r, db, cfg)

	return r
}

// setupGlobalMiddleware 配置全局中间件
// 包括 CORS、Session 等
func setupGlobalMiddleware(r *gin.Engine, cfg *config.Config) {
	// CORS 跨域中间件
	r.Use(middleware.CORSMiddleware(cfg))

	// Session 中间件
	store := cookie.NewStore([]byte(cfg.Session.Secret))
	store.Options(sessions.Options{
		MaxAge:   cfg.Session.MaxAge,    // Session 有效期（秒）
		Path:     "/",                    // Cookie 路径
		HttpOnly: true,                   // 防止 XSS 攻击
		Secure:   false,                  // 开发环境使用 HTTP，生产环境应改为 true（HTTPS）
		SameSite: 2,                      // Lax 模式，允许部分跨站请求
	})
	r.Use(sessions.Sessions(cfg.Session.CookieName, store))
}

// setupStaticFiles 配置静态文件服务
// 用于访问上传的图片和检测结果
func setupStaticFiles(r *gin.Engine, cfg *config.Config) {
	// 访问路径: http://localhost:8080/uploads/images/xxx.jpg
	r.Static("/uploads", "./uploads")
}

// setupAPIRoutes 配置 API 路由
// 将路由分组为：公开、认证、管理员三类
func setupAPIRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// API 路由组（/api）
	api := r.Group("/api")

	// 1. 公开路由（无需登录）
	registerPublicRoutes(api, db)

	// 2. 认证路由（需要登录）
	auth := api.Group("")
	auth.Use(middleware.AuthRequired(db))
	registerAuthRoutes(auth, db, cfg)

	// 3. 管理员路由（需要管理员权限）
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(db))
	admin.Use(middleware.AdminRequired())
	registerAdminRoutes(admin, db)
}

// registerPublicRoutes 注册公开路由
// 这些接口无需登录即可访问
func registerPublicRoutes(r *gin.RouterGroup, db *gorm.DB) {
	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "Bridge Detection System API is running",
		})
	})

	// 用户认证相关接口（将在后续实现）
	// r.POST("/register", handler.Register(db))  // 用户注册
	// r.POST("/login", handler.Login(db))        // 用户登录

	// 注意：实际的 handler 实现将在下一步完成
	// 这里先保留注释，说明接口规划
}

// registerAuthRoutes 注册认证路由
// 这些接口需要用户登录后才能访问
func registerAuthRoutes(r *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	// ========== 用户相关 ==========
	// r.POST("/logout", handler.Logout)                    // 退出登录
	// r.GET("/user/info", handler.GetUserInfo)             // 获取当前用户信息
	// r.PUT("/user/info", handler.UpdateUserInfo(db))      // 更新用户信息
	// r.PUT("/user/password", handler.ChangePassword(db))  // 修改密码

	// ========== 桥梁管理 ==========
	// r.GET("/bridges", handler.GetBridges(db))                              // 获取桥梁列表
	// r.POST("/bridges", handler.CreateBridge(db))                           // 创建桥梁
	// r.GET("/bridges/:id", handler.GetBridge(db))                           // 获取桥梁详情
	// r.PUT("/bridges/:id", handler.UpdateBridge(db))                        // 更新桥梁
	// r.DELETE("/bridges/:id", handler.DeleteBridge(db))                     // 删除桥梁
	// r.GET("/bridges/:id/defects", handler.GetBridgeDefects(db))            // 获取桥梁的缺陷列表

	// ========== 无人机管理 ==========
	// r.GET("/drones", handler.GetDrones(db))                                // 获取无人机列表
	// r.POST("/drones", handler.CreateDrone(db))                             // 创建无人机
	// r.GET("/drones/:id", handler.GetDrone(db))                             // 获取无人机详情
	// r.PUT("/drones/:id", handler.UpdateDrone(db))                          // 更新无人机
	// r.DELETE("/drones/:id", handler.DeleteDrone(db))                       // 删除无人机

	// ========== 缺陷检测 ==========
	// r.POST("/detect/image", handler.DetectImage(db, cfg))                  // 图片检测
	// r.POST("/detect/video/start", handler.StartVideoDetection(db, cfg))    // 开始视频检测
	// r.POST("/detect/video/stop", handler.StopVideoDetection(db))           // 停止视频检测
	// r.GET("/defects", handler.GetDefects(db))                              // 获取缺陷列表
	// r.GET("/defects/:id", handler.GetDefect(db))                           // 获取缺陷详情
	// r.DELETE("/defects/:id", handler.DeleteDefect(db))                     // 删除缺陷记录

	// ========== 统计分析 ==========
	// r.GET("/stats/overview", handler.GetStatsOverview(db))                 // 获取统计概览
	// r.GET("/stats/defect-trends", handler.GetDefectTrends(db))             // 获取缺陷趋势
	// r.GET("/stats/bridge/:id", handler.GetBridgeStats(db))                 // 获取单个桥梁统计

	// 注意：Handler 实现将在后续完成
	// 这里先保留注释，展示完整的接口设计
}

// registerAdminRoutes 注册管理员路由
// 这些接口需要管理员权限才能访问
func registerAdminRoutes(r *gin.RouterGroup, db *gorm.DB) {
	// ========== 用户管理 ==========
	// r.GET("/users", handler.GetAllUsers(db))                // 获取所有用户
	// r.GET("/users/:id", handler.GetUser(db))                // 获取用户详情
	// r.PUT("/users/:id/role", handler.UpdateUserRole(db))    // 修改用户角色
	// r.DELETE("/users/:id", handler.DeleteUser(db))          // 删除用户

	// ========== 全局统计 ==========
	// r.GET("/stats/global", handler.GetGlobalStats(db))      // 获取全局统计数据
	// r.GET("/stats/users", handler.GetUserStats(db))         // 获取用户统计
	// r.GET("/stats/bridges", handler.GetBridgeStats(db))     // 获取桥梁统计

	// ========== 系统管理 ==========
	// r.GET("/system/info", handler.GetSystemInfo)            // 获取系统信息
	// r.POST("/system/backup", handler.BackupDatabase(db))    // 备份数据库

	// 注意：Handler 实现将在后续完成
}

// GetRouterSummary 获取路由摘要信息
// 用于调试和文档生成，返回所有已注册的路由
// 参数：
//   - r: Gin 引擎
//
// 返回值：
//   - []gin.RouteInfo: 路由信息列表
//
// 使用示例：
//
//	routes := router.GetRouterSummary(r)
//	for _, route := range routes {
//	    fmt.Printf("%s %s\n", route.Method, route.Path)
//	}
func GetRouterSummary(r *gin.Engine) []gin.RouteInfo {
	return r.Routes()
}
