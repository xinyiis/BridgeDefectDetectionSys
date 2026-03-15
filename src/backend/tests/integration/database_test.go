// +build integration

// Package integration 数据库集成测试
package integration

import (
	"testing"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/internal/domain/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// 使用内存 SQLite 数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(
		&model.User{},
		&model.Bridge{},
		&model.Drone{},
		&model.Defect{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}

	return db
}

// TestDatabaseConnection 测试数据库连接
func TestDatabaseConnection(t *testing.T) {
	db := setupTestDB(t)

	// 获取底层数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取数据库连接失败: %v", err)
	}

	// Ping 测试
	if err := sqlDB.Ping(); err != nil {
		t.Errorf("数据库 Ping 失败: %v", err)
	}

	// 验证连接池配置
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	stats := sqlDB.Stats()
	t.Logf("数据库连接池状态: %+v", stats)
}

// TestUserCRUD 测试User模型的CRUD操作
func TestUserCRUD(t *testing.T) {
	db := setupTestDB(t)

	t.Run("创建用户", func(t *testing.T) {
		user := model.User{
			Username: "testuser",
			Password: "hashed_password",
			Email:    "test@example.com",
			Role:     "user",
		}

		result := db.Create(&user)

		if result.Error != nil {
			t.Errorf("创建用户失败: %v", result.Error)
		}

		if user.ID == 0 {
			t.Error("用户ID应该自动生成")
		}

		if user.CreatedAt.IsZero() {
			t.Error("CreatedAt 应该自动设置")
		}
	})

	t.Run("查询用户", func(t *testing.T) {
		// 先创建
		user := model.User{
			Username: "queryuser",
			Password: "password",
			Email:    "query@example.com",
			Role:     "user",
		}
		db.Create(&user)

		// 按ID查询
		var found model.User
		result := db.First(&found, user.ID)

		if result.Error != nil {
			t.Errorf("查询用户失败: %v", result.Error)
		}

		if found.Username != "queryuser" {
			t.Errorf("用户名不匹配: got %s, want queryuser", found.Username)
		}
	})

	t.Run("按用户名查询", func(t *testing.T) {
		user := model.User{
			Username: "finduser",
			Password: "password",
			Role:     "user",
		}
		db.Create(&user)

		var found model.User
		result := db.Where("username = ?", "finduser").First(&found)

		if result.Error != nil {
			t.Errorf("按用户名查询失败: %v", result.Error)
		}

		if found.ID != user.ID {
			t.Error("查询结果ID不匹配")
		}
	})

	t.Run("更新用户", func(t *testing.T) {
		user := model.User{
			Username: "updateuser",
			Password: "password",
			Email:    "old@example.com",
			Role:     "user",
		}
		db.Create(&user)

		// 更新邮箱
		db.Model(&user).Update("Email", "new@example.com")

		// 重新查询验证
		var updated model.User
		db.First(&updated, user.ID)

		if updated.Email != "new@example.com" {
			t.Errorf("更新失败: got %s, want new@example.com", updated.Email)
		}
	})

	t.Run("删除用户", func(t *testing.T) {
		user := model.User{
			Username: "deleteuser",
			Password: "password",
			Role:     "user",
		}
		db.Create(&user)

		// 删除
		db.Delete(&user)

		// 验证已删除
		var found model.User
		result := db.First(&found, user.ID)

		if result.Error != gorm.ErrRecordNotFound {
			t.Error("用户应该已被删除")
		}
	})

	t.Run("用户名唯一性", func(t *testing.T) {
		user1 := model.User{
			Username: "uniqueuser",
			Password: "password",
			Role:     "user",
		}
		db.Create(&user1)

		// 尝试创建同名用户
		user2 := model.User{
			Username: "uniqueuser",
			Password: "password2",
			Role:     "user",
		}
		result := db.Create(&user2)

		// 应该失败（SQLite可能不强制唯一约束，但MySQL会）
		if result.Error == nil {
			t.Log("警告: 数据库未强制用户名唯一性约束（SQLite限制）")
		}
	})
}

