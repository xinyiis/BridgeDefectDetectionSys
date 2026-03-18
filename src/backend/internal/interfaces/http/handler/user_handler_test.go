// Package handler_test 用户管理接口集成测试
package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"golang.org/x/crypto/bcrypt"
)


// TestListUsers_Success 测试获取用户列表（管理员权限）
func TestListUsers_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 创建管理员用户
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := model.User{
		Username: "admin",
		Password: string(hashedPassword),
		RealName: "管理员",
		Email:    "admin@example.com",
		Role:     "admin",
	}
	db.Create(&admin)

	// 创建几个普通用户
	for i := 1; i <= 3; i++ {
		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		user := model.User{
			Username: fmt.Sprintf("user%d", i),
			Password: string(hashedPwd),
			RealName: fmt.Sprintf("用户%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Role:     "user",
		}
		db.Create(&user)
	}

	// 管理员登录
	cookies := loginAndGetCookies(t, router, "admin", "admin123")

	// 获取用户列表
	req, _ := http.NewRequest("GET", "/api/admin/users?page=1&page_size=10", nil)
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
	total := data["total"].(float64)
	if total != 4 { // 1个管理员 + 3个普通用户
		t.Errorf("Expected 4 users, got %v", total)
	}
}

// TestListUsers_Forbidden 测试普通用户访问用户列表（权限不足）
func TestListUsers_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 注册并登录普通用户
	registerAndLogin(t, router)
	cookies := loginAndGetCookies(t, router, "testuser", "123456")

	// 尝试获取用户列表
	req, _ := http.NewRequest("GET", "/api/admin/users", nil)
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

// TestGetUserByID_Success 测试获取指定用户信息（管理员权限）
func TestGetUserByID_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 创建管理员
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := model.User{
		Username: "admin",
		Password: string(hashedPassword),
		RealName: "管理员",
		Email:    "admin@example.com",
		Role:     "admin",
	}
	db.Create(&admin)

	// 创建一个普通用户
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user := model.User{
		Username: "testuser",
		Password: string(hashedPwd),
		RealName: "测试用户",
		Email:    "test@example.com",
		Role:     "user",
	}
	db.Create(&user)

	// 管理员登录
	cookies := loginAndGetCookies(t, router, "admin", "admin123")

	// 获取用户信息
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/admin/users/%d", user.ID), nil)
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
	if data["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got %v", data["username"])
	}
}

// TestDeleteUser_Success 测试删除用户（管理员权限）
func TestDeleteUser_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 创建管理员
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := model.User{
		Username: "admin",
		Password: string(hashedPassword),
		RealName: "管理员",
		Email:    "admin@example.com",
		Role:     "admin",
	}
	db.Create(&admin)

	// 创建要删除的用户
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user := model.User{
		Username: "tobedeleted",
		Password: string(hashedPwd),
		RealName: "待删除用户",
		Email:    "delete@example.com",
		Role:     "user",
	}
	db.Create(&user)

	// 管理员登录
	cookies := loginAndGetCookies(t, router, "admin", "admin123")

	// 删除用户
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/admin/users/%d", user.ID), nil)
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

	if response["message"] != "删除成功" {
		t.Errorf("Expected message '删除成功', got %v", response["message"])
	}

	// 验证用户确实被删除
	var count int64
	db.Model(&model.User{}).Where("id = ?", user.ID).Count(&count)
	if count != 0 {
		t.Error("User should be deleted but still exists")
	}
}

// TestDeleteUser_Forbidden 测试普通用户删除用户（权限不足）
func TestDeleteUser_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 创建两个普通用户
	hashedPwd1, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user1 := model.User{
		Username: "user1",
		Password: string(hashedPwd1),
		RealName: "用户1",
		Email:    "user1@example.com",
		Role:     "user",
	}
	db.Create(&user1)

	hashedPwd2, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user2 := model.User{
		Username: "user2",
		Password: string(hashedPwd2),
		RealName: "用户2",
		Email:    "user2@example.com",
		Role:     "user",
	}
	db.Create(&user2)

	// user1登录
	cookies := loginAndGetCookies(t, router, "user1", "123456")

	// user1尝试删除user2
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/admin/users/%d", user2.ID), nil)
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

