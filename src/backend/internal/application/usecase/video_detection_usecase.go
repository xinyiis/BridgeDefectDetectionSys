package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	videopkg "github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/video"
)

const uploadsRoot = "./uploads"

// VideoDetectionUseCase 视频分析用例。
type VideoDetectionUseCase struct {
	defectService   *service.DefectService
	bridgeService   *service.BridgeService
	pythonService   service.PythonService
	fileService     service.FileService
	frameExtractor  videopkg.FrameExtractor
	taskRepo        repository.VideoTaskRepository
	observationRepo repository.DefectObservationRepository
	tasks           sync.Map
	tracks          sync.Map
	cancelFns       sync.Map
	running         sync.Map
}

// NewVideoDetectionUseCase 创建视频分析用例。
func NewVideoDetectionUseCase(
	defectService *service.DefectService,
	bridgeService *service.BridgeService,
	pythonService service.PythonService,
	fileService service.FileService,
	frameExtractor videopkg.FrameExtractor,
	taskRepo repository.VideoTaskRepository,
	observationRepo repository.DefectObservationRepository,
) *VideoDetectionUseCase {
	return &VideoDetectionUseCase{
		defectService:   defectService,
		bridgeService:   bridgeService,
		pythonService:   pythonService,
		fileService:     fileService,
		frameExtractor:  frameExtractor,
		taskRepo:        taskRepo,
		observationRepo: observationRepo,
	}
}

// UploadVideo 仅上传视频并创建任务。
func (uc *VideoDetectionUseCase) UploadVideo(req *dto.VideoUploadRequest, currentUser *model.User) (*dto.VideoUploadResponse, error) {
	bridge, err := uc.bridgeService.GetByID(req.BridgeID)
	if err != nil {
		return nil, err
	}
	if bridge == nil {
		return nil, errors.New("桥梁不存在")
	}
	if !currentUser.IsAdmin() && bridge.UserID != currentUser.ID {
		return nil, errors.New("无权访问此桥梁")
	}

	videoPath, err := uc.fileService.SaveUploadedFile(req.Video, "videos")
	if err != nil {
		return nil, fmt.Errorf("视频保存失败: %w", err)
	}

	modelName := req.ModelName
	if modelName == "" {
		modelName = "baseline"
	}

	fps := req.FPS
	if fps <= 0 {
		fps = 1
	}

	task := &model.VideoAnalysisTask{
		TaskID:        "video_" + uuid.NewString(),
		UserID:        currentUser.ID,
		BridgeID:      req.BridgeID,
		VideoPath:     videoPath,
		FPS:           fps,
		ModelName:     modelName,
		PixelRatio:    req.PixelRatio,
		EnableSegment: req.EnableSegment,
		Status:        model.VideoTaskUploaded,
	}

	if err := uc.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("创建视频任务失败: %w", err)
	}

	uc.tasks.Store(task.TaskID, task)

	return &dto.VideoUploadResponse{
		TaskID:    task.TaskID,
		Status:    task.Status,
		VideoPath: task.VideoPath,
	}, nil
}

// StartTask 创建分析会话。
func (uc *VideoDetectionUseCase) StartTask(taskID string, currentUser *model.User) (*dto.VideoStartResponse, error) {
	task, err := uc.loadOwnedTask(taskID, currentUser)
	if err != nil {
		return nil, err
	}

	if task.Status != model.VideoTaskUploaded {
		return nil, errors.New("当前任务状态不允许开始分析")
	}

	task.SessionID = "session_" + uuid.NewString()
	task.Status = model.VideoTaskReadyToProcess
	task.ErrorMessage = ""
	task.CanceledAt = nil
	task.CompletedAt = nil
	task.ProcessedFrames = 0
	task.ConfirmedDefects = 0

	if err := uc.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("更新任务状态失败: %w", err)
	}

	uc.tasks.Store(task.TaskID, task)

	return &dto.VideoStartResponse{
		TaskID:    task.TaskID,
		SessionID: task.SessionID,
		Status:    task.Status,
		WSURL:     fmt.Sprintf("/api/v1/detection/video/ws?task_id=%s&session_id=%s", task.TaskID, task.SessionID),
	}, nil
}

