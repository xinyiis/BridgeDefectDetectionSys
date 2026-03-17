// Package handler_test 桥梁管理接口集成测试
package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/persistence"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/interfaces/http/handler"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/interfaces/http/middleware"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupBridgeTestDB 创建桥梁测试数据库
func setupBridgeTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&model.User{}, &model.Bridge{}, &model.Defect{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// setupBridgeTestRouter 创建桥梁测试路由
func setupBridgeTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 配置Session
	store := cookie.NewStore([]byte("test-secret-key"))
	r.Use(sessions.Sessions("test_session", store))

	// 依赖注入
	userRepo := persistence.NewUserRepository(db)
	bridgeRepo := persistence.NewBridgeRepository(db)
	fileService := persistence.NewLocalFileStorage("./test_uploads")

	userService := service.NewUserService(userRepo)
	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)

	authUseCase := usecase.NewAuthUseCase(userService)
	userUseCase := usecase.NewUserUseCase(userService)
	bridgeUseCase := usecase.NewBridgeUseCase(bridgeService)

	authHandler := handler.NewAuthHandler(authUseCase)
	userHandler := handler.NewUserHandler(userUseCase)
	bridgeHandler := handler.NewBridgeHandler(bridgeUseCase, fileService)

	// 注册路由
	api := r.Group("/api")

	// 公开路由
	api.POST("/register", authHandler.Register)
	api.POST("/login", authHandler.Login)

	// 认证路由
	auth := api.Group("")
	auth.Use(middleware.AuthRequired(db))
	auth.POST("/logout", authHandler.Logout)
	auth.GET("/user/info", userHandler.GetUserInfo)

	// 桥梁路由
	bridges := auth.Group("/bridges")
	{
		bridges.GET("", bridgeHandler.ListBridges)
		bridges.POST("", bridgeHandler.CreateBridge)

		bridgeResource := bridges.Group("/:id")
		bridgeResource.Use(middleware.BridgeOwnershipRequired(bridgeRepo))
		{
			bridgeResource.GET("", bridgeHandler.GetBridge)
			bridgeResource.PUT("", bridgeHandler.UpdateBridge)
			bridgeResource.DELETE("", bridgeHandler.DeleteBridge)
		}
	}

	return r
}

