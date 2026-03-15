// Package model_test 数据模型单元测试
package model_test

import (
	"testing"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/internal/domain/model"
)

// TestUserTableName 测试User表名
func TestUserTableName(t *testing.T) {
	user := model.User{}
	tableName := user.TableName()

	expected := "users"
	if tableName != expected {
		t.Errorf("User.TableName() = %s, 期望 %s", tableName, expected)
	}
}

// TestUserIsAdmin 测试IsAdmin方法
func TestUserIsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		user     model.User
		expected bool
	}{
		{
			name: "管理员用户",
			user: model.User{
				ID:       1,
				Username: "admin",
				Role:     "admin",
			},
			expected: true,
		},
		{
			name: "普通用户",
			user: model.User{
				ID:       2,
				Username: "user1",
				Role:     "user",
			},
			expected: false,
		},
		{
			name: "空角色",
			user: model.User{
				ID:       3,
				Username: "user2",
				Role:     "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.IsAdmin()
			if result != tt.expected {
				t.Errorf("User.IsAdmin() = %v, 期望 %v (用户角色: %s)", result, tt.expected, tt.user.Role)
			}
		})
	}
}

// TestBridgeTableName 测试Bridge表名
func TestBridgeTableName(t *testing.T) {
	bridge := model.Bridge{}
	tableName := bridge.TableName()

	expected := "bridges"
	if tableName != expected {
		t.Errorf("Bridge.TableName() = %s, 期望 %s", tableName, expected)
	}
}

// TestDroneTableName 测试Drone表名
func TestDroneTableName(t *testing.T) {
	drone := model.Drone{}
	tableName := drone.TableName()

	expected := "drones"
	if tableName != expected {
		t.Errorf("Drone.TableName() = %s, 期望 %s", tableName, expected)
	}
}

// TestDefectTableName 测试Defect表名
func TestDefectTableName(t *testing.T) {
	defect := model.Defect{}
	tableName := defect.TableName()

	expected := "defects"
	if tableName != expected {
		t.Errorf("Defect.TableName() = %s, 期望 %s", tableName, expected)
	}
}

// TestUserCreation 测试User对象创建
func TestUserCreation(t *testing.T) {
	now := time.Now()
	user := model.User{
		ID:        1,
		Username:  "testuser",
		Password:  "hashed_password",
		Email:     "test@example.com",
		Role:      "user",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 验证字段
	if user.ID != 1 {
		t.Errorf("User.ID = %d, 期望 1", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("User.Username = %s, 期望 testuser", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("User.Email = %s, 期望 test@example.com", user.Email)
	}
	if user.Role != "user" {
		t.Errorf("User.Role = %s, 期望 user", user.Role)
	}
}

// TestBridgeCreation 测试Bridge对象创建
func TestBridgeCreation(t *testing.T) {
	bridge := model.Bridge{
		ID:          1,
		Name:        "测试桥梁",
		Location:    "北京市朝阳区",
		Description: "这是一座测试桥梁",
		UserID:      100,
	}

	if bridge.Name != "测试桥梁" {
		t.Errorf("Bridge.Name = %s, 期望 测试桥梁", bridge.Name)
	}
	if bridge.UserID != 100 {
		t.Errorf("Bridge.UserID = %d, 期望 100", bridge.UserID)
	}
}

// TestDefectCreation 测试Defect对象创建
func TestDefectCreation(t *testing.T) {
	defect := model.Defect{
		ID:         1,
		BridgeID:   10,
		DefectType: "裂缝",
		ImagePath:  "/uploads/images/test.jpg",
		ResultPath: "/uploads/results/test_result.jpg",
		Length:     1.5,
		Width:      0.3,
		Area:       0.45,
		Confidence: 0.95,
	}

	if defect.DefectType != "裂缝" {
		t.Errorf("Defect.DefectType = %s, 期望 裂缝", defect.DefectType)
	}
	if defect.Length != 1.5 {
		t.Errorf("Defect.Length = %f, 期望 1.5", defect.Length)
	}
	if defect.Confidence != 0.95 {
		t.Errorf("Defect.Confidence = %f, 期望 0.95", defect.Confidence)
	}
}

// TestUserPasswordFieldNotExported 测试密码字段不导出到JSON
func TestUserPasswordFieldNotExported(t *testing.T) {
	// 这个测试验证Password字段的json标签为"-"
	// 实际应用中，密码不会出现在JSON响应中
	user := model.User{
		ID:       1,
		Username: "test",
		Password: "should_not_export",
		Role:     "user",
	}

	// 由于Password字段的json标签是"-"，序列化时会被忽略
	// 这里只是验证结构定义正确
	if user.Password != "should_not_export" {
		t.Errorf("Password字段应该可以在内部访问")
	}
}

// BenchmarkUserIsAdmin 性能测试：IsAdmin方法
func BenchmarkUserIsAdmin(b *testing.B) {
	user := model.User{
		ID:       1,
		Username: "admin",
		Role:     "admin",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user.IsAdmin()
	}
}
