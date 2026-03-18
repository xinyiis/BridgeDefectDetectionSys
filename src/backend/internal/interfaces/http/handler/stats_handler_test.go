// Package handler_test 统计模块集成测试
package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/persistence"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/interfaces/http/handler"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/interfaces/http/middleware"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/cache"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupStatsTestDB 创建统计测试数据库
func setupStatsTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&model.User{}, &model.Bridge{}, &model.Drone{}, &model.Defect{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// setupStatsTestRouter 创建统计测试路由
func setupStatsTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 配置Session
	store := cookie.NewStore([]byte("test-secret-key"))
	r.Use(sessions.Sessions("test_session", store))

	// 依赖注入
	userRepo := persistence.NewUserRepository(db)
	statsService := persistence.NewStatsService(db)

	userService := service.NewUserService(userRepo)
	authUseCase := usecase.NewAuthUseCase(userService)
	statsUseCase := usecase.NewStatsUseCase(statsService)

	authHandler := handler.NewAuthHandler(authUseCase)
	statsHandler := handler.NewStatsHandler(statsUseCase)

	// 注册路由
	api := r.Group("/api/v1")

	// 公开路由
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	// 认证路由
	auth := api.Group("")
	auth.Use(middleware.AuthRequired(db))
	{
		// 统计模块
		stats := auth.Group("/stats")
		{
			stats.GET("/overview", statsHandler.GetOverview)
			stats.GET("/defect-types", statsHandler.GetDefectTypeDistribution)
			stats.GET("/defect-trend", statsHandler.GetDefectTrend)
			stats.GET("/bridge-ranking", statsHandler.GetBridgeRanking)
			stats.GET("/recent-detections", statsHandler.GetRecentDetections)
			stats.GET("/high-risk-alerts", statsHandler.GetHighRiskAlerts)
		}
	}

	return r
}

// createStatsTestUser 创建测试用户
func createStatsTestUser(db *gorm.DB, username, role string) *model.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		RealName: username,
		Email:    username + "@example.com",
		Role:     role,
	}
	db.Create(user)
	return user
}

// createStatsTestBridge 创建测试桥梁
func createStatsTestBridge(db *gorm.DB, userID uint, bridgeName string) *model.Bridge {
	bridge := &model.Bridge{
		BridgeName: bridgeName,
		BridgeCode: fmt.Sprintf("BRIDGE_%d_%d", userID, time.Now().UnixNano()),
		UserID:     userID,
		Address:    "测试地址",
		Longitude:  118.78,
		Latitude:   32.04,
		BridgeType: "梁桥",
	}
	db.Create(bridge)
	return bridge
}

// createStatsTestDrone 创建测试无人机
func createStatsTestDrone(db *gorm.DB, userID uint, droneName string) *model.Drone {
	drone := &model.Drone{
		Name:   droneName,
		UserID: userID,
		Model:  "Mavic 3",
	}
	db.Create(drone)
	return drone
}

// createStatsTestDefect 创建测试缺陷
func createStatsTestDefect(db *gorm.DB, bridgeID uint, defectType string, confidence float64, area float64, detectedAt time.Time) *model.Defect {
	defect := &model.Defect{
		BridgeID:   bridgeID,
		DefectType: defectType,
		ImagePath:  fmt.Sprintf("images/test_%d.jpg", time.Now().UnixNano()),
		ResultPath: "results/test_result.jpg",
		Confidence: confidence,
		Area:       area,
		DetectedAt: detectedAt,
	}
	db.Create(defect)
	return defect
}

// loginStatsTest 登录并获取Cookies
func loginStatsTest(t *testing.T, router *gin.Engine, username, password string) []*http.Cookie {
	loginBody := map[string]interface{}{
		"username": username,
		"password": password,
	}
	bodyBytes, _ := json.Marshal(loginBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Login failed with status %d: %s", w.Code, w.Body.String())
	}

	return w.Result().Cookies()
}

// clearStatsCache 清除统计缓存
func clearStatsCache() {
	cache.StatsCache.Clear()
}