// GetTask 获取任务状态。
func (uc *VideoDetectionUseCase) GetTask(taskID string, currentUser *model.User) (*dto.VideoTaskResponse, error) {
	task, err := uc.loadOwnedTask(taskID, currentUser)
	if err != nil {
		return nil, err
	}
	return uc.toTaskResponse(task), nil
}

// CancelTask 取消任务。
func (uc *VideoDetectionUseCase) CancelTask(taskID string, currentUser *model.User) (*dto.VideoTaskResponse, error) {
	task, err := uc.loadOwnedTask(taskID, currentUser)
	if err != nil {
		return nil, err
	}

	switch task.Status {
	case model.VideoTaskUploaded, model.VideoTaskReadyToProcess, model.VideoTaskProcessing:
	default:
		return nil, errors.New("当前任务状态不允许取消")
	}

	if cancelFn, ok := uc.cancelFns.Load(task.TaskID); ok {
		if fn, ok := cancelFn.(context.CancelFunc); ok {
			fn()
		}
	}

	now := time.Now()
	task.Status = model.VideoTaskCanceled
	task.CanceledAt = &now
	task.ErrorMessage = ""

	if err := uc.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("更新取消状态失败: %w", err)
	}

	uc.tasks.Store(task.TaskID, task)

	return uc.toTaskResponse(task), nil
}

// StreamTask 在 WebSocket 建连后开始处理任务。
func (uc *VideoDetectionUseCase) StreamTask(taskID, sessionID string, currentUser *model.User, send func(any) error) error {
	task, err := uc.loadOwnedTask(taskID, currentUser)
	if err != nil {
		return err
	}
	if task.SessionID == "" || task.SessionID != sessionID {
		return errors.New("分析会话不存在")
	}

	uc.sendMessage(send, dto.VideoConnectedMessage{
		Type:      "connected",
		TaskID:    task.TaskID,
		SessionID: task.SessionID,
		Message:   "websocket connected",
	})

	switch task.Status {
	case model.VideoTaskReadyToProcess:
		if _, loaded := uc.running.LoadOrStore(task.TaskID, struct{}{}); loaded {
			return nil
		}
		defer uc.running.Delete(task.TaskID)
		return uc.runTask(task, send)
	case model.VideoTaskProcessing:
		return nil
	case model.VideoTaskCompleted:
		uc.sendMessage(send, dto.VideoCompletedMessage{
			Type:            "completed",
			TaskID:          task.TaskID,
			SessionID:       task.SessionID,
			ProcessedFrames: task.ProcessedFrames,
			TotalFrames:     task.TotalFrames,
			TotalDefects:    task.ConfirmedDefects,
			DefectSummary:   map[string]int{},
		})
		return nil
	case model.VideoTaskCanceled:
		return errors.New("任务已取消")
	default:
		return errors.New("当前任务状态不允许建立分析流")
	}
}

