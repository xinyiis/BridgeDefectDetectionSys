// Package handler HTTP集成测试
// 测试用户认证和管理模块的所有接口
package handler_test

import (
	"bytes"
	"encoding/json"
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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestRouter 创建测试用的路由引擎
func setupTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 配置Session
	store := cookie.NewStore([]byte("test-secret-key"))
	r.Use(sessions.Sessions("test_session", store))

	// 依赖注入
	userRepo := persistence.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	authUseCase := usecase.NewAuthUseCase(userService)
	userUseCase := usecase.NewUserUseCase(userService)

	authHandler := handler.NewAuthHandler(authUseCase)
	userHandler := handler.NewUserHandler(userUseCase)

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
	auth.PUT("/user/info", userHandler.UpdateUserInfo)

	// 管理员路由
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(db))
	admin.Use(middleware.AdminRequired())
	admin.GET("/users", userHandler.ListUsers)
	admin.GET("/users/:id", userHandler.GetUserByID)
	admin.DELETE("/users/:id", userHandler.DeleteUser)
	admin.POST("/users/promote", userHandler.PromoteToAdmin)

	return r
}

// setupTestDB 创建测试数据库（使用SQLite内存数据库）
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移表结构（直接使用model.User）
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// TestRegister_Success 测试用户注册成功
func TestRegister_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 准备请求数据
	reqBody := map[string]interface{}{
		"username":  "testuser",
		"password":  "123456",
		"real_name": "测试用户",
		"email":     "test@example.com",
		"phone":     "13800138000",
	}
	jsonData, _ := json.Marshal(reqBody)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["code"].(float64) != 200 {
		t.Errorf("Expected code 200, got %v", response["code"])
	}

	if response["message"] != "注册成功" {
		t.Errorf("Expected message '注册成功', got %v", response["message"])
	}

	// 验证返回的用户数据
	data := response["data"].(map[string]interface{})
	if data["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got %v", data["username"])
	}
	if data["email"] != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %v", data["email"])
	}
	if data["role"] != "user" {
		t.Errorf("Expected role 'user', got %v", data["role"])
	}
}

// TestRegister_DuplicateUsername 测试注册重复用户名
func TestRegister_DuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 第一次注册
	reqBody := map[string]interface{}{
		"username":  "testuser",
		"password":  "123456",
		"real_name": "测试用户",
		"email":     "test1@example.com",
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 第二次注册（相同用户名）
	reqBody["email"] = "test2@example.com"
	jsonData, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "用户名已存在" {
		t.Errorf("Expected error message about duplicate username, got %v", response["message"])
	}
}

// TestRegister_DuplicateEmail 测试注册重复邮箱
func TestRegister_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 第一次注册
	reqBody := map[string]interface{}{
		"username":  "user1",
		"password":  "123456",
		"real_name": "用户1",
		"email":     "test@example.com",
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 第二次注册（相同邮箱）
	reqBody["username"] = "user2"
	jsonData, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "邮箱已被注册" {
		t.Errorf("Expected error message about duplicate email, got %v", response["message"])
	}
}

// TestRegister_InvalidData 测试注册数据验证
func TestRegister_InvalidData(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	testCases := []struct {
		name     string
		reqBody  map[string]interface{}
		expected string
	}{
		{
			name: "缺少用户名",
			reqBody: map[string]interface{}{
				"password":  "123456",
				"real_name": "测试用户",
				"email":     "test@example.com",
			},
			expected: "参数错误",
		},
		{
			name: "密码过短",
			reqBody: map[string]interface{}{
				"username":  "testuser",
				"password":  "123",
				"real_name": "测试用户",
				"email":     "test@example.com",
			},
			expected: "参数错误",
		},
		{
			name: "邮箱格式错误",
			reqBody: map[string]interface{}{
				"username":  "testuser",
				"password":  "123456",
				"real_name": "测试用户",
				"email":     "invalid-email",
			},
			expected: "参数错误",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tc.reqBody)
			req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", w.Code)
			}
		})
	}
}

