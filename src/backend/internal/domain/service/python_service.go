// Package service 定义领域服务接口
package service

import (
	"encoding/json"
	"math"
)

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

	// Detect 调用 YOLO 目标检测接口。
	Detect(imagePath string, req *DetectRequest) (*DetectResult, error)

	// Preprocess 调用图像预处理接口。
	Preprocess(imagePath string, mode string) (*PreprocessResult, error)

	// Segment 调用实例分割接口。
	Segment(imagePath string, req *SegmentRequest) (*SegmentResult, error)

	// EnqueueVideoFrameDetect 提交视频单帧异步检测任务。
	// 该接口只负责入队，实际结果通过回调返回到后端。
	EnqueueVideoFrameDetect(req *VideoFrameDetectEnqueueRequest) (*VideoFrameDetectEnqueueResponse, error)
}

// PythonDetectionResult Python检测返回结果
type PythonDetectionResult struct {
	Success        bool              `json:"success"`                 // 是否成功
	TotalDefects   int               `json:"total_defects"`           // 检测到的缺陷总数
	Defects        []DefectDetection `json:"defects"`                 // 缺陷列表
	ResultImage    string            `json:"result_image"`            // 结果图base64（可选）
	ProcessingTime float64           `json:"processing_time"`         // 处理时间（秒）
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
	X          int        `json:"x"`           // X坐标（像素）
	Y          int        `json:"y"`           // Y坐标（像素）
	Width      int        `json:"width"`       // 宽度（像素）
	Height     int        `json:"height"`      // 高度（像素）
	YOLOCoords [4]float64 `json:"yolo_coords"` // 算法原始返回的归一化坐标
}

// DetectRequest YOLO 检测请求。
type DetectRequest struct {
	ModelName string  // 模型名称
	Conf      float64 // 置信度阈值
}

// SegmentRequest 实例分割请求。
type SegmentRequest struct {
	BBoxesJSON string  // bbox JSON 字符串
	Alpha      float64 // 掩码透明度
}

// PreprocessResult 图像预处理响应。
type PreprocessResult struct {
	Status      string `json:"status"`
	ImageBase64 string `json:"image_base64"`
}

// DetectResult 检测响应。
type DetectResult struct {
	Status      string     `json:"status"`
	ModelUsed   string     `json:"model_used"`
	YOLOBBoxes  []BBoxItem `json:"yolo_bboxes"`
	ImageResult string     `json:"image_results"`
}

// BBoxItem 单个检测框。
type BBoxItem struct {
	BoxID      int        `json:"box_id"`
	ClassIdx   int        `json:"class_idx"`
	ClassName  string     `json:"class_name"`
	YOLOCoords [4]float64 `json:"yolo_coords"`
	Confidence float64    `json:"confidence"`
}

// SegmentResult 实例分割响应。
type SegmentResult struct {
	Status          string           `json:"status"`
	FusionImage     string           `json:"fusion_image"`
	IndividualMasks []IndividualMask `json:"individual_masks"`
}

// VideoFrameDetectEnqueueRequest 视频单帧异步检测入队请求。
type VideoFrameDetectEnqueueRequest struct {
	RequestID       string  `json:"request_id"`
	TaskID          string  `json:"task_id"`
	BridgeID        uint    `json:"bridge_id"`
	FrameNo         int     `json:"frame_no"`
	TimestampMS     int     `json:"timestamp_ms"`
	FrameRef        string  `json:"frame_ref"`
	ModelName       string  `json:"model_name"`
	Conf            float64 `json:"conf"`
	CallbackBaseURL string  `json:"callback_base_url"`
}

// VideoFrameDetectEnqueueResponse 视频单帧异步检测入队响应。
type VideoFrameDetectEnqueueResponse struct {
	Status       string `json:"status"`
	RequestID    string `json:"request_id"`
	TaskID       string `json:"task_id"`
	QueuedAt     string `json:"queued_at,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// IndividualMask 单个分割掩码。
type IndividualMask struct {
	BoxIndex   int    `json:"box_index"`
	Label      string `json:"label"`
	MaskBase64 string `json:"mask_base64"`
}

// ClassNameToChinese 病害类别映射。
var ClassNameToChinese = map[int]string{
	0: "裂缝",
	1: "破损",
	2: "剥落",
	3: "孔洞",
	4: "钢筋外露",
	5: "渗水",
}

// BBoxJSON 转换为JSON字符串（用于存储到数据库）
func (d *DefectDetection) BBoxJSON() string {
	data, _ := json.Marshal(d.BBox)
	return string(data)
}

// CalculatePhysicalDimensions 将 YOLO 归一化坐标转换为像素框和物理尺寸。
func CalculatePhysicalDimensions(
	coords [4]float64,
	imgW, imgH int,
	pixelRatio float64,
) (x, y, w, h int, length, width, area float64) {
	boxW := coords[2] * float64(imgW)
	boxH := coords[3] * float64(imgH)
	left := (coords[0] - coords[2]/2) * float64(imgW)
	top := (coords[1] - coords[3]/2) * float64(imgH)

	x = int(math.Round(left))
	y = int(math.Round(top))
	w = int(math.Round(boxW))
	h = int(math.Round(boxH))

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}

	length = float64(w) * pixelRatio
	width = float64(h) * pixelRatio
	area = length * width
	return
}
