package persistence

import (
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"gorm.io/gorm"
)

type defectObservationRepositoryImpl struct {
	db *gorm.DB
}

// NewDefectObservationRepository 创建帧级观测仓储。
func NewDefectObservationRepository(db *gorm.DB) repository.DefectObservationRepository {
	return &defectObservationRepositoryImpl{db: db}
}

func (r *defectObservationRepositoryImpl) Create(observation *model.DefectObservation) error {
	return r.db.Create(observation).Error
}