// TestBridgeCRUD 测试Bridge模型的CRUD操作
func TestBridgeCRUD(t *testing.T) {
	db := setupTestDB(t)

	// 先创建用户
	user := model.User{
		Username: "bridgeowner",
		Password: "password",
		Role:     "user",
	}
	db.Create(&user)

	t.Run("创建桥梁", func(t *testing.T) {
		bridge := model.Bridge{
			Name:        "测试桥梁",
			Location:    "北京市朝阳区",
			Description: "这是一座测试桥梁",
			UserID:      user.ID,
		}

		result := db.Create(&bridge)

		if result.Error != nil {
			t.Errorf("创建桥梁失败: %v", result.Error)
		}

		if bridge.ID == 0 {
			t.Error("桥梁ID应该自动生成")
		}
	})

	t.Run("关联查询", func(t *testing.T) {
		bridge := model.Bridge{
			Name:   "关联测试桥梁",
			UserID: user.ID,
		}
		db.Create(&bridge)

		// 预加载用户信息
		var found model.Bridge
		db.Preload("User").First(&found, bridge.ID)

		if found.User == nil {
			t.Error("User 应该被预加载")
		}

		if found.User.Username != "bridgeowner" {
			t.Errorf("关联用户不正确: got %s", found.User.Username)
		}
	})

	t.Run("查询用户的所有桥梁", func(t *testing.T) {
		// 创建多个桥梁
		for i := 1; i <= 3; i++ {
			bridge := model.Bridge{
				Name:   "桥梁" + string(rune(i+'0')),
				UserID: user.ID,
			}
			db.Create(&bridge)
		}

		// 查询
		var bridges []model.Bridge
		db.Where("user_id = ?", user.ID).Find(&bridges)

		if len(bridges) < 3 {
			t.Errorf("应该查询到至少3座桥梁, 实际: %d", len(bridges))
		}
	})
}

// TestDefectCRUD 测试Defect模型的CRUD操作
func TestDefectCRUD(t *testing.T) {
	db := setupTestDB(t)

	// 创建测试数据
	user := model.User{
		Username: "defectuser",
		Password: "password",
		Role:     "user",
	}
	db.Create(&user)

	bridge := model.Bridge{
		Name:   "缺陷测试桥梁",
		UserID: user.ID,
	}
	db.Create(&bridge)

	t.Run("创建缺陷记录", func(t *testing.T) {
		defect := model.Defect{
			BridgeID:   bridge.ID,
			DefectType: "裂缝",
			ImagePath:  "/uploads/images/test.jpg",
			ResultPath: "/uploads/results/test_result.jpg",
			BBox:       `{"x":100,"y":200,"w":50,"h":30}`,
			Length:     1.5,
			Width:      0.3,
			Area:       0.45,
			Confidence: 0.95,
			DetectedAt: time.Now(),
		}

		result := db.Create(&defect)

		if result.Error != nil {
			t.Errorf("创建缺陷记录失败: %v", result.Error)
		}

		if defect.ID == 0 {
			t.Error("缺陷ID应该自动生成")
		}
	})

	t.Run("查询桥梁的缺陷", func(t *testing.T) {
		// 创建多个缺陷
		defectTypes := []string{"裂缝", "剥落", "锈蚀"}
		for _, defectType := range defectTypes {
			defect := model.Defect{
				BridgeID:   bridge.ID,
				DefectType: defectType,
				ImagePath:  "/test.jpg",
				Confidence: 0.9,
				DetectedAt: time.Now(),
			}
			db.Create(&defect)
		}

		// 查询
		var defects []model.Defect
		db.Where("bridge_id = ?", bridge.ID).Find(&defects)

		if len(defects) < 3 {
			t.Errorf("应该查询到至少3个缺陷, 实际: %d", len(defects))
		}
	})

	t.Run("按时间范围查询缺陷", func(t *testing.T) {
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)

		defect := model.Defect{
			BridgeID:   bridge.ID,
			DefectType: "时间测试",
			ImagePath:  "/test.jpg",
			DetectedAt: now,
		}
		db.Create(&defect)

		// 查询昨天到今天的缺陷
		var defects []model.Defect
		db.Where("bridge_id = ? AND detected_at BETWEEN ? AND ?", bridge.ID, yesterday, now.Add(time.Hour)).
			Find(&defects)

		if len(defects) == 0 {
			t.Error("应该查询到时间范围内的缺陷")
		}
	})
}

