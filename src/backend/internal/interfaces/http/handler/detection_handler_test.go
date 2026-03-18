// Package handler_test 缺陷检测接口集成测试
package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/external"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/persistence"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/interfaces/http/handler"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/interfaces/http/middleware"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupDetectionTestDB 创建检测测试数据库
func setupDetectionTestDB(t *testing.T) *gorm.DB {
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

// setupDetectionTestRouter 创建检测测试路由
func setupDetectionTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 配置Session
	store := cookie.NewStore([]byte("test-secret-key"))
	r.Use(sessions.Sessions("test_session", store))

	// 依赖注入
	userRepo := persistence.NewUserRepository(db)
	bridgeRepo := persistence.NewBridgeRepository(db)
	defectRepo := persistence.NewDefectRepository(db)
	fileService := persistence.NewLocalFileStorage("./test_uploads")

	userService := service.NewUserService(userRepo)
	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)
	pythonService := external.NewMockPythonService()
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)

	authUseCase := usecase.NewAuthUseCase(userService)
	detectionUseCase := usecase.NewDetectionUseCase(defectService, bridgeService, pythonService, fileService)
	defectUseCase := usecase.NewDefectUseCase(defectService, fileService)

	authHandler := handler.NewAuthHandler(authUseCase)
	detectionHandler := handler.NewDetectionHandler(detectionUseCase)
	defectHandler := handler.NewDefectHandler(defectUseCase)

	// 注册路由
	api := r.Group("/api/v1")

	// 公开路由
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	// 认证路由
	auth := api.Group("")
	auth.Use(middleware.AuthRequired(db))
	{
		// 缺陷检测
		detection := auth.Group("/detection")
		{
			detection.POST("/upload", detectionHandler.UploadAndDetect)
		}

		// 缺陷管理
		defects := auth.Group("/defects")
		{
			defects.GET("", defectHandler.ListDefects)

			defectResource := defects.Group("/:id")
			defectResource.Use(middleware.DefectOwnershipRequired(defectService))
			{
				defectResource.GET("", defectHandler.GetDefect)
				defectResource.DELETE("", defectHandler.DeleteDefect)
			}
		}
	}

	return r
}

