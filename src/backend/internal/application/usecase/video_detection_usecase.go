package usecase

import (
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	appconfig "github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/config"
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
	frameTaskRepo   repository.VideoFrameTaskRepository
	observationRepo repository.DefectObservationRepository
	tasks           sync.Map
	tracks          sync.Map
	cancelFns       sync.Map
	running         sync.Map
	runtimes        sync.Map
	videoConfig     appconfig.VideoDetectionConfig
}

type videoTaskRuntime struct {
	mu      sync.Mutex
	send    func(any) error
	summary map[string]int
	done    chan struct{}
	doneOnce sync.Once
}

// NewVideoDetectionUseCase 创建视频分析用例。
func NewVideoDetectionUseCase(
	defectService *service.DefectService,
	bridgeService *service.BridgeService,
	pythonService service.PythonService,
	fileService service.FileService,
	frameExtractor videopkg.FrameExtractor,
	taskRepo repository.VideoTaskRepository,
	frameTaskRepo repository.VideoFrameTaskRepository,
	observationRepo repository.DefectObservationRepository,
) *VideoDetectionUseCase {
	return NewVideoDetectionUseCaseWithConfig(
		defectService,
		bridgeService,
		pythonService,
		fileService,
		frameExtractor,
		taskRepo,
		frameTaskRepo,
		observationRepo,
		appconfig.VideoDetectionConfig{
			SampleFPS:                    1,
			PlaybackDelaySeconds:         6,
			TrackWindowSeconds:           3,
			TrackCloseSeconds:            4,
			TrackIOUThreshold:            0.3,
			TrackCenterDistanceThreshold: 0.08,
			ConfirmHits:                  3,
			CandidateConfidenceThreshold: 0.45,
			PersistConfidenceThreshold:   0.55,
			MaxQueueInflight:             8,
		},
	)
}

