package usecase_test

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/persistence"
	appconfig "github.com/xinyiis/BridgeDefectDetectionSys/src/backend/pkg/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type stubFrameExtractor struct {
	frames []string
}

func (s *stubFrameExtractor) CountFrames(string, float64) (int, error) {
	return len(s.frames), nil
}

func (s *stubFrameExtractor) ExtractFramesStream(_ string, _ string, _ float64, onFrame func(framePath string, frameNo int) error) error {
	for idx, frame := range s.frames {
		if err := onFrame(frame, idx+1); err != nil {
			return err
		}
	}
	return nil
}

type stubPythonService struct {
	onEnqueue func(req *service.VideoFrameDetectEnqueueRequest)
}

func (s *stubPythonService) DetectDefect(string, string, string, float64) (*service.PythonDetectionResult, error) {
	return nil, nil
}

func (s *stubPythonService) Detect(string, *service.DetectRequest) (*service.DetectResult, error) {
	return &service.DetectResult{
		Status:    "success",
		ModelUsed: "baseline",
		YOLOBBoxes: []service.BBoxItem{
			{
				BoxID:      1,
				ClassIdx:   0,
				ClassName:  "Crack",
				YOLOCoords: [4]float64{0.5, 0.5, 0.3, 0.1},
				Confidence: 0.92,
			},
		},
		ImageResult: base64.StdEncoding.EncodeToString([]byte("frame-result")),
	}, nil
}

func (s *stubPythonService) Preprocess(string, string) (*service.PreprocessResult, error) {
	return &service.PreprocessResult{Status: "success"}, nil
}

func (s *stubPythonService) Segment(string, *service.SegmentRequest) (*service.SegmentResult, error) {
	return &service.SegmentResult{Status: "success"}, nil
}

func (s *stubPythonService) EnqueueVideoFrameDetect(req *service.VideoFrameDetectEnqueueRequest) (*service.VideoFrameDetectEnqueueResponse, error) {
	if req == nil {
		return nil, nil
	}
	if s.onEnqueue != nil {
		s.onEnqueue(req)
	}
	return &service.VideoFrameDetectEnqueueResponse{
		Status:    "accepted",
		RequestID: req.RequestID,
		TaskID:    req.TaskID,
	}, nil
}

func wireSuccessfulCallbacks(t *testing.T, videoUC *usecase.VideoDetectionUseCase, pythonService *stubPythonService) {
	t.Helper()
	requests := make(chan *service.VideoFrameDetectEnqueueRequest, 32)
	t.Cleanup(func() {
		close(requests)
	})

	pythonService.onEnqueue = func(req *service.VideoFrameDetectEnqueueRequest) {
		if req == nil {
			return
		}
		requests <- req
	}

	go func() {
		var sentStarted sync.Map
		for req := range requests {
			if _, err := videoUC.HandleAlgoTaskQueued(&dto.AlgoVideoTaskQueuedCallbackRequest{
				TaskID:   req.TaskID,
				Status:   "queued",
				QueuedAt: time.Now().UTC().Format(time.RFC3339),
			}); err != nil {
				t.Errorf("task queued callback failed: %v", err)
				continue
			}

			if _, loaded := sentStarted.LoadOrStore(req.TaskID, struct{}{}); !loaded {
				if _, err := videoUC.HandleAlgoTaskStarted(&dto.AlgoVideoTaskStartedCallbackRequest{
					TaskID:    req.TaskID,
					Status:    "processing",
					StartedAt: time.Now().UTC().Format(time.RFC3339),
				}); err != nil {
					t.Errorf("task started callback failed: %v", err)
					continue
				}
			}

			if _, err := videoUC.HandleAlgoFrameResult(&dto.AlgoVideoFrameResultCallbackRequest{
				RequestID:      req.RequestID,
				TaskID:         req.TaskID,
				FrameNo:        req.FrameNo,
				TimestampMS:    req.TimestampMS,
				Status:         "success",
				QueueLatencyMS: 10,
				DetectTotalMS:  100,
				DecodeMS:       10,
				InferMS:        70,
				PostprocessMS:  20,
				FrameRef:       req.FrameRef,
				YOLOBBoxes: []dto.AlgoVideoBBox{
					{
						BoxID:      1,
						ClassIdx:   0,
						ClassName:  "Crack",
						YOLOCoords: [4]float64{0.5, 0.5, 0.3, 0.1},
						Confidence: 0.92,
					},
				},
			}); err != nil {
				t.Errorf("frame result callback failed: %v", err)
			}
		}
	}()
}

func setupVideoUseCaseTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db failed: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	err = db.AutoMigrate(
		&model.User{},
		&model.Bridge{},
		&model.Defect{},
		&model.VideoAnalysisTask{},
		&model.DefectObservation{},
		&model.VideoFrameTaskRequest{},
	)
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	return db
}

func createVideoTestUser(t *testing.T, db *gorm.DB) *model.User {
	t.Helper()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password failed: %v", err)
	}

	user := &model.User{
		Username: "video_user",
		Password: string(hashedPassword),
		RealName: "Video User",
		Email:    "video_user@example.com",
		Role:     "user",
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	return user
}

func createVideoTestBridge(t *testing.T, db *gorm.DB, userID uint) *model.Bridge {
	t.Helper()

	bridge := &model.Bridge{
		BridgeName: "测试桥梁",
		BridgeCode: "VIDEO_BRIDGE_001",
		UserID:     userID,
	}

	if err := db.Create(bridge).Error; err != nil {
		t.Fatalf("create bridge failed: %v", err)
	}

	return bridge
}

func createJPEGFrame(t *testing.T, dir string, name string) string {
	t.Helper()

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	framePath := filepath.Join(dir, name)
	file, err := os.Create(framePath)
	if err != nil {
		t.Fatalf("create frame failed: %v", err)
	}
	defer file.Close()

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 200, B: 200, A: 255})
		}
	}

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode jpeg failed: %v", err)
	}

	return framePath
}

func TestVideoDetectionUseCase_StreamTaskConfirmsDefect(t *testing.T) {
	db := setupVideoUseCaseTestDB(t)
	user := createVideoTestUser(t, db)
	bridge := createVideoTestBridge(t, db, user.ID)

	userRepo := persistence.NewUserRepository(db)
	bridgeRepo := persistence.NewBridgeRepository(db)
	defectRepo := persistence.NewDefectRepository(db)
	taskRepo := persistence.NewVideoTaskRepository(db)
	frameTaskRepo := persistence.NewVideoFrameTaskRepository(db)
	observationRepo := persistence.NewDefectObservationRepository(db)
	fileService := persistence.NewLocalFileStorage("./uploads")

	userService := service.NewUserService(userRepo)
	_ = userService
	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)

	tempDir := t.TempDir()
	frames := []string{
		createJPEGFrame(t, tempDir, "frame_001.jpg"),
		createJPEGFrame(t, tempDir, "frame_002.jpg"),
		createJPEGFrame(t, tempDir, "frame_003.jpg"),
	}

	pythonService := &stubPythonService{}
	videoUC := usecase.NewVideoDetectionUseCase(
		defectService,
		bridgeService,
		pythonService,
		fileService,
		&stubFrameExtractor{frames: frames},
		taskRepo,
		frameTaskRepo,
		observationRepo,
	)
	wireSuccessfulCallbacks(t, videoUC, pythonService)

	now := time.Now()
	task := &model.VideoAnalysisTask{
		TaskID:        "video_task_test",
		SessionID:     "session_test",
		UserID:        user.ID,
		BridgeID:      bridge.ID,
		VideoPath:     "videos/test.mp4",
		FPS:           1,
		ModelName:     "baseline",
		PixelRatio:    0.01,
		EnableSegment: false,
		Status:        model.VideoTaskReadyToProcess,
		CreatedAt:     now,
	}

	if err := taskRepo.Create(task); err != nil {
		t.Fatalf("create task failed: %v", err)
	}

	var rawMessages []any
	err := videoUC.StreamTask(task.TaskID, task.SessionID, user, func(message any) error {
		rawMessages = append(rawMessages, message)
		return nil
	})
	if err != nil {
		t.Fatalf("stream task failed: %v", err)
	}

	storedTask, err := taskRepo.FindByTaskID(task.TaskID)
	if err != nil {
		t.Fatalf("find task failed: %v", err)
	}
	if storedTask.Status != model.VideoTaskCompleted {
		t.Fatalf("expected completed task, got %s", storedTask.Status)
	}
	if storedTask.ProcessedFrames != 3 {
		t.Fatalf("expected 3 processed frames, got %d", storedTask.ProcessedFrames)
	}
	if storedTask.ConfirmedDefects != 1 {
		t.Fatalf("expected 1 confirmed defect, got %d", storedTask.ConfirmedDefects)
	}

	var defects []model.Defect
	if err := db.Find(&defects).Error; err != nil {
		t.Fatalf("list defects failed: %v", err)
	}
	if len(defects) != 1 {
		t.Fatalf("expected 1 defect record, got %d", len(defects))
	}
	if defects[0].SourceType != "video" {
		t.Fatalf("expected source_type video, got %s", defects[0].SourceType)
	}
	if defects[0].ObservationCount != 3 {
		t.Fatalf("expected observation count 3, got %d", defects[0].ObservationCount)
	}
	if defects[0].SourceTaskID == nil || *defects[0].SourceTaskID != task.TaskID {
		t.Fatalf("expected source task id %s, got %#v", task.TaskID, defects[0].SourceTaskID)
	}
	if defects[0].BestFramePath == "" {
		t.Fatalf("expected persisted frame evidence path, got frame=%s", defects[0].BestFramePath)
	}
	if defects[0].BestResultPath != "" {
		t.Fatalf("expected detect-only video path to keep best result path empty, got %s", defects[0].BestResultPath)
	}
	if defects[0].Area != 0 || defects[0].Length != 0 || defects[0].Width != 0 {
		t.Fatalf("expected video defects to omit physical metrics, got area=%v length=%v width=%v", defects[0].Area, defects[0].Length, defects[0].Width)
	}
	if defects[0].MeasurementSource != "none" || defects[0].MeasurementStatus != "unavailable" {
		t.Fatalf("expected unavailable measurement markers, got source=%s status=%s", defects[0].MeasurementSource, defects[0].MeasurementStatus)
	}

	var observationCount int64
	if err := db.Model(&model.DefectObservation{}).Count(&observationCount).Error; err != nil {
		t.Fatalf("count observations failed: %v", err)
	}
	if observationCount != 3 {
		t.Fatalf("expected 3 observations, got %d", observationCount)
	}

	var frameTaskCount int64
	if err := db.Model(&model.VideoFrameTaskRequest{}).Count(&frameTaskCount).Error; err != nil {
		t.Fatalf("count frame tasks failed: %v", err)
	}
	if frameTaskCount != 3 {
		t.Fatalf("expected 3 frame task requests, got %d", frameTaskCount)
	}

	if len(rawMessages) < 3 {
		t.Fatalf("expected streamed messages, got %d", len(rawMessages))
	}
}

