package persistence

import (
	"errors"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"gorm.io/gorm"
)

type videoTaskRepositoryImpl struct {
	db *gorm.DB
}

// NewVideoTaskRepository 创建视频任务仓储。
func NewVideoTaskRepository(db *gorm.DB) repository.VideoTaskRepository {
	return &videoTaskRepositoryImpl{db: db}
}

func (r *videoTaskRepositoryImpl) Create(task *model.VideoAnalysisTask) error {
	return r.db.Create(task).Error
}

func (r *videoTaskRepositoryImpl) FindByTaskID(taskID string) (*model.VideoAnalysisTask, error) {
	var task model.VideoAnalysisTask
	if err := r.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

func (r *videoTaskRepositoryImpl) Update(task *model.VideoAnalysisTask) error {
	return r.db.Save(task).Error
}