func (uc *VideoDetectionUseCase) runTask(task *model.VideoAnalysisTask, send func(any) error) error {
	ctx, cancel := context.WithCancel(context.Background())
	uc.cancelFns.Store(task.TaskID, cancel)
	defer uc.cancelFns.Delete(task.TaskID)
	defer uc.tracks.Delete(task.TaskID)

	now := time.Now()
	task.Status = model.VideoTaskProcessing
	task.StartedAt = &now
	task.ErrorMessage = ""

	if totalFrames, err := uc.frameExtractor.CountFrames(uc.resolveStoredPath(task.VideoPath), task.FPS); err == nil && totalFrames > 0 {
		task.TotalFrames = totalFrames
	}
	if err := uc.taskRepo.Update(task); err != nil {
		return fmt.Errorf("更新处理状态失败: %w", err)
	}
	uc.tasks.Store(task.TaskID, task)

	summary := make(map[string]int)
	tempDir := filepath.Join(os.TempDir(), "bridge_detect_video_frames", task.TaskID)
	defer os.RemoveAll(tempDir)

	err := uc.frameExtractor.ExtractFramesStream(
		uc.resolveStoredPath(task.VideoPath),
		tempDir,
		task.FPS,
		func(framePath string, frameNo int) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return uc.processFrame(task, framePath, frameNo, send, summary)
		},
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			if task.Status != model.VideoTaskCanceled {
				canceledAt := time.Now()
				task.Status = model.VideoTaskCanceled
				task.CanceledAt = &canceledAt
				task.ErrorMessage = ""
				_ = uc.taskRepo.Update(task)
				uc.tasks.Store(task.TaskID, task)
			}
			return nil
		}

		task.Status = model.VideoTaskFailed
		task.ErrorMessage = err.Error()
		failedAt := time.Now()
		task.CompletedAt = &failedAt
		_ = uc.taskRepo.Update(task)
		uc.tasks.Store(task.TaskID, task)
		uc.sendMessage(send, dto.VideoErrorMessage{
			Type:      "error",
			TaskID:    task.TaskID,
			SessionID: task.SessionID,
			Message:   err.Error(),
		})
		return err
	}

	if task.TotalFrames == 0 {
		task.TotalFrames = task.ProcessedFrames
	}
	completedAt := time.Now()
	task.Status = model.VideoTaskCompleted
	task.CompletedAt = &completedAt
	task.ErrorMessage = ""
	if err := uc.taskRepo.Update(task); err != nil {
		return fmt.Errorf("更新完成状态失败: %w", err)
	}
	uc.tasks.Store(task.TaskID, task)

	uc.sendMessage(send, dto.VideoCompletedMessage{
		Type:            "completed",
		TaskID:          task.TaskID,
		SessionID:       task.SessionID,
		ProcessedFrames: task.ProcessedFrames,
		TotalFrames:     task.TotalFrames,
		TotalDefects:    task.ConfirmedDefects,
		DefectSummary:   summary,
	})

	return nil
}