// TestGetOverview_Success 测试获取概览统计成功
func TestGetOverview_Success(t *testing.T) {
	db := setupStatsTestDB(t)
	router := setupStatsTestRouter(db)
	defer clearStatsCache()

	// 创建测试数据
	user := createStatsTestUser(db, "testuser", "user")
	bridge := createStatsTestBridge(db, user.ID, "测试桥梁")
	createStatsTestDrone(db, user.ID, "测试无人机")

	// 创建今天的缺陷
	today := time.Now()
	createStatsTestDefect(db, bridge.ID, "裂缝", 0.95, 0.05, today)
	createStatsTestDefect(db, bridge.ID, "剥落", 0.88, 0.03, today)

	// 创建昨天的缺陷
	yesterday := today.AddDate(0, 0, -1)
	createStatsTestDefect(db, bridge.ID, "裂缝", 0.90, 0.02, yesterday)

	// 登录并请求
	cookies := loginStatsTest(t, router, "testuser", "123456")
	req := httptest.NewRequest("GET", "/api/v1/stats/overview", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["code"].(float64) != 200 {
		t.Errorf("Expected code 200, got %v", response["code"])
	}

	// 验证返回数据
	data := response["data"].(map[string]interface{})
	if int(data["bridge_count"].(float64)) != 1 {
		t.Errorf("Expected bridge_count 1, got %v", data["bridge_count"])
	}
	if int(data["drone_count"].(float64)) != 1 {
		t.Errorf("Expected drone_count 1, got %v", data["drone_count"])
	}
	if int(data["defect_count"].(float64)) != 3 {
		t.Errorf("Expected defect_count 3, got %v", data["defect_count"])
	}
	if int(data["today_defects"].(float64)) != 2 {
		t.Errorf("Expected today_defects 2, got %v", data["today_defects"])
	}

	// 验证趋势方向
	trendDirection := data["trend_direction"].(string)
	if trendDirection != "up" && trendDirection != "down" && trendDirection != "stable" {
		t.Errorf("Invalid trend_direction: %v", trendDirection)
	}
}

// TestGetOverview_PermissionFilter 测试概览统计权限过滤
func TestGetOverview_PermissionFilter(t *testing.T) {
	db := setupStatsTestDB(t)
	router := setupStatsTestRouter(db)
	defer clearStatsCache()

	// 创建两个用户和各自的数据
	user1 := createStatsTestUser(db, "user1", "user")
	user2 := createStatsTestUser(db, "user2", "user")
	createStatsTestUser(db, "admin", "admin")

	bridge1 := createStatsTestBridge(db, user1.ID, "用户1的桥梁")
	bridge2 := createStatsTestBridge(db, user2.ID, "用户2的桥梁")

	createStatsTestDefect(db, bridge1.ID, "裂缝", 0.95, 0.05, time.Now())
	createStatsTestDefect(db, bridge1.ID, "剥落", 0.88, 0.03, time.Now())
	createStatsTestDefect(db, bridge2.ID, "裂缝", 0.90, 0.02, time.Now())

	// 用户1登录，只能看到自己的数据
	cookies1 := loginStatsTest(t, router, "user1", "123456")
	req1 := httptest.NewRequest("GET", "/api/v1/stats/overview", nil)
	for _, cookie := range cookies1 {
		req1.AddCookie(cookie)
	}
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	var response1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &response1)
	data1 := response1["data"].(map[string]interface{})

	if int(data1["bridge_count"].(float64)) != 1 {
		t.Errorf("User1 expected bridge_count 1, got %v", data1["bridge_count"])
	}
	if int(data1["defect_count"].(float64)) != 2 {
		t.Errorf("User1 expected defect_count 2, got %v", data1["defect_count"])
	}

	// 管理员登录，能看到所有数据
	cookiesAdmin := loginStatsTest(t, router, "admin", "123456")
	reqAdmin := httptest.NewRequest("GET", "/api/v1/stats/overview", nil)
	for _, cookie := range cookiesAdmin {
		reqAdmin.AddCookie(cookie)
	}
	wAdmin := httptest.NewRecorder()
	router.ServeHTTP(wAdmin, reqAdmin)

	var responseAdmin map[string]interface{}
	json.Unmarshal(wAdmin.Body.Bytes(), &responseAdmin)
	dataAdmin := responseAdmin["data"].(map[string]interface{})

	if int(dataAdmin["bridge_count"].(float64)) != 2 {
		t.Errorf("Admin expected bridge_count 2, got %v", dataAdmin["bridge_count"])
	}
	if int(dataAdmin["defect_count"].(float64)) != 3 {
		t.Errorf("Admin expected defect_count 3, got %v", dataAdmin["defect_count"])
	}
}

