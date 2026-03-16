// Package service 定义领域服务层
package service

import (
	"errors"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"gorm.io/gorm"
)

// DefectService 缺陷领域服务
// 处理缺陷相关的核心业务逻辑
type DefectService struct {
	db         *gorm.DB                   // GORM数据库连接
	defectRepo repository.DefectRepository // 缺陷仓储
	bridgeRepo repository.BridgeRepository // 桥梁仓储（用于验证桥梁存在性）
}

// NewDefectService 创建缺陷服务实例
// 参数：
//   - db: GORM数据库连接
//   - defectRepo: 缺陷Repository接口
//   - bridgeRepo: 桥梁Repository接口
// 返回：
//   - *DefectService: 缺陷服务实例
func NewDefectService(db *gorm.DB, defectRepo repository.DefectRepository, bridgeRepo repository.BridgeRepository) *DefectService {
	return &DefectService{
		db:         db,
		defectRepo: defectRepo,
		bridgeRepo: bridgeRepo,
	}
}

// CreateDefect 创建缺陷记录
// 参数：
//   - defect: 缺陷实体
// 返回：
//   - error: 操作错误
func (s *DefectService) CreateDefect(defect *model.Defect) error {
	// 1. 验证桥梁存在性
	bridge, err := s.bridgeRepo.FindByID(defect.BridgeID)
	if err != nil {
		return err
	}
	if bridge == nil {
		return errors.New("桥梁不存在")
	}

	// 2. 创建缺陷记录
	return s.defectRepo.Create(defect)
}

// GetDefect 根据ID获取缺陷
// 参数：
//   - id: 缺陷ID
// 返回：
//   - *model.Defect: 缺陷实体（包含关联的Bridge信息）
//   - error: 操作错误
func (s *DefectService) GetDefect(id uint) (*model.Defect, error) {
	return s.defectRepo.FindByID(id)
}

// ListDefects 分页获取缺陷列表（带权限过滤）
// 参数：
//   - filters: 过滤条件（包含用户信息）
// 返回：
//   - []model.Defect: 缺陷列表
//   - int64: 总数量
//   - error: 操作错误
func (s *DefectService) ListDefects(filters repository.DefectListFilters) ([]model.Defect, int64, error) {
	// 参数验证
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 10
	}

	// Repository层会自动进行权限过滤（JOIN bridges表）
	return s.defectRepo.List(filters)
}

// DeleteDefect 删除缺陷（软删除）
// 参数：
//   - id: 缺陷ID
// 返回：
//   - error: 操作错误
func (s *DefectService) DeleteDefect(id uint) error {
	return s.defectRepo.Delete(id)
}

// VerifyDefectOwnership 验证缺陷所有权（用于中间件）
// 参数：
//   - defectID: 缺陷ID
//   - userID: 用户ID
//   - isAdmin: 是否为管理员
// 返回：
//   - *model.Defect: 缺陷实体（验证通过）
//   - error: 权限错误
func (s *DefectService) VerifyDefectOwnership(defectID uint, userID uint, isAdmin bool) (*model.Defect, error) {
	// 1. 查询缺陷
	defect, err := s.defectRepo.FindByID(defectID)
	if err != nil {
		return nil, err
	}
	if defect == nil {
		return nil, errors.New("缺陷不存在")
	}

	// 2. 管理员直接放行
	if isAdmin {
		return defect, nil
	}

	// 3. 查询关联的桥梁
	bridge, err := s.bridgeRepo.FindByID(defect.BridgeID)
	if err != nil {
		return nil, err
	}
	if bridge == nil {
		return nil, errors.New("关联桥梁不存在")
	}

	// 4. 验证所有权
	if bridge.UserID != userID {
		return nil, errors.New("无权访问此缺陷")
	}

	return defect, nil
}