// NewVideoDetectionUseCaseWithConfig 创建带视频配置的视频分析用例。
func NewVideoDetectionUseCaseWithConfig(
	defectService *service.DefectService,
	bridgeService *service.BridgeService,
	pythonService service.PythonService,
	fileService service.FileService,
	frameExtractor videopkg.FrameExtractor,
	taskRepo repository.VideoTaskRepository,
	frameTaskRepo repository.VideoFrameTaskRepository,
	observationRepo repository.DefectObservationRepository,
	videoConfig appconfig.VideoDetectionConfig,
) *VideoDetectionUseCase {
	return &VideoDetectionUseCase{
		defectService:   defectService,
		bridgeService:   bridgeService,
		pythonService:   pythonService,
		fileService:     fileService,
		frameExtractor:  frameExtractor,
		taskRepo:        taskRepo,
		frameTaskRepo:   frameTaskRepo,
		observationRepo: observationRepo,
		videoConfig:     videoConfig,
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
		fps = uc.videoConfig.SampleFPS
	}

	task := &model.VideoAnalysisTask{
		TaskID:        "video_" + uuid.NewString(),
		UserID:        currentUser.ID,
		BridgeID:      req.BridgeID,
		VideoPath:     videoPath,
		FPS:           fps,
		ModelName:     modelName,
		PixelRatio:    req.PixelRatio,
		EnableSegment: false,
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

// HandleAlgoTaskQueued 处理算法端任务入队回调。
func (uc *VideoDetectionUseCase) HandleAlgoTaskQueued(req *dto.AlgoVideoTaskQueuedCallbackRequest) (*dto.VideoCallbackAckResponse, error) {
	task, err := uc.taskRepo.FindByTaskID(req.TaskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("任务不存在")
	}

	task.Status = model.VideoTaskQueued
	task.ErrorMessage = ""
	if err := uc.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("更新任务入队状态失败: %w", err)
	}
	uc.tasks.Store(task.TaskID, task)

	return &dto.VideoCallbackAckResponse{
		TaskID:  task.TaskID,
		Status:  "accepted",
		Message: "task queued callback processed",
	}, nil
}

// HandleAlgoTaskStarted 处理算法端任务开始回调。
func (uc *VideoDetectionUseCase) HandleAlgoTaskStarted(req *dto.AlgoVideoTaskStartedCallbackRequest) (*dto.VideoCallbackAckResponse, error) {
	task, err := uc.taskRepo.FindByTaskID(req.TaskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("任务不存在")
	}

	if task.Status != model.VideoTaskProcessing {
		now := time.Now()
		task.Status = model.VideoTaskProcessing
		task.StartedAt = &now
		task.ErrorMessage = ""
		if err := uc.taskRepo.Update(task); err != nil {
			return nil, fmt.Errorf("更新任务开始状态失败: %w", err)
		}
		uc.tasks.Store(task.TaskID, task)
	}

	return &dto.VideoCallbackAckResponse{
		TaskID:  task.TaskID,
		Status:  "accepted",
		Message: "task started callback processed",
	}, nil
}

// HandleAlgoFrameResult 处理算法端单帧结果回调。
func (uc *VideoDetectionUseCase) HandleAlgoFrameResult(req *dto.AlgoVideoFrameResultCallbackRequest) (*dto.VideoCallbackAckResponse, error) {
	task, err := uc.taskRepo.FindByTaskID(req.TaskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("任务不存在")
	}

	existing, err := uc.frameTaskRepo.FindByRequestID(req.RequestID)
	if err != nil {
		return nil, fmt.Errorf("查询帧请求状态失败: %w", err)
	}
	if existing != nil && (existing.Status == model.VideoFrameTaskCompleted || existing.Status == model.VideoFrameTaskFailed) {
		return &dto.VideoCallbackAckResponse{
			TaskID:    task.TaskID,
			RequestID: req.RequestID,
			Status:    "duplicate",
			Message:   "frame result already processed",
		}, nil
	}

	frameTask, err := uc.upsertFrameTask(req, existing)
	if err != nil {
		return nil, err
	}

	if req.Status == "failed" {
		now := time.Now()
		frameTask.Status = model.VideoFrameTaskFailed
		frameTask.ErrorMessage = req.ErrorMessage
		frameTask.FinishedAt = &now
		if err := uc.frameTaskRepo.Update(frameTask); err != nil {
			return nil, fmt.Errorf("更新帧失败状态失败: %w", err)
		}
		return &dto.VideoCallbackAckResponse{
			TaskID:    task.TaskID,
			RequestID: req.RequestID,
			Status:    "accepted",
			Message:   "frame failure callback processed",
		}, nil
	}

	runtime := uc.loadRuntime(task.TaskID)
	if runtime != nil {
		runtime.mu.Lock()
		defer runtime.mu.Unlock()
	}
	task, err = uc.taskRepo.FindByTaskID(req.TaskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("任务不存在")
	}

	if err := uc.applyFrameResult(task, req, runtime); err != nil {
		return nil, err
	}

	return &dto.VideoCallbackAckResponse{
		TaskID:    task.TaskID,
		RequestID: req.RequestID,
		Status:    "accepted",
		Message:   "frame result callback processed",
	}, nil
}

// HandleAlgoTaskCompleted 处理算法端任务完成回调。
func (uc *VideoDetectionUseCase) HandleAlgoTaskCompleted(req *dto.AlgoVideoTaskCompletedCallbackRequest) (*dto.VideoCallbackAckResponse, error) {
	task, err := uc.taskRepo.FindByTaskID(req.TaskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("任务不存在")
	}

	if err := uc.finalizeOpenTracks(task); err != nil {
		return nil, err
	}

	now := time.Now()
	task.Status = model.VideoTaskCompleted
	if req.TotalFrames > 0 {
		task.TotalFrames = req.TotalFrames
	}
	if req.ProcessedFrames > task.ProcessedFrames {
		task.ProcessedFrames = req.ProcessedFrames
	}
	task.CompletedAt = &now
	task.ErrorMessage = ""
	if err := uc.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("更新任务完成状态失败: %w", err)
	}
	uc.tasks.Store(task.TaskID, task)
	if runtime := uc.loadRuntime(task.TaskID); runtime != nil {
		runtime.mu.Lock()
		defectSummary := make(map[string]int, len(runtime.summary))
		for k, v := range runtime.summary {
			defectSummary[k] = v
		}
		send := runtime.send
		runtime.mu.Unlock()
		uc.sendMessage(send, dto.VideoCompletedMessage{
			Type:            "completed",
			TaskID:          task.TaskID,
			SessionID:       task.SessionID,
			ProcessedFrames: task.ProcessedFrames,
			TotalFrames:     task.TotalFrames,
			TotalDefects:    task.ConfirmedDefects,
			DefectSummary:   defectSummary,
		})
	}
	uc.completeRuntime(task.TaskID)

	return &dto.VideoCallbackAckResponse{
		TaskID:  task.TaskID,
		Status:  "accepted",
		Message: "task completed callback processed",
	}, nil
}