// TestGetDefectTypeDistribution_Success 测试缺陷类型分布统计
func TestGetDefectTypeDistribution_Success(t *testing.T) {
	db := setupStatsTestDB(t)
	router := setupStatsTestRouter(db)
	defer clearStatsCache()

	user := createStatsTestUser(db, "testuser", "user")
	bridge := createStatsTestBridge(db, user.ID, "测试桥梁")

	// 创建不同类型的缺陷
	createStatsTestDefect(db, bridge.ID, "裂缝", 0.95, 0.05, time.Now())
	createStatsTestDefect(db, bridge.ID, "裂缝", 0.92, 0.04, time.Now())
	createStatsTestDefect(db, bridge.ID, "裂缝", 0.90, 0.03, time.Now())
	createStatsTestDefect(db, bridge.ID, "剥落", 0.88, 0.03, time.Now())
	createStatsTestDefect(db, bridge.ID, "剥落", 0.85, 0.02, time.Now())
	createStatsTestDefect(db, bridge.ID, "钢筋锈蚀", 0.93, 0.06, time.Now())

	cookies := loginStatsTest(t, router, "testuser", "123456")
	req := httptest.NewRequest("GET", "/api/v1/stats/defect-types?days=30", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})

	if int(data["total"].(float64)) != 6 {
		t.Errorf("Expected total 6, got %v", data["total"])
	}

	distribution := data["distribution"].([]interface{})
	if len(distribution) != 3 {
		t.Errorf("Expected 3 defect types, got %d", len(distribution))
	}

	// 验证第一个类型（应该是裂缝，数量最多）
	firstType := distribution[0].(map[string]interface{})
	if firstType["defect_type"].(string) != "裂缝" {
		t.Errorf("Expected first type '裂缝', got %v", firstType["defect_type"])
	}
	if int(firstType["count"].(float64)) != 3 {
		t.Errorf("Expected count 3, got %v", firstType["count"])
	}
	if firstType["percentage"].(float64) < 49 || firstType["percentage"].(float64) > 51 {
		t.Errorf("Expected percentage ~50, got %v", firstType["percentage"])
	}
}

// TestGetDefectTrend_Success 测试缺陷趋势统计
func TestGetDefectTrend_Success(t *testing.T) {
	db := setupStatsTestDB(t)
	router := setupStatsTestRouter(db)
	defer clearStatsCache()

	user := createStatsTestUser(db, "testuser", "user")
	bridge := createStatsTestBridge(db, user.ID, "测试桥梁")

	// 创建7天的缺陷数据
	today := time.Now()
	for i := 0; i < 7; i++ {
		date := today.AddDate(0, 0, -i)
		count := i + 1 // 越早天数越多
		for j := 0; j < count; j++ {
			createStatsTestDefect(db, bridge.ID, "裂缝", 0.90, 0.03, date)
		}
	}

	cookies := loginStatsTest(t, router, "testuser", "123456")
	req := httptest.NewRequest("GET", "/api/v1/stats/defect-trend?days=7&granularity=day", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})

	if data["period"].(string) != "7days" {
		t.Errorf("Expected period '7days', got %v", data["period"])
	}

	trend := data["trend"].([]interface{})
	if len(trend) == 0 {
		t.Errorf("Expected trend data, got empty array")
	}

	// 验证总数
	total := int(data["total"].(float64))
	expectedTotal := 1 + 2 + 3 + 4 + 5 + 6 + 7
	if total != expectedTotal {
		t.Errorf("Expected total %d, got %d", expectedTotal, total)
	}
}

