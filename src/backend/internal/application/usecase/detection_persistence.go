package usecase

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
)

const (
	detectionPersistenceStatusPending   = "pending"
	detectionPersistenceStatusCompleted = "completed"
	detectionPersistenceStatusFailed    = "failed"
)

type detectionPersistenceTask struct {
	TaskID            string
	BridgeID          uint
	TempImagePath     string
	FinalImagePath    string
	FinalResultPath   string
	ResultImageBase64 string
	Defects           []*model.Defect
	CreatedAt         time.Time
}

type detectionPersistenceRecord struct {
	TaskID         string
	BridgeID       uint
	Status         string
	ImagePath      string
	ResultPath     string
	SavedDefectIDs []uint
	ErrorMessage   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (uc *DetectionUseCase) startPersistenceWorkers(count int) {
	if count <= 0 {
		count = 1
	}
	for i := 0; i < count; i++ {
		uc.workerWG.Add(1)
		go func() {
			defer uc.workerWG.Done()
			for {
				select {
				case <-uc.stopCh:
					return
				case task := <-uc.persistenceQueue:
					uc.processPersistenceTask(task)
				}
			}
		}()
	}
}

func (uc *DetectionUseCase) processPersistenceTask(task detectionPersistenceTask) {
	record := uc.getPersistenceRecord(task.TaskID)
	if record == nil {
		record = uc.registerPersistenceTask(task.TaskID, task.BridgeID, task.FinalImagePath, task.FinalResultPath)
	}

	if err := uc.fileService.PersistTempFile(task.TempImagePath, task.FinalImagePath); err != nil {
		uc.failPersistenceTask(record, fmt.Errorf("持久化原图失败: %w", err))
		_ = uc.fileService.DeleteFile(task.TempImagePath)
		return
	}

	if task.FinalResultPath != "" && task.ResultImageBase64 != "" {
		if err := uc.fileService.PersistResultImage(task.ResultImageBase64, task.FinalResultPath); err != nil {
			uc.failPersistenceTask(record, fmt.Errorf("持久化结果图失败: %w", err))
			_ = uc.fileService.DeleteFile(task.TempImagePath)
			return
		}
	}

	if err := uc.defectService.CreateDefectsBatch(task.Defects); err != nil {
		uc.failPersistenceTask(record, fmt.Errorf("批量保存缺陷失败: %w", err))
		_ = uc.fileService.DeleteFile(task.TempImagePath)
		return
	}

	savedIDs := make([]uint, 0, len(task.Defects))
	for _, defect := range task.Defects {
		if defect != nil && defect.ID != 0 {
			savedIDs = append(savedIDs, defect.ID)
		}
	}

	uc.persistenceMu.Lock()
	record.Status = detectionPersistenceStatusCompleted
	record.SavedDefectIDs = savedIDs
	record.ErrorMessage = ""
	record.UpdatedAt = time.Now()
	uc.persistenceRecords[record.TaskID] = record
	uc.persistenceMu.Unlock()

	if err := uc.fileService.DeleteFile(task.TempImagePath); err != nil {
		log.Printf("删除检测临时图片失败: task=%s err=%v", task.TaskID, err)
	}
}

func (uc *DetectionUseCase) failPersistenceTask(record *detectionPersistenceRecord, err error) {
	log.Printf("检测持久化失败: task=%s err=%v", record.TaskID, err)
	uc.persistenceMu.Lock()
	record.Status = detectionPersistenceStatusFailed
	record.ErrorMessage = err.Error()
	record.UpdatedAt = time.Now()
	uc.persistenceRecords[record.TaskID] = record
	uc.persistenceMu.Unlock()
}

func (uc *DetectionUseCase) enqueuePersistenceTask(task detectionPersistenceTask) error {
	select {
	case <-uc.stopCh:
		return fmt.Errorf("检测持久化服务已停止")
	case uc.persistenceQueue <- task:
		return nil
	default:
		return fmt.Errorf("检测持久化队列已满")
	}
}

func (uc *DetectionUseCase) newPersistenceTaskID() string {
	return uuid.NewString()
}

func (uc *DetectionUseCase) registerPersistenceTask(taskID string, bridgeID uint, imagePath, resultPath string) *detectionPersistenceRecord {
	record := &detectionPersistenceRecord{
		TaskID:     taskID,
		BridgeID:   bridgeID,
		Status:     detectionPersistenceStatusPending,
		ImagePath:  imagePath,
		ResultPath: resultPath,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	uc.persistenceMu.Lock()
	uc.persistenceRecords[taskID] = record
	uc.persistenceMu.Unlock()
	return record
}

func (uc *DetectionUseCase) getPersistenceRecord(taskID string) *detectionPersistenceRecord {
	uc.persistenceMu.RLock()
	defer uc.persistenceMu.RUnlock()
	record, ok := uc.persistenceRecords[taskID]
	if !ok {
		return nil
	}
	clone := *record
	clone.SavedDefectIDs = append([]uint(nil), record.SavedDefectIDs...)
	return &clone
}

func (uc *DetectionUseCase) currentPersistenceStatus(taskID string) string {
	record := uc.getPersistenceRecord(taskID)
	if record == nil {
		return detectionPersistenceStatusFailed
	}
	return record.Status
}

func (uc *DetectionUseCase) buildPersistenceStatusResponse(record *detectionPersistenceRecord) *dto.DetectionPersistenceStatusResponse {
	return &dto.DetectionPersistenceStatusResponse{
		TaskID:         record.TaskID,
		Status:         record.Status,
		ImagePath:      record.ImagePath,
		ResultPath:     record.ResultPath,
		SavedDefectIDs: append([]uint(nil), record.SavedDefectIDs...),
		ErrorMessage:   record.ErrorMessage,
	}
}
