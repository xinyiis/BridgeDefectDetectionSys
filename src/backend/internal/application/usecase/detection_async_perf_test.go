package usecase

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/repository"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/persistence"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type perfPythonServiceStub struct {
	detectDelay time.Duration
	defectCount int
}

func (s *perfPythonServiceStub) DetectDefect(imagePath, modelName string, pixelRatio float64) (*service.PythonDetectionResult, error) {
	time.Sleep(s.detectDelay)

	defects := make([]service.DefectDetection, 0, s.defectCount)
	for i := 0; i < s.defectCount; i++ {
		defects = append(defects, service.DefectDetection{
			DefectType: "Crack",
			BBox: service.BBoxData{
				YOLOCoords: [4]float64{0.5, 0.5, 0.2, 0.4},
			},
			Confidence: 0.9,
		})
	}

	return &service.PythonDetectionResult{
		Success:      true,
		TotalDefects: s.defectCount,
		Defects:      defects,
		ResultImage:  base64.StdEncoding.EncodeToString([]byte("perf-result-image")),
	}, nil
}

func (s *perfPythonServiceStub) Detect(string, *service.DetectRequest) (*service.DetectResult, error) {
	return nil, nil
}

func (s *perfPythonServiceStub) Preprocess(string, string) (*service.PreprocessResult, error) {
	return nil, nil
}

func (s *perfPythonServiceStub) Segment(string, *service.SegmentRequest) (*service.SegmentResult, error) {
	return nil, nil
}

func (s *perfPythonServiceStub) EnqueueVideoFrameDetect(*service.VideoFrameDetectEnqueueRequest) (*service.VideoFrameDetectEnqueueResponse, error) {
	return nil, nil
}

type delayedFileService struct {
	base             service.FileService
	saveTempDelay    time.Duration
	persistFileDelay time.Duration
	persistResDelay  time.Duration
}

func (s *delayedFileService) SaveUploadedFile(file *multipart.FileHeader, dir string) (string, error) {
	return s.base.SaveUploadedFile(file, dir)
}

func (s *delayedFileService) SaveImage(file *multipart.FileHeader, dir string) (string, error) {
	return s.base.SaveImage(file, dir)
}

func (s *delayedFileService) SaveTempImage(file *multipart.FileHeader) (string, error) {
	if s.saveTempDelay > 0 {
		time.Sleep(s.saveTempDelay)
	}
	return s.base.SaveTempImage(file)
}

func (s *delayedFileService) AllocateImagePath(dir, ext string) (string, error) {
	return s.base.AllocateImagePath(dir, ext)
}

func (s *delayedFileService) PersistTempFile(tempPath, finalPath string) error {
	if s.persistFileDelay > 0 {
		time.Sleep(s.persistFileDelay)
	}
	return s.base.PersistTempFile(tempPath, finalPath)
}

func (s *delayedFileService) PersistResultImage(base64Data, finalPath string) error {
	if s.persistResDelay > 0 {
		time.Sleep(s.persistResDelay)
	}
	return s.base.PersistResultImage(base64Data, finalPath)
}

func (s *delayedFileService) SaveResultImage(base64Data string, dir string) (string, error) {
	return s.base.SaveResultImage(base64Data, dir)
}

func (s *delayedFileService) DeleteFile(path string) error {
	return s.base.DeleteFile(path)
}

func (s *delayedFileService) ResolvePath(path string) string {
	return s.base.ResolvePath(path)
}

func (s *delayedFileService) ValidateFileFormat(file *multipart.FileHeader, allowedExts []string) error {
	return s.base.ValidateFileFormat(file, allowedExts)
}

func (s *delayedFileService) ValidateFileSize(file *multipart.FileHeader, maxSize int64) error {
	return s.base.ValidateFileSize(file, maxSize)
}

type delayedDefectRepo struct {
	base        repository.DefectRepository
	createDelay time.Duration
	batchDelay  time.Duration
}

func (r *delayedDefectRepo) Create(defect *model.Defect) error {
	if r.createDelay > 0 {
		time.Sleep(r.createDelay)
	}
	return r.base.Create(defect)
}

func (r *delayedDefectRepo) CreateBatch(defects []*model.Defect) error {
	if r.batchDelay > 0 {
		time.Sleep(r.batchDelay)
	}
	return r.base.CreateBatch(defects)
}