// TestGetBridgeRanking_Success 测试桥梁健康度排名
func TestGetBridgeRanking_Success(t *testing.T) {
	db := setupStatsTestDB(t)
	router := setupStatsTestRouter(db)
	defer clearStatsCache()

	user := createStatsTestUser(db, "testuser", "user")

	// 创建3座桥梁，缺陷数量不同
	bridge1 := createStatsTestBridge(db, user.ID, "健康桥梁")
	bridge2 := createStatsTestBridge(db, user.ID, "一般桥梁")
	bridge3 := createStatsTestBridge(db, user.ID, "危险桥梁")

	// 健康桥梁：1个低危缺陷
	createStatsTestDefect(db, bridge1.ID, "裂缝", 0.70, 0.01, time.Now())

	// 一般桥梁：3个中危缺陷
	createStatsTestDefect(db, bridge2.ID, "裂缝", 0.85, 0.03, time.Now())
	createStatsTestDefect(db, bridge2.ID, "剥落", 0.88, 0.02, time.Now())
	createStatsTestDefect(db, bridge2.ID, "钢筋锈蚀", 0.82, 0.04, time.Now())

	// 危险桥梁：5个缺陷，含2个高危
	createStatsTestDefect(db, bridge3.ID, "裂缝", 0.96, 0.12, time.Now()) // 高危
	createStatsTestDefect(db, bridge3.ID, "剥落", 0.92, 0.08, time.Now()) // 高危
	createStatsTestDefect(db, bridge3.ID, "钢筋锈蚀", 0.85, 0.03, time.Now())
	createStatsTestDefect(db, bridge3.ID, "混凝土开裂", 0.88, 0.04, time.Now())
	createStatsTestDefect(db, bridge3.ID, "表面损伤", 0.80, 0.02, time.Now())

	cookies := loginStatsTest(t, router, "testuser", "123456")
	req := httptest.NewRequest("GET", "/api/v1/stats/bridge-ranking?limit=10&order=worst", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})

	ranking := data["ranking"].([]interface{})
	if len(ranking) != 3 {
		t.Errorf("Expected 3 bridges, got %d", len(ranking))
	}

	// 验证排名顺序（最差的在前）
	firstBridge := ranking[0].(map[string]interface{})
	if firstBridge["bridge_name"].(string) != "危险桥梁" {
		t.Errorf("Expected first bridge '危险桥梁', got %v", firstBridge["bridge_name"])
	}
	if int(firstBridge["defect_count"].(float64)) != 5 {
		t.Errorf("Expected defect_count 5, got %v", firstBridge["defect_count"])
	}
	if int(firstBridge["high_risk_count"].(float64)) != 2 {
		t.Errorf("Expected high_risk_count 2, got %v", firstBridge["high_risk_count"])
	}

	// 验证健康等级（评分85 = 良好）
	healthLevel := firstBridge["health_level"].(string)
	if healthLevel != "良好" && healthLevel != "一般" && healthLevel != "较差" {
		t.Errorf("Expected health_level '良好', '一般' or '较差', got %v", healthLevel)
	}
}

// TestGetBridgeRanking_EmptyData 测试无数据情况
func TestGetBridgeRanking_EmptyData(t *testing.T) {
	db := setupStatsTestDB(t)
	router := setupStatsTestRouter(db)
	defer clearStatsCache()

	// 创建用户但不创建桥梁
	createStatsTestUser(db, "testuser", "user")

	cookies := loginStatsTest(t, router, "testuser", "123456")
	req := httptest.NewRequest("GET", "/api/v1/stats/bridge-ranking?limit=10&order=worst", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})

	// ranking 可能是 nil 或空数组
	ranking, ok := data["ranking"].([]interface{})
	if !ok || ranking == nil {
		ranking = []interface{}{}
	}
	if len(ranking) != 0 {
		t.Errorf("Expected empty ranking, got %d bridges", len(ranking))
	}
}