// HandleAlgoTaskFailed 处理算法端任务失败回调。
func (uc *VideoDetectionUseCase) HandleAlgoTaskFailed(req *dto.AlgoVideoTaskFailedCallbackRequest) (*dto.VideoCallbackAckResponse, error) {
	task, err := uc.taskRepo.FindByTaskID(req.TaskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("任务不存在")
	}

	now := time.Now()
	task.Status = model.VideoTaskFailed
	task.CompletedAt = &now
	task.ErrorMessage = req.ErrorMessage
	if err := uc.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("更新任务失败状态失败: %w", err)
	}
	uc.tasks.Store(task.TaskID, task)
	if runtime := uc.loadRuntime(task.TaskID); runtime != nil {
		runtime.mu.Lock()
		send := runtime.send
		runtime.mu.Unlock()
		uc.sendMessage(send, dto.VideoErrorMessage{
			Type:      "error",
			TaskID:    task.TaskID,
			SessionID: task.SessionID,
			Message:   req.ErrorMessage,
		})
	}
	uc.completeRuntime(task.TaskID)

	return &dto.VideoCallbackAckResponse{
		TaskID:  task.TaskID,
		Status:  "accepted",
		Message: "task failed callback processed",
	}, nil
}

// CancelTask 取消任务。
func (uc *VideoDetectionUseCase) CancelTask(taskID string, currentUser *model.User) (*dto.VideoTaskResponse, error) {
	task, err := uc.loadOwnedTask(taskID, currentUser)
	if err != nil {
		return nil, err
	}

	switch task.Status {
	case model.VideoTaskUploaded, model.VideoTaskFrameExtracting, model.VideoTaskDispatching, model.VideoTaskReadyToProcess, model.VideoTaskQueued, model.VideoTaskProcessing:
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
	runtime := uc.ensureRuntime(task.TaskID, send)

	switch task.Status {
	case model.VideoTaskReadyToProcess:
		if _, loaded := uc.running.LoadOrStore(task.TaskID, struct{}{}); loaded {
			return nil
		}
		defer uc.running.Delete(task.TaskID)
		return uc.runTask(task, send)
	case model.VideoTaskFrameExtracting, model.VideoTaskDispatching, model.VideoTaskQueued:
		<-runtime.done
		return nil
	case model.VideoTaskProcessing:
		<-runtime.done
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
	task.Status = model.VideoTaskFrameExtracting
	task.StartedAt = &now
	task.ErrorMessage = ""

	if totalFrames, err := uc.frameExtractor.CountFrames(uc.resolveStoredPath(task.VideoPath), task.FPS); err == nil && totalFrames > 0 {
		task.TotalFrames = totalFrames
	}
	if err := uc.taskRepo.Update(task); err != nil {
		return fmt.Errorf("更新处理状态失败: %w", err)
	}
	uc.tasks.Store(task.TaskID, task)

	tempDir := filepath.Join(uploadsRoot, "video_tasks", task.TaskID, "frames")
	_ = os.RemoveAll(tempDir)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("创建视频帧工作目录失败: %w", err)
	}
	maxInflight := uc.videoConfig.MaxQueueInflight
	sem := make(chan struct{}, maxInflight)
	var wg sync.WaitGroup
	var asyncErr error
	var asyncErrMu sync.Mutex

	setAsyncErr := func(err error) {
		if err == nil {
			return
		}
		asyncErrMu.Lock()
		if asyncErr == nil {
			asyncErr = err
			cancel()
		}
		asyncErrMu.Unlock()
	}

	task.Status = model.VideoTaskDispatching
	if err := uc.taskRepo.Update(task); err != nil {
		return fmt.Errorf("更新任务派发状态失败: %w", err)
	}
	uc.tasks.Store(task.TaskID, task)

	err := uc.frameExtractor.ExtractFramesStream(
		uc.resolveStoredPath(task.VideoPath),
		tempDir,
		task.FPS,
		func(framePath string, frameNo int) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			requestID := fmt.Sprintf("%s_frame_%06d", task.TaskID, frameNo)
			timestampMS := frameTimestampMS(task.FPS, frameNo)
			if err := uc.registerQueuedFrameTask(task, requestID, frameNo, timestampMS, framePath); err != nil {
				return err
			}

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return ctx.Err()
			}

			wg.Add(1)
			go func(framePath string, frameNo int, requestID string, timestampMS int) {
				defer wg.Done()
				defer func() { <-sem }()
				if err := uc.dispatchFrameDetect(ctx, task, framePath, frameNo, requestID, timestampMS); err != nil {
					setAsyncErr(err)
				}
			}(framePath, frameNo, requestID, timestampMS)
			return nil
		},
	)
	wg.Wait()
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
			uc.completeRuntime(task.TaskID)
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
		uc.completeRuntime(task.TaskID)
		return err
	}
	if asyncErr != nil {
		task.Status = model.VideoTaskFailed
		task.ErrorMessage = asyncErr.Error()
		failedAt := time.Now()
		task.CompletedAt = &failedAt
		_ = uc.taskRepo.Update(task)
		uc.tasks.Store(task.TaskID, task)
		uc.sendMessage(send, dto.VideoErrorMessage{
			Type:      "error",
			TaskID:    task.TaskID,
			SessionID: task.SessionID,
			Message:   asyncErr.Error(),
		})
		uc.completeRuntime(task.TaskID)
		return asyncErr
	}

	refreshedTask, err := uc.taskRepo.FindByTaskID(task.TaskID)
	if err != nil {
		return err
	}
	if refreshedTask != nil {
		task = refreshedTask
	}
	if _, err := uc.HandleAlgoTaskCompleted(&dto.AlgoVideoTaskCompletedCallbackRequest{
		TaskID:          task.TaskID,
		Status:          "completed",
		TotalFrames:     task.TotalFrames,
		ProcessedFrames: task.ProcessedFrames,
		FailedFrames:    0,
	}); err != nil {
		return err
	}

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

