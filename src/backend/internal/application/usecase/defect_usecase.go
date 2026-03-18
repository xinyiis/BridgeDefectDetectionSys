// Package usecase 定义应用层用例
package usecase

import (
	"errors"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
)

// DefectUseCase 缺陷用例
// 处理缺陷查询、详情、删除等业务流程
type DefectUseCase struct {
	defectService *service.DefectService  // 缺陷领域服务
	fileService   service.FileService     // 文件服务
}

// NewDefectUseCase 创建缺陷用例实例
// 参数：
//   - defectService: 缺陷领域服务
//   - fileService: 文件服务
// 返回：
//   - *DefectUseCase: 缺陷用例实例
func NewDefectUseCase(defectService *service.DefectService, fileService service.FileService) *DefectUseCase {
	return &DefectUseCase{
		defectService: defectService,
		fileService:   fileService,
	}
}

// ListDefects 获取缺陷列表
// 参数：
//   - req: 列表查询请求
//   - currentUser: 当前用户
// 返回：
//   - *dto.DefectListResponse: 缺陷列表响应
//   - error: 操作错误
func (uc *DefectUseCase) ListDefects(req *dto.DefectListRequest, currentUser *model.User) (*dto.DefectListResponse, error) {
	// 1. 解析时间过滤条件
	var startTime, endTime *time.Time
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			startTime = &t
		}
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			// 结束时间设置为当天的23:59:59
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endTime = &t
		}
	}

	// 2. 构建过滤器
	filters := repository.DefectListFilters{
		Page:        req.Page,
		PageSize:    req.PageSize,
		BridgeID:    req.BridgeID,
		DefectType:  req.DefectType,
		StartTime:   startTime,
		EndTime:     endTime,
		CurrentUser: currentUser, // 关键：传递用户信息用于权限过滤
	}

	// 3. 查询列表（Service层会自动进行权限过滤）
	defects, total, err := uc.defectService.ListDefects(filters)
	if err != nil {
		return nil, err
	}

	// 4. 转换为DTO
	defectDTOs := make([]dto.DefectDTO, 0, len(defects))
	for _, defect := range defects {
		bridgeName := ""
		if defect.Bridge != nil {
			bridgeName = defect.Bridge.BridgeName
		}

		defectDTOs = append(defectDTOs, dto.DefectDTO{
			ID:         defect.ID,
			BridgeID:   defect.BridgeID,
			BridgeName: bridgeName,
			DefectType: defect.DefectType,
			ImagePath:  defect.ImagePath,
			ResultPath: defect.ResultPath,
			BBox:       defect.BBox,
			Length:     defect.Length,
			Width:      defect.Width,
			Area:       defect.Area,
			Confidence: defect.Confidence,
			DetectedAt: defect.DetectedAt,
			CreatedAt:  defect.CreatedAt,
		})
	}

	return &dto.DefectListResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     defectDTOs,
	}, nil
}

// GetDefect 获取缺陷详情
// 参数：
//   - id: 缺陷ID
// 返回：
//   - *dto.DefectDetailResponse: 缺陷详情响应
//   - error: 操作错误
func (uc *DefectUseCase) GetDefect(id uint) (*dto.DefectDetailResponse, error) {
	// 1. 查询缺陷
	defect, err := uc.defectService.GetDefect(id)
	if err != nil {
		return nil, err
	}
	if defect == nil {
		return nil, errors.New("缺陷不存在")
	}

	// 2. 构建响应
	response := &dto.DefectDetailResponse{
		DefectDTO: dto.DefectDTO{
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
			CreatedAt:  defect.CreatedAt,
		},
	}

	// 3. 添加桥梁信息
	if defect.Bridge != nil {
		response.Bridge = &dto.BridgeSimpleInfo{
			ID:         defect.Bridge.ID,
			BridgeName: defect.Bridge.BridgeName,
			BridgeCode: defect.Bridge.BridgeCode,
		}
	}

	return response, nil
}

// DeleteDefect 删除缺陷
// 参数：
//   - id: 缺陷ID
// 返回：
//   - error: 操作错误
func (uc *DefectUseCase) DeleteDefect(id uint) error {
	// 1. 查询缺陷信息（用于删除文件）
	defect, err := uc.defectService.GetDefect(id)
	if err != nil {
		return err
	}
	if defect == nil {
		return errors.New("缺陷不存在")
	}

	// 2. 软删除缺陷记录
	if err := uc.defectService.DeleteDefect(id); err != nil {
		return err
	}

	// 3. 异步删除文件（不阻塞响应）
	go func() {
		if defect.ImagePath != "" {
			uc.fileService.DeleteFile(defect.ImagePath)
		}
		if defect.ResultPath != "" {
			uc.fileService.DeleteFile(defect.ResultPath)
		}
	}()

	return nil
}