// createTestUser 创建测试用户
func createTestUser(db *gorm.DB, username, role string) *model.User {
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

// createTestBridge 创建测试桥梁
func createTestBridge(db *gorm.DB, userID uint, bridgeName, bridgeCode string) *model.Bridge {
	bridge := &model.Bridge{
		BridgeName:  bridgeName,
		BridgeCode:  bridgeCode,
		Address:     "测试地址",
		Longitude:   118.78,
		Latitude:    32.04,
		BridgeType:  "梁桥",
		BuildYear:   2020,
		Length:      100.5,
		Width:       15.0,
		Status:      "正常",
		Model3DPath: "",
		Remark:      "测试桥梁",
		UserID:      userID,
	}
	db.Create(bridge)
	return bridge
}

// createTestDefect 创建测试缺陷
func createTestDefect(db *gorm.DB, bridgeID uint, defectType string) *model.Defect {
	defect := &model.Defect{
		BridgeID:   bridgeID,
		DefectType: defectType,
	}
	db.Create(defect)
	return defect
}

// loginBridgeTest 登录并获取Cookies
func loginBridgeTest(t *testing.T, router *gin.Engine, username, password string) []*http.Cookie {
	loginBody := map[string]interface{}{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(loginBody)
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Result().Cookies()
}

// TestCreateBridge_Success 测试创建桥梁成功
func TestCreateBridge_Success(t *testing.T) {
	db := setupBridgeTestDB(t)
	router := setupBridgeTestRouter(db)

	// 创建测试用户
	user := createTestUser(db, "testuser", "user")

	// 登录
	cookies := loginBridgeTest(t, router, "testuser", "123456")

	// 准备请求数据
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("bridge_name", "长江大桥")
	writer.WriteField("bridge_code", "BR001")
	writer.WriteField("address", "湖北省武汉市")
	writer.WriteField("longitude", "114.31")
	writer.WriteField("latitude", "30.52")
	writer.WriteField("bridge_type", "梁桥")
	writer.WriteField("build_year", "2020")
	writer.WriteField("length", "1500.5")
	writer.WriteField("width", "30.0")
	writer.WriteField("remark", "测试桥梁")
	writer.Close()

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/bridges", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	if data["bridge_name"] != "长江大桥" {
		t.Errorf("Expected bridge_name '长江大桥', got %v", data["bridge_name"])
	}
	if data["bridge_code"] != "BR001" {
		t.Errorf("Expected bridge_code 'BR001', got %v", data["bridge_code"])
	}

	// 验证数据库中存在该桥梁
	var bridge model.Bridge
	db.Where("bridge_code = ?", "BR001").First(&bridge)
	if bridge.ID == 0 {
		t.Error("Bridge not created in database")
	}
	if bridge.UserID != user.ID {
		t.Errorf("Expected user_id %d, got %d", user.ID, bridge.UserID)
	}
}

// TestCreateBridge_DuplicateCode 测试创建重复编号桥梁
func TestCreateBridge_DuplicateCode(t *testing.T) {
	db := setupBridgeTestDB(t)
	router := setupBridgeTestRouter(db)

	// 创建测试用户
	user := createTestUser(db, "testuser", "user")

	// 先创建一个桥梁
	createTestBridge(db, user.ID, "桥梁1", "BR001")

	// 登录
	cookies := loginBridgeTest(t, router, "testuser", "123456")

	// 尝试创建相同编号的桥梁
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("bridge_name", "桥梁2")
	writer.WriteField("bridge_code", "BR001") // 重复编号
	writer.WriteField("address", "测试地址")
	writer.WriteField("longitude", "118.78")
	writer.WriteField("latitude", "32.04")
	writer.WriteField("bridge_type", "梁桥")
	writer.WriteField("build_year", "2020")
	writer.WriteField("length", "100.5")
	writer.WriteField("width", "15.0")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/bridges", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应 - 应该返回400错误
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "桥梁编号已存在" {
		t.Errorf("Expected error about duplicate code, got %v", response["message"])
	}
}

// TestListBridges_PermissionFilter 测试桥梁列表权限过滤
func TestListBridges_PermissionFilter(t *testing.T) {
	db := setupBridgeTestDB(t)
	router := setupBridgeTestRouter(db)

	// 创建两个用户
	user1 := createTestUser(db, "user1", "user")
	user2 := createTestUser(db, "user2", "user")

	// 用户1创建2个桥梁，用户2创建1个桥梁
	createTestBridge(db, user1.ID, "Bridge1", "BR001")
	createTestBridge(db, user1.ID, "Bridge2", "BR002")
	createTestBridge(db, user2.ID, "Bridge3", "BR003")

	// 用户1登录并查询列表
	cookies := loginBridgeTest(t, router, "user1", "123456")
	req, _ := http.NewRequest("GET", "/api/bridges", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证用户1只能看到自己的2个桥梁
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	total := data["total"].(float64)
	if total != 2 {
		t.Errorf("User1 should see 2 bridges, got %v", total)
	}
}

// TestListBridges_AdminSeeAll 测试管理员查看所有桥梁
func TestListBridges_AdminSeeAll(t *testing.T) {
	db := setupBridgeTestDB(t)
	router := setupBridgeTestRouter(db)

	// 创建管理员和普通用户
	admin := createTestUser(db, "admin", "admin")
	user1 := createTestUser(db, "user1", "user")
	user2 := createTestUser(db, "user2", "user")

	// 不同用户创建桥梁
	createTestBridge(db, admin.ID, "AdminBridge", "BR001")
	createTestBridge(db, user1.ID, "User1Bridge", "BR002")
	createTestBridge(db, user2.ID, "User2Bridge", "BR003")

	// 管理员登录并查询列表
	cookies := loginBridgeTest(t, router, "admin", "123456")
	req, _ := http.NewRequest("GET", "/api/bridges", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证管理员能看到所有3个桥梁
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	total := data["total"].(float64)
	if total != 3 {
		t.Errorf("Admin should see 3 bridges, got %v", total)
	}
}

// TestGetBridge_Success 测试获取桥梁详情
func TestGetBridge_Success(t *testing.T) {
	db := setupBridgeTestDB(t)
	router := setupBridgeTestRouter(db)

	// 创建测试用户和桥梁
	user := createTestUser(db, "testuser", "user")
	bridge := createTestBridge(db, user.ID, "TestBridge", "BR001")

	// 登录
	cookies := loginBridgeTest(t, router, "testuser", "123456")

	// 获取桥梁详情
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/bridges/%d", bridge.ID), nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	if data["bridge_name"] != "TestBridge" {
		t.Errorf("Expected bridge_name 'TestBridge', got %v", data["bridge_name"])
	}
	if data["bridge_code"] != "BR001" {
		t.Errorf("Expected bridge_code 'BR001', got %v", data["bridge_code"])
	}
}

// TestGetBridge_Forbidden 测试访问他人桥梁被拒绝
func TestGetBridge_Forbidden(t *testing.T) {
	db := setupBridgeTestDB(t)
	router := setupBridgeTestRouter(db)

	// 创建两个用户
	user1 := createTestUser(db, "user1", "user")
	_ = createTestUser(db, "user2", "user") // 创建user2但不需要保存引用

	// user1创建桥梁
	bridge := createTestBridge(db, user1.ID, "User1Bridge", "BR001")

	// user2登录
	cookies := loginBridgeTest(t, router, "user2", "123456")

	// user2尝试访问user1的桥梁
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/bridges/%d", bridge.ID), nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应 - 应该返回403禁止访问
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

// TestUpdateBridge_Success 测试更新桥梁成功
func TestUpdateBridge_Success(t *testing.T) {
	db := setupBridgeTestDB(t)
	router := setupBridgeTestRouter(db)

	// 创建测试用户和桥梁
	user := createTestUser(db, "testuser", "user")
	bridge := createTestBridge(db, user.ID, "OldName", "BR001")

	// 登录
	cookies := loginBridgeTest(t, router, "testuser", "123456")

	// 更新桥梁
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("bridge_name", "NewName")
	writer.WriteField("status", "维修中")
	writer.Close()

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/bridges/%d", bridge.ID), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	// 验证数据库中的数据已更新
	var updatedBridge model.Bridge
	db.First(&updatedBridge, bridge.ID)
	if updatedBridge.BridgeName != "NewName" {
		t.Errorf("Expected bridge_name 'NewName', got %s", updatedBridge.BridgeName)
	}
	if updatedBridge.Status != "维修中" {
		t.Errorf("Expected status '维修中', got %s", updatedBridge.Status)
	}
	// 验证未更新的字段保持不变
	if updatedBridge.BridgeCode != "BR001" {
		t.Errorf("BridgeCode should remain 'BR001', got %s", updatedBridge.BridgeCode)
	}
}

// TestDeleteBridge_Success 测试删除桥梁成功
func TestDeleteBridge_Success(t *testing.T) {
	db := setupBridgeTestDB(t)
	router := setupBridgeTestRouter(db)

	// 创建测试用户和桥梁
	user := createTestUser(db, "testuser", "user")
	bridge := createTestBridge(db, user.ID, "TestBridge", "BR001")

	// 登录
	cookies := loginBridgeTest(t, router, "testuser", "123456")

	// 删除桥梁
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/bridges/%d", bridge.ID), nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	// 验证桥梁已软删除（查询不到）
	var deletedBridge model.Bridge
	err := db.First(&deletedBridge, bridge.ID).Error
	if err == nil {
		t.Error("Bridge should be deleted (soft delete)")
	}

	// 验证使用Unscoped可以查到（确认是软删除）
	err = db.Unscoped().First(&deletedBridge, bridge.ID).Error
	if err != nil {
		t.Error("Bridge should exist with soft delete")
	}
	if deletedBridge.DeletedAt.Time.IsZero() {
		t.Error("DeletedAt should be set")
	}
}

// TestDeleteBridge_CascadeDelete 测试级联删除缺陷
func TestDeleteBridge_CascadeDelete(t *testing.T) {
	db := setupBridgeTestDB(t)
	router := setupBridgeTestRouter(db)

	// 创建测试用户、桥梁和缺陷
	user := createTestUser(db, "testuser", "user")
	bridge := createTestBridge(db, user.ID, "TestBridge", "BR001")
	_ = createTestDefect(db, bridge.ID, "裂缝") // 创建defect但不需要保存引用
	_ = createTestDefect(db, bridge.ID, "剥落") // 创建defect但不需要保存引用

	// 登录
	cookies := loginBridgeTest(t, router, "testuser", "123456")

	// 删除桥梁
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/bridges/%d", bridge.ID), nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 验证桥梁已软删除
	var deletedBridge model.Bridge
	err := db.First(&deletedBridge, bridge.ID).Error
	if err == nil {
		t.Error("Bridge should be soft deleted")
	}

	// 验证缺陷也已软删除（查询不到）
	var defectCount int64
	db.Model(&model.Defect{}).Where("bridge_id = ?", bridge.ID).Count(&defectCount)
	if defectCount != 0 {
		t.Errorf("Expected 0 defects, got %d", defectCount)
	}

	// 验证使用Unscoped可以查到缺陷（确认是软删除）
	var deletedDefects []model.Defect
	db.Unscoped().Where("bridge_id = ?", bridge.ID).Find(&deletedDefects)
	if len(deletedDefects) != 2 {
		t.Errorf("Expected 2 deleted defects, got %d", len(deletedDefects))
	}

	// 验证缺陷的deleted_at已设置
	for _, defect := range deletedDefects {
		if defect.DeletedAt.Time.IsZero() {
			t.Errorf("Defect %d DeletedAt should be set", defect.ID)
		}
	}
}

// TestDeleteBridge_UniqueIndexRelease 测试删除后编号释放
func TestDeleteBridge_UniqueIndexRelease(t *testing.T) {
	db := setupBridgeTestDB(t)
	router := setupBridgeTestRouter(db)

	// 创建测试用户和桥梁
	user := createTestUser(db, "testuser", "user")
	bridge := createTestBridge(db, user.ID, "Bridge1", "BR001")

	// 登录
	cookies := loginBridgeTest(t, router, "testuser", "123456")

	// 删除桥梁
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/bridges/%d", bridge.ID), nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 验证可以创建相同编号的新桥梁
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("bridge_name", "Bridge2")
	writer.WriteField("bridge_code", "BR001") // 相同编号
	writer.WriteField("address", "测试地址")
	writer.WriteField("longitude", "118.78")
	writer.WriteField("latitude", "32.04")
	writer.WriteField("bridge_type", "梁桥")
	writer.WriteField("build_year", "2020")
	writer.WriteField("length", "100.5")
	writer.WriteField("width", "15.0")
	writer.Close()

	req, _ = http.NewRequest("POST", "/api/bridges", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证可以成功创建
	if w.Code != http.StatusOK {
		t.Errorf("Should be able to create bridge with same code after delete, got status %d. Response: %s", w.Code, w.Body.String())
	}

	// 验证旧桥梁的编号已被修改
	var oldBridge model.Bridge
	db.Unscoped().First(&oldBridge, bridge.ID)
	if oldBridge.BridgeCode == "BR001" {
		t.Error("Old bridge code should be modified after delete")
	}
	if !deletedAt_contains(oldBridge.BridgeCode, "deleted") {
		t.Errorf("Old bridge code should contain 'deleted', got %s", oldBridge.BridgeCode)
	}
}

// deletedAt_contains 辅助函数：检查字符串是否包含子串
func deletedAt_contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] != substr && containsString(s, substr)
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// 清理函数（在测试结束后清理测试文件）
func cleanupTestUploads() {
	os.RemoveAll("./test_uploads")
}

func TestMain(m *testing.M) {
	// 运行测试
	code := m.Run()

	// 清理测试文件
	cleanupTestUploads()

	os.Exit(code)
}