func (uc *VideoDetectionUseCase) loadRuntime(taskID string) *videoTaskRuntime {
	runtime, ok := uc.runtimes.Load(taskID)
	if !ok {
		return nil
	}
	typed, _ := runtime.(*videoTaskRuntime)
	return typed
}

func (uc *VideoDetectionUseCase) ensureRuntime(taskID string, send func(any) error) *videoTaskRuntime {
	if runtime := uc.loadRuntime(taskID); runtime != nil {
		runtime.mu.Lock()
		runtime.send = send
		runtime.mu.Unlock()
		return runtime
	}

	runtime := &videoTaskRuntime{
		send:    send,
		summary: make(map[string]int),
		done:    make(chan struct{}),
	}
	uc.runtimes.Store(taskID, runtime)
	return runtime
}

func (uc *VideoDetectionUseCase) completeRuntime(taskID string) {
	runtime := uc.loadRuntime(taskID)
	if runtime == nil {
		return
	}
	runtime.doneOnce.Do(func() {
		close(runtime.done)
	})
}

func (uc *VideoDetectionUseCase) registerQueuedFrameTask(task *model.VideoAnalysisTask, requestID string, frameNo int, timestampMS int, framePath string) error {
	now := time.Now()
	frameTask := &model.VideoFrameTaskRequest{
		RequestID:   requestID,
		TaskID:      task.TaskID,
		FrameNo:     frameNo,
		TimestampMS: timestampMS,
		Status:      model.VideoFrameTaskQueued,
		FrameRef:    "file://" + framePath,
		QueuedAt:    &now,
	}
	if err := uc.frameTaskRepo.Create(frameTask); err != nil {
		return fmt.Errorf("创建排队帧任务失败: %w", err)
	}
	return nil
}

