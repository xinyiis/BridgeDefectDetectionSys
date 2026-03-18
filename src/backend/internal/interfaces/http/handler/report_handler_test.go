// Package handler_test 报表模块集成测试
package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/pdf"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/persistence"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/interfaces/http/handler"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/interfaces/http/middleware"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupReportTestDB 创建报表测试数据库
func setupReportTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移表结构（SQLite会将enum自动转为text）
	err = db.AutoMigrate(&model.User{}, &model.Bridge{}, &model.Defect{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// 手动创建reports表（简化的SQLite版本）
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS reports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			report_name VARCHAR(200) NOT NULL,
			report_type VARCHAR(50) NOT NULL,
			user_id INTEGER NOT NULL,
			bridge_id INTEGER,
			bridge_ids TEXT,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			file_path VARCHAR(500),
			file_size INTEGER DEFAULT 0,
			status VARCHAR(20) DEFAULT 'generating',
			error_message TEXT,
			total_pages INTEGER DEFAULT 0,
			defect_count INTEGER DEFAULT 0,
			high_risk_count INTEGER DEFAULT 0,
			health_score REAL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create reports table: %v", err)
	}

	return db
}

// setupReportTestRouter 创建报表测试路由
func setupReportTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, repository.ReportRepository) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 配置Session
	store := cookie.NewStore([]byte("test-secret-key"))
	r.Use(sessions.Sessions("test_session", store))

	// 依赖注入
	userRepo := persistence.NewUserRepository(db)
	bridgeRepo := persistence.NewBridgeRepository(db)
	defectRepo := persistence.NewDefectRepository(db)
	reportRepo := persistence.NewReportRepository(db)

	userService := service.NewUserService(userRepo)
	bridgeService := service.NewBridgeService(db, bridgeRepo, nil)
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)
	reportService := service.NewReportService(db, reportRepo)

	authUseCase := usecase.NewAuthUseCase(userService)

	// 创建测试用的PDF生成器（使用临时字体路径）
	testFontPath := createTestFont(t)
	pdfGenerator := pdf.NewReportGenerator(testFontPath)
	reportUseCase := usecase.NewReportUseCase(reportService, bridgeService, defectService, pdfGenerator, os.TempDir())

	authHandler := handler.NewAuthHandler(authUseCase)
	reportHandler := handler.NewReportHandler(reportUseCase)

	// 注册路由
	api := r.Group("/api/v1")

	// 公开路由
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	// 认证路由
	auth := api.Group("")
	auth.Use(middleware.AuthRequired(db))
	{
		// 报表模块
		reports := auth.Group("/reports")
		{
			reports.GET("", reportHandler.ListReports)
			reports.POST("", reportHandler.CreateReport)

			reportResource := reports.Group("/:id")
			reportResource.Use(middleware.ReportOwnershipRequired(reportRepo))
			{
				reportResource.GET("", reportHandler.GetReport)
				reportResource.GET("/download", reportHandler.DownloadReport)
				reportResource.DELETE("", reportHandler.DeleteReport)
			}
		}
	}

	return r, reportRepo
}

// createTestFont 创建测试字体文件
func createTestFont(t *testing.T) string {
	// 对于测试，我们可以创建一个空文件或者跳过字体
	// 实际PDF生成会失败，但对于测试API逻辑足够了
	fontPath := "/tmp/test_font.ttf"
	file, err := os.Create(fontPath)
	if err != nil {
		t.Logf("Warning: Could not create test font: %v", err)
		return ""
	}
	file.Close()
	return fontPath
}