func TestVideoDetectionUseCase_StreamTaskRejectsWrongUser(t *testing.T) {
	db := setupVideoUseCaseTestDB(t)
	user := createVideoTestUser(t, db)
	bridge := createVideoTestBridge(t, db, user.ID)

	otherHashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password failed: %v", err)
	}
	otherUser := &model.User{
		Username: "other_video_user",
		Password: string(otherHashedPassword),
		RealName: "Other",
		Email:    "other_video_user@example.com",
		Role:     "user",
	}
	if err := db.Create(otherUser).Error; err != nil {
		t.Fatalf("create other user failed: %v", err)
	}

	bridgeRepo := persistence.NewBridgeRepository(db)
	defectRepo := persistence.NewDefectRepository(db)
	taskRepo := persistence.NewVideoTaskRepository(db)
	frameTaskRepo := persistence.NewVideoFrameTaskRepository(db)
	observationRepo := persistence.NewDefectObservationRepository(db)
	fileService := persistence.NewLocalFileStorage("./uploads")
	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)

	videoUC := usecase.NewVideoDetectionUseCase(
		defectService,
		bridgeService,
		&stubPythonService{},
		fileService,
		&stubFrameExtractor{},
		taskRepo,
		frameTaskRepo,
		observationRepo,
	)

	task := &model.VideoAnalysisTask{
		TaskID:     "video_task_forbidden",
		SessionID:  "session_forbidden",
		UserID:     user.ID,
		BridgeID:   bridge.ID,
		VideoPath:  "videos/test.mp4",
		FPS:        1,
		ModelName:  "baseline",
		PixelRatio: 0.01,
		Status:     model.VideoTaskReadyToProcess,
	}
	if err := taskRepo.Create(task); err != nil {
		t.Fatalf("create task failed: %v", err)
	}

	err = videoUC.StreamTask(task.TaskID, task.SessionID, otherUser, func(any) error { return nil })
	if err == nil {
		t.Fatal("expected forbidden error, got nil")
	}
}