func (uc *VideoDetectionUseCase) dispatchFrameDetect(ctx context.Context, task *model.VideoAnalysisTask, framePath string, frameNo int, requestID string, timestampMS int) error {
	frameTask, err := uc.frameTaskRepo.FindByRequestID(requestID)
	if err != nil {
		return fmt.Errorf("查询排队帧任务失败: %w", err)
	}
	if frameTask == nil {
		return fmt.Errorf("排队帧任务不存在: %s", requestID)
	}

	startedAt := time.Now()
	frameTask.Status = model.VideoFrameTaskProcessing
	frameTask.StartedAt = &startedAt
	if frameTask.QueuedAt != nil {
		frameTask.QueueLatencyMS = int(startedAt.Sub(*frameTask.QueuedAt).Milliseconds())
	}
	if err := uc.frameTaskRepo.Update(frameTask); err != nil {
		return fmt.Errorf("更新帧任务处理中状态失败: %w", err)
	}

	if _, err := uc.HandleAlgoTaskStarted(&dto.AlgoVideoTaskStartedCallbackRequest{
		TaskID: task.TaskID,
		Status: "processing",
	}); err != nil {
		return err
	}

	detectStartedAt := time.Now()
	detectResult, err := uc.pythonService.Detect(framePath, &service.DetectRequest{
		ModelName: task.ModelName,
		Conf:      0.25,
	})
	detectTotalMS := int(time.Since(detectStartedAt).Milliseconds())
	if err != nil {
		_, handleErr := uc.HandleAlgoFrameResult(&dto.AlgoVideoFrameResultCallbackRequest{
			RequestID:      requestID,
			TaskID:         task.TaskID,
			FrameNo:        frameNo,
			TimestampMS:    timestampMS,
			Status:         "failed",
			QueueLatencyMS: frameTask.QueueLatencyMS,
			DetectTotalMS:  detectTotalMS,
			FrameRef:       "file://" + framePath,
			ErrorMessage:   err.Error(),
		})
		if handleErr != nil {
			return handleErr
		}
		return fmt.Errorf("AI检测失败: %w", err)
	}

	yoloBBoxes := make([]dto.AlgoVideoBBox, 0, len(detectResult.YOLOBBoxes))
	for _, bbox := range detectResult.YOLOBBoxes {
		yoloBBoxes = append(yoloBBoxes, dto.AlgoVideoBBox{
			BoxID:      bbox.BoxID,
			ClassIdx:   bbox.ClassIdx,
			ClassName:  bbox.ClassName,
			YOLOCoords: bbox.YOLOCoords,
			Confidence: bbox.Confidence,
		})
	}

	_, err = uc.HandleAlgoFrameResult(&dto.AlgoVideoFrameResultCallbackRequest{
		RequestID:      requestID,
		TaskID:         task.TaskID,
		FrameNo:        frameNo,
		TimestampMS:    timestampMS,
		Status:         "success",
		QueueLatencyMS: frameTask.QueueLatencyMS,
		DetectTotalMS:  detectTotalMS,
		InferMS:        detectTotalMS,
		FrameRef:       "file://" + framePath,
		YOLOBBoxes:     yoloBBoxes,
	})
	return err
}