// TestGetRecentDetections_Success 测试最近检测记录
func TestGetRecentDetections_Success(t *testing.T) {
	db := setupStatsTestDB(t)
	router := setupStatsTestRouter(db)
	defer clearStatsCache()

	user := createStatsTestUser(db, "testuser", "user")
	bridge1 := createStatsTestBridge(db, user.ID, "桥梁1")
	bridge2 := createStatsTestBridge(db, user.ID, "桥梁2")

	// 创建不同时间的检测记录（相同image_path视为一次检测）
	now := time.Now()
	imagePath1 := "images/test_1.jpg"
	imagePath2 := "images/test_2.jpg"
	imagePath3 := "images/test_3.jpg"

	// 检测1：桥梁1，3个缺陷
	for i := 0; i < 3; i++ {
		defect := createStatsTestDefect(db, bridge1.ID, "裂缝", 0.90, 0.03, now.Add(-1*time.Hour))
		db.Model(defect).Update("image_path", imagePath1)
	}

	// 检测2：桥梁2，2个缺陷
	for i := 0; i < 2; i++ {
		defect := createStatsTestDefect(db, bridge2.ID, "剥落", 0.85, 0.02, now.Add(-2*time.Hour))
		db.Model(defect).Update("image_path", imagePath2)
	}

	// 检测3：桥梁1，1个缺陷
	defect := createStatsTestDefect(db, bridge1.ID, "钢筋锈蚀", 0.92, 0.05, now.Add(-3*time.Hour))
	db.Model(defect).Update("image_path", imagePath3)

	cookies := loginStatsTest(t, router, "testuser", "123456")
	req := httptest.NewRequest("GET", "/api/v1/stats/recent-detections?limit=10", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})

	detections := data["detections"].([]interface{})
	if len(detections) != 3 {
		t.Errorf("Expected 3 detections, got %d", len(detections))
	}

	// 验证第一条记录（最新的）
	firstDetection := detections[0].(map[string]interface{})
	if firstDetection["image_path"].(string) != imagePath1 {
		t.Errorf("Expected image_path %s, got %v", imagePath1, firstDetection["image_path"])
	}
	if int(firstDetection["defect_count"].(float64)) != 3 {
		t.Errorf("Expected defect_count 3, got %v", firstDetection["defect_count"])
	}
}

// TestGetHighRiskAlerts_Success 测试高危缺陷告警
func TestGetHighRiskAlerts_Success(t *testing.T) {
	db := setupStatsTestDB(t)
	router := setupStatsTestRouter(db)
	defer clearStatsCache()

	user := createStatsTestUser(db, "testuser", "user")
	bridge := createStatsTestBridge(db, user.ID, "测试桥梁")

	// 创建不同风险等级的缺陷
	createStatsTestDefect(db, bridge.ID, "裂缝", 0.96, 0.12, time.Now()) // 紧急：confidence≥0.95 AND area≥0.1
	createStatsTestDefect(db, bridge.ID, "剥落", 0.93, 0.08, time.Now()) // 严重：confidence≥0.90 OR area≥0.05
	createStatsTestDefect(db, bridge.ID, "钢筋锈蚀", 0.88, 0.03, time.Now()) // 高危：confidence≥0.85 OR area≥0.02
	createStatsTestDefect(db, bridge.ID, "表面损伤", 0.75, 0.01, time.Now()) // 一般：不满足高危条件

	cookies := loginStatsTest(t, router, "testuser", "123456")
	req := httptest.NewRequest("GET", "/api/v1/stats/high-risk-alerts?limit=20", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})

	alerts := data["alerts"].([]interface{})
	if len(alerts) != 3 {
		t.Errorf("Expected 3 high-risk alerts, got %d", len(alerts))
	}

	// 验证第一个告警（置信度最高的）
	firstAlert := alerts[0].(map[string]interface{})
	if firstAlert["defect_type"].(string) != "裂缝" {
		t.Errorf("Expected defect_type '裂缝', got %v", firstAlert["defect_type"])
	}
	if firstAlert["confidence"].(float64) < 0.95 {
		t.Errorf("Expected confidence ≥0.95, got %v", firstAlert["confidence"])
	}

	// 验证严重程度
	severity := firstAlert["severity"].(string)
	if severity != "紧急" && severity != "严重" && severity != "高危" {
		t.Errorf("Invalid severity: %v", severity)
	}
}