func TestCreateJPEGFrameProducesValidJPEG(t *testing.T) {
	path := createJPEGFrame(t, t.TempDir(), "frame.jpg")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read frame failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected jpeg data")
	}
	if !bytes.HasPrefix(data, []byte{0xFF, 0xD8}) {
		t.Fatal("expected jpeg header")
	}
}

func TestVideoDetectionUseCase_HonorsCandidateConfidenceThreshold(t *testing.T) {
	db := setupVideoUseCaseTestDB(t)
	user := createVideoTestUser(t, db)
	bridge := createVideoTestBridge(t, db, user.ID)

	bridgeRepo := persistence.NewBridgeRepository(db)
	defectRepo := persistence.NewDefectRepository(db)
	taskRepo := persistence.NewVideoTaskRepository(db)
	frameTaskRepo := persistence.NewVideoFrameTaskRepository(db)
	observationRepo := persistence.NewDefectObservationRepository(db)
	fileService := persistence.NewLocalFileStorage("./uploads")
	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)

	tempDir := t.TempDir()
	frames := []string{
		createJPEGFrame(t, tempDir, "frame_001.jpg"),
		createJPEGFrame(t, tempDir, "frame_002.jpg"),
		createJPEGFrame(t, tempDir, "frame_003.jpg"),
	}

	pythonService := &stubPythonService{}
	videoUC := usecase.NewVideoDetectionUseCaseWithConfig(
		defectService,
		bridgeService,
		pythonService,
		fileService,
		&stubFrameExtractor{frames: frames},
		taskRepo,
		frameTaskRepo,
		observationRepo,
		appconfig.VideoDetectionConfig{
			SampleFPS:                    1,
			PlaybackDelaySeconds:         6,
			TrackWindowSeconds:           3,
			TrackCloseSeconds:            4,
			TrackIOUThreshold:            0.3,
			CandidateConfidenceThreshold: 0.95,
			PersistConfidenceThreshold:   0.55,
			MaxQueueInflight:             4,
			QueuedTimeoutSeconds:         1,
			ProgressIdleTimeoutSeconds:   1,
		},
	)
	wireSuccessfulCallbacks(t, videoUC, pythonService)

	task := &model.VideoAnalysisTask{
		TaskID:     "video_task_threshold",
		SessionID:  "session_threshold",
		UserID:     user.ID,
		BridgeID:   bridge.ID,
		VideoPath:  "videos/test.mp4",
		FPS:        1,
		ModelName:  "baseline",
		PixelRatio: 0.01,
		Status:     model.VideoTaskReadyToProcess,
	}
	if err := taskRepo.Create(task); err != nil {
		t.Fatalf("create task failed: %v", err)
	}

	if err := videoUC.StreamTask(task.TaskID, task.SessionID, user, func(any) error { return nil }); err != nil {
		t.Fatalf("stream task failed: %v", err)
	}

	var observationCount int64
	if err := db.Model(&model.DefectObservation{}).Count(&observationCount).Error; err != nil {
		t.Fatalf("count observations failed: %v", err)
	}
	if observationCount != 0 {
		t.Fatalf("expected no observations under high candidate threshold, got %d", observationCount)
	}

	var defectCount int64
	if err := db.Model(&model.Defect{}).Count(&defectCount).Error; err != nil {
		t.Fatalf("count defects failed: %v", err)
	}
	if defectCount != 0 {
		t.Fatalf("expected no confirmed defects under high candidate threshold, got %d", defectCount)
	}
}

