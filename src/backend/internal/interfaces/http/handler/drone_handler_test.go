// Package handler_test 无人机管理接口集成测试
package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

// setupDroneTestDB 创建无人机测试数据库
func setupDroneTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&model.User{}, &model.Drone{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// setupDroneTestRouter 创建无人机测试路由
func setupDroneTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 配置Session
	store := cookie.NewStore([]byte("test-secret-key"))
	r.Use(sessions.Sessions("test_session", store))

	// 依赖注入
	userRepo := persistence.NewUserRepository(db)
	droneRepo := persistence.NewDroneRepository(db)

	userService := service.NewUserService(userRepo)
	droneService := service.NewDroneService(db, droneRepo)

	authUseCase := usecase.NewAuthUseCase(userService)
	userUseCase := usecase.NewUserUseCase(userService)
	droneUseCase := usecase.NewDroneUseCase(droneService)

	authHandler := handler.NewAuthHandler(authUseCase)
	userHandler := handler.NewUserHandler(userUseCase)
	droneHandler := handler.NewDroneHandler(droneUseCase)

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

	// 无人机路由
	drones := auth.Group("/drones")
	{
		drones.GET("", droneHandler.ListDrones)
		drones.POST("", droneHandler.CreateDrone)

		droneResource := drones.Group("/:id")
		droneResource.Use(middleware.DroneOwnershipRequired(droneRepo))
		{
			droneResource.GET("", droneHandler.GetDrone)
			droneResource.PUT("", droneHandler.UpdateDrone)
			droneResource.DELETE("", droneHandler.DeleteDrone)
		}
	}

	return r
}

// createTestDroneUser 创建测试用户（专用于无人机测试）
func createTestDroneUser(db *gorm.DB, username, role string) *model.User {
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

// createTestDrone 创建测试无人机
func createTestDrone(db *gorm.DB, userID uint, name, droneModel string) *model.Drone {
	drone := &model.Drone{
		Name:      name,
		Model:     droneModel,
		StreamURL: "rtsp://192.168.1.100:554/stream",
		UserID:    userID,
	}
	db.Create(drone)
	return drone
}

// loginDroneTest 登录并获取Cookies
func loginDroneTest(t *testing.T, router *gin.Engine, username, password string) []*http.Cookie {
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

// TestCreateDrone_Success 测试创建无人机成功
func TestCreateDrone_Success(t *testing.T) {
	db := setupDroneTestDB(t)
	router := setupDroneTestRouter(db)

	// 创建测试用户
	user := createTestDroneUser(db, "testuser", "user")

	// 登录
	cookies := loginDroneTest(t, router, "testuser", "123456")

	// 准备请求数据（使用JSON格式）
	reqBody := map[string]interface{}{
		"name":       "大疆 Mavic 3",
		"model":      "Mavic 3",
		"stream_url": "rtsp://192.168.1.100:554/stream",
	}
	jsonData, _ := json.Marshal(reqBody)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/drones", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
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
	if data["name"] != "大疆 Mavic 3" {
		t.Errorf("Expected name '大疆 Mavic 3', got %v", data["name"])
	}
	if data["model"] != "Mavic 3" {
		t.Errorf("Expected model 'Mavic 3', got %v", data["model"])
	}

	// 验证数据库中存在该无人机
	var drone model.Drone
	db.Where("name = ?", "大疆 Mavic 3").First(&drone)
	if drone.ID == 0 {
		t.Error("Drone not created in database")
	}
	if drone.UserID != user.ID {
		t.Errorf("Expected user_id %d, got %d", user.ID, drone.UserID)
	}
}

// TestCreateDrone_InvalidName 测试创建无人机参数验证
func TestCreateDrone_InvalidName(t *testing.T) {
	db := setupDroneTestDB(t)
	router := setupDroneTestRouter(db)

	// 创建测试用户
	createTestDroneUser(db, "testuser", "user")

	// 登录
	cookies := loginDroneTest(t, router, "testuser", "123456")

	// 测试空名称
	reqBody := map[string]interface{}{
		"name": "",
	}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/drones", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应 - 应该返回400错误
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty name, got %d", w.Code)
	}

	// 测试超长名称（101个字符）
	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}
	reqBody = map[string]interface{}{
		"name": longName,
	}
	jsonData, _ = json.Marshal(reqBody)

	req, _ = http.NewRequest("POST", "/api/drones", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应 - 应该返回400错误
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for long name, got %d", w.Code)
	}
}

