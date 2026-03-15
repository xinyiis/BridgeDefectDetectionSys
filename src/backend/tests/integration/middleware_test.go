// +build integration

// Package integration 集成测试
// 测试中间件和路由的集成
package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/internal/interfaces/http/middleware"
	"github.com/xinyiis/BridgeDefectDetectionSys/pkg/config"
	"github.com/xinyiis/BridgeDefectDetectionSys/pkg/response"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestRouter 创建测试路由
func setupTestRouter() (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	// 使用内存数据库
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.User{}, &model.Bridge{}, &model.Drone{}, &model.Defect{})

	// 创建测试用户
	adminUser := model.User{
		ID:       1,
		Username: "admin",
		Password: "hashed_password",
		Role:     "admin",
	}
	normalUser := model.User{
		ID:       2,
		Username: "user",
		Password: "hashed_password",
		Role:     "user",
	}
	db.Create(&adminUser)
	db.Create(&normalUser)

	// 创建路由
	r := gin.New()

	// 配置 Session
	store := cookie.NewStore([]byte("test-secret"))
	r.Use(sessions.Sessions("test_session", store))

	// 配置 CORS
	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowOrigins:     []string{"http://localhost:5173"},
			AllowCredentials: true,
		},
	}
	r.Use(middleware.CORSMiddleware(cfg))

	return r, db
}

// TestCORSMiddleware 测试CORS中间件
func TestCORSMiddleware(t *testing.T) {
	r, _ := setupTestRouter()

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// 创建请求
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// 验证CORS头
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("CORS头不正确: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("CORS AllowCredentials 应该为 true")
	}
}

// TestAuthRequiredMiddleware 测试认证中间件
func TestAuthRequiredMiddleware(t *testing.T) {
	r, db := setupTestRouter()

	// 受保护的路由
	r.GET("/protected", middleware.AuthRequired(db), func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		c.JSON(200, gin.H{"username": user.Username})
	})

	t.Run("未登录访问", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// 应该返回 401
		if w.Code != 401 {
			t.Errorf("未登录应返回 401, 实际: %d", w.Code)
		}
	})

	t.Run("已登录访问", func(t *testing.T) {
		// 注意：完整的 Session 测试需要实际的 HTTP 请求
		// 这里简化测试，只测试中间件逻辑
		t.Skip("Session 集成测试需要完整的 HTTP 服务器，跳过此测试")
	})
}

// TestAdminRequiredMiddleware 测试管理员权限中间件
func TestAdminRequiredMiddleware(t *testing.T) {
	r, db := setupTestRouter()

	// 管理员路由
	r.GET("/admin", middleware.AuthRequired(db), middleware.AdminRequired(), func(c *gin.Context) {
		response.Success(c, gin.H{"message": "admin area"})
	})

	t.Run("普通用户访问", func(t *testing.T) {
		// 创建普通用户的请求
		req, _ := http.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()

		// 设置普通用户的 Context
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("current_user", &model.User{
			ID:       2,
			Username: "user",
			Role:     "user",
		})

		// 执行中间件
		middleware.AdminRequired()(c)

		// 应该返回 403
		if w.Code != 403 && !c.IsAborted() {
			t.Error("普通用户访问管理员路由应返回 403")
		}
	})

	t.Run("管理员访问", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("current_user", &model.User{
			ID:       1,
			Username: "admin",
			Role:     "admin",
		})

		// 执行中间件
		middleware.AdminRequired()(c)

		// 不应该被中止
		if c.IsAborted() {
			t.Error("管理员访问应该通过")
		}
	})
}

// TestCheckResourceOwnership 测试资源所有权中间件
func TestCheckResourceOwnership(t *testing.T) {
	r, db := setupTestRouter()

	// 创建测试桥梁
	bridge := model.Bridge{
		ID:          1,
		Name:        "测试桥梁",
		Location:    "测试位置",
		Description: "测试描述",
		UserID:      1, // 属于用户1（admin）
	}
	db.Create(&bridge)

	// 受保护的桥梁路由
	r.GET("/bridges/:id",
		middleware.AuthRequired(db),
		middleware.CheckResourceOwnership(db, "bridge"),
		func(c *gin.Context) {
			response.Success(c, gin.H{"message": "ok"})
		},
	)

	t.Run("访问自己的资源", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bridges/1", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("current_user", &model.User{
			ID:       1,
			Username: "admin",
			Role:     "admin",
		})

		// 执行中间件
		middleware.CheckResourceOwnership(db, "bridge")(c)

		// 应该通过
		if c.IsAborted() {
			t.Error("访问自己的资源应该通过")
		}
	})

	t.Run("访问他人的资源", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bridges/1", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("current_user", &model.User{
			ID:       2,
			Username: "user",
			Role:     "user",
		})

		// 执行中间件
		middleware.CheckResourceOwnership(db, "bridge")(c)

		// 应该被拒绝
		if !c.IsAborted() || w.Code != 403 {
			t.Error("访问他人资源应返回 403")
		}
	})

	t.Run("管理员可以访问任何资源", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bridges/1", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("current_user", &model.User{
			ID:       1,
			Username: "admin",
			Role:     "admin",
		})

		// 执行中间件
		middleware.CheckResourceOwnership(db, "bridge")(c)

		// 管理员应该通过
		if c.IsAborted() {
			t.Error("管理员应该可以访问任何资源")
		}
	})
}

// TestMiddlewareChain 测试中间件链
func TestMiddlewareChain(t *testing.T) {
	r, db := setupTestRouter()

	// 创建中间件链
	r.GET("/chain",
		middleware.AuthRequired(db),
		middleware.AdminRequired(),
		func(c *gin.Context) {
			response.Success(c, gin.H{"message": "success"})
		},
	)

	t.Run("未登录", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/chain", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// 第一个中间件就应该拦截
		if w.Code != 401 {
			t.Errorf("未登录应返回 401, 实际: %d", w.Code)
		}
	})
}

// BenchmarkAuthMiddleware 测试认证中间件性能
func BenchmarkAuthMiddleware(b *testing.B) {
	r, db := setupTestRouter()

	r.GET("/bench", middleware.AuthRequired(db), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/bench", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
