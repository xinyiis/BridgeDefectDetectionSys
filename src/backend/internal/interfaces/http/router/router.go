// Package router 提供路由管理功能
// 负责注册所有 HTTP 路由和中间件
package router

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/external"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/persistence"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/interfaces/http/handler"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/interfaces/http/middleware"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/config"
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
//   - /api/v1/auth       认证路由（公开，无需登录）
//   - /api/v1/user       用户相关（需要登录）
//   - /api/v1/admin      管理员路由（需要管理员权限）
//   - /api/v1/bridges    桥梁管理（需要登录）
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
	// ========== 依赖注入 ==========
	// 1. Repository 层
	userRepo := persistence.NewUserRepository(db)
	bridgeRepo := persistence.NewBridgeRepository(db)
	droneRepo := persistence.NewDroneRepository(db)
	defectRepo := persistence.NewDefectRepository(db)

	// 2. Service 层
	userService := service.NewUserService(userRepo)
	fileService := persistence.NewLocalFileStorage("./uploads")
	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)
	droneService := service.NewDroneService(db, droneRepo)

	// Python service (Mock or HTTP based on config)
	var pythonService service.PythonService
	if cfg.PythonService.Enabled {
		pythonService = external.NewHTTPPythonService(cfg.PythonService.URL)
	} else {
		pythonService = external.NewMockPythonService()
	}
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)
	statsService := persistence.NewStatsService(db)

	// 3. UseCase 层
	authUseCase := usecase.NewAuthUseCase(userService)
	userUseCase := usecase.NewUserUseCase(userService)
	bridgeUseCase := usecase.NewBridgeUseCase(bridgeService)
	droneUseCase := usecase.NewDroneUseCase(droneService)
	detectionUseCase := usecase.NewDetectionUseCase(defectService, bridgeService, pythonService, fileService)
	defectUseCase := usecase.NewDefectUseCase(defectService, fileService)
	statsUseCase := usecase.NewStatsUseCase(statsService)

	// 4. Handler 层
	authHandler := handler.NewAuthHandler(authUseCase)
	userHandler := handler.NewUserHandler(userUseCase)
	bridgeHandler := handler.NewBridgeHandler(bridgeUseCase, fileService)
	droneHandler := handler.NewDroneHandler(droneUseCase)
	detectionHandler := handler.NewDetectionHandler(detectionUseCase)
	defectHandler := handler.NewDefectHandler(defectUseCase)
	statsHandler := handler.NewStatsHandler(statsUseCase)

	// ========== 路由注册 ==========
	// API 路由组（/api/v1）
	api := r.Group("/api/v1")

	// 1. 公开路由（无需登录）- 认证相关
	registerPublicRoutes(api, authHandler)

	// 2. 认证路由（需要登录）
	auth := api.Group("")
	auth.Use(middleware.AuthRequired(db))
	registerAuthRoutes(auth, authHandler, userHandler, bridgeHandler, droneHandler, detectionHandler, defectHandler, statsHandler, bridgeRepo, droneRepo, defectService, cfg)

	// 3. 管理员路由（需要管理员权限）
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(db))
	admin.Use(middleware.AdminRequired())
	registerAdminRoutes(admin, userHandler)
}

// registerPublicRoutes 注册公开路由
// 这些接口无需登录即可访问
func registerPublicRoutes(r *gin.RouterGroup, authHandler *handler.AuthHandler) {
	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Bridge Detection System API is running",
			"version": "v1",
		})
	})

	// ========== 用户认证（/auth前缀）==========
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register) // POST /api/v1/auth/register
		auth.POST("/login", authHandler.Login)       // POST /api/v1/auth/login
	}
}

