// Package main 是应用程序的入口点
// 负责初始化所有组件并启动 HTTP 服务器
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/internal/infrastructure/persistence"
	"github.com/xinyiis/BridgeDefectDetectionSys/internal/interfaces/http/router"
	"github.com/xinyiis/BridgeDefectDetectionSys/pkg/config"
)

func main() {
	// 打印启动横幅
	printBanner()

	// 1. 加载配置
	log.Println("========== 初始化配置 ==========")
	cfg := config.LoadConfig()

	// 2. 初始化数据库
	log.Println("\n========== 初始化数据库 ==========")
	db := persistence.InitDatabase(cfg)
	defer persistence.CloseDatabase(db)

	// 3. 初始化路由
	log.Println("\n========== 初始化路由 ==========")
	r := router.SetupRouter(db, cfg)
	printRoutes(r)

	// 4. 启动 HTTP 服务器
	log.Println("\n========== 启动服务器 ==========")
	startServer(r, cfg)
}

// startServer 启动 HTTP 服务器（支持优雅关闭）
// 参数：
//   - r: Gin 引擎
//   - cfg: 配置对象
//
// 功能：
//   - 启动 HTTP 服务器
//   - 监听系统信号（SIGINT, SIGTERM）
//   - 收到退出信号时优雅关闭服务器
func startServer(r http.Handler, cfg *config.Config) {
	// 配置 HTTP 服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// 在 goroutine 中启动服务器
	go func() {
		log.Printf("✓ 服务器启动成功")
		log.Printf("✓ 监听地址: http://localhost%s\n", addr)
		log.Printf("✓ 健康检查: http://localhost%s/api/health\n", addr)
		log.Println("✓ 按 Ctrl+C 优雅退出")
		log.Println("\n========== 服务器运行中 ==========")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 开始优雅关闭
	log.Println("\n========== 正在关闭服务器 ==========")

	// 设置 5 秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️  服务器关闭超时: %v", err)
	}

	log.Println("✓ 服务器已安全关闭")
}

// printBanner 打印启动横幅
func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║     ____       _     _             ____       _              ║
║    | __ ) _ __(_) __| | __ _  ___|  _ \  ___| |_            ║
║    |  _ \| '__| |/ _  |/ _  |/ _ \ | | |/ _ \ __|           ║
║    | |_) | |  | | (_| | (_| |  __/ |_| |  __/ |_            ║
║    |____/|_|  |_|\__,_|\__, |\___|____/ \___|\__|           ║
║                        |___/                                 ║
║                                                              ║
║         桥梁缺陷检测系统 - Bridge Defect Detection           ║
║                    Version: 1.0.0                            ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

// printRoutes 打印已注册的路由（仅在 debug 模式下）
// 参数：
//   - r: Gin 引擎
func printRoutes(r interface{}) {
	// 注意：这里简化了路由打印逻辑
	// 在实际使用中，可以通过 gin.Engine 的 Routes() 方法获取所有路由
	log.Println("✓ 路由注册完成")
	log.Println("  公开路由:")
	log.Println("    GET  /api/health          - 健康检查")
	log.Println("    POST /api/register        - 用户注册（待实现）")
	log.Println("    POST /api/login           - 用户登录（待实现）")
	log.Println("")
	log.Println("  认证路由:")
	log.Println("    POST   /api/logout        - 退出登录（待实现）")
	log.Println("    GET    /api/user/info     - 获取用户信息（待实现）")
	log.Println("    GET    /api/bridges       - 桥梁列表（待实现）")
	log.Println("    POST   /api/detect/image  - 图片检测（待实现）")
	log.Println("    ...    更多接口待实现")
	log.Println("")
	log.Println("  管理员路由:")
	log.Println("    GET  /api/admin/users     - 用户管理（待实现）")
	log.Println("    GET  /api/admin/stats     - 全局统计（待实现）")
	log.Println("")
	log.Println("  静态文件:")
	log.Println("    /uploads/*                - 上传文件访问")
}

/*
========== 项目启动流程说明 ==========

1. 加载配置（config.yaml）
   ├─ 服务器配置（端口、运行模式）
   ├─ 数据库配置（连接串、连接池）
   ├─ Python 服务配置（算法服务地址）
   ├─ 文件上传配置（保存目录）
   └─ Session/CORS 配置

2. 初始化数据库
   ├─ 连接 MySQL
   ├─ 配置 GORM
   ├─ 自动迁移表结构（users, bridges, drones, defects）
   ├─ 创建默认管理员账户
   └─ 配置连接池

3. 初始化路由
   ├─ 配置全局中间件（CORS, Session）
   ├─ 注册公开路由（注册、登录）
   ├─ 注册认证路由（业务接口）
   ├─ 注册管理员路由（管理接口）
   └─ 配置静态文件服务

4. 启动 HTTP 服务器
   ├─ 监听指定端口（默认 8080）
   ├─ 等待客户端请求
   └─ 支持优雅关闭（Ctrl+C）

========== 下一步开发计划 ==========

基础设施已完成，接下来需要实现具体的业务逻辑：

1. 用户认证模块
   - 实现用户注册（密码 bcrypt 加密）
   - 实现用户登录（Session 管理）
   - 实现退出登录

2. 桥梁管理模块
   - CRUD 操作
   - 权限控制（用户只能管理自己的桥梁）

3. 缺陷检测模块
   - 图片上传和检测
   - 调用 Python 算法服务
   - 保存检测结果

4. 统计分析模块
   - 统计概览
   - 缺陷趋势分析

========== 启动方式 ==========

开发环境：
  cd src/backend
  go run cmd/server/main.go

生产环境：
  cd src/backend
  go build -o bridge-server cmd/server/main.go
  ./bridge-server

访问地址：
  健康检查: http://localhost:8080/api/health
  API 文档: http://localhost:8080/api/docs (待添加)

========================================
*/
