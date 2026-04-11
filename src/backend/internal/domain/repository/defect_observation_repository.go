package repository

import "github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"

// DefectObservationRepository 帧级观测仓储接口。
type DefectObservationRepository interface {
	Create(observation *model.DefectObservation) error
}
