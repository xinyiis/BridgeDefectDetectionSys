// Package cache 实现内存缓存
package cache

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MemoryCache 内存缓存（线程安全）
type MemoryCache struct {
	data map[string][]byte    // 存储序列化后的数据
	ttl  map[string]time.Time // 存储过期时间
	mu   sync.RWMutex         // 读写锁
}

// NewMemoryCache 创建内存缓存实例
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string][]byte),
		ttl:  make(map[string]time.Time),
	}

	// 启动后台清理任务（每分钟清理过期数据）
	go cache.cleanupExpired()

	return cache
}

// Get 获取缓存（自动反序列化）
// 参数：
//   - key: 缓存键
//   - dest: 目标对象指针（用于反序列化）
// 返回：
//   - error: 缓存未命中或已过期返回错误
func (c *MemoryCache) Get(key string, dest interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 检查是否存在
	data, exists := c.data[key]
	if !exists {
		return ErrCacheMiss
	}

	// 检查是否过期
	if expireTime, ok := c.ttl[key]; ok {
		if time.Now().After(expireTime) {
			return ErrCacheExpired
		}
	}

	// 反序列化
	return json.Unmarshal(data, dest)
}

// Set 设置缓存（自动序列化）
// 参数：
//   - key: 缓存键
//   - value: 缓存值（任意类型，会被JSON序列化）
//   - ttl: 过期时间（0表示永不过期）
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	// 序列化
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = data

	if ttl > 0 {
		c.ttl[key] = time.Now().Add(ttl)
	}

	return nil
}

// Del 删除缓存
// 参数：
//   - key: 缓存键
func (c *MemoryCache) Del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	delete(c.ttl, key)
}

// DelPattern 删除匹配模式的所有缓存
// 支持通配符 * （例如：stats:overview:* 删除所有用户的概览统计）
// 参数：
//   - pattern: 匹配模式
func (c *MemoryCache) DelPattern(pattern string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 将通配符 * 替换为空，用于简单的字符串匹配
	pattern = strings.ReplaceAll(pattern, "*", "")

	count := 0
	for key := range c.data {
		if strings.Contains(key, pattern) {
			delete(c.data, key)
			delete(c.ttl, key)
			count++
		}
	}

	return count
}

// Clear 清空所有缓存
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string][]byte)
	c.ttl = make(map[string]time.Time)
}

// Size 获取缓存条目数量
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.data)
}

// cleanupExpired 后台清理过期数据（每分钟执行一次）
func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()

		for key, expireTime := range c.ttl {
			if now.After(expireTime) {
				delete(c.data, key)
				delete(c.ttl, key)
			}
		}

		c.mu.Unlock()
	}
}

// 错误定义
var (
	ErrCacheMiss    = fmt.Errorf("cache miss")
	ErrCacheExpired = fmt.Errorf("cache expired")
)

// 全局缓存实例
var StatsCache = NewMemoryCache()

// 缓存键常量
const (
	CacheKeyOverview        = "stats:overview:%d"        // stats:overview:1 (用户ID)
	CacheKeyDefectTypes     = "stats:defect_types:%d:%d" // stats:defect_types:1:30 (用户ID:天数)
	CacheKeyDefectTrend     = "stats:trend:%d:%d:%s"     // stats:trend:1:7:day (用户ID:天数:粒度)
	CacheKeyBridgeRanking   = "stats:ranking:%d:%d"      // stats:ranking:1:10 (用户ID:数量)
	CacheKeyRecentDetection = "stats:recent:%d:%d"       // stats:recent:1:10 (用户ID:数量)
	CacheKeyHighRiskAlert   = "stats:alerts:%d:%s:%d"    // stats:alerts:1:urgent:20 (用户ID:严重度:数量)
)

// TTL常量
const (
	CacheTTLOverview   = 5 * time.Minute  // 概览统计：5分钟
	CacheTTLDefectType = 10 * time.Minute // 类型分布：10分钟
	CacheTTLTrend      = 1 * time.Hour    // 趋势统计：1小时
	CacheTTLRanking    = 10 * time.Minute // 排名数据：10分钟
	CacheTTLRecent     = 3 * time.Minute  // 最近检测：3分钟
	CacheTTLAlert      = 5 * time.Minute  // 高危告警：5分钟
)