// TestLogin_Success 测试登录成功
func TestLogin_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 先注册一个用户
	registerBody := map[string]interface{}{
		"username":  "testuser",
		"password":  "123456",
		"real_name": "测试用户",
		"email":     "test@example.com",
	}
	jsonData, _ := json.Marshal(registerBody)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 登录
	loginBody := map[string]interface{}{
		"username": "testuser",
		"password": "123456",
	}
	jsonData, _ = json.Marshal(loginBody)
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["code"].(float64) != 200 {
		t.Errorf("Expected code 200, got %v", response["code"])
	}

	if response["message"] != "登录成功" {
		t.Errorf("Expected message '登录成功', got %v", response["message"])
	}

	// 验证Session Cookie存在
	cookies := w.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == "test_session" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected session cookie not found")
	}
}

// TestLogin_WrongPassword 测试登录密码错误
func TestLogin_WrongPassword(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 先注册一个用户
	registerBody := map[string]interface{}{
		"username":  "testuser",
		"password":  "123456",
		"real_name": "测试用户",
		"email":     "test@example.com",
	}
	jsonData, _ := json.Marshal(registerBody)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 使用错误密码登录
	loginBody := map[string]interface{}{
		"username": "testuser",
		"password": "wrongpassword",
	}
	jsonData, _ = json.Marshal(loginBody)
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "用户名或密码错误" {
		t.Errorf("Expected error message about wrong credentials, got %v", response["message"])
	}
}

// TestLogin_UserNotExist 测试登录用户不存在
func TestLogin_UserNotExist(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	loginBody := map[string]interface{}{
		"username": "nonexistent",
		"password": "123456",
	}
	jsonData, _ := json.Marshal(loginBody)
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestLogout_Success 测试登出成功
func TestLogout_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 先注册并登录
	registerBody := map[string]interface{}{
		"username":  "testuser",
		"password":  "123456",
		"real_name": "测试用户",
		"email":     "test@example.com",
	}
	jsonData, _ := json.Marshal(registerBody)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginBody := map[string]interface{}{
		"username": "testuser",
		"password": "123456",
	}
	jsonData, _ = json.Marshal(loginBody)
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 获取Session Cookie
	cookies := w.Result().Cookies()

	// 登出
	req, _ = http.NewRequest("POST", "/api/logout", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "登出成功" {
		t.Errorf("Expected message '登出成功', got %v", response["message"])
	}
}

// TestGetUserInfo_Success 测试获取用户信息成功
func TestGetUserInfo_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 注册并登录
	registerBody := map[string]interface{}{
		"username":  "testuser",
		"password":  "123456",
		"real_name": "测试用户",
		"email":     "test@example.com",
	}
	jsonData, _ := json.Marshal(registerBody)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginBody := map[string]interface{}{
		"username": "testuser",
		"password": "123456",
	}
	jsonData, _ = json.Marshal(loginBody)
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// 获取用户信息
	req, _ = http.NewRequest("GET", "/api/user/info", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	if data["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got %v", data["username"])
	}
}

// TestGetUserInfo_Unauthorized 测试未登录获取用户信息
func TestGetUserInfo_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	req, _ := http.NewRequest("GET", "/api/user/info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestUpdateUserInfo_Success 测试更新用户信息成功
func TestUpdateUserInfo_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 注册并登录
	registerAndLogin(t, router)
	cookies := loginAndGetCookies(t, router, "testuser", "123456")

	// 更新用户信息
	updateBody := map[string]interface{}{
		"real_name": "新名字",
		"phone":     "13900139000",
	}
	jsonData, _ := json.Marshal(updateBody)
	req, _ := http.NewRequest("PUT", "/api/user/info", bytes.NewBuffer(jsonData))
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
	if data["real_name"] != "新名字" {
		t.Errorf("Expected real_name '新名字', got %v", data["real_name"])
	}
	if data["phone"] != "13900139000" {
		t.Errorf("Expected phone '13900139000', got %v", data["phone"])
	}
}

// 辅助函数：注册并登录
func registerAndLogin(t *testing.T, router *gin.Engine) {
	registerBody := map[string]interface{}{
		"username":  "testuser",
		"password":  "123456",
		"real_name": "测试用户",
		"email":     "test@example.com",
	}
	jsonData, _ := json.Marshal(registerBody)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
}

// 辅助函数：登录并获取Cookies
func loginAndGetCookies(t *testing.T, router *gin.Engine, username, password string) []*http.Cookie {
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
