// Package external 处理外部服务集成
package external

import (
	"encoding/base64"
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
func (s *MockPythonService) DetectDefect(imagePath, modelName string, pixelRatio float64) (*service.PythonDetectionResult, error) {
	// 模拟处理时间
	time.Sleep(100 * time.Millisecond)

	// 随机生成1-3个缺陷
	rand.Seed(time.Now().UnixNano())
	numDefects := rand.Intn(3) + 1

	defects := make([]service.DefectDetection, numDefects)
	defectTypes := []string{"裂缝", "剥落", "破损", "渗水", "锈蚀"}

	for i := 0; i < numDefects; i++ {
		// 随机选择缺陷类型
		defectType := defectTypes[rand.Intn(len(defectTypes))]

		// 生成随机边界框
		x := rand.Intn(1000)
		y := rand.Intn(800)
		bboxWidth := rand.Intn(200) + 50
		bboxHeight := rand.Intn(100) + 20

		// 计算实际尺寸（像素 * 像素系数）
		lengthPixels := float64(bboxWidth)
		widthPixels := float64(bboxHeight)
		defectLength := lengthPixels * pixelRatio
		defectWidth := widthPixels * pixelRatio
		defectArea := defectLength * defectWidth

		defects[i] = service.DefectDetection{
			DefectType: defectType,
			BBox: service.BBoxData{
				X:      x,
				Y:      y,
				Width:  bboxWidth,
				Height: bboxHeight,
			},
			Length:     defectLength,
			Width:      defectWidth,
			Area:       defectArea,
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
