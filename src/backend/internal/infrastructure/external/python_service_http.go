// Package external 处理外部服务集成
package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
)

// HTTPPythonService Python服务HTTP实现
// 通过HTTP请求调用真实的Python AI服务
type HTTPPythonService struct {
	baseURL string       // Python服务基础URL
	client  *http.Client // HTTP客户端
}

// NewHTTPPythonService 创建HTTP Python服务实例
// 参数：
//   - baseURL: Python服务基础URL（如：http://localhost:8000）
func NewHTTPPythonService(baseURL string) *HTTPPythonService {
	return &HTTPPythonService{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second, // 30秒超时
		},
	}
}

// DetectDefect 调用Python服务进行缺陷检测
func (s *HTTPPythonService) DetectDefect(imagePath, modelName string, pixelRatio float64) (*service.PythonDetectionResult, error) {
	// 1. 构建请求体
	requestBody := map[string]interface{}{
		"image_path":  imagePath,
		"model_name":  modelName,
		"pixel_ratio": pixelRatio,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	// 2. 发送HTTP POST请求
	url := fmt.Sprintf("%s/api/detect", s.baseURL)
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("调用Python服务失败: %w", err)
	}
	defer resp.Body.Close()

	// 3. 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Python服务返回错误: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 4. 解析响应
	var result service.PythonDetectionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析Python响应失败: %w", err)
	}

	// 5. 检查业务成功标志
	if !result.Success {
		return nil, fmt.Errorf("Python检测失败: %s", result.ErrorMessage)
	}

	return &result, nil
}