func (uc *VideoDetectionUseCase) processFrame(
	task *model.VideoAnalysisTask,
	framePath string,
	frameNo int,
	send func(any) error,
	summary map[string]int,
) error {
	detectResult, err := uc.pythonService.Detect(framePath, &service.DetectRequest{
		ModelName: task.ModelName,
		Conf:      0.25,
	})
	if err != nil {
		return fmt.Errorf("AI检测失败: %w", err)
	}

	outputImageBase64 := detectResult.ImageResult
	if task.EnableSegment && len(detectResult.YOLOBBoxes) > 0 {
		bboxesJSON, err := json.Marshal(detectResult.YOLOBBoxes)
		if err == nil {
			segmentResult, segErr := uc.pythonService.Segment(framePath, &service.SegmentRequest{
				BBoxesJSON: string(bboxesJSON),
				Alpha:      0.5,
			})
			if segErr == nil && segmentResult.FusionImage != "" {
				outputImageBase64 = segmentResult.FusionImage
			}
		}
	}

	imgW, imgH, err := getImageDimensions(framePath)
	if err != nil {
		return fmt.Errorf("读取帧尺寸失败: %w", err)
	}

	tracks := uc.loadTracks(task.TaskID)
	matchedTrackIDs := make(map[string]struct{})
	defects := make([]dto.VideoFrameDefect, 0, len(detectResult.YOLOBBoxes))
	timestampMS := frameTimestampMS(task.FPS, frameNo)
	observedAt := time.Now()
	persistedResultPath := ""

	for _, bbox := range detectResult.YOLOBBoxes {
		defectType := translateDefectType(bbox)
		x, y, w, h, length, width, area := service.CalculatePhysicalDimensions(
			bbox.YOLOCoords,
			imgW,
			imgH,
			task.PixelRatio,
		)

		currentDefect := dto.VideoFrameDefect{
			DefectType: defectType,
			BBox: dto.VideoBBox{
				X:      x,
				Y:      y,
				Width:  w,
				Height: h,
			},
			Length:            length,
			Width:             width,
			Area:              area,
			MeasurementSource: "bbox_estimate",
			MeasurementStatus: "estimated",
		}

		track, becameBest := uc.upsertTrack(task, tracks, currentDefect, bbox.Confidence, frameNo, observedAt)
		matchedTrackIDs[track.TrackID] = struct{}{}
		if !containsTrack(tracks, track.TrackID) {
			tracks = append(tracks, track)
		}

		resultPathForObservation := ""
		if becameBest {
			bestFramePath, err := uc.persistFrameEvidence(framePath)
			if err == nil {
				if track.BestFramePath != "" && track.BestFramePath != bestFramePath {
					_ = uc.fileService.DeleteFile(track.BestFramePath)
				}
				track.BestFramePath = bestFramePath
			}
			if outputImageBase64 != "" {
				bestResultPath, resultErr := uc.fileService.SaveResultImage(outputImageBase64, "video_results")
				if resultErr == nil {
					if track.BestResultPath != "" && track.BestResultPath != bestResultPath {
						_ = uc.fileService.DeleteFile(track.BestResultPath)
					}
					track.BestResultPath = bestResultPath
					resultPathForObservation = bestResultPath
					if persistedResultPath == "" {
						persistedResultPath = bestResultPath
					}
				}
			}
		}

		if err := uc.observationRepo.Create(&model.DefectObservation{
			TaskID:      task.TaskID,
			BridgeID:    task.BridgeID,
			FrameNo:     frameNo,
			TimestampMS: timestampMS,
			DefectType:  currentDefect.DefectType,
			BBox:        currentDefect.BBox.JSON(),
			Confidence:  bbox.Confidence,
			FramePath:   framePath,
			ResultPath:  resultPathForObservation,
			TrackID:     track.TrackID,
		}); err != nil {
			return fmt.Errorf("保存帧级观测失败: %w", err)
		}

		if uc.shouldConfirmTrack(track) && track.DefectID == nil {
			defect := uc.buildConfirmedDefect(task, track, observedAt)
			created, err := uc.defectService.ConfirmVideoDefect(defect)
			if err != nil {
				return fmt.Errorf("确认视频缺陷失败: %w", err)
			}

			track.DefectID = &created.ID
			track.Status = trackConfirmed
			task.ConfirmedDefects++
			summary[track.DefectType]++
		}

		if track.DefectID != nil {
			if err := uc.defectService.UpdateVideoDefectEvidence(*track.DefectID, &service.DefectEvidenceUpdate{
				LastSeenAt:        track.LastSeenAt,
				ObservationCount:  track.Hits,
				BestFramePath:     track.BestFramePath,
				BestResultPath:    track.BestResultPath,
				BestConfidence:    track.BestConfidence,
				FinalBBox:         track.BestBBox.JSON(),
				FinalLength:       track.FinalLength,
				FinalWidth:        track.FinalWidth,
				FinalArea:         track.FinalArea,
				MeasurementSource: track.MeasurementSource,
				MeasurementStatus: track.MeasurementStatus,
			}); err != nil {
				return fmt.Errorf("更新历史缺陷证据失败: %w", err)
			}
		}

		defects = append(defects, currentDefect)
	}

	uc.closeMissedTracks(tracks, matchedTrackIDs)
	uc.storeTracks(task.TaskID, tracks)

	task.ProcessedFrames = frameNo
	if task.TotalFrames < frameNo {
		task.TotalFrames = frameNo
	}
	if err := uc.taskRepo.Update(task); err != nil {
		return fmt.Errorf("更新任务进度失败: %w", err)
	}
	uc.tasks.Store(task.TaskID, task)

	progress := currentProgress(task)
	uc.sendMessage(send, dto.VideoFrameMessage{
		Type:                "frame",
		TaskID:              task.TaskID,
		SessionID:           task.SessionID,
		FrameNo:             frameNo,
		ProcessedFrames:     task.ProcessedFrames,
		TotalFrames:         task.TotalFrames,
		Progress:            progress,
		TimestampMS:         timestampMS,
		ImageBase64:         outputImageBase64,
		PersistedResultPath: persistedResultPath,
		Defects:             defects,
		DefectCount:         len(defects),
	})
	uc.sendMessage(send, dto.VideoProgressMessage{
		Type:            "progress",
		TaskID:          task.TaskID,
		SessionID:       task.SessionID,
		ProcessedFrames: task.ProcessedFrames,
		TotalFrames:     task.TotalFrames,
		Progress:        progress,
		TotalDefects:    task.ConfirmedDefects,
	})

	return nil
}

