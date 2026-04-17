// Package usecase 定义应用层用例
package usecase

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
)

const defaultSegmentModelType = "student"

// DetectionUseCase 检测用例
// 处理图片上传检测业务流程
type DetectionUseCase struct {
	defectService *service.DefectService // 缺陷领域服务
	bridgeService *service.BridgeService // 桥梁领域服务
	pythonService service.PythonService  // Python检测服务
	fileService   service.FileService    // 文件服务

	persistenceQueue   chan detectionPersistenceTask
	persistenceWorkers int
	persistenceMu      sync.RWMutex
	persistenceRecords map[string]*detectionPersistenceRecord
	workerWG           sync.WaitGroup
	closeOnce          sync.Once
	stopCh             chan struct{}
}

// NewDetectionUseCase 创建检测用例实例
// 参数：
//   - defectService: 缺陷领域服务
//   - bridgeService: 桥梁领域服务
//   - pythonService: Python检测服务
//   - fileService: 文件服务
//
// 返回：
//   - *DetectionUseCase: 检测用例实例
func NewDetectionUseCase(
	defectService *service.DefectService,
	bridgeService *service.BridgeService,
	pythonService service.PythonService,
	fileService service.FileService,
) *DetectionUseCase {
	return NewDetectionUseCaseWithPersistenceWorkers(defectService, bridgeService, pythonService, fileService, 2)
}

// NewDetectionUseCaseWithPersistenceWorkers 创建支持自定义持久化 worker 数量的检测用例实例。
func NewDetectionUseCaseWithPersistenceWorkers(
	defectService *service.DefectService,
	bridgeService *service.BridgeService,
	pythonService service.PythonService,
	fileService service.FileService,
	persistenceWorkers int,
) *DetectionUseCase {
	if persistenceWorkers <= 0 {
		persistenceWorkers = 2
	}

	uc := &DetectionUseCase{
		defectService:      defectService,
		bridgeService:      bridgeService,
		pythonService:      pythonService,
		fileService:        fileService,
		persistenceQueue:   make(chan detectionPersistenceTask, 32),
		persistenceWorkers: persistenceWorkers,
		persistenceRecords: make(map[string]*detectionPersistenceRecord),
		stopCh:             make(chan struct{}),
	}

	uc.startPersistenceWorkers(persistenceWorkers)
	return uc
}

