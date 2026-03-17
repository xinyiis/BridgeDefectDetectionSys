// Package service 定义领域服务层
// 包含核心业务逻辑和领域规则
package service

import (
	"errors"
	"fmt"
	"log"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"gorm.io/gorm"
)

// BridgeService 桥梁领域服务
// 处理桥梁相关的核心业务逻辑
type BridgeService struct {
	db          *gorm.DB                     // GORM数据库连接
	bridgeRepo  repository.BridgeRepository  // 桥梁仓储
	fileService FileService                  // 文件服务
}

// NewBridgeService 创建桥梁服务实例
// 参数：
//   - db: GORM数据库连接
//   - bridgeRepo: 桥梁Repository接口
//   - fileService: 文件服务接口
// 返回：
//   - *BridgeService: 桥梁服务实例
func NewBridgeService(db *gorm.DB, bridgeRepo repository.BridgeRepository, fileService FileService) *BridgeService {
	return &BridgeService{
		db:          db,
		bridgeRepo:  bridgeRepo,
		fileService: fileService,
	}
}

// CreateBridge 创建桥梁
// 参数：
//   - bridge: 桥梁实体
// 返回：
//   - error: 操作错误
func (s *BridgeService) CreateBridge(bridge *model.Bridge) error {
	// 1. 检查桥梁编号是否已存在
	existingBridge, err := s.bridgeRepo.FindByCode(bridge.BridgeCode)
	if err != nil {
		return err
	}
	if existingBridge != nil {
		return errors.New("桥梁编号已存在")
	}

	// 2. 创建桥梁
	return s.bridgeRepo.Create(bridge)
}

// GetByID 根据ID获取桥梁
// 参数：
//   - id: 桥梁ID
// 返回：
//   - *model.Bridge: 桥梁实体
//   - error: 操作错误
func (s *BridgeService) GetByID(id uint) (*model.Bridge, error) {
	return s.bridgeRepo.FindByID(id)
}

// ListBridges 分页获取桥梁列表（带权限过滤）
// 参数：
//   - currentUser: 当前用户
//   - page: 页码
//   - pageSize: 每页数量
//   - status: 状态过滤（可选）
// 返回：
//   - []model.Bridge: 桥梁列表
//   - int64: 总数量
//   - error: 操作错误
func (s *BridgeService) ListBridges(currentUser *model.User, page, pageSize int, status string) ([]model.Bridge, int64, error) {
	var bridges []model.Bridge
	var total int64

	// 参数验证
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := s.db.Model(&model.Bridge{})

	// 🔑 关键：根据用户角色自动过滤
	if !currentUser.IsAdmin() {
		// 普通用户：只查询自己的桥梁
		query = query.Where("user_id = ?", currentUser.ID)
	}
	// 管理员：不添加user_id条件，查询所有桥梁

	// 状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&bridges).Error; err != nil {
		return nil, 0, err
	}

	return bridges, total, nil
}

// UpdateBridge 更新桥梁信息
// 参数：
//   - bridge: 桥梁实体（包含更新后的数据）
// 返回：
//   - error: 操作错误
func (s *BridgeService) UpdateBridge(bridge *model.Bridge) error {
	return s.bridgeRepo.Update(bridge)
}

// DeleteBridge 删除桥梁（软删除 + 级联删除）
// 参数：
//   - bridgeID: 桥梁ID
//   - currentUser: 当前用户
// 返回：
//   - error: 操作错误
func (s *BridgeService) DeleteBridge(bridgeID uint, currentUser *model.User) error {
	// 1. 查询桥梁（验证存在性）
	bridge, err := s.bridgeRepo.FindByID(bridgeID)
	if err != nil {
		return err
	}
	if bridge == nil {
		return errors.New("桥梁不存在")
	}

	// 2. 权限验证
	if !currentUser.IsAdmin() && !bridge.IsOwnedBy(currentUser.ID) {
		return errors.New("权限不足")
	}

	// 3. 开启事务（保证数据一致性）
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 3.1 修改桥梁编号，释放唯一索引
		timestamp := bridge.CreatedAt.Unix()
		bridge.BridgeCode = fmt.Sprintf("%s_deleted_%d", bridge.BridgeCode, timestamp)
		if err := tx.Save(bridge).Error; err != nil {
			return fmt.Errorf("修改桥梁编号失败: %w", err)
		}

		// 3.2 软删除桥梁
		if err := tx.Delete(bridge).Error; err != nil {
			return fmt.Errorf("删除桥梁失败: %w", err)
		}

		// 3.3 级联软删除所有关联的缺陷记录
		if err := tx.Where("bridge_id = ?", bridgeID).Delete(&model.Defect{}).Error; err != nil {
			return fmt.Errorf("删除缺陷记录失败: %w", err)
		}

		// 3.4 删除3D模型文件（同步删除，不影响事务）
		if bridge.Model3DPath != "" {
			if err := s.fileService.DeleteFile(bridge.Model3DPath); err != nil {
				// 🔑 关键：文件删除失败只记录日志，不回滚事务
				log.Printf("警告：删除3D模型文件失败: %s, error: %v", bridge.Model3DPath, err)
			}
		}

		return nil
	})
}
