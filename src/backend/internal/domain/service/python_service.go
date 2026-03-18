// Package service 定义领域服务接口
package service

import "encoding/json"

// PythonService Python检测服务接口
// 提供AI模型检测功能的抽象接口，支持Mock和HTTP两种实现
type PythonService interface {
	// DetectDefect 检测缺陷
	// 参数：
	//   - imagePath: 图片路径
	//   - modelName: 模型名称/版本
	//   - pixelRatio: 像素实际系数（米/像素）
	// 返回：
	//   - *PythonDetectionResult: 检测结果（包含多个缺陷）
	//   - error: 错误信息
	DetectDefect(imagePath, modelName string, pixelRatio float64) (*PythonDetectionResult, error)
}

// PythonDetectionResult Python检测返回结果
type PythonDetectionResult struct {
	Success        bool              `json:"success"`         // 是否成功
	TotalDefects   int               `json:"total_defects"`   // 检测到的缺陷总数
	Defects        []DefectDetection `json:"defects"`         // 缺陷列表
	ResultImage    string            `json:"result_image"`    // 结果图base64（可选）
	ProcessingTime float64           `json:"processing_time"` // 处理时间（秒）
	ErrorMessage   string            `json:"error_message,omitempty"` // 错误信息
}

// DefectDetection 单个缺陷检测结果
type DefectDetection struct {
	DefectType string   `json:"defect_type"` // 缺陷类型
	BBox       BBoxData `json:"bbox"`        // 边界框
	Length     float64  `json:"length"`      // 长度（米）
	Width      float64  `json:"width"`       // 宽度（米）
	Area       float64  `json:"area"`        // 面积（平方米）
	Confidence float64  `json:"confidence"`  // 置信度（0-1）
}

// BBoxData 边界框数据
type BBoxData struct {
	X      int `json:"x"`      // X坐标（像素）
	Y      int `json:"y"`      // Y坐标（像素）
	Width  int `json:"width"`  // 宽度（像素）
	Height int `json:"height"` // 高度（像素）
}

// BBoxJSON 转换为JSON字符串（用于存储到数据库）
func (d *DefectDetection) BBoxJSON() string {
	data, _ := json.Marshal(d.BBox)
	return string(data)
}
