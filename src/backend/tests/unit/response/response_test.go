// Package response_test 响应工具单元测试
package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/pkg/response"
)

// setupTestContext 创建测试用的Gin上下文
func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

// TestSuccess 测试成功响应
func TestSuccess(t *testing.T) {
	c, w := setupTestContext()

	testData := gin.H{
		"user": "testuser",
		"id":   123,
	}

	response.Success(c, testData)

	// 验证HTTP状态码
	if w.Code != http.StatusOK {
		t.Errorf("Success() HTTP状态码 = %d, 期望 %d", w.Code, http.StatusOK)
	}

	// 解析响应
	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证响应结构
	if resp.Code != 200 {
		t.Errorf("Response.Code = %d, 期望 200", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("Response.Message = %s, 期望 success", resp.Message)
	}
	if resp.Data == nil {
		t.Error("Response.Data 不应为空")
	}
}

// TestSuccessWithMessage 测试带自定义消息的成功响应
func TestSuccessWithMessage(t *testing.T) {
	c, w := setupTestContext()

	testMessage := "创建成功"
	testData := gin.H{"id": 1}

	response.SuccessWithMessage(c, testMessage, testData)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Message != testMessage {
		t.Errorf("Response.Message = %s, 期望 %s", resp.Message, testMessage)
	}
	if resp.Code != 200 {
		t.Errorf("Response.Code = %d, 期望 200", resp.Code)
	}
}

// TestError 测试错误响应
func TestError(t *testing.T) {
	c, w := setupTestContext()

	testCode := 400
	testMessage := "参数错误"

	response.Error(c, testCode, testMessage)

	// 验证HTTP状态码
	if w.Code != testCode {
		t.Errorf("Error() HTTP状态码 = %d, 期望 %d", w.Code, testCode)
	}

	// 解析响应
	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != testCode {
		t.Errorf("Response.Code = %d, 期望 %d", resp.Code, testCode)
	}
	if resp.Message != testMessage {
		t.Errorf("Response.Message = %s, 期望 %s", resp.Message, testMessage)
	}
}

// TestBadRequest 测试400错误
func TestBadRequest(t *testing.T) {
	c, w := setupTestContext()

	testMessage := "用户名不能为空"

	response.BadRequest(c, testMessage)

	if w.Code != http.StatusBadRequest {
		t.Errorf("BadRequest() HTTP状态码 = %d, 期望 %d", w.Code, http.StatusBadRequest)
	}

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Message != testMessage {
		t.Errorf("Response.Message = %s, 期望 %s", resp.Message, testMessage)
	}
}

// TestUnauthorized 测试401错误
func TestUnauthorized(t *testing.T) {
	c, w := setupTestContext()

	response.Unauthorized(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Unauthorized() HTTP状态码 = %d, 期望 %d", w.Code, http.StatusUnauthorized)
	}

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != 401 {
		t.Errorf("Response.Code = %d, 期望 401", resp.Code)
	}
	if resp.Message != "未登录或登录已过期" {
		t.Errorf("Response.Message = %s, 期望 '未登录或登录已过期'", resp.Message)
	}
}

// TestForbidden 测试403错误
func TestForbidden(t *testing.T) {
	c, w := setupTestContext()

	response.Forbidden(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("Forbidden() HTTP状态码 = %d, 期望 %d", w.Code, http.StatusForbidden)
	}

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Message != "权限不足" {
		t.Errorf("Response.Message = %s, 期望 '权限不足'", resp.Message)
	}
}

// TestNotFound 测试404错误
func TestNotFound(t *testing.T) {
	c, w := setupTestContext()

	resource := "用户"

	response.NotFound(c, resource)

	if w.Code != http.StatusNotFound {
		t.Errorf("NotFound() HTTP状态码 = %d, 期望 %d", w.Code, http.StatusNotFound)
	}

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	expectedMessage := "用户不存在"
	if resp.Message != expectedMessage {
		t.Errorf("Response.Message = %s, 期望 %s", resp.Message, expectedMessage)
	}
}

// TestInternalError 测试500错误
func TestInternalError(t *testing.T) {
	c, w := setupTestContext()

	response.InternalError(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("InternalError() HTTP状态码 = %d, 期望 %d", w.Code, http.StatusInternalServerError)
	}

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Message != "服务器内部错误" {
		t.Errorf("Response.Message = %s, 期望 '服务器内部错误'", resp.Message)
	}
}

// TestErrorWithDetail 测试带详情的错误响应
func TestErrorWithDetail(t *testing.T) {
	c, w := setupTestContext()

	testCode := 500
	testMessage := "数据库错误"
	testDetail := "connection refused"

	response.ErrorWithDetail(c, testCode, testMessage, testDetail)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Error != testDetail {
		t.Errorf("Response.Error = %s, 期望 %s", resp.Error, testDetail)
	}
}

// TestResponseStructure 测试响应结构的JSON序列化
func TestResponseStructure(t *testing.T) {
	resp := response.Response{
		Code:    200,
		Message: "test",
		Data:    gin.H{"key": "value"},
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("序列化Response失败: %v", err)
	}

	// 验证JSON包含必要字段
	jsonStr := string(jsonBytes)
	requiredFields := []string{"code", "message", "data"}

	for _, field := range requiredFields {
		if !contains(jsonStr, field) {
			t.Errorf("JSON响应缺少字段: %s", field)
		}
	}
}

// contains 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestUnauthorizedWithMessage 测试带自定义消息的401错误
func TestUnauthorizedWithMessage(t *testing.T) {
	c, w := setupTestContext()

	customMessage := "Token已过期"

	response.UnauthorizedWithMessage(c, customMessage)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Message != customMessage {
		t.Errorf("Response.Message = %s, 期望 %s", resp.Message, customMessage)
	}
}

// TestForbiddenWithMessage 测试带自定义消息的403错误
func TestForbiddenWithMessage(t *testing.T) {
	c, w := setupTestContext()

	customMessage := "无权访问此资源"

	response.ForbiddenWithMessage(c, customMessage)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Message != customMessage {
		t.Errorf("Response.Message = %s, 期望 %s", resp.Message, customMessage)
	}
}

// BenchmarkSuccess 性能测试：Success函数
func BenchmarkSuccess(b *testing.B) {
	gin.SetMode(gin.TestMode)

	testData := gin.H{"test": "data"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := setupTestContext()
		response.Success(c, testData)
	}
}

// BenchmarkError 性能测试：Error函数
func BenchmarkError(b *testing.B) {
	gin.SetMode(gin.TestMode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := setupTestContext()
		response.Error(c, 400, "test error")
	}
}