// TestPromoteToAdmin_Success 测试提升用户为管理员
func TestPromoteToAdmin_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 创建管理员
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := model.User{
		Username: "admin",
		Password: string(hashedPassword),
		RealName: "管理员",
		Email:    "admin@example.com",
		Role:     "admin",
	}
	db.Create(&admin)

	// 创建普通用户
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user := model.User{
		Username: "normaluser",
		Password: string(hashedPwd),
		RealName: "普通用户",
		Email:    "normal@example.com",
		Role:     "user",
	}
	db.Create(&user)

	// 管理员登录
	cookies := loginAndGetCookies(t, router, "admin", "admin123")

	// 提升用户为管理员
	promoteBody := map[string]interface{}{
		"user_id": user.ID,
	}
	jsonData, _ := json.Marshal(promoteBody)
	req, _ := http.NewRequest("POST", "/api/admin/users/promote", bytes.NewBuffer(jsonData))
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

	if response["message"] != "提升成功" {
		t.Errorf("Expected message '提升成功', got %v", response["message"])
	}

	// 验证用户角色确实变为admin
	var updatedUser model.User
	db.First(&updatedUser, user.ID)
	if updatedUser.Role != "admin" {
		t.Errorf("Expected role 'admin', got %v", updatedUser.Role)
	}
}

// TestPromoteToAdmin_UserNotExist 测试提升不存在的用户
func TestPromoteToAdmin_UserNotExist(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 创建管理员
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := model.User{
		Username: "admin",
		Password: string(hashedPassword),
		RealName: "管理员",
		Email:    "admin@example.com",
		Role:     "admin",
	}
	db.Create(&admin)

	// 管理员登录
	cookies := loginAndGetCookies(t, router, "admin", "admin123")

	// 尝试提升不存在的用户ID
	promoteBody := map[string]interface{}{
		"user_id": 99999,
	}
	jsonData, _ := json.Marshal(promoteBody)
	req, _ := http.NewRequest("POST", "/api/admin/users/promote", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "用户不存在" {
		t.Errorf("Expected error about user not found, got %v", response["message"])
	}
}

// TestUpdatePassword_Success 测试修改密码
func TestUpdatePassword_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 注册并登录
	registerAndLogin(t, router)
	cookies := loginAndGetCookies(t, router, "testuser", "123456")

	// 修改密码
	updateBody := map[string]interface{}{
		"password": "newpassword123",
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

	// 验证新密码可以登录
	loginBody := map[string]interface{}{
		"username": "testuser",
		"password": "newpassword123",
	}
	jsonData, _ = json.Marshal(loginBody)
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Error("New password should work for login")
	}
}

// TestListUsers_Pagination 测试用户列表分页
func TestListUsers_Pagination(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// 创建管理员
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := model.User{
		Username: "admin",
		Password: string(hashedPassword),
		RealName: "管理员",
		Email:    "admin@example.com",
		Role:     "admin",
	}
	db.Create(&admin)

	// 创建15个用户
	for i := 1; i <= 15; i++ {
		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		user := model.User{
			Username: fmt.Sprintf("user%d", i),
			Password: string(hashedPwd),
			RealName: fmt.Sprintf("用户%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Role:     "user",
		}
		db.Create(&user)
	}

	// 管理员登录
	cookies := loginAndGetCookies(t, router, "admin", "admin123")

	// 测试第一页（每页10条）
	req, _ := http.NewRequest("GET", "/api/admin/users?page=1&page_size=10", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	total := data["total"].(float64)
	page := data["page"].(float64)
	pageSize := data["page_size"].(float64)
	users := data["users"].([]interface{})

	if total != 16 { // 1管理员 + 15普通用户
		t.Errorf("Expected total 16, got %v", total)
	}
	if page != 1 {
		t.Errorf("Expected page 1, got %v", page)
	}
	if pageSize != 10 {
		t.Errorf("Expected page_size 10, got %v", pageSize)
	}
	if len(users) != 10 {
		t.Errorf("Expected 10 users in page 1, got %d", len(users))
	}

	// 测试第二页
	req, _ = http.NewRequest("GET", "/api/admin/users?page=2&page_size=10", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &response)
	data = response["data"].(map[string]interface{})
	users = data["users"].([]interface{})

	if len(users) != 6 { // 剩余6个用户
		t.Errorf("Expected 6 users in page 2, got %d", len(users))
	}
}
