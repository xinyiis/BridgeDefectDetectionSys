// Package persistence 提供数据持久化功能
// 负责数据库连接、初始化和迁移
package persistence

import (
	"log"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// InitDatabase 初始化数据库连接
// 创建数据库连接池，配置 GORM，执行表结构迁移
// 参数：
//   - cfg: 配置对象，包含数据库连接信息
//
// 返回值：
//   - *gorm.DB: GORM 数据库连接对象
//
// 使用示例：
//
//	cfg := config.LoadConfig()
//	db := persistence.InitDatabase(cfg)
func InitDatabase(cfg *config.Config) *gorm.DB {
	// 1. 配置 GORM
	gormConfig := &gorm.Config{
		// 禁用默认事务（提高性能，需要时手动开启）
		SkipDefaultTransaction: true,

		// 命名策略：使用单数表名
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 表名不加 s（user 而不是 users）
		},

		// 日志配置
		Logger: logger.Default.LogMode(getLogMode(cfg.Server.Mode)),
	}

	// 2. 打开数据库连接
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), gormConfig)
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v\n提示: 请检查 MySQL 是否运行，以及 config.yaml 中的连接配置", err)
	}

	// 3. 获取底层 sql.DB 对象，配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ 获取数据库连接池失败: %v", err)
	}

	// 配置连接池参数
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)                  // 最大空闲连接数
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)                  // 最大打开连接数
	sqlDB.SetConnMaxLifetime(cfg.Database.GetConnMaxLifetime())       // 连接最大生命周期

	// 4. 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ 数据库连接测试失败: %v", err)
	}

	log.Println("✓ 数据库连接成功")

	// 5. 自动迁移表结构
	AutoMigrate(db)

	return db
}

// AutoMigrate 自动迁移数据库表结构
// 根据 model 定义自动创建或更新表结构（不会删除字段）
// 参数：
//   - db: GORM 数据库连接对象
//
// 注意：
//   - 只会添加新字段，不会删除或修改已存在的字段
//   - 生产环境建议使用专业的迁移工具（如 golang-migrate）
func AutoMigrate(db *gorm.DB) {
	log.Println("开始数据库表结构迁移...")

	// 要迁移的模型列表
	models := []interface{}{
		&model.User{},
		&model.Bridge{},
		&model.Drone{},
		&model.Defect{},
	}

	// 执行迁移
	if err := db.AutoMigrate(models...); err != nil {
		log.Fatalf("❌ 数据库表结构迁移失败: %v", err)
	}

	log.Println("✓ 数据库表结构迁移完成")

	// 创建默认管理员账户（如果不存在）
	createDefaultAdmin(db)

	// 创建索引（GORM 的 AutoMigrate 会自动创建，这里仅作说明）
	createIndexes(db)
}

// createDefaultAdmin 创建默认管理员账户
// 如果数据库中没有管理员账户，则创建一个默认的
// 参数：
//   - db: GORM 数据库连接对象
func createDefaultAdmin(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Where("role = ?", "admin").Count(&count)

	if count == 0 {
		// 使用 bcrypt 加密默认密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("⚠️  密码加密失败: %v", err)
			return
		}

		admin := model.User{
			Username: "admin",
			Password: string(hashedPassword), // bcrypt 加密后的密码
			RealName: "系统管理员",
			Email:    "admin@example.com",
			Role:     "admin",
		}

		if err := db.Create(&admin).Error; err != nil {
			log.Printf("⚠️  创建默认管理员失败: %v", err)
		} else {
			log.Println("✓ 已创建默认管理员账户 (用户名: admin, 密码: admin123)")
			log.Println("  ⚠️  警告: 请尽快修改默认密码！")
		}
	}
}

// createIndexes 创建数据库索引
// GORM 的 AutoMigrate 会根据 struct tag 自动创建索引，这里仅作补充说明
// 参数：
//   - db: GORM 数据库连接对象
func createIndexes(db *gorm.DB) {
	// GORM 会根据以下 tag 自动创建索引：
	// - `gorm:"uniqueIndex"` : 唯一索引
	// - `gorm:"index"`       : 普通索引
	// - `gorm:"primaryKey"`  : 主键索引

	// 如果需要手动创建复合索引，可以这样：
	// db.Exec("CREATE INDEX idx_bridge_user ON bridges(user_id, created_at)")

	log.Println("✓ 数据库索引检查完成")
}

// getLogMode 根据运行模式返回日志级别
// 参数：
//   - mode: 运行模式（debug/release）
//
// 返回值：
//   - logger.LogLevel: GORM 日志级别
func getLogMode(mode string) logger.LogLevel {
	switch mode {
	case "debug":
		return logger.Info // 开发环境：显示所有 SQL 语句
	case "release":
		return logger.Warn // 生产环境：只显示警告和错误
	default:
		return logger.Info
	}
}

// CloseDatabase 关闭数据库连接
// 应在程序退出时调用，确保资源正确释放
// 参数：
//   - db: GORM 数据库连接对象
//
// 使用示例：
//
//	defer persistence.CloseDatabase(db)
func CloseDatabase(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("⚠️  获取数据库连接失败: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("⚠️  关闭数据库连接失败: %v", err)
	} else {
		log.Println("✓ 数据库连接已关闭")
	}
}
