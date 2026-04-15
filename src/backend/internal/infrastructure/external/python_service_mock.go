// Package external 处理外部服务集成
package external

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
)

// MockPythonService Python服务Mock实现
// 用于开发和测试阶段，模拟AI检测返回固定结果
type MockPythonService struct{}

// NewMockPythonService 创建Mock Python服务实例
func NewMockPythonService() *MockPythonService {
	return &MockPythonService{}
}

// DetectDefect 模拟缺陷检测
// 返回随机生成的1-3个缺陷，用于测试多缺陷处理逻辑
func (s *MockPythonService) DetectDefect(imagePath, modelName, segmentModelType string, pixelRatio float64) (*service.PythonDetectionResult, error) {
	// 模拟处理时间
	time.Sleep(100 * time.Millisecond)

	// 随机生成1-3个缺陷
	rand.Seed(time.Now().UnixNano())
	numDefects := rand.Intn(3) + 1

	defects := make([]service.DefectDetection, numDefects)
	defectTypes := []string{"Crack", "Comb", "Breakage", "Seepage", "Reinforcement"}

	for i := 0; i < numDefects; i++ {
		// 随机选择缺陷类型
		defectType := defectTypes[rand.Intn(len(defectTypes))]

		boxWidthRatio := 0.1 + rand.Float64()*0.2
		boxHeightRatio := 0.05 + rand.Float64()*0.15
		xCenter := boxWidthRatio/2 + rand.Float64()*(1-boxWidthRatio)
		yCenter := boxHeightRatio/2 + rand.Float64()*(1-boxHeightRatio)

		defects[i] = service.DefectDetection{
			DefectType: defectType,
			BBox: service.BBoxData{
				YOLOCoords: [4]float64{xCenter, yCenter, boxWidthRatio, boxHeightRatio},
			},
			Length:     0,
			Width:      0,
			Area:       0,
			Confidence: 0.80 + rand.Float64()*0.2, // 0.8-1.0
		}
	}

	// 模拟结果图（空Base64）
	mockImage := []byte("mock_result_image_data")
	resultImageBase64 := base64.StdEncoding.EncodeToString(mockImage)

	return &service.PythonDetectionResult{
		Success:        true,
		TotalDefects:   numDefects,
		Defects:        defects,
		ResultImage:    resultImageBase64,
		ProcessingTime: 0.100 + rand.Float64()*0.1, // 0.1-0.2秒
	}, nil
}

// Detect 模拟视频帧检测。
func (s *MockPythonService) Detect(string, *service.DetectRequest) (*service.DetectResult, error) {
	mockImage := []byte("mock_video_detection_image")

	return &service.DetectResult{
		Status:    "success",
		ModelUsed: "baseline",
		YOLOBBoxes: []service.BBoxItem{
			{
				BoxID:      1,
				ClassIdx:   0,
				ClassName:  "Crack",
				YOLOCoords: [4]float64{0.5, 0.5, 0.2, 0.1},
				Confidence: 0.91,
			},
		},
		ImageResult: base64.StdEncoding.EncodeToString(mockImage),
	}, nil
}

// Preprocess 模拟图像预处理。
func (s *MockPythonService) Preprocess(string, string) (*service.PreprocessResult, error) {
	return &service.PreprocessResult{
		Status:      "success",
		ImageBase64: base64.StdEncoding.EncodeToString([]byte("mock_preprocess_image")),
	}, nil
}

// Segment 模拟实例分割。
func (s *MockPythonService) Segment(string, *service.SegmentRequest) (*service.SegmentResult, error) {
	return &service.SegmentResult{
		Status:      "success",
		FusionImage: base64.StdEncoding.EncodeToString([]byte("mock_segment_image")),
		IndividualMasks: []service.IndividualMask{
			{
				BoxIndex:   0,
				Label:      "裂缝",
				MaskBase64: base64.StdEncoding.EncodeToString([]byte("mock_mask")),
			},
		},
	}, nil
}

// EnqueueVideoFrameDetect 模拟视频帧异步入队。
func (s *MockPythonService) EnqueueVideoFrameDetect(req *service.VideoFrameDetectEnqueueRequest) (*service.VideoFrameDetectEnqueueResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("视频帧入队请求不能为空")
	}
	if req.RequestID == "" || req.TaskID == "" {
		return nil, fmt.Errorf("request_id 和 task_id 不能为空")
	}

	return &service.VideoFrameDetectEnqueueResponse{
		Status:    "accepted",
		RequestID: req.RequestID,
		TaskID:    req.TaskID,
		QueuedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}