// UploadAndDetect 上传图片并进行缺陷检测
// 参数：
//   - req: 检测上传请求
//   - currentUser: 当前用户
//
// 返回：
//   - *dto.DetectionResponse: 检测结果响应（包含多个缺陷）
//   - error: 操作错误
func (uc *DetectionUseCase) UploadAndDetect(req *dto.DetectionUploadRequest, currentUser *model.User) (*dto.DetectionResponse, error) {
	startTime := time.Now()

	// 1. 验证桥梁权限
	bridge, err := uc.bridgeService.GetByID(req.BridgeID)
	if err != nil {
		return nil, err
	}
	if bridge == nil {
		return nil, errors.New("桥梁不存在")
	}

	// 普通用户只能上传到自己的桥梁
	if !currentUser.IsAdmin() && !bridge.IsOwnedBy(currentUser.ID) {
		return nil, errors.New("无权访问此桥梁")
	}

	// 2. 同步保存临时图片，供算法读取
	tempImagePath, err := uc.fileService.SaveTempImage(req.Image)
	if err != nil {
		return nil, fmt.Errorf("图片保存失败: %w", err)
	}

	resolvedImagePath := uc.fileService.ResolvePath(tempImagePath)
	imgW, imgH, err := getImageDimensions(resolvedImagePath)
	if err != nil {
		_ = uc.fileService.DeleteFile(tempImagePath)
		return nil, fmt.Errorf("读取图片尺寸失败: %w", err)
	}

	// 3. 调用Python服务检测（返回多个缺陷）
	segmentModelType := strings.TrimSpace(req.ModelType)
	if segmentModelType == "" {
		segmentModelType = defaultSegmentModelType
	}

	pythonResult, err := uc.pythonService.DetectDefect(
		resolvedImagePath,
		req.ModelName,
		segmentModelType,
		req.PixelRatio,
	)
	if err != nil {
		// 回滚：删除临时图片
		uc.fileService.DeleteFile(tempImagePath)
		return nil, fmt.Errorf("AI检测失败: %w", err)
	}

	imageExt := strings.ToLower(filepath.Ext(req.Image.Filename))
	if imageExt == "" {
		imageExt = ".jpg"
	}
	imagePath, err := uc.fileService.AllocateImagePath("images", imageExt)
	if err != nil {
		_ = uc.fileService.DeleteFile(tempImagePath)
		return nil, fmt.Errorf("分配原图路径失败: %w", err)
	}

	var resultPath string
	if pythonResult.ResultImage != "" {
		resultPath, err = uc.fileService.AllocateImagePath("results", ".jpg")
		if err != nil {
			_ = uc.fileService.DeleteFile(tempImagePath)
			return nil, fmt.Errorf("分配结果图路径失败: %w", err)
		}
	}

	// 5. 在内存中构造缺陷结果，不阻塞数据库写入
	defects := make([]*model.Defect, 0, len(pythonResult.Defects))

	for _, detectedDefect := range pythonResult.Defects {
		defectType := translateAlgorithmDefectType(detectedDefect.DefectType)
		if hasYOLOCoords(detectedDefect.BBox.YOLOCoords) {
			x, y, w, h, length, width, area := service.CalculatePhysicalDimensions(
				detectedDefect.BBox.YOLOCoords,
				imgW,
				imgH,
				req.PixelRatio,
			)
			detectedDefect.BBox.X = x
			detectedDefect.BBox.Y = y
			detectedDefect.BBox.Width = w
			detectedDefect.BBox.Height = h
			detectedDefect.Length = length
			detectedDefect.Width = width
			detectedDefect.Area = area
		}

		defect := &model.Defect{
			BridgeID:   req.BridgeID,
			DefectType: defectType,
			ImagePath:  imagePath,  // 共享同一张原图
			ResultPath: resultPath, // 共享同一张结果图
			BBox:       detectedDefect.BBoxJSON(),
			Length:     detectedDefect.Length,
			Width:      detectedDefect.Width,
			Area:       detectedDefect.Area,
			Confidence: detectedDefect.Confidence,
			DetectedAt: time.Now(),
		}
		defects = append(defects, defect)
	}

	taskID := uc.newPersistenceTaskID()
	record := uc.registerPersistenceTask(taskID, req.BridgeID, imagePath, resultPath)
	task := detectionPersistenceTask{
		TaskID:            taskID,
		BridgeID:          req.BridgeID,
		TempImagePath:     tempImagePath,
		FinalImagePath:    imagePath,
		FinalResultPath:   resultPath,
		ResultImageBase64: pythonResult.ResultImage,
		Defects:           defects,
		CreatedAt:         time.Now(),
	}
	if err := uc.enqueuePersistenceTask(task); err != nil {
		record.Status = detectionPersistenceStatusFailed
		record.ErrorMessage = err.Error()
		record.UpdatedAt = time.Now()
		uc.persistenceMu.Lock()
		uc.persistenceRecords[taskID] = record
		uc.persistenceMu.Unlock()
		_ = uc.fileService.DeleteFile(tempImagePath)
		log.Printf("投递检测持久化任务失败: %v", err)
	} else {
		_ = record
	}

	// 6. 计算处理时间（仅同步检测阶段）
	processingTime := time.Since(startTime).Seconds()

	// 7. 返回多个缺陷结果
	response := &dto.DetectionResponse{
		TotalDefects:      len(defects),
		ImagePath:         imagePath,
		ResultPath:        resultPath,
		ResultImageBase64: pythonResult.ResultImage,
		ProcessingTime:    processingTime,
		PersistenceStatus: uc.currentPersistenceStatus(taskID),
		PersistenceTaskID: taskID,
		Defects:           uc.toDefectDTOs(defects),
		DefectSummary:     uc.buildDefectSummary(defects),
	}
	if response.PersistenceStatus == detectionPersistenceStatusFailed {
		if status, statusErr := uc.GetPersistenceTaskStatus(taskID, currentUser); statusErr == nil {
			response.PersistenceMessage = status.ErrorMessage
		}
	}

	return response, nil
}