// TestGetHighRiskAlerts_SeverityFilter 测试高危告警严重程度过滤
func TestGetHighRiskAlerts_SeverityFilter(t *testing.T) {
	db := setupStatsTestDB(t)
	router := setupStatsTestRouter(db)
	defer clearStatsCache()

	user := createStatsTestUser(db, "testuser", "user")
	bridge := createStatsTestBridge(db, user.ID, "测试桥梁")

	// 创建紧急级别缺陷
	createStatsTestDefect(db, bridge.ID, "裂缝", 0.96, 0.12, time.Now())
	// 创建严重级别缺陷
	createStatsTestDefect(db, bridge.ID, "剥落", 0.93, 0.08, time.Now())
	// 创建高危级别缺陷
	createStatsTestDefect(db, bridge.ID, "钢筋锈蚀", 0.88, 0.03, time.Now())

	cookies := loginStatsTest(t, router, "testuser", "123456")

	// 测试过滤紧急级别
	reqUrgent := httptest.NewRequest("GET", "/api/v1/stats/high-risk-alerts?severity=urgent&limit=20", nil)
	for _, cookie := range cookies {
		reqUrgent.AddCookie(cookie)
	}
	wUrgent := httptest.NewRecorder()
	router.ServeHTTP(wUrgent, reqUrgent)

	var responseUrgent map[string]interface{}
	json.Unmarshal(wUrgent.Body.Bytes(), &responseUrgent)
	dataUrgent := responseUrgent["data"].(map[string]interface{})

	alertsUrgent := dataUrgent["alerts"].([]interface{})
	// 只有confidence≥0.95 AND area≥0.1的才是紧急
	if len(alertsUrgent) != 1 {
		t.Errorf("Expected 1 urgent alert, got %d", len(alertsUrgent))
	}
}

// TestStatsCache_Working 测试缓存机制工作正常
func TestStatsCache_Working(t *testing.T) {
	db := setupStatsTestDB(t)
	router := setupStatsTestRouter(db)
	defer clearStatsCache()

	user := createStatsTestUser(db, "testuser", "user")
	bridge := createStatsTestBridge(db, user.ID, "测试桥梁")
	createStatsTestDefect(db, bridge.ID, "裂缝", 0.95, 0.05, time.Now())

	cookies := loginStatsTest(t, router, "testuser", "123456")

	// 第一次请求（缓存未命中，从数据库查询）
	req1 := httptest.NewRequest("GET", "/api/v1/stats/overview", nil)
	for _, cookie := range cookies {
		req1.AddCookie(cookie)
	}
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	var response1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &response1)
	data1 := response1["data"].(map[string]interface{})

	// 添加新缺陷
	createStatsTestDefect(db, bridge.ID, "剥落", 0.88, 0.03, time.Now())

	// 第二次请求（缓存命中，数据应该和第一次相同）
	req2 := httptest.NewRequest("GET", "/api/v1/stats/overview", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	var response2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response2)
	data2 := response2["data"].(map[string]interface{})

	// 验证缓存生效（defect_count应该还是1，因为缓存了）
	if int(data2["defect_count"].(float64)) != int(data1["defect_count"].(float64)) {
		t.Errorf("Cache should return same data, but defect_count changed from %v to %v",
			data1["defect_count"], data2["defect_count"])
	}

	// 清除缓存
	clearStatsCache()

	// 第三次请求（缓存已清除，应该从数据库获取最新数据）
	req3 := httptest.NewRequest("GET", "/api/v1/stats/overview", nil)
	for _, cookie := range cookies {
		req3.AddCookie(cookie)
	}
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	var response3 map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &response3)
	data3 := response3["data"].(map[string]interface{})

	// 验证缓存清除后获取到最新数据（defect_count应该是2）
	if int(data3["defect_count"].(float64)) != 2 {
		t.Errorf("Expected defect_count 2 after cache clear, got %v", data3["defect_count"])
	}
}