func (r *delayedDefectRepo) FindByID(id uint) (*model.Defect, error) {
	return r.base.FindByID(id)
}

func (r *delayedDefectRepo) Delete(id uint) error {
	return r.base.Delete(id)
}

func (r *delayedDefectRepo) List(filters repository.DefectListFilters) ([]model.Defect, int64, error) {
	return r.base.List(filters)
}

func (r *delayedDefectRepo) ListByBridgeID(bridgeID uint, page, pageSize int) ([]model.Defect, int64, error) {
	return r.base.ListByBridgeID(bridgeID, page, pageSize)
}

func (r *delayedDefectRepo) Update(defect *model.Defect) error {
	return r.base.Update(defect)
}

type perfEnv struct {
	db            *gorm.DB
	bridge        *model.Bridge
	user          *model.User
	bridgeService *service.BridgeService
	defectService *service.DefectService
	fileService   service.FileService
	pythonService service.PythonService
}

func setupPerfEnv(t *testing.T, fileDelay, resultDelay, createDelay, batchDelay, detectDelay time.Duration, defectCount int) *perfEnv {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "perf.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	if err := db.AutoMigrate(&model.User{}, &model.Bridge{}, &model.Defect{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	uploadsDir := t.TempDir()
	baseFileService := persistence.NewLocalFileStorage(uploadsDir)
	fileService := &delayedFileService{
		base:             baseFileService,
		saveTempDelay:    10 * time.Millisecond,
		persistFileDelay: fileDelay,
		persistResDelay:  resultDelay,
	}

	bridgeRepo := persistence.NewBridgeRepository(db)
	baseDefectRepo := persistence.NewDefectRepository(db)
	defectRepo := &delayedDefectRepo{
		base:        baseDefectRepo,
		createDelay: createDelay,
		batchDelay:  batchDelay,
	}

	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)
	pythonService := &perfPythonServiceStub{detectDelay: detectDelay, defectCount: defectCount}

	user := &model.User{Username: fmt.Sprintf("perf_%d", time.Now().UnixNano()), Password: "x", RealName: "perf", Email: fmt.Sprintf("perf_%d@test.com", time.Now().UnixNano()), Role: "user"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	bridge := &model.Bridge{
		BridgeName: "性能桥梁",
		BridgeCode: fmt.Sprintf("PERF_%d", time.Now().UnixNano()),
		UserID:     user.ID,
		Address:    "性能地址",
		Longitude:  118.78,
		Latitude:   32.04,
		BridgeType: "梁桥",
	}
	if err := db.Create(bridge).Error; err != nil {
		t.Fatalf("create bridge: %v", err)
	}

	return &perfEnv{
		db:            db,
		bridge:        bridge,
		user:          user,
		bridgeService: bridgeService,
		defectService: defectService,
		fileService:   fileService,
		pythonService: pythonService,
	}
}

func newDetectionUseCaseWithWorkers(env *perfEnv, workers int) *DetectionUseCase {
	uc := &DetectionUseCase{
		defectService:      env.defectService,
		bridgeService:      env.bridgeService,
		pythonService:      env.pythonService,
		fileService:        env.fileService,
		persistenceQueue:   make(chan detectionPersistenceTask, 64),
		persistenceRecords: make(map[string]*detectionPersistenceRecord),
		stopCh:             make(chan struct{}),
	}
	uc.startPersistenceWorkers(workers)
	return uc
}

func createPerfUploadImage(t *testing.T, width, height int) *multipart.FileHeader {
	t.Helper()

	tempDir := t.TempDir()
	imagePath := filepath.Join(tempDir, "perf.png")
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 180, G: 180, B: 180, A: 255})
		}
	}

	file, err := os.Create(imagePath)
	if err != nil {
		t.Fatalf("create sample image: %v", err)
	}
	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		t.Fatalf("encode sample image: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close sample image: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", filepath.Base(imagePath))
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		t.Fatalf("read sample image: %v", err)
	}
	if _, err := part.Write(imageBytes); err != nil {
		t.Fatalf("write form image: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest("POST", "/", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if err := request.ParseMultipartForm(int64(len(body.Bytes()) + 1024)); err != nil {
		t.Fatalf("parse multipart form: %v", err)
	}

	fileHeader := request.MultipartForm.File["image"][0]
	t.Cleanup(func() {
		_ = request.MultipartForm.RemoveAll()
	})
	return fileHeader
}