// GetPersistenceTaskStatus 查询持久化任务状态。
func (uc *DetectionUseCase) GetPersistenceTaskStatus(taskID string, currentUser *model.User) (*dto.DetectionPersistenceStatusResponse, error) {
	record := uc.getPersistenceRecord(taskID)
	if record == nil {
		return nil, errors.New("持久化任务不存在")
	}

	bridge, err := uc.bridgeService.GetByID(record.BridgeID)
	if err != nil {
		return nil, err
	}
	if bridge == nil {
		return nil, errors.New("桥梁不存在")
	}
	if !currentUser.IsAdmin() && !bridge.IsOwnedBy(currentUser.ID) {
		return nil, errors.New("无权访问此桥梁")
	}

	return uc.buildPersistenceStatusResponse(record), nil
}

// WaitForPersistenceTask 等待指定持久化任务结束（仅供测试/调试使用）。
func (uc *DetectionUseCase) WaitForPersistenceTask(taskID string, timeout time.Duration) (*dto.DetectionPersistenceStatusResponse, error) {
	deadline := time.Now().Add(timeout)
	for {
		record := uc.getPersistenceRecord(taskID)
		if record == nil {
			return nil, errors.New("持久化任务不存在")
		}
		if record.Status == detectionPersistenceStatusCompleted || record.Status == detectionPersistenceStatusFailed {
			return uc.buildPersistenceStatusResponse(record), nil
		}
		if time.Now().After(deadline) {
			return nil, errors.New("等待持久化任务超时")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// Shutdown 停止后台持久化 worker。
func (uc *DetectionUseCase) Shutdown() {
	uc.closeOnce.Do(func() {
		close(uc.stopCh)
		uc.workerWG.Wait()
	})
}

// PersistenceWorkerCount 返回当前配置的持久化 worker 数量。
func (uc *DetectionUseCase) PersistenceWorkerCount() int {
	return uc.persistenceWorkers
}

// toDefectDTOs 转换为DTO列表
func (uc *DetectionUseCase) toDefectDTOs(defects []*model.Defect) []dto.DefectDTO {
	dtos := make([]dto.DefectDTO, 0, len(defects))
	for _, defect := range defects {
		dtos = append(dtos, dto.DefectDTO{
			ID:         defect.ID,
			BridgeID:   defect.BridgeID,
			DefectType: defect.DefectType,
			ImagePath:  defect.ImagePath,
			ResultPath: defect.ResultPath,
			BBox:       defect.BBox,
			Length:     defect.Length,
			Width:      defect.Width,
			Area:       defect.Area,
			Confidence: defect.Confidence,
			DetectedAt: defect.DetectedAt,
		})
	}
	return dtos
}

// buildDefectSummary 构建缺陷统计
func (uc *DetectionUseCase) buildDefectSummary(defects []*model.Defect) map[string]int {
	summary := make(map[string]int)
	for _, defect := range defects {
		summary[defect.DefectType]++
	}
	return summary
}

func hasYOLOCoords(coords [4]float64) bool {
	return coords != [4]float64{}
}

func translateAlgorithmDefectType(defectType string) string {
	switch defectType {
	case "Crack":
		return "裂缝"
	case "Breakage":
		return "破损"
	case "Comb":
		return "剥落"
	case "Hole":
		return "孔洞"
	case "Reinforcement":
		return "钢筋外露"
	case "Seepage":
		return "渗水"
	default:
		return defectType
	}
}
