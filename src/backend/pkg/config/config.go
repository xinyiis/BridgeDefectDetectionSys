// Package config 提供配置管理功能
// 负责加载和解析 config.yaml 配置文件
package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 全局配置结构体
// 包含服务器、数据库、第三方服务等所有配置项
type Config struct {
	Server         ServerConfig         `yaml:"server"`          // 服务器配置
	Database       DatabaseConfig       `yaml:"database"`        // 数据库配置
	PythonService  PythonServiceConfig  `yaml:"python_service"`  // Python 算法服务配置
	Detection      DetectionConfig      `yaml:"detection"`       // 检测流程配置
	VideoDetection VideoDetectionConfig `yaml:"video_detection"` // 视频检测流程配置
	Upload         UploadConfig         `yaml:"upload"`          // 文件上传配置
	Session        SessionConfig        `yaml:"session"`         // Session 配置
	CORS           CORSConfig           `yaml:"cors"`            // CORS 跨域配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"` // HTTP 服务端口，默认 8080
	Mode string `yaml:"mode"` // 运行模式: debug/release
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	DSN             string `yaml:"dsn"`               // 数据库连接串
	MaxIdleConns    int    `yaml:"max_idle_conns"`    // 最大空闲连接数
	MaxOpenConns    int    `yaml:"max_open_conns"`    // 最大打开连接数
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // 连接最大生命周期（秒）
}

// PythonServiceConfig Python 算法服务配置
type PythonServiceConfig struct {
	Enabled bool   `yaml:"enabled"` // 是否启用真实Python服务（false=Mock，true=HTTP）
	URL     string `yaml:"url"`     // Python 服务地址
	Timeout int    `yaml:"timeout"` // 请求超时时间（秒）
}

// DetectionConfig 检测流程配置
type DetectionConfig struct {
	PersistenceWorkers int `yaml:"persistence_workers"` // 检测持久化后台 worker 数量
}

// VideoDetectionConfig 视频检测流程配置
type VideoDetectionConfig struct {
	SampleFPS                    float64 `yaml:"sample_fps"`                      // 默认抽帧频率
	PlaybackDelaySeconds         int     `yaml:"playback_delay_seconds"`          // 前端播放缓冲时间
	TrackWindowSeconds           int     `yaml:"track_window_seconds"`            // 轨迹匹配窗口
	TrackCloseSeconds            int     `yaml:"track_close_seconds"`             // 轨迹关闭窗口
	TrackIOUThreshold            float64 `yaml:"track_iou_threshold"`             // 轨迹匹配 IoU 阈值
	TrackCenterDistanceThreshold float64 `yaml:"track_center_distance_threshold"` // 轨迹匹配中心点距离阈值
	ConfirmHits                  int     `yaml:"confirm_hits"`                    // 正式缺陷确认的最少命中次数
	CandidateConfidenceThreshold float64 `yaml:"candidate_confidence_threshold"`  // 候选观测最低置信度
	PersistConfidenceThreshold   float64 `yaml:"persist_confidence_threshold"`    // 正式缺陷入库最低置信度
	MaxQueueInflight             int     `yaml:"max_queue_inflight"`              // 帧任务最大并发数
	QueuedTimeoutSeconds         int     `yaml:"queued_timeout_seconds"`          // 启动后等待 queued 的超时时间
	ProgressIdleTimeoutSeconds   int     `yaml:"progress_idle_timeout_seconds"`   // queued 后无进度的超时时间
	CallbackBaseURL              string  `yaml:"callback_base_url"`               // 算法端回调后端基础地址
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	ImageDir  string `yaml:"image_dir"`  // 原始图片保存目录
	ResultDir string `yaml:"result_dir"` // 检测结果保存目录
	MaxSize   int    `yaml:"max_size"`   // 最大文件大小（MB）
}

// SessionConfig Session 配置
type SessionConfig struct {
	Secret     string `yaml:"secret"`      // Session 密钥
	MaxAge     int    `yaml:"max_age"`     // Session 有效期（秒）
	CookieName string `yaml:"cookie_name"` // Cookie 名称
}

// CORSConfig CORS 跨域配置
type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins"`     // 允许的源列表
	AllowCredentials bool     `yaml:"allow_credentials"` // 是否允许携带 Cookie
}

// globalConfig 全局配置实例
var globalConfig *Config