// TestListDrones_PermissionFilter 测试无人机列表权限过滤
func TestListDrones_PermissionFilter(t *testing.T) {
	db := setupDroneTestDB(t)
	router := setupDroneTestRouter(db)

	// 创建两个用户
	user1 := createTestDroneUser(db, "user1", "user")
	user2 := createTestDroneUser(db, "user2", "user")

	// 用户1创建2个无人机，用户2创建1个无人机
	createTestDrone(db, user1.ID, "Drone1", "Model1")
	createTestDrone(db, user1.ID, "Drone2", "Model2")
	createTestDrone(db, user2.ID, "Drone3", "Model3")

	// 用户1登录并查询列表
	cookies := loginDroneTest(t, router, "user1", "123456")
	req, _ := http.NewRequest("GET", "/api/drones", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证用户1只能看到自己的2个无人机
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	total := data["total"].(float64)
	if total != 2 {
		t.Errorf("User1 should see 2 drones, got %v", total)
	}
}

// TestListDrones_AdminSeeAll 测试管理员查看所有无人机
func TestListDrones_AdminSeeAll(t *testing.T) {
	db := setupDroneTestDB(t)
	router := setupDroneTestRouter(db)

	// 创建管理员和普通用户
	admin := createTestDroneUser(db, "admin", "admin")
	user1 := createTestDroneUser(db, "user1", "user")
	user2 := createTestDroneUser(db, "user2", "user")

	// 不同用户创建无人机
	createTestDrone(db, admin.ID, "AdminDrone", "AdminModel")
	createTestDrone(db, user1.ID, "User1Drone", "User1Model")
	createTestDrone(db, user2.ID, "User2Drone", "User2Model")

	// 管理员登录并查询列表
	cookies := loginDroneTest(t, router, "admin", "123456")
	req, _ := http.NewRequest("GET", "/api/drones", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证管理员能看到所有3个无人机
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	total := data["total"].(float64)
	if total != 3 {
		t.Errorf("Admin should see 3 drones, got %v", total)
	}
}

// TestGetDrone_Success 测试获取无人机详情成功
func TestGetDrone_Success(t *testing.T) {
	db := setupDroneTestDB(t)
	router := setupDroneTestRouter(db)

	// 创建测试用户和无人机
	user := createTestDroneUser(db, "testuser", "user")
	drone := createTestDrone(db, user.ID, "TestDrone", "TestModel")

	// 登录
	cookies := loginDroneTest(t, router, "testuser", "123456")

	// 获取无人机详情
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/drones/%d", drone.ID), nil)
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
	if data["name"] != "TestDrone" {
		t.Errorf("Expected name 'TestDrone', got %v", data["name"])
	}
	if data["model"] != "TestModel" {
		t.Errorf("Expected model 'TestModel', got %v", data["model"])
	}
}

// TestGetDrone_Forbidden 测试访问他人无人机被拒绝
func TestGetDrone_Forbidden(t *testing.T) {
	db := setupDroneTestDB(t)
	router := setupDroneTestRouter(db)

	// 创建两个用户
	user1 := createTestDroneUser(db, "user1", "user")
	_ = createTestDroneUser(db, "user2", "user")

	// user1创建无人机
	drone := createTestDrone(db, user1.ID, "User1Drone", "User1Model")

	// user2登录
	cookies := loginDroneTest(t, router, "user2", "123456")

	// user2尝试访问user1的无人机
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/drones/%d", drone.ID), nil)
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

// TestUpdateDrone_Success 测试更新无人机成功
func TestUpdateDrone_Success(t *testing.T) {
	db := setupDroneTestDB(t)
	router := setupDroneTestRouter(db)

	// 创建测试用户和无人机
	user := createTestDroneUser(db, "testuser", "user")
	drone := createTestDrone(db, user.ID, "OldName", "OldModel")

	// 登录
	cookies := loginDroneTest(t, router, "testuser", "123456")

	// 更新无人机（使用JSON格式）
	reqBody := map[string]interface{}{
		"name":       "NewName",
		"model":      "NewModel",
		"stream_url": "rtsp://192.168.1.101:554/stream",
	}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/drones/%d", drone.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
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
	var updatedDrone model.Drone
	db.First(&updatedDrone, drone.ID)
	if updatedDrone.Name != "NewName" {
		t.Errorf("Expected name 'NewName', got %s", updatedDrone.Name)
	}
	if updatedDrone.Model != "NewModel" {
		t.Errorf("Expected model 'NewModel', got %s", updatedDrone.Model)
	}
	if updatedDrone.StreamURL != "rtsp://192.168.1.101:554/stream" {
		t.Errorf("Expected stream_url 'rtsp://192.168.1.101:554/stream', got %s", updatedDrone.StreamURL)
	}
}

// TestDeleteDrone_Success 测试删除无人机成功
func TestDeleteDrone_Success(t *testing.T) {
	db := setupDroneTestDB(t)
	router := setupDroneTestRouter(db)

	// 创建测试用户和无人机
	user := createTestDroneUser(db, "testuser", "user")
	drone := createTestDrone(db, user.ID, "TestDrone", "TestModel")

	// 登录
	cookies := loginDroneTest(t, router, "testuser", "123456")

	// 删除无人机
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/drones/%d", drone.ID), nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	// 验证无人机已被物理删除（查询不到）
	var deletedDrone model.Drone
	err := db.First(&deletedDrone, drone.ID).Error
	if err == nil {
		t.Error("Drone should be deleted (physical delete)")
	}

	// 验证使用Unscoped也查不到（确认是物理删除）
	err = db.Unscoped().First(&deletedDrone, drone.ID).Error
	if err == nil {
		t.Error("Drone should be physically deleted, not soft deleted")
	}
}

// TestDeleteDrone_PhysicalDelete 测试验证物理删除
func TestDeleteDrone_PhysicalDelete(t *testing.T) {
	db := setupDroneTestDB(t)
	router := setupDroneTestRouter(db)

	// 创建测试用户和无人机
	user := createTestDroneUser(db, "testuser", "user")
	drone := createTestDrone(db, user.ID, "TestDrone", "TestModel")

	// 记录删除前的ID
	droneID := drone.ID

	// 登录
	cookies := loginDroneTest(t, router, "testuser", "123456")

	// 删除无人机
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/drones/%d", drone.ID), nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 验证记录真正被删除（不是软删除）
	var count int64
	db.Unscoped().Model(&model.Drone{}).Where("id = ?", droneID).Count(&count)
	if count != 0 {
		t.Errorf("Drone should be physically deleted, found %d records", count)
	}
}

// TestListDrones_Pagination 测试分页功能
func TestListDrones_Pagination(t *testing.T) {
	db := setupDroneTestDB(t)
	router := setupDroneTestRouter(db)

	// 创建测试用户
	user := createTestDroneUser(db, "testuser", "user")

	// 创建15个无人机
	for i := 1; i <= 15; i++ {
		createTestDrone(db, user.ID, fmt.Sprintf("Drone%d", i), fmt.Sprintf("Model%d", i))
	}

	// 登录
	cookies := loginDroneTest(t, router, "testuser", "123456")

	// 测试第1页（每页10条）
	req, _ := http.NewRequest("GET", "/api/drones?page=1&page_size=10", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	total := data["total"].(float64)
	page := data["page"].(float64)
	pageSize := data["page_size"].(float64)
	list := data["list"].([]interface{})

	if total != 15 {
		t.Errorf("Expected total 15, got %v", total)
	}
	if page != 1 {
		t.Errorf("Expected page 1, got %v", page)
	}
	if pageSize != 10 {
		t.Errorf("Expected page_size 10, got %v", pageSize)
	}
	if len(list) != 10 {
		t.Errorf("Expected 10 items in page 1, got %d", len(list))
	}

	// 测试第2页（应该有5条）
	req, _ = http.NewRequest("GET", "/api/drones?page=2&page_size=10", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &response)
	data = response["data"].(map[string]interface{})
	list = data["list"].([]interface{})

	if len(list) != 5 {
		t.Errorf("Expected 5 items in page 2, got %d", len(list))
	}
}