// createDetectionTestUser 创建测试用户
func createDetectionTestUser(db *gorm.DB, username, role string) *model.User {
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

// createDetectionTestBridge 创建测试桥梁
func createDetectionTestBridge(db *gorm.DB, userID uint, bridgeName string) *model.Bridge {
	bridge := &model.Bridge{
		BridgeName: bridgeName,
		BridgeCode: fmt.Sprintf("BRIDGE_%d", userID),
		UserID:     userID,
		Address:    "测试地址",
		Longitude:  118.78,
		Latitude:   32.04,
		BridgeType: "梁桥",
	}
	db.Create(bridge)
	return bridge
}

// createDetectionTestDefect 创建测试缺陷
func createDetectionTestDefect(db *gorm.DB, bridgeID uint, defectType string) *model.Defect {
	defect := &model.Defect{
		BridgeID:   bridgeID,
		DefectType: defectType,
		ImagePath:  "images/test.jpg",
		ResultPath: "results/test_result.jpg",
		Confidence: 0.95,
	}
	db.Create(defect)
	return defect
}

// loginDetectionTest 登录并获取Cookies
func loginDetectionTest(t *testing.T, router *gin.Engine, username, password string) []*http.Cookie {
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

// createTestImageFile 创建测试图片文件
func createTestImageFile(t *testing.T) string {
	// 创建临时测试图片
	tmpDir := "./test_uploads/images"
	os.MkdirAll(tmpDir, 0755)
	tmpFile := filepath.Join(tmpDir, "test_image.jpg")

	// 写入简单的JPEG头（模拟图片）
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	err := os.WriteFile(tmpFile, jpegHeader, 0644)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	return tmpFile
}

// cleanupTestFiles 清理测试文件
func cleanupTestFiles() {
	os.RemoveAll("./test_uploads")
}

// TestUploadAndDetect_Success 测试上传检测成功
func TestUploadAndDetect_Success(t *testing.T) {
	db := setupDetectionTestDB(t)
	router := setupDetectionTestRouter(db)
	defer cleanupTestFiles()

	// 创建测试数据
	user := createDetectionTestUser(db, "testuser", "user")
	bridge := createDetectionTestBridge(db, user.ID, "测试桥梁")
	cookies := loginDetectionTest(t, router, "testuser", "123456")

	// 创建测试图片
	testImagePath := createTestImageFile(t)
	defer os.Remove(testImagePath)

	// 构建multipart请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加图片文件
	file, err := os.Open(testImagePath)
	if err != nil {
		t.Fatalf("Failed to open test image: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("image", "test.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	io.Copy(part, file)

	// 添加其他字段
	writer.WriteField("bridge_id", fmt.Sprintf("%d", bridge.ID))
	writer.WriteField("model_name", "yolov8")
	writer.WriteField("pixel_ratio", "0.001")
	writer.Close()

	// 发送请求
	req := httptest.NewRequest("POST", "/api/v1/detection/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
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
	if data["total_defects"].(float64) < 1 {
		t.Errorf("Expected at least 1 defect, got %v", data["total_defects"])
	}

	// 验证数据库中创建了缺陷记录
	var count int64
	db.Model(&model.Defect{}).Where("bridge_id = ?", bridge.ID).Count(&count)
	if count < 1 {
		t.Errorf("Expected at least 1 defect in database, got %d", count)
	}
}

// TestUploadAndDetect_InvalidFile 测试上传无效文件
func TestUploadAndDetect_InvalidFile(t *testing.T) {
	db := setupDetectionTestDB(t)
	router := setupDetectionTestRouter(db)
	defer cleanupTestFiles()

	user := createDetectionTestUser(db, "testuser", "user")
	bridge := createDetectionTestBridge(db, user.ID, "测试桥梁")
	cookies := loginDetectionTest(t, router, "testuser", "123456")

	// 构建multipart请求（上传txt文件）
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("image", "test.txt")
	part.Write([]byte("This is not an image"))

	writer.WriteField("bridge_id", fmt.Sprintf("%d", bridge.ID))
	writer.WriteField("model_name", "yolov8")
	writer.WriteField("pixel_ratio", "0.001")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/detection/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回400错误
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestUploadAndDetect_InvalidBridge 测试桥梁不存在
func TestUploadAndDetect_InvalidBridge(t *testing.T) {
	db := setupDetectionTestDB(t)
	router := setupDetectionTestRouter(db)
	defer cleanupTestFiles()

	createDetectionTestUser(db, "testuser", "user")
	cookies := loginDetectionTest(t, router, "testuser", "123456")

	testImagePath := createTestImageFile(t)
	defer os.Remove(testImagePath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, _ := os.Open(testImagePath)
	defer file.Close()
	part, _ := writer.CreateFormFile("image", "test.jpg")
	io.Copy(part, file)

	writer.WriteField("bridge_id", "9999") // 不存在的桥梁ID
	writer.WriteField("model_name", "yolov8")
	writer.WriteField("pixel_ratio", "0.001")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/detection/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回400或404错误
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 400 or 404, got %d", w.Code)
	}
}

// TestUploadAndDetect_NotOwnedBridge 测试访问他人桥梁
func TestUploadAndDetect_NotOwnedBridge(t *testing.T) {
	db := setupDetectionTestDB(t)
	router := setupDetectionTestRouter(db)
	defer cleanupTestFiles()

	createDetectionTestUser(db, "user1", "user")
	user2 := createDetectionTestUser(db, "user2", "user")
	bridge := createDetectionTestBridge(db, user2.ID, "用户2的桥梁")

	// 用户1登录
	cookies := loginDetectionTest(t, router, "user1", "123456")

	testImagePath := createTestImageFile(t)
	defer os.Remove(testImagePath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, _ := os.Open(testImagePath)
	defer file.Close()
	part, _ := writer.CreateFormFile("image", "test.jpg")
	io.Copy(part, file)

	writer.WriteField("bridge_id", fmt.Sprintf("%d", bridge.ID))
	writer.WriteField("model_name", "yolov8")
	writer.WriteField("pixel_ratio", "0.001")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/detection/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回403错误
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

// TestListDefects_PermissionFilter 测试缺陷列表权限过滤
func TestListDefects_PermissionFilter(t *testing.T) {
	db := setupDetectionTestDB(t)
	router := setupDetectionTestRouter(db)

	// 创建两个用户和各自的桥梁、缺陷
	user1 := createDetectionTestUser(db, "user1", "user")
	user2 := createDetectionTestUser(db, "user2", "user")
	createDetectionTestUser(db, "admin", "admin")

	bridge1 := createDetectionTestBridge(db, user1.ID, "用户1的桥梁")
	bridge2 := createDetectionTestBridge(db, user2.ID, "用户2的桥梁")

	createDetectionTestDefect(db, bridge1.ID, "裂缝")
	createDetectionTestDefect(db, bridge1.ID, "剥落")
	createDetectionTestDefect(db, bridge2.ID, "裂缝")

	// 用户1登录，只能看到2条缺陷
	cookies1 := loginDetectionTest(t, router, "user1", "123456")
	req1 := httptest.NewRequest("GET", "/api/v1/defects", nil)
	for _, cookie := range cookies1 {
		req1.AddCookie(cookie)
	}
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("User1 request failed with status %d", w1.Code)
	}

	var response1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &response1)
	data1 := response1["data"].(map[string]interface{})

	if int(data1["total"].(float64)) != 2 {
		t.Errorf("User1 expected 2 defects, got %v", data1["total"])
	}

	// 管理员登录，能看到所有3条缺陷
	cookiesAdmin := loginDetectionTest(t, router, "admin", "123456")
	reqAdmin := httptest.NewRequest("GET", "/api/v1/defects", nil)
	for _, cookie := range cookiesAdmin {
		reqAdmin.AddCookie(cookie)
	}
	wAdmin := httptest.NewRecorder()
	router.ServeHTTP(wAdmin, reqAdmin)

	var responseAdmin map[string]interface{}
	json.Unmarshal(wAdmin.Body.Bytes(), &responseAdmin)
	dataAdmin := responseAdmin["data"].(map[string]interface{})

	if int(dataAdmin["total"].(float64)) != 3 {
		t.Errorf("Admin expected 3 defects, got %v", dataAdmin["total"])
	}
}

// TestListDefects_FilterByBridge 测试按桥梁过滤
func TestListDefects_FilterByBridge(t *testing.T) {
	db := setupDetectionTestDB(t)
	router := setupDetectionTestRouter(db)

	user := createDetectionTestUser(db, "testuser", "user")
	bridge1 := createDetectionTestBridge(db, user.ID, "桥梁1")
	bridge2 := createDetectionTestBridge(db, user.ID, "桥梁2")

	createDetectionTestDefect(db, bridge1.ID, "裂缝")
	createDetectionTestDefect(db, bridge1.ID, "剥落")
	createDetectionTestDefect(db, bridge2.ID, "裂缝")

	cookies := loginDetectionTest(t, router, "testuser", "123456")

	// 过滤bridge1的缺陷
	url := fmt.Sprintf("/api/v1/defects?bridge_id=%d", bridge1.ID)
	req := httptest.NewRequest("GET", url, nil)
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

	if int(data["total"].(float64)) != 2 {
		t.Errorf("Expected 2 defects for bridge1, got %v", data["total"])
	}
}

// TestGetDefect_Success 测试获取缺陷详情成功
func TestGetDefect_Success(t *testing.T) {
	db := setupDetectionTestDB(t)
	router := setupDetectionTestRouter(db)

	user := createDetectionTestUser(db, "testuser", "user")
	bridge := createDetectionTestBridge(db, user.ID, "测试桥梁")
	defect := createDetectionTestDefect(db, bridge.ID, "裂缝")
	cookies := loginDetectionTest(t, router, "testuser", "123456")

	url := fmt.Sprintf("/api/v1/defects/%d", defect.ID)
	req := httptest.NewRequest("GET", url, nil)
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

	if response["code"].(float64) != 200 {
		t.Errorf("Expected code 200, got %v", response["code"])
	}

	data := response["data"].(map[string]interface{})
	if data["defect_type"].(string) != "裂缝" {
		t.Errorf("Expected defect_type '裂缝', got %v", data["defect_type"])
	}
}

// TestGetDefect_Forbidden 测试访问他人缺陷被拒
func TestGetDefect_Forbidden(t *testing.T) {
	db := setupDetectionTestDB(t)
	router := setupDetectionTestRouter(db)

	createDetectionTestUser(db, "user1", "user")
	user2 := createDetectionTestUser(db, "user2", "user")
	bridge := createDetectionTestBridge(db, user2.ID, "用户2的桥梁")
	defect := createDetectionTestDefect(db, bridge.ID, "裂缝")

	// 用户1尝试访问用户2的缺陷
	cookies := loginDetectionTest(t, router, "user1", "123456")

	url := fmt.Sprintf("/api/v1/defects/%d", defect.ID)
	req := httptest.NewRequest("GET", url, nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回403错误
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

// TestDeleteDefect_Success 测试删除缺陷成功
func TestDeleteDefect_Success(t *testing.T) {
	db := setupDetectionTestDB(t)
	router := setupDetectionTestRouter(db)
	defer cleanupTestFiles()

	user := createDetectionTestUser(db, "testuser", "user")
	bridge := createDetectionTestBridge(db, user.ID, "测试桥梁")
	defect := createDetectionTestDefect(db, bridge.ID, "裂缝")
	cookies := loginDetectionTest(t, router, "testuser", "123456")

	url := fmt.Sprintf("/api/v1/defects/%d", defect.ID)
	req := httptest.NewRequest("DELETE", url, nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// 验证软删除（记录仍在数据库，但DeletedAt不为空）
	var deletedDefect model.Defect
	result := db.Unscoped().First(&deletedDefect, defect.ID)
	if result.Error != nil {
		t.Errorf("Defect should still exist in database after soft delete")
	}
	if deletedDefect.DeletedAt.Time.IsZero() {
		t.Errorf("Expected DeletedAt to be set, but it's zero")
	}

	// 验证正常查询找不到该记录
	var normalQuery model.Defect
	result = db.First(&normalQuery, defect.ID)
	if result.Error == nil {
		t.Errorf("Soft deleted defect should not be found in normal query")
	}
}
