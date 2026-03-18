// Package usecase 定义应用层用例
package usecase

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
)

// DetectionUseCase 检测用例
// 处理图片上传检测业务流程
type DetectionUseCase struct {
	defectService  *service.DefectService  // 缺陷领域服务
	bridgeService  *service.BridgeService  // 桥梁领域服务
	pythonService  service.PythonService   // Python检测服务
	fileService    service.FileService     // 文件服务
}

// NewDetectionUseCase 创建检测用例实例
// 参数：
//   - defectService: 缺陷领域服务
//   - bridgeService: 桥梁领域服务
//   - pythonService: Python检测服务
//   - fileService: 文件服务
// 返回：
//   - *DetectionUseCase: 检测用例实例
func NewDetectionUseCase(
	defectService *service.DefectService,
	bridgeService *service.BridgeService,
	pythonService service.PythonService,
	fileService service.FileService,
) *DetectionUseCase {
	return &DetectionUseCase{
		defectService:  defectService,
		bridgeService:  bridgeService,
		pythonService:  pythonService,
		fileService:    fileService,
	}
}

// UploadAndDetect 上传图片并进行缺陷检测
// 参数：
//   - req: 检测上传请求
//   - currentUser: 当前用户
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

	// 2. 保存上传的图片
	imagePath, err := uc.fileService.SaveImage(req.Image, "images")
	if err != nil {
		return nil, fmt.Errorf("图片保存失败: %w", err)
	}

	// 3. 调用Python服务检测（返回多个缺陷）
	pythonResult, err := uc.pythonService.DetectDefect(imagePath, req.ModelName, req.PixelRatio)
	if err != nil {
		// 回滚：删除已上传的图片
		uc.fileService.DeleteFile(imagePath)
		return nil, fmt.Errorf("AI检测失败: %w", err)
	}

	// 4. 保存结果图（如果有）- 只有一张，包含所有缺陷标注
	var resultPath string
	if pythonResult.ResultImage != "" {
		resultPath, err = uc.fileService.SaveResultImage(pythonResult.ResultImage, "results")
		if err != nil {
			// 回滚：删除原图
			uc.fileService.DeleteFile(imagePath)
			return nil, fmt.Errorf("结果图保存失败: %w", err)
		}
	}

	// 5. 为每个检测到的缺陷创建数据库记录
	defects := make([]*model.Defect, 0, len(pythonResult.Defects))

	for _, detectedDefect := range pythonResult.Defects {
		defect := &model.Defect{
			BridgeID:   req.BridgeID,
			DefectType: detectedDefect.DefectType,
			ImagePath:  imagePath,      // 共享同一张原图
			ResultPath: resultPath,     // 共享同一张结果图
			BBox:       detectedDefect.BBoxJSON(),
			Length:     detectedDefect.Length,
			Width:      detectedDefect.Width,
			Area:       detectedDefect.Area,
			Confidence: detectedDefect.Confidence,
			DetectedAt: time.Now(),
		}

		if err := uc.defectService.CreateDefect(defect); err != nil {
			log.Printf("保存缺陷记录失败: %v", err)
			continue // 继续处理其他缺陷
		}

		defects = append(defects, defect)
	}

	// 6. 如果一个缺陷都没保存成功，删除文件
	if len(defects) == 0 && len(pythonResult.Defects) > 0 {
		uc.fileService.DeleteFile(imagePath)
		if resultPath != "" {
			uc.fileService.DeleteFile(resultPath)
		}
		return nil, errors.New("保存缺陷记录失败")
	}

	// 7. 计算处理时间
	processingTime := time.Since(startTime).Seconds()

	// 8. 返回多个缺陷结果
	return &dto.DetectionResponse{
		TotalDefects:   len(defects),
		ImagePath:      imagePath,
		ResultPath:     resultPath,
		ProcessingTime: processingTime,
		Defects:        uc.toDefectDTOs(defects),
		DefectSummary:  uc.buildDefectSummary(defects),
	}, nil
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