func (uc *VideoDetectionUseCase) applyFrameResult(task *model.VideoAnalysisTask, req *dto.AlgoVideoFrameResultCallbackRequest, runtime *videoTaskRuntime) error {
	framePath := normalizeFrameRef(req.FrameRef)
	imgW, imgH := 0, 0
	if framePath != "" {
		if width, height, err := getImageDimensions(framePath); err == nil {
			imgW = width
			imgH = height
		}
	}

	tracks := uc.loadTracks(task.TaskID)
	matchedTrackIDs := make(map[string]struct{})
	defects := make([]dto.VideoFrameDefect, 0, len(req.YOLOBBoxes))
	observedAt := time.Now()
	persistedResultPath := ""

	for _, item := range req.YOLOBBoxes {
		if item.Confidence < uc.videoConfig.CandidateConfidenceThreshold {
			continue
		}
		defectType := translateDefectType(service.BBoxItem{
			ClassIdx:   item.ClassIdx,
			ClassName:  item.ClassName,
			YOLOCoords: item.YOLOCoords,
			Confidence: item.Confidence,
		})

		currentDefect := dto.VideoFrameDefect{
			DefectType: defectType,
			MeasurementSource: "bbox_estimate",
			MeasurementStatus: "estimated",
		}
		if imgW > 0 && imgH > 0 {
			x, y, w, h, length, width, area := service.CalculatePhysicalDimensions(item.YOLOCoords, imgW, imgH, task.PixelRatio)
			currentDefect.BBox = dto.VideoBBox{X: x, Y: y, Width: w, Height: h}
			currentDefect.Length = length
			currentDefect.Width = width
			currentDefect.Area = area
		}

		track, becameBest := uc.upsertTrack(task, tracks, currentDefect, item.Confidence, req.FrameNo, observedAt)
		matchedTrackIDs[track.TrackID] = struct{}{}
		if !containsTrack(tracks, track.TrackID) {
			tracks = append(tracks, track)
		}

		if becameBest && framePath != "" {
			bestFramePath, err := uc.persistFrameEvidence(framePath)
			if err == nil {
				if track.BestFramePath != "" && track.BestFramePath != bestFramePath {
					_ = uc.fileService.DeleteFile(track.BestFramePath)
				}
				track.BestFramePath = bestFramePath
			}
		}

		if err := uc.observationRepo.Create(&model.DefectObservation{
			TaskID:      task.TaskID,
			BridgeID:    task.BridgeID,
			FrameNo:     req.FrameNo,
			TimestampMS: req.TimestampMS,
			DefectType:  currentDefect.DefectType,
			BBox:        currentDefect.BBox.JSON(),
			Confidence:  item.Confidence,
			FramePath:   framePath,
			ResultPath:  "",
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
			if runtime != nil {
				runtime.summary[track.DefectType]++
			}
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

	uc.closeMissedTracks(task, tracks, matchedTrackIDs, req.FrameNo)
	uc.storeTracks(task.TaskID, tracks)

	task.ProcessedFrames++
	task.Status = model.VideoTaskProcessing
	task.ErrorMessage = ""
	if err := uc.taskRepo.Update(task); err != nil {
		return fmt.Errorf("更新任务帧进度失败: %w", err)
	}
	uc.tasks.Store(task.TaskID, task)

	if runtime != nil {
		progress := currentProgress(task)
		uc.sendMessage(runtime.send, dto.VideoFrameMessage{
			Type:                "frame",
			TaskID:              task.TaskID,
			SessionID:           task.SessionID,
			FrameNo:             req.FrameNo,
			ProcessedFrames:     task.ProcessedFrames,
			TotalFrames:         task.TotalFrames,
			Progress:            progress,
			TimestampMS:         req.TimestampMS,
			ImageBase64:         "",
			PersistedResultPath: persistedResultPath,
			Defects:             defects,
			DefectCount:         len(defects),
			QueueLatencyMS:      req.QueueLatencyMS,
			DetectLatencyMS:     req.DetectTotalMS,
		})
		uc.sendMessage(runtime.send, dto.VideoProgressMessage{
			Type:            "progress",
			TaskID:          task.TaskID,
			SessionID:       task.SessionID,
			ProcessedFrames: task.ProcessedFrames,
			TotalFrames:     task.TotalFrames,
			Progress:        progress,
			TotalDefects:    task.ConfirmedDefects,
		})
	}

	return nil
}

func (uc *VideoDetectionUseCase) finalizeOpenTracks(task *model.VideoAnalysisTask) error {
	runtime := uc.loadRuntime(task.TaskID)
	if runtime != nil {
		runtime.mu.Lock()
		defer runtime.mu.Unlock()
	}

	tracks := uc.loadTracks(task.TaskID)
	changed := false
	for _, track := range tracks {
		if track.Status == trackClosed {
			continue
		}
		track.Status = trackClosed
		changed = true
		if !uc.shouldConfirmTrack(track) || track.DefectID != nil {
			continue
		}

		defect := uc.buildConfirmedDefect(task, track, track.LastSeenAt)
		created, err := uc.defectService.ConfirmVideoDefect(defect)
		if err != nil {
			return fmt.Errorf("结算未关闭轨迹失败: %w", err)
		}
		track.DefectID = &created.ID
		task.ConfirmedDefects++
		if runtime != nil {
			runtime.summary[track.DefectType]++
		}
	}

	if changed {
		uc.storeTracks(task.TaskID, tracks)
		if err := uc.taskRepo.Update(task); err != nil {
			return fmt.Errorf("更新任务轨迹结算结果失败: %w", err)
		}
		uc.tasks.Store(task.TaskID, task)
	}
	return nil
}

func (uc *VideoDetectionUseCase) upsertFrameTask(req *dto.AlgoVideoFrameResultCallbackRequest, existing *model.VideoFrameTaskRequest) (*model.VideoFrameTaskRequest, error) {
	now := time.Now()
	if existing == nil {
		frameTask := &model.VideoFrameTaskRequest{
			RequestID:      req.RequestID,
			TaskID:         req.TaskID,
			FrameNo:        req.FrameNo,
			TimestampMS:    req.TimestampMS,
			Status:         model.VideoFrameTaskCompleted,
			FrameRef:       req.FrameRef,
			QueueLatencyMS: req.QueueLatencyMS,
			DetectTotalMS:  req.DetectTotalMS,
			DecodeMS:       req.DecodeMS,
			InferMS:        req.InferMS,
			PostprocessMS:  req.PostprocessMS,
			ErrorMessage:   req.ErrorMessage,
			FinishedAt:     &now,
		}
		if req.Status == "failed" {
			frameTask.Status = model.VideoFrameTaskFailed
		}
		if err := uc.frameTaskRepo.Create(frameTask); err != nil {
			return nil, fmt.Errorf("创建帧请求记录失败: %w", err)
		}
		return frameTask, nil
	}

	existing.FrameNo = req.FrameNo
	existing.TimestampMS = req.TimestampMS
	existing.FrameRef = req.FrameRef
	existing.QueueLatencyMS = req.QueueLatencyMS
	existing.DetectTotalMS = req.DetectTotalMS
	existing.DecodeMS = req.DecodeMS
	existing.InferMS = req.InferMS
	existing.PostprocessMS = req.PostprocessMS
	existing.ErrorMessage = req.ErrorMessage
	existing.FinishedAt = &now
	if req.Status == "failed" {
		existing.Status = model.VideoFrameTaskFailed
	} else {
		existing.Status = model.VideoFrameTaskCompleted
	}
	if err := uc.frameTaskRepo.Update(existing); err != nil {
		return nil, fmt.Errorf("更新帧请求记录失败: %w", err)
	}
	return existing, nil
}

func normalizeFrameRef(frameRef string) string {
	return strings.TrimPrefix(frameRef, "file://")
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
		status == model.VideoTaskFrameExtracting ||
		status == model.VideoTaskDispatching ||
		status == model.VideoTaskReadyToProcess ||
		status == model.VideoTaskQueued ||
		status == model.VideoTaskProcessing
}