func (uc *VideoDetectionUseCase) buildConfirmedDefect(task *model.VideoAnalysisTask, track *activeDefectTrack, observedAt time.Time) *model.Defect {
	return &model.Defect{
		BridgeID:          task.BridgeID,
		DefectType:        track.DefectType,
		ImagePath:         track.BestFramePath,
		ResultPath:        track.BestResultPath,
		BBox:              track.BestBBox.JSON(),
		Length:            track.FinalLength,
		Width:             track.FinalWidth,
		Area:              track.FinalArea,
		Confidence:        track.BestConfidence,
		SourceType:        "video",
		SourceTaskID:      &task.TaskID,
		TrackID:           &track.TrackID,
		FirstSeenAt:       &track.FirstSeenAt,
		LastSeenAt:        &track.LastSeenAt,
		ObservationCount:  track.Hits,
		BestFramePath:     track.BestFramePath,
		BestResultPath:    track.BestResultPath,
		MeasurementSource: track.MeasurementSource,
		MeasurementStatus: track.MeasurementStatus,
		DetectedAt:        observedAt,
	}
}

func (uc *VideoDetectionUseCase) loadOwnedTask(taskID string, currentUser *model.User) (*model.VideoAnalysisTask, error) {
	task, err := uc.taskRepo.FindByTaskID(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("任务不存在")
	}
	if !currentUser.IsAdmin() && task.UserID != currentUser.ID {
		return nil, errors.New("无权访问此任务")
	}
	return task, nil
}

func (uc *VideoDetectionUseCase) toTaskResponse(task *model.VideoAnalysisTask) *dto.VideoTaskResponse {
	return &dto.VideoTaskResponse{
		TaskID:           task.TaskID,
		SessionID:        task.SessionID,
		BridgeID:         task.BridgeID,
		Status:           task.Status,
		ModelName:        task.ModelName,
		PixelRatio:       task.PixelRatio,
		FPS:              task.FPS,
		EnableSegment:    task.EnableSegment,
		VideoPath:        task.VideoPath,
		TotalFrames:      task.TotalFrames,
		ProcessedFrames:  task.ProcessedFrames,
		ConfirmedDefects: task.ConfirmedDefects,
		ErrorMessage:     task.ErrorMessage,
		CanCancel:        canCancelTask(task.Status),
		CanRetry:         task.Status == model.VideoTaskFailed || task.Status == model.VideoTaskCanceled,
		CreatedAt:        task.CreatedAt,
		StartedAt:        task.StartedAt,
		CompletedAt:      task.CompletedAt,
		CanceledAt:       task.CanceledAt,
	}
}

func (uc *VideoDetectionUseCase) resolveStoredPath(storedPath string) string {
	if filepath.IsAbs(storedPath) {
		return storedPath
	}
	return filepath.Join(uploadsRoot, storedPath)
}

func (uc *VideoDetectionUseCase) persistFrameEvidence(framePath string) (string, error) {
	data, err := os.ReadFile(framePath)
	if err != nil {
		return "", fmt.Errorf("读取证据帧失败: %w", err)
	}

	filename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(framePath))
	relativePath := filepath.Join("video_frames", filename)
	fullPath := filepath.Join(uploadsRoot, relativePath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("创建证据目录失败: %w", err)
	}
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("写入证据帧失败: %w", err)
	}

	return relativePath, nil
}

func (uc *VideoDetectionUseCase) sendMessage(send func(any) error, message any) {
	if send == nil {
		return
	}
	_ = send(message)
}

func translateDefectType(item service.BBoxItem) string {
	if name, ok := service.ClassNameToChinese[item.ClassIdx]; ok {
		return name
	}

	switch item.ClassName {
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
		return item.ClassName
	}
}

func frameTimestampMS(fps float64, frameNo int) int {
	if fps <= 0 {
		return 0
	}
	return int(math.Round(float64(frameNo-1) / fps * 1000))
}

func currentProgress(task *model.VideoAnalysisTask) int {
	if task.TotalFrames <= 0 {
		return 0
	}
	progress := int(math.Round(float64(task.ProcessedFrames) / float64(task.TotalFrames) * 100))
	if progress > 100 {
		return 100
	}
	if progress < 0 {
		return 0
	}
	return progress
}

func getImageDimensions(imagePath string) (int, int, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}

func containsTrack(tracks []*activeDefectTrack, trackID string) bool {
	for _, track := range tracks {
		if track.TrackID == trackID {
			return true
		}
	}
	return false
}

func canCancelTask(status model.VideoTaskStatus) bool {
	return status == model.VideoTaskUploaded ||
		status == model.VideoTaskReadyToProcess ||
		status == model.VideoTaskProcessing
}