func runLegacySyncLikeFlow(uc *DetectionUseCase, req *dto.DetectionUploadRequest, currentUser *model.User) (*dto.DetectionResponse, error) {
	startTime := time.Now()

	bridge, err := uc.bridgeService.GetByID(req.BridgeID)
	if err != nil {
		return nil, err
	}
	if bridge == nil {
		return nil, fmt.Errorf("桥梁不存在")
	}
	if !currentUser.IsAdmin() && !bridge.IsOwnedBy(currentUser.ID) {
		return nil, fmt.Errorf("无权访问此桥梁")
	}

	tempImagePath, err := uc.fileService.SaveTempImage(req.Image)
	if err != nil {
		return nil, err
	}
	resolvedImagePath := uc.fileService.ResolvePath(tempImagePath)
	imgW, imgH, err := getImageDimensions(resolvedImagePath)
	if err != nil {
		return nil, err
	}

	pythonResult, err := uc.pythonService.DetectDefect(resolvedImagePath, req.ModelName, req.PixelRatio)
	if err != nil {
		return nil, err
	}

	imagePath, err := uc.fileService.AllocateImagePath("images", ".png")
	if err != nil {
		return nil, err
	}
	resultPath, err := uc.fileService.AllocateImagePath("results", ".jpg")
	if err != nil {
		return nil, err
	}

	defects := make([]*model.Defect, 0, len(pythonResult.Defects))
	for _, detectedDefect := range pythonResult.Defects {
		x, y, w, h, length, width, area := service.CalculatePhysicalDimensions(
			detectedDefect.BBox.YOLOCoords, imgW, imgH, req.PixelRatio,
		)
		detectedDefect.BBox.X = x
		detectedDefect.BBox.Y = y
		detectedDefect.BBox.Width = w
		detectedDefect.BBox.Height = h
		detectedDefect.Length = length
		detectedDefect.Width = width
		detectedDefect.Area = area

		defects = append(defects, &model.Defect{
			BridgeID:   req.BridgeID,
			DefectType: translateAlgorithmDefectType(detectedDefect.DefectType),
			ImagePath:  imagePath,
			ResultPath: resultPath,
			BBox:       detectedDefect.BBoxJSON(),
			Length:     detectedDefect.Length,
			Width:      detectedDefect.Width,
			Area:       detectedDefect.Area,
			Confidence: detectedDefect.Confidence,
			DetectedAt: time.Now(),
		})
	}

	if err := uc.fileService.PersistTempFile(tempImagePath, imagePath); err != nil {
		return nil, err
	}
	if err := uc.fileService.PersistResultImage(pythonResult.ResultImage, resultPath); err != nil {
		return nil, err
	}
	for _, defect := range defects {
		if err := uc.defectService.CreateDefect(defect); err != nil {
			return nil, err
		}
	}
	_ = uc.fileService.DeleteFile(tempImagePath)

	return &dto.DetectionResponse{
		TotalDefects:      len(defects),
		ImagePath:         imagePath,
		ResultPath:        resultPath,
		ProcessingTime:    time.Since(startTime).Seconds(),
		PersistenceStatus: detectionPersistenceStatusCompleted,
		Defects:           uc.toDefectDTOs(defects),
		DefectSummary:     uc.buildDefectSummary(defects),
	}, nil
}

func averageDuration(ds []time.Duration) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	var total time.Duration
	for _, d := range ds {
		total += d
	}
	return total / time.Duration(len(ds))
}

