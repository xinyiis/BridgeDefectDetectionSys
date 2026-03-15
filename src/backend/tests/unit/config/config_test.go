// Package config_test 配置管理单元测试
package config_test

import (
	"os"
	"testing"

	"github.com/xinyiis/BridgeDefectDetectionSys/pkg/config"
)

// TestLoadConfig 测试配置加载功能
func TestLoadConfig(t *testing.T) {
	// 创建测试配置文件
	testConfig := `
server:
  port: 8080
  mode: debug

database:
  dsn: "root:123456@tcp(localhost:3306)/test_db"
  max_idle_conns: 5
  max_open_conns: 50
  conn_max_lifetime: 1800

python_service:
  url: "http://localhost:8000"
  timeout: 30

upload:
  image_dir: "./test_uploads/images"
  result_dir: "./test_uploads/results"
  max_size: 10

session:
  secret: "test-secret-key"
  max_age: 3600
  cookie_name: "test_session"

cors:
  allow_origins:
    - "http://localhost:5173"
  allow_credentials: true
`

	// 保存原始配置文件（如果存在）
	originalConfig, _ := os.ReadFile("config.yaml")
	defer func() {
		if originalConfig != nil {
			os.WriteFile("config.yaml", originalConfig, 0644)
		}
	}()

	// 写入测试配置
	err := os.WriteFile("config.yaml", []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("创建测试配置文件失败: %v", err)
	}

	// 测试加载配置
	cfg := config.LoadConfig()

	// 验证配置是否正确加载
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"服务器端口", cfg.Server.Port, 8080},
		{"服务器模式", cfg.Server.Mode, "debug"},
		{"数据库DSN", cfg.Database.DSN, "root:123456@tcp(localhost:3306)/test_db"},
		{"最大空闲连接", cfg.Database.MaxIdleConns, 5},
		{"最大打开连接", cfg.Database.MaxOpenConns, 50},
		{"Python服务URL", cfg.PythonService.URL, "http://localhost:8000"},
		{"Python超时时间", cfg.PythonService.Timeout, 30},
		{"图片目录", cfg.Upload.ImageDir, "./test_uploads/images"},
		{"结果目录", cfg.Upload.ResultDir, "./test_uploads/results"},
		{"Session密钥", cfg.Session.Secret, "test-secret-key"},
		{"Session有效期", cfg.Session.MaxAge, 3600},
		{"CORS允许携带Cookie", cfg.CORS.AllowCredentials, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, 期望 %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	// 测试CORS Origins数组
	if len(cfg.CORS.AllowOrigins) != 1 {
		t.Errorf("CORS AllowOrigins 长度 = %d, 期望 1", len(cfg.CORS.AllowOrigins))
	}
	if cfg.CORS.AllowOrigins[0] != "http://localhost:5173" {
		t.Errorf("CORS AllowOrigins[0] = %s, 期望 http://localhost:5173", cfg.CORS.AllowOrigins[0])
	}

	// 清理测试创建的目录
	os.RemoveAll("./test_uploads")
}

// TestGetConfig 测试全局配置获取
func TestGetConfig(t *testing.T) {
	// 直接调用获取配置
	cfg := config.GetConfig()

	if cfg == nil {
		t.Error("GetConfig() 不应返回 nil")
	}
}

// TestDatabaseConfigMethods 测试数据库配置方法
func TestDatabaseConfigMethods(t *testing.T) {
	// 测试配置加载后的方法
	cfg := config.GetConfig()

	lifetime := cfg.Database.GetConnMaxLifetime()

	// 验证返回的是 time.Duration 类型
	if lifetime.Seconds() <= 0 {
		t.Error("GetConnMaxLifetime() 应返回正数")
	}
}

// TestPythonServiceConfigMethods 测试Python服务配置方法
func TestPythonServiceConfigMethods(t *testing.T) {
	// 测试配置加载后的方法
	cfg := config.GetConfig()

	timeout := cfg.PythonService.GetTimeout()

	// 验证返回的是 time.Duration 类型
	if timeout.Seconds() <= 0 {
		t.Error("GetTimeout() 应返回正数")
	}
}

// BenchmarkGetConfig 性能测试：配置获取
func BenchmarkGetConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.GetConfig()
	}
}