func TestVideoDetectionUseCase_QueuedTimeoutFailsTask(t *testing.T) {
	db := setupVideoUseCaseTestDB(t)
	user := createVideoTestUser(t, db)
	bridge := createVideoTestBridge(t, db, user.ID)

	bridgeRepo := persistence.NewBridgeRepository(db)
	defectRepo := persistence.NewDefectRepository(db)
	taskRepo := persistence.NewVideoTaskRepository(db)
	frameTaskRepo := persistence.NewVideoFrameTaskRepository(db)
	observationRepo := persistence.NewDefectObservationRepository(db)
	fileService := persistence.NewLocalFileStorage("./uploads")
	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)

	tempDir := t.TempDir()
	frames := []string{createJPEGFrame(t, tempDir, "frame_001.jpg")}

	pythonService := &stubPythonService{}
	videoUC := usecase.NewVideoDetectionUseCaseWithConfig(
		defectService,
		bridgeService,
		pythonService,
		fileService,
		&stubFrameExtractor{frames: frames},
		taskRepo,
		frameTaskRepo,
		observationRepo,
		appconfig.VideoDetectionConfig{
			SampleFPS:                  1,
			MaxQueueInflight:           1,
			QueuedTimeoutSeconds:       1,
			ProgressIdleTimeoutSeconds: 1,
		},
	)

	task := &model.VideoAnalysisTask{
		TaskID:     "video_task_queued_timeout",
		SessionID:  "session_queued_timeout",
		UserID:     user.ID,
		BridgeID:   bridge.ID,
		VideoPath:  "videos/test.mp4",
		FPS:        1,
		ModelName:  "baseline",
		PixelRatio: 0.01,
		Status:     model.VideoTaskReadyToProcess,
	}
	if err := taskRepo.Create(task); err != nil {
		t.Fatalf("create task failed: %v", err)
	}

	err := videoUC.StreamTask(task.TaskID, task.SessionID, user, func(any) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "queued 回调超时") {
		t.Fatalf("expected queued timeout error, got %v", err)
	}

	storedTask, err := taskRepo.FindByTaskID(task.TaskID)
	if err != nil {
		t.Fatalf("find task failed: %v", err)
	}
	if storedTask.Status != model.VideoTaskFailed {
		t.Fatalf("expected failed task, got %s", storedTask.Status)
	}
}

func TestVideoDetectionUseCase_ProgressIdleTimeoutFailsTask(t *testing.T) {
	db := setupVideoUseCaseTestDB(t)
	user := createVideoTestUser(t, db)
	bridge := createVideoTestBridge(t, db, user.ID)

	bridgeRepo := persistence.NewBridgeRepository(db)
	defectRepo := persistence.NewDefectRepository(db)
	taskRepo := persistence.NewVideoTaskRepository(db)
	frameTaskRepo := persistence.NewVideoFrameTaskRepository(db)
	observationRepo := persistence.NewDefectObservationRepository(db)
	fileService := persistence.NewLocalFileStorage("./uploads")
	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)

	tempDir := t.TempDir()
	frames := []string{createJPEGFrame(t, tempDir, "frame_001.jpg")}

	pythonService := &stubPythonService{}
	videoUC := usecase.NewVideoDetectionUseCaseWithConfig(
		defectService,
		bridgeService,
		pythonService,
		fileService,
		&stubFrameExtractor{frames: frames},
		taskRepo,
		frameTaskRepo,
		observationRepo,
		appconfig.VideoDetectionConfig{
			SampleFPS:                  1,
			MaxQueueInflight:           1,
			QueuedTimeoutSeconds:       1,
			ProgressIdleTimeoutSeconds: 1,
		},
	)
	pythonService.onEnqueue = func(req *service.VideoFrameDetectEnqueueRequest) {
		go func() {
			_, _ = videoUC.HandleAlgoTaskQueued(&dto.AlgoVideoTaskQueuedCallbackRequest{
				TaskID:   req.TaskID,
				Status:   "queued",
				QueuedAt: time.Now().UTC().Format(time.RFC3339),
			})
		}()
	}

	task := &model.VideoAnalysisTask{
		TaskID:     "video_task_progress_idle_timeout",
		SessionID:  "session_progress_idle_timeout",
		UserID:     user.ID,
		BridgeID:   bridge.ID,
		VideoPath:  "videos/test.mp4",
		FPS:        1,
		ModelName:  "baseline",
		PixelRatio: 0.01,
		Status:     model.VideoTaskReadyToProcess,
	}
	if err := taskRepo.Create(task); err != nil {
		t.Fatalf("create task failed: %v", err)
	}

	err := videoUC.StreamTask(task.TaskID, task.SessionID, user, func(any) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "无进度消息超时") {
		t.Fatalf("expected progress idle timeout error, got %v", err)
	}

	storedTask, err := taskRepo.FindByTaskID(task.TaskID)
	if err != nil {
		t.Fatalf("find task failed: %v", err)
	}
	if storedTask.Status != model.VideoTaskFailed {
		t.Fatalf("expected failed task, got %s", storedTask.Status)
	}
}