// createReportTestUser 创建测试用户
func createReportTestUser(db *gorm.DB, username, role string) *model.User {
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

// createReportTestBridge 创建测试桥梁
func createReportTestBridge(db *gorm.DB, userID uint, bridgeName string) *model.Bridge {
	bridge := &model.Bridge{
		BridgeName: bridgeName,
		BridgeCode: fmt.Sprintf("BRIDGE_%d_%d", userID, time.Now().UnixNano()),
		UserID:     userID,
		Address:    "测试地址",
		Longitude:  118.78,
		Latitude:   32.04,
		BridgeType: "梁桥",
		BuildYear:  2020,
		Length:     100.0,
		Width:      20.0,
		Status:     "正常",
	}
	db.Create(bridge)
	return bridge
}

// createReportTestDefect 创建测试缺陷
func createReportTestDefect(db *gorm.DB, bridgeID uint, defectType string, confidence float64, area float64, detectedAt time.Time) *model.Defect {
	defect := &model.Defect{
		BridgeID:   bridgeID,
		DefectType: defectType,
		ImagePath:  fmt.Sprintf("images/test_%d.jpg", time.Now().UnixNano()),
		ResultPath: "results/test_result.jpg",
		BBox:       "[10, 20, 30, 40]",
		Confidence: confidence,
		Area:       area,
		Length:     1.0,
		Width:      area,
		DetectedAt: detectedAt,
	}
	db.Create(defect)
	return defect
}

// loginReportTest 登录并获取Cookies
func loginReportTest(t *testing.T, router *gin.Engine, username, password string) []*http.Cookie {
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

// TestCreateReport_Success 测试创建报表成功
func TestCreateReport_Success(t *testing.T) {
	db := setupReportTestDB(t)
	router, _ := setupReportTestRouter(t, db)

	// 创建测试数据
	user := createReportTestUser(db, "testuser", "user")
	bridge := createReportTestBridge(db, user.ID, "测试桥梁")

	// 创建一些缺陷数据
	now := time.Now()
	createReportTestDefect(db, bridge.ID, "裂缝", 0.95, 0.05, now.AddDate(0, 0, -1))
	createReportTestDefect(db, bridge.ID, "剥落", 0.88, 0.03, now.AddDate(0, 0, -2))

	// 登录
	cookies := loginReportTest(t, router, "testuser", "123456")

	// 创建报表
	reportBody := map[string]interface{}{
		"report_name": "测试报表",
		"report_type": "bridge_inspection",
		"bridge_id":   bridge.ID,
		"start_time":  now.AddDate(0, 0, -7).Format("2006-01-02"),
		"end_time":    now.Format("2006-01-02"),
	}
	bodyBytes, _ := json.Marshal(reportBody)

	req := httptest.NewRequest("POST", "/api/v1/reports", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		return
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["code"].(float64) != 200 {
		t.Errorf("Expected code 200, got %v", response["code"])
	}

	// 验证返回的报表数据
	data := response["data"].(map[string]interface{})
	if data["report_name"].(string) != "测试报表" {
		t.Errorf("Expected report_name '测试报表', got %v", data["report_name"])
	}
	if data["report_type"].(string) != "bridge_inspection" {
		t.Errorf("Expected report_type 'bridge_inspection', got %v", data["report_type"])
	}
	if data["status"].(string) != "generating" {
		t.Errorf("Expected status 'generating', got %v", data["status"])
	}

	t.Logf("✓ 报表创建成功，ID: %v", data["id"])
}

// TestCreateReport_InvalidParams 测试创建报表参数错误
func TestCreateReport_InvalidParams(t *testing.T) {
	db := setupReportTestDB(t)
	router, _ := setupReportTestRouter(t, db)

	createReportTestUser(db, "testuser", "user")
	cookies := loginReportTest(t, router, "testuser", "123456")

	// 测试缺少必填字段
	reportBody := map[string]interface{}{
		"report_name": "测试报表",
		"report_type": "bridge_inspection",
		// 缺少 bridge_id
		"start_time": "2024-01-01",
		"end_time":   "2024-01-31",
	}
	bodyBytes, _ := json.Marshal(reportBody)

	req := httptest.NewRequest("POST", "/api/v1/reports", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回400错误
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	t.Logf("✓ 参数验证正常，拒绝了无效请求")
}

// TestListReports_Success 测试查询报表列表
func TestListReports_Success(t *testing.T) {
	db := setupReportTestDB(t)
	router, _ := setupReportTestRouter(t, db)

	user := createReportTestUser(db, "testuser", "user")
	bridge := createReportTestBridge(db, user.ID, "测试桥梁")

	// 创建几个报表记录
	for i := 1; i <= 3; i++ {
		report := &model.Report{
			ReportName: fmt.Sprintf("测试报表%d", i),
			ReportType: model.ReportTypeBridgeInspection,
			UserID:     user.ID,
			BridgeID:   &bridge.ID,
			StartTime:  time.Now().AddDate(0, 0, -30),
			EndTime:    time.Now(),
			Status:     model.ReportStatusCompleted,
		}
		db.Create(report)
	}

	// 登录并查询
	cookies := loginReportTest(t, router, "testuser", "123456")
	req := httptest.NewRequest("GET", "/api/v1/reports?page=1&page_size=10", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		return
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	list := data["list"].([]interface{})

	if len(list) != 3 {
		t.Errorf("Expected 3 reports, got %d", len(list))
	}

	if int(data["total"].(float64)) != 3 {
		t.Errorf("Expected total 3, got %v", data["total"])
	}

	t.Logf("✓ 查询到%d个报表", len(list))
}

// TestListReports_PermissionFilter 测试报表列表权限过滤
func TestListReports_PermissionFilter(t *testing.T) {
	db := setupReportTestDB(t)
	router, _ := setupReportTestRouter(t, db)

	// 创建两个用户
	user1 := createReportTestUser(db, "user1", "user")
	user2 := createReportTestUser(db, "user2", "user")

	bridge1 := createReportTestBridge(db, user1.ID, "用户1的桥梁")
	bridge2 := createReportTestBridge(db, user2.ID, "用户2的桥梁")

	// 为user1创建2个报表
	for i := 1; i <= 2; i++ {
		report := &model.Report{
			ReportName: fmt.Sprintf("用户1报表%d", i),
			ReportType: model.ReportTypeBridgeInspection,
			UserID:     user1.ID,
			BridgeID:   &bridge1.ID,
			StartTime:  time.Now().AddDate(0, 0, -30),
			EndTime:    time.Now(),
			Status:     model.ReportStatusCompleted,
		}
		db.Create(report)
	}

	// 为user2创建3个报表
	for i := 1; i <= 3; i++ {
		report := &model.Report{
			ReportName: fmt.Sprintf("用户2报表%d", i),
			ReportType: model.ReportTypeBridgeInspection,
			UserID:     user2.ID,
			BridgeID:   &bridge2.ID,
			StartTime:  time.Now().AddDate(0, 0, -30),
			EndTime:    time.Now(),
			Status:     model.ReportStatusCompleted,
		}
		db.Create(report)
	}

	// user1登录查询，应该只能看到自己的2个报表
	cookies := loginReportTest(t, router, "user1", "123456")
	req := httptest.NewRequest("GET", "/api/v1/reports", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	list := data["list"].([]interface{})

	if len(list) != 2 {
		t.Errorf("User1 should see 2 reports, got %d", len(list))
	}

	// 验证报表名称
	for _, item := range list {
		reportData := item.(map[string]interface{})
		reportName := reportData["report_name"].(string)
		if reportName == "用户2报表1" || reportName == "用户2报表2" || reportName == "用户2报表3" {
			t.Errorf("User1 should not see User2's reports, but got: %s", reportName)
		}
	}

	t.Logf("✓ 权限过滤正常，user1只能看到自己的%d个报表", len(list))
}

// TestGetReport_Success 测试获取报表详情
func TestGetReport_Success(t *testing.T) {
	db := setupReportTestDB(t)
	router, _ := setupReportTestRouter(t, db)

	user := createReportTestUser(db, "testuser", "user")
	bridge := createReportTestBridge(db, user.ID, "测试桥梁")

	// 创建一个报表
	report := &model.Report{
		ReportName:    "测试报表",
		ReportType:    model.ReportTypeBridgeInspection,
		UserID:        user.ID,
		BridgeID:      &bridge.ID,
		StartTime:     time.Now().AddDate(0, 0, -30),
		EndTime:       time.Now(),
		Status:        model.ReportStatusCompleted,
		DefectCount:   10,
		HighRiskCount: 2,
		HealthScore:   85.5,
	}
	db.Create(report)

	// 登录并查询
	cookies := loginReportTest(t, router, "testuser", "123456")
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/reports/%d", report.ID), nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		return
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	if data["report_name"].(string) != "测试报表" {
		t.Errorf("Expected report_name '测试报表', got %v", data["report_name"])
	}
	if int(data["defect_count"].(float64)) != 10 {
		t.Errorf("Expected defect_count 10, got %v", data["defect_count"])
	}
	if int(data["high_risk_count"].(float64)) != 2 {
		t.Errorf("Expected high_risk_count 2, got %v", data["high_risk_count"])
	}
	if data["health_score"].(float64) != 85.5 {
		t.Errorf("Expected health_score 85.5, got %v", data["health_score"])
	}

	t.Logf("✓ 获取报表详情成功")
}

// TestGetReport_Forbidden 测试无权访问他人报表
func TestGetReport_Forbidden(t *testing.T) {
	db := setupReportTestDB(t)
	router, _ := setupReportTestRouter(t, db)

	user1 := createReportTestUser(db, "user1", "user")
	createReportTestUser(db, "user2", "user")
	bridge1 := createReportTestBridge(db, user1.ID, "用户1的桥梁")

	// user1创建报表
	report := &model.Report{
		ReportName: "用户1的报表",
		ReportType: model.ReportTypeBridgeInspection,
		UserID:     user1.ID,
		BridgeID:   &bridge1.ID,
		StartTime:  time.Now().AddDate(0, 0, -30),
		EndTime:    time.Now(),
		Status:     model.ReportStatusCompleted,
	}
	db.Create(report)

	// user2尝试访问user1的报表
	cookies := loginReportTest(t, router, "user2", "123456")
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/reports/%d", report.ID), nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回403 Forbidden
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	t.Logf("✓ 权限验证正常，user2无法访问user1的报表")
}

// TestDeleteReport_Success 测试删除报表
func TestDeleteReport_Success(t *testing.T) {
	db := setupReportTestDB(t)
	router, _ := setupReportTestRouter(t, db)

	user := createReportTestUser(db, "testuser", "user")
	bridge := createReportTestBridge(db, user.ID, "测试桥梁")

	// 创建报表
	report := &model.Report{
		ReportName: "待删除的报表",
		ReportType: model.ReportTypeBridgeInspection,
		UserID:     user.ID,
		BridgeID:   &bridge.ID,
		StartTime:  time.Now().AddDate(0, 0, -30),
		EndTime:    time.Now(),
		Status:     model.ReportStatusCompleted,
	}
	db.Create(report)

	// 登录并删除
	cookies := loginReportTest(t, router, "testuser", "123456")
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/reports/%d", report.ID), nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		return
	}

	// 验证报表已被软删除
	var deletedReport model.Report
	result := db.Unscoped().First(&deletedReport, report.ID)
	if result.Error != nil {
		t.Errorf("Failed to query deleted report: %v", result.Error)
	}
	if deletedReport.DeletedAt.Time.IsZero() {
		t.Errorf("Report should be soft deleted, but DeletedAt is zero")
	}

	t.Logf("✓ 报表删除成功（软删除）")
}

// TestAdminCanSeeAllReports 测试管理员可以查看所有报表
func TestAdminCanSeeAllReports(t *testing.T) {
	db := setupReportTestDB(t)
	router, _ := setupReportTestRouter(t, db)

	// 创建管理员和普通用户
	createReportTestUser(db, "admin", "admin")
	user1 := createReportTestUser(db, "user1", "user")
	user2 := createReportTestUser(db, "user2", "user")

	bridge1 := createReportTestBridge(db, user1.ID, "用户1的桥梁")
	bridge2 := createReportTestBridge(db, user2.ID, "用户2的桥梁")

	// 为不同用户创建报表
	report1 := &model.Report{
		ReportName: "用户1的报表",
		ReportType: model.ReportTypeBridgeInspection,
		UserID:     user1.ID,
		BridgeID:   &bridge1.ID,
		StartTime:  time.Now().AddDate(0, 0, -30),
		EndTime:    time.Now(),
		Status:     model.ReportStatusCompleted,
	}
	db.Create(report1)

	report2 := &model.Report{
		ReportName: "用户2的报表",
		ReportType: model.ReportTypeBridgeInspection,
		UserID:     user2.ID,
		BridgeID:   &bridge2.ID,
		StartTime:  time.Now().AddDate(0, 0, -30),
		EndTime:    time.Now(),
		Status:     model.ReportStatusCompleted,
	}
	db.Create(report2)

	// 管理员登录查询
	cookies := loginReportTest(t, router, "admin", "123456")
	req := httptest.NewRequest("GET", "/api/v1/reports", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	list := data["list"].([]interface{})

	// 管理员应该能看到所有报表
	if len(list) != 2 {
		t.Errorf("Admin should see all 2 reports, got %d", len(list))
	}

	t.Logf("✓ 管理员可以查看所有%d个报表", len(list))
}
