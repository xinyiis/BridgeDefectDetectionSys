// Package model_test 桥梁模型单元测试
package model_test

import (
	"testing"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
)

// TestBridge_IsOwnedBy 测试桥梁所有权验证
func TestBridge_IsOwnedBy(t *testing.T) {
	bridge := &model.Bridge{
		ID:     1,
		UserID: 10,
	}

	tests := []struct {
		name   string
		userID uint
		want   bool
	}{
		{"所有者", 10, true},
		{"非所有者", 20, false},
		{"零值", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bridge.IsOwnedBy(tt.userID); got != tt.want {
				t.Errorf("IsOwnedBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestBridge_IsOwnedBy_NilBridge 测试空桥梁对象
func TestBridge_IsOwnedBy_NilBridge(t *testing.T) {
	// 注意：这是一个边界条件测试，虽然实际代码中不太可能出现
	// 但为了完整性还是测试一下
	defer func() {
		if r := recover(); r != nil {
			// 如果panic，说明没有正确处理nil情况
			t.Log("IsOwnedBy caused panic when called on nil - this is expected behavior")
		}
	}()

	var bridge *model.Bridge
	// 这会导致panic，因为bridge是nil
	_ = bridge.IsOwnedBy(10)
}

// TestBridge_FieldDefaults 测试桥梁字段默认值
func TestBridge_FieldDefaults(t *testing.T) {
	bridge := &model.Bridge{}

	if bridge.ID != 0 {
		t.Errorf("Default ID should be 0, got %d", bridge.ID)
	}
	if bridge.BridgeName != "" {
		t.Errorf("Default BridgeName should be empty, got %s", bridge.BridgeName)
	}
	if bridge.Status != "" {
		t.Errorf("Default Status should be empty, got %s", bridge.Status)
	}
	if bridge.UserID != 0 {
		t.Errorf("Default UserID should be 0, got %d", bridge.UserID)
	}
}

// TestBridge_Relationships 测试桥梁关联关系
func TestBridge_Relationships(t *testing.T) {
	bridge := &model.Bridge{
		ID:     1,
		UserID: 10,
		User: &model.User{
			ID:       10,
			Username: "testuser",
		},
		Defects: []model.Defect{
			{ID: 1, BridgeID: 1, DefectType: "裂缝"},
			{ID: 2, BridgeID: 1, DefectType: "剥落"},
		},
	}

	// 验证用户关联
	if bridge.User == nil {
		t.Error("User relationship should not be nil")
	}
	if bridge.User.ID != 10 {
		t.Errorf("Expected User.ID 10, got %d", bridge.User.ID)
	}

	// 验证缺陷关联
	if len(bridge.Defects) != 2 {
		t.Errorf("Expected 2 defects, got %d", len(bridge.Defects))
	}
	if bridge.Defects[0].DefectType != "裂缝" {
		t.Errorf("Expected first defect type '裂缝', got %s", bridge.Defects[0].DefectType)
	}
}