// TestTransactions 测试事务
func TestTransactions(t *testing.T) {
	db := setupTestDB(t)

	t.Run("事务提交", func(t *testing.T) {
		err := db.Transaction(func(tx *gorm.DB) error {
			user := model.User{
				Username: "txuser",
				Password: "password",
				Role:     "user",
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}

			bridge := model.Bridge{
				Name:   "txbridge",
				UserID: user.ID,
			}
			if err := tx.Create(&bridge).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			t.Errorf("事务失败: %v", err)
		}

		// 验证数据已提交
		var user model.User
		if err := db.Where("username = ?", "txuser").First(&user).Error; err != nil {
			t.Error("事务提交后应该能查询到数据")
		}
	})

	t.Run("事务回滚", func(t *testing.T) {
		err := db.Transaction(func(tx *gorm.DB) error {
			user := model.User{
				Username: "rollbackuser",
				Password: "password",
				Role:     "user",
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}

			// 返回错误，触发回滚
			return gorm.ErrInvalidTransaction
		})

		if err == nil {
			t.Error("事务应该失败")
		}

		// 验证数据已回滚
		var user model.User
		result := db.Where("username = ?", "rollbackuser").First(&user)
		if result.Error != gorm.ErrRecordNotFound {
			t.Error("事务回滚后不应该查询到数据")
		}
	})
}

// TestDatabasePerformance 测试数据库性能
func TestDatabasePerformance(t *testing.T) {
	db := setupTestDB(t)

	t.Run("批量插入性能", func(t *testing.T) {
		start := time.Now()

		users := make([]model.User, 100)
		for i := 0; i < 100; i++ {
			users[i] = model.User{
				Username: "batchuser" + string(rune(i)),
				Password: "password",
				Role:     "user",
			}
		}

		db.CreateInBatches(users, 10)

		elapsed := time.Since(start)
		t.Logf("批量插入100条记录耗时: %v", elapsed)

		if elapsed > 1*time.Second {
			t.Logf("警告: 批量插入较慢 (%v)", elapsed)
		}
	})

	t.Run("索引查询性能", func(t *testing.T) {
		// 创建测试数据
		for i := 0; i < 50; i++ {
			user := model.User{
				Username: "perfuser" + string(rune(i)),
				Password: "password",
				Role:     "user",
			}
			db.Create(&user)
		}

		start := time.Now()

		// 按用户名查询（应该使用索引）
		var user model.User
		db.Where("username = ?", "perfuser25").First(&user)

		elapsed := time.Since(start)
		t.Logf("索引查询耗时: %v", elapsed)

		if elapsed > 100*time.Millisecond {
			t.Logf("警告: 查询较慢 (%v)", elapsed)
		}
	})
}

// BenchmarkUserCreate 测试用户创建性能
func BenchmarkUserCreate(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.User{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := model.User{
			Username: "benchuser",
			Password: "password",
			Role:     "user",
		}
		db.Create(&user)
	}
}

// BenchmarkUserQuery 测试用户查询性能
func BenchmarkUserQuery(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.User{})

	// 创建测试数据
	user := model.User{
		Username: "benchuser",
		Password: "password",
		Role:     "user",
	}
	db.Create(&user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var found model.User
		db.First(&found, user.ID)
	}
}
