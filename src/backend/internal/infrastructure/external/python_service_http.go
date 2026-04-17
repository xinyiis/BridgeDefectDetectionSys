// Package external 处理外部服务集成
package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
)

const defaultSegmentModelType = "student"

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
func (s *HTTPPythonService) DetectDefect(imagePath, modelName, segmentModelType string, pixelRatio float64) (*service.PythonDetectionResult, error) {
	_ = pixelRatio
	segmentModelType = normalizeSegmentModelType(segmentModelType)

	detectResult, err := s.Detect(imagePath, &service.DetectRequest{
		ModelName: modelName,
		Conf:      0.25,
	})
	if err != nil {
		return nil, err
	}
	if detectResult == nil {
		return nil, fmt.Errorf("检测接口返回空响应")
	}
	if detectResult.Status != "success" {
		return nil, fmt.Errorf("检测接口业务失败: %s", detectResult.Status)
	}

	defects := make([]service.DefectDetection, 0, len(detectResult.YOLOBBoxes))
	for _, item := range detectResult.YOLOBBoxes {
		defects = append(defects, service.DefectDetection{
			DefectType: item.ClassName,
			BBox: service.BBoxData{
				YOLOCoords: item.YOLOCoords,
			},
			Confidence: item.Confidence,
		})
	}

	if len(defects) == 0 {
		return &service.PythonDetectionResult{
			Success:      true,
			TotalDefects: 0,
			Defects:      defects,
			ResultImage:  detectResult.ImageResult,
		}, nil
	}

	bboxesJSON, err := json.Marshal(detectResult.YOLOBBoxes)
	if err != nil {
		return nil, fmt.Errorf("序列化检测框失败: %w", err)
	}

	// 计算 area_ratio: 米/像素 -> cm²/像素
	// 1米 = 100厘米，面积 = (长度)²
	areaRatio := pixelRatio * 100 * pixelRatio * 100

	segmentResult, err := s.Segment(imagePath, &service.SegmentRequest{
		BBoxesJSON: string(bboxesJSON),
		Alpha:      0.5,
		ModelType:  segmentModelType,
		AreaRatio:  areaRatio,
	})
	if err != nil {
		return nil, err
	}
	if segmentResult == nil {
		return nil, fmt.Errorf("分割接口返回空响应")
	}
	if segmentResult.Status != "success" {
		return nil, fmt.Errorf("分割接口业务失败: %s", segmentResult.Status)
	}

	// 将分割结果中的面积数据映射到缺陷列表
	maskMap := make(map[int]*service.IndividualMask)
	for i := range segmentResult.IndividualMasks {
		mask := &segmentResult.IndividualMasks[i]
		maskMap[mask.BoxIndex] = mask
	}

	// 更新缺陷的面积信息（从算法端返回的 actual_area，单位：cm²）
	for i := range defects {
		if mask, ok := maskMap[i]; ok {
			// actual_area 单位是 cm²，转换为 m²
			defects[i].Area = mask.ActualArea / 10000.0
			// 假设缺陷是矩形，根据面积估算长宽（这是一个简化，实际可能需要更复杂的计算）
			// 这里保持原有逻辑，或者算法端也可以返回长宽
			if defects[i].Area > 0 {
				side := math.Sqrt(defects[i].Area)
				defects[i].Length = side
				defects[i].Width = side
			}
		}
	}

	return &service.PythonDetectionResult{
		Success:      true,
		TotalDefects: len(defects),
		Defects:      defects,
		ResultImage:  segmentResult.FusionImage,
	}, nil
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
		modelType := defaultSegmentModelType
		if req == nil {
			if err := writer.WriteField("model_type", modelType); err != nil {
				return err
			}
			return nil
		}
		modelType = normalizeSegmentModelType(req.ModelType)
		if req.BBoxesJSON != "" {
			if err := writer.WriteField("bboxes_json", req.BBoxesJSON); err != nil {
				return err
			}
		}
		if req.Alpha > 0 {
			if err := writer.WriteField("alpha", fmt.Sprintf("%.2f", req.Alpha)); err != nil {
				return err
			}
		}
		if req.AreaRatio > 0 {
			if err := writer.WriteField("area_ratio", fmt.Sprintf("%.6f", req.AreaRatio)); err != nil {
				return err
			}
		}
		return writer.WriteField("model_type", modelType)
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

func normalizeSegmentModelType(modelType string) string {
	candidate := strings.TrimSpace(modelType)
	if candidate == "" {
		return defaultSegmentModelType
	}
	return candidate
}

// EnqueueVideoFrameDetect 提交视频单帧异步检测入队请求。
func (s *HTTPPythonService) EnqueueVideoFrameDetect(req *service.VideoFrameDetectEnqueueRequest) (*service.VideoFrameDetectEnqueueResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("视频帧入队请求不能为空")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化视频帧入队请求失败: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, s.baseURL+"/algo/video/frames/detect", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("创建视频帧入队请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用视频帧入队接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("视频帧入队接口返回错误: %d %s", resp.StatusCode, string(bodyBytes))
	}

	var result service.VideoFrameDetectEnqueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析视频帧入队响应失败: %w", err)
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