// registerAuthRoutes 注册认证路由
// 这些接口需要用户登录后才能访问
func registerAuthRoutes(r *gin.RouterGroup, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, bridgeHandler *handler.BridgeHandler, droneHandler *handler.DroneHandler, detectionHandler *handler.DetectionHandler, defectHandler *handler.DefectHandler, statsHandler *handler.StatsHandler, bridgeRepo repository.BridgeRepository, droneRepo repository.DroneRepository, defectService *service.DefectService, cfg *config.Config) {
	// ========== 用户认证相关 ==========
	auth := r.Group("/auth")
	{
		auth.POST("/logout", authHandler.Logout) // POST /api/v1/auth/logout
	}

	// ========== 用户个人信息 ==========
	user := r.Group("/user")
	{
		user.GET("/profile", userHandler.GetUserInfo)     // GET /api/v1/user/profile
		user.PUT("/profile", userHandler.UpdateUserInfo)  // PUT /api/v1/user/profile
	}

	// ========== 桥梁管理 ==========
	bridges := r.Group("/bridges")
	{
		// 列表和创建不需要所有权验证
		bridges.GET("", bridgeHandler.ListBridges)    // 获取桥梁列表
		bridges.POST("", bridgeHandler.CreateBridge)  // 创建桥梁

		// 单个资源操作需要所有权验证
		bridgeResource := bridges.Group("/:id")
		bridgeResource.Use(middleware.BridgeOwnershipRequired(bridgeRepo))
		{
			bridgeResource.GET("", bridgeHandler.GetBridge)       // 获取桥梁详情
			bridgeResource.PUT("", bridgeHandler.UpdateBridge)    // 更新桥梁
			bridgeResource.DELETE("", bridgeHandler.DeleteBridge) // 删除桥梁
		}
	}

	// ========== 无人机管理 ==========
	drones := r.Group("/drones")
	{
		// 列表和创建不需要所有权验证
		drones.GET("", droneHandler.ListDrones)    // GET /api/v1/drones
		drones.POST("", droneHandler.CreateDrone)  // POST /api/v1/drones

		// 单个资源操作需要所有权验证
		droneResource := drones.Group("/:id")
		droneResource.Use(middleware.DroneOwnershipRequired(droneRepo))
		{
			droneResource.GET("", droneHandler.GetDrone)       // GET /api/v1/drones/:id
			droneResource.PUT("", droneHandler.UpdateDrone)    // PUT /api/v1/drones/:id
			droneResource.DELETE("", droneHandler.DeleteDrone) // DELETE /api/v1/drones/:id
		}
	}

	// ========== 缺陷检测 ==========
	detection := r.Group("/detection")
	{
		detection.POST("/upload", detectionHandler.UploadAndDetect) // POST /api/v1/detection/upload
	}

	// ========== 缺陷管理 ==========
	defects := r.Group("/defects")
	{
		defects.GET("", defectHandler.ListDefects) // GET /api/v1/defects

		defectResource := defects.Group("/:id")
		defectResource.Use(middleware.DefectOwnershipRequired(defectService))
		{
			defectResource.GET("", defectHandler.GetDefect)       // GET /api/v1/defects/:id
			defectResource.DELETE("", defectHandler.DeleteDefect) // DELETE /api/v1/defects/:id
		}
	}

	// ========== 统计分析 ==========
	stats := r.Group("/stats")
	{
		stats.GET("/overview", statsHandler.GetOverview)                        // GET /api/v1/stats/overview
		stats.GET("/defect-types", statsHandler.GetDefectTypeDistribution)      // GET /api/v1/stats/defect-types
		stats.GET("/defect-trend", statsHandler.GetDefectTrend)                 // GET /api/v1/stats/defect-trend
		stats.GET("/bridge-ranking", statsHandler.GetBridgeRanking)             // GET /api/v1/stats/bridge-ranking
		stats.GET("/recent-detections", statsHandler.GetRecentDetections)       // GET /api/v1/stats/recent-detections
		stats.GET("/high-risk-alerts", statsHandler.GetHighRiskAlerts)          // GET /api/v1/stats/high-risk-alerts
	}
}

// registerAdminRoutes 注册管理员路由
// 这些接口需要管理员权限才能访问
func registerAdminRoutes(r *gin.RouterGroup, userHandler *handler.UserHandler) {
	// ========== 用户管理 ==========
	r.GET("/users", userHandler.ListUsers)              // 获取所有用户（分页）
	r.GET("/users/:id", userHandler.GetUserByID)        // 获取用户详情
	r.DELETE("/users/:id", userHandler.DeleteUser)      // 删除用户
	r.POST("/users/promote", userHandler.PromoteToAdmin) // 提升用户为管理员

	// ========== 全局统计（待实现）==========
	// r.GET("/stats/global", handler.GetGlobalStats(db))      // 获取全局统计数据
	// r.GET("/stats/users", handler.GetUserStats(db))         // 获取用户统计
	// r.GET("/stats/bridges", handler.GetBridgeStats(db))     // 获取桥梁统计

	// ========== 系统管理（待实现）==========
	// r.GET("/system/info", handler.GetSystemInfo)            // 获取系统信息
	// r.POST("/system/backup", handler.BackupDatabase(db))    // 备份数据库
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