// LoadConfig 加载配置文件
// 从 config.yaml 读取配置并解析到 Config 结构体
// 返回值：
//   - *Config: 配置对象指针
//
// 使用示例：
//
//	config := config.LoadConfig()
//	fmt.Println("服务端口:", config.Server.Port)
func LoadConfig() *Config {
	if globalConfig != nil {
		return globalConfig
	}

	// 1. 读取配置文件
	configFile := "config.yaml"
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("读取配置文件失败: %v\n提示: 请确保 config.yaml 文件存在于项目根目录", err)
	}

	// 2. 解析 YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	// 3. 验证必要配置项
	applyEnvOverrides(&cfg)
	if err := validateConfig(&cfg); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	// 4. 创建上传目录（如果不存在）
	createUploadDirs(&cfg)

	globalConfig = &cfg
	log.Println("✓ 配置加载成功")
	return globalConfig
}

func applyEnvOverrides(cfg *Config) {
	if cfg == nil {
		return
	}

	if pythonServiceURL := os.Getenv("PYTHON_SERVICE_URL"); pythonServiceURL != "" {
		cfg.PythonService.URL = pythonServiceURL
	}
}

// GetConfig 获取全局配置实例
// 返回已加载的配置对象，如果未加载则先调用 LoadConfig
// 返回值：
//   - *Config: 配置对象指针
func GetConfig() *Config {
	if globalConfig == nil {
		return LoadConfig()
	}
	return globalConfig
}

// validateConfig 验证配置的有效性
// 检查必要的配置项是否存在且合法
func validateConfig(cfg *Config) error {
	// 检查数据库连接串
	if cfg.Database.DSN == "" {
		log.Fatal("数据库连接串（database.dsn）不能为空")
	}

	// 检查服务端口
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		log.Fatal("服务端口（server.port）必须在 1-65535 之间")
	}

	// 检查 Session 密钥
	if cfg.Session.Secret == "" {
		log.Fatal("Session 密钥（session.secret）不能为空")
	}

	// 生产环境警告
	if cfg.Server.Mode == "release" && cfg.Session.Secret == "bridge-detection-secret-key-change-in-production" {
		log.Println("⚠️  警告: 生产环境请修改 Session 密钥（session.secret）")
	}

	if cfg.Detection.PersistenceWorkers <= 0 {
		cfg.Detection.PersistenceWorkers = 2
	}
	if cfg.VideoDetection.SampleFPS <= 0 {
		cfg.VideoDetection.SampleFPS = 1
	}
	if cfg.VideoDetection.PlaybackDelaySeconds <= 0 {
		cfg.VideoDetection.PlaybackDelaySeconds = 6
	}
	if cfg.VideoDetection.TrackWindowSeconds <= 0 {
		cfg.VideoDetection.TrackWindowSeconds = 3
	}
	if cfg.VideoDetection.TrackCloseSeconds <= 0 {
		cfg.VideoDetection.TrackCloseSeconds = 4
	}
	if cfg.VideoDetection.TrackIOUThreshold <= 0 {
		cfg.VideoDetection.TrackIOUThreshold = 0.3
	}
	if cfg.VideoDetection.TrackCenterDistanceThreshold <= 0 {
		cfg.VideoDetection.TrackCenterDistanceThreshold = 0.08
	}
	if cfg.VideoDetection.ConfirmHits <= 0 {
		cfg.VideoDetection.ConfirmHits = 3
	}
	if cfg.VideoDetection.CandidateConfidenceThreshold <= 0 {
		cfg.VideoDetection.CandidateConfidenceThreshold = 0.45
	}
	if cfg.VideoDetection.PersistConfidenceThreshold <= 0 {
		cfg.VideoDetection.PersistConfidenceThreshold = 0.55
	}
	if cfg.VideoDetection.MaxQueueInflight <= 0 {
		cfg.VideoDetection.MaxQueueInflight = 8
	}
	if cfg.VideoDetection.QueuedTimeoutSeconds <= 0 {
		cfg.VideoDetection.QueuedTimeoutSeconds = 5
	}
	if cfg.VideoDetection.ProgressIdleTimeoutSeconds <= 0 {
		cfg.VideoDetection.ProgressIdleTimeoutSeconds = 15
	}
	if cfg.VideoDetection.CallbackBaseURL == "" {
		cfg.VideoDetection.CallbackBaseURL = fmt.Sprintf("http://localhost:%d/api/v1/detection/video/callback", cfg.Server.Port)
	}

	return nil
}

// createUploadDirs 创建上传目录
// 如果目录不存在则自动创建
func createUploadDirs(cfg *Config) {
	dirs := []string{
		cfg.Upload.ImageDir,
		cfg.Upload.ResultDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("⚠️  创建目录失败 %s: %v", dir, err)
		}
	}
}

// GetConnMaxLifetime 获取连接最大生命周期（time.Duration 类型）
// 用于设置数据库连接池的连接生命周期
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}

// GetTimeout 获取超时时间（time.Duration 类型）
// 用于设置 HTTP 客户端的请求超时
func (p *PythonServiceConfig) GetTimeout() time.Duration {
	return time.Duration(p.Timeout) * time.Second
}
