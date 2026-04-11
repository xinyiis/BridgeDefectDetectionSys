// Package external 处理外部服务集成
package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
		body, _ := io.ReadAll(resp.Body)
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

// Detect 调用 Python YOLO 检测接口。
func (s *HTTPPythonService) Detect(imagePath string, req *service.DetectRequest) (*service.DetectResult, error) {
	body, contentType, err := buildMultipartRequest(imagePath, func(writer *multipart.Writer) error {
		if req == nil {
			return nil
		}
		if req.ModelName != "" {
			if err := writer.WriteField("model_name", req.ModelName); err != nil {
				return err
			}
		}
		if req.Conf > 0 {
			if err := writer.WriteField("conf", fmt.Sprintf("%.2f", req.Conf)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, s.baseURL+"/algo/detect", body)
	if err != nil {
		return nil, fmt.Errorf("创建检测请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", contentType)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用检测接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("检测接口返回错误: %d %s", resp.StatusCode, string(bodyBytes))
	}

	var result service.DetectResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析检测响应失败: %w", err)
	}

	return &result, nil
}

// Preprocess 调用 Python 预处理接口。
func (s *HTTPPythonService) Preprocess(imagePath string, mode string) (*service.PreprocessResult, error) {
	body, contentType, err := buildMultipartRequest(imagePath, func(writer *multipart.Writer) error {
		if mode == "" {
			return nil
		}
		return writer.WriteField("mode", mode)
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, s.baseURL+"/algo/preprocess", body)
	if err != nil {
		return nil, fmt.Errorf("创建预处理请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", contentType)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用预处理接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("预处理接口返回错误: %d %s", resp.StatusCode, string(bodyBytes))
	}

	var result service.PreprocessResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析预处理响应失败: %w", err)
	}

	return &result, nil
}

// Segment 调用 Python 分割接口。
func (s *HTTPPythonService) Segment(imagePath string, req *service.SegmentRequest) (*service.SegmentResult, error) {
	body, contentType, err := buildMultipartRequest(imagePath, func(writer *multipart.Writer) error {
		if req == nil {
			return nil
		}
		if req.BBoxesJSON != "" {
			if err := writer.WriteField("yolo_bboxes", req.BBoxesJSON); err != nil {
				return err
			}
		}
		if req.Alpha > 0 {
			if err := writer.WriteField("alpha", fmt.Sprintf("%.2f", req.Alpha)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, s.baseURL+"/algo/segment", body)
	if err != nil {
		return nil, fmt.Errorf("创建分割请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", contentType)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用分割接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("分割接口返回错误: %d %s", resp.StatusCode, string(bodyBytes))
	}

	var result service.SegmentResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析分割响应失败: %w", err)
	}

	return &result, nil
}

func buildMultipartRequest(imagePath string, writeFields func(writer *multipart.Writer) error) (*bytes.Buffer, string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, "", fmt.Errorf("打开图片失败: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(imagePath))
	if err != nil {
		return nil, "", fmt.Errorf("创建文件字段失败: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, "", fmt.Errorf("写入上传文件失败: %w", err)
	}

	if writeFields != nil {
		if err := writeFields(writer); err != nil {
			return nil, "", fmt.Errorf("写入表单字段失败: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("关闭 multipart writer 失败: %w", err)
	}

	return body, writer.FormDataContentType(), nil
}