func TestDetectionAsyncPerformanceComparison(t *testing.T) {
	env := setupPerfEnv(
		t,
		90*time.Millisecond, 110*time.Millisecond,
		35*time.Millisecond, 45*time.Millisecond,
		100*time.Millisecond, 3,
	)

	asyncUC := newDetectionUseCaseWithWorkers(env, 2)
	defer asyncUC.Shutdown()

	syncUC := &DetectionUseCase{
		defectService: env.defectService,
		bridgeService: env.bridgeService,
		pythonService: env.pythonService,
		fileService:   env.fileService,
	}

	const runs = 8
	asyncDurations := make([]time.Duration, 0, runs)
	syncDurations := make([]time.Duration, 0, runs)
	taskIDs := make([]string, 0, runs)

	for i := 0; i < runs; i++ {
		req := &dto.DetectionUploadRequest{
			Image:      createPerfUploadImage(t, 128, 72),
			BridgeID:   env.bridge.ID,
			ModelName:  "perf",
			PixelRatio: 0.01,
		}

		start := time.Now()
		asyncResp, err := asyncUC.UploadAndDetect(req, env.user)
		if err != nil {
			t.Fatalf("async upload failed: %v", err)
		}
		asyncDurations = append(asyncDurations, time.Since(start))
		taskIDs = append(taskIDs, asyncResp.PersistenceTaskID)

		start = time.Now()
		if _, err := runLegacySyncLikeFlow(syncUC, req, env.user); err != nil {
			t.Fatalf("sync baseline failed: %v", err)
		}
		syncDurations = append(syncDurations, time.Since(start))
	}

	for _, taskID := range taskIDs {
		if _, err := asyncUC.WaitForPersistenceTask(taskID, 5*time.Second); err != nil {
			t.Fatalf("wait for async persistence task %s: %v", taskID, err)
		}
	}

	asyncAvg := averageDuration(asyncDurations)
	syncAvg := averageDuration(syncDurations)
	improvement := (float64(syncAvg-asyncAvg) / float64(syncAvg)) * 100

	t.Logf("async avg response: %v", asyncAvg)
	t.Logf("sync  avg response: %v", syncAvg)
	t.Logf("response latency improvement: %.1f%%", improvement)

	if asyncAvg >= syncAvg {
		t.Fatalf("expected async response to be faster than sync: async=%v sync=%v", asyncAvg, syncAvg)
	}
	if improvement < 35 {
		t.Fatalf("expected substantial improvement >=35%%, got %.1f%%", improvement)
	}
}

func TestDetectionAsyncWorkerSizing(t *testing.T) {
	type result struct {
		workers         int
		responseWindow  time.Duration
		persistenceDone time.Duration
	}

	runScenario := func(workers int) result {
		env := setupPerfEnv(
			t,
			120*time.Millisecond, 140*time.Millisecond,
			40*time.Millisecond, 60*time.Millisecond,
			90*time.Millisecond, 3,
		)
		uc := newDetectionUseCaseWithWorkers(env, workers)
		defer uc.Shutdown()

		const concurrentRequests = 12
		taskIDs := make([]string, concurrentRequests)
		start := time.Now()
		var wg sync.WaitGroup
		errCh := make(chan error, concurrentRequests)

		for i := 0; i < concurrentRequests; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				req := &dto.DetectionUploadRequest{
					Image:      createPerfUploadImage(t, 160, 90),
					BridgeID:   env.bridge.ID,
					ModelName:  "perf",
					PixelRatio: 0.01,
				}
				resp, err := uc.UploadAndDetect(req, env.user)
				if err != nil {
					errCh <- err
					return
				}
				taskIDs[idx] = resp.PersistenceTaskID
			}(i)
		}
		wg.Wait()
		close(errCh)
		for err := range errCh {
			if err != nil {
				t.Fatalf("workers=%d upload failed: %v", workers, err)
			}
		}

		responseWindow := time.Since(start)
		for _, taskID := range taskIDs {
			if _, err := uc.WaitForPersistenceTask(taskID, 10*time.Second); err != nil {
				t.Fatalf("workers=%d wait task %s: %v", workers, taskID, err)
			}
		}

		return result{
			workers:         workers,
			responseWindow:  responseWindow,
			persistenceDone: time.Since(start),
		}
	}

	one := runScenario(1)
	two := runScenario(2)
	four := runScenario(4)

	t.Logf("workers=%d response=%v persistence_done=%v", one.workers, one.responseWindow, one.persistenceDone)
	t.Logf("workers=%d response=%v persistence_done=%v", two.workers, two.responseWindow, two.persistenceDone)
	t.Logf("workers=%d response=%v persistence_done=%v", four.workers, four.responseWindow, four.persistenceDone)

	if !(two.persistenceDone < one.persistenceDone) {
		t.Fatalf("expected 2 workers to drain faster than 1 worker")
	}
	if !(four.persistenceDone < two.persistenceDone) {
		t.Fatalf("expected 4 workers to drain faster than 2 workers")
	}
	if !(two.persistenceDone <= one.persistenceDone*65/100) {
		t.Fatalf("expected 2 workers to materially reduce drain time vs 1: one=%v two=%v", one.persistenceDone, two.persistenceDone)
	}
}
