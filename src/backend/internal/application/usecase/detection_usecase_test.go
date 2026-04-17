package usecase_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/dto"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/persistence"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type detectionPythonServiceStub struct {
	receivedImagePath        string
	receivedSegmentModelType string
	result                   *service.PythonDetectionResult
}

func (s *detectionPythonServiceStub) DetectDefect(imagePath, modelName, segmentModelType string, pixelRatio float64) (*service.PythonDetectionResult, error) {
	s.receivedImagePath = imagePath
	s.receivedSegmentModelType = segmentModelType
	if s.result != nil {
		return s.result, nil
	}
	return &service.PythonDetectionResult{
		Success:      true,
		TotalDefects: 1,
		Defects: []service.DefectDetection{
			{
				DefectType: "Crack",
				BBox: service.BBoxData{
					YOLOCoords: [4]float64{0.5, 0.5, 0.2, 0.4},
				},
				Confidence: 0.88,
			},
		},
		ResultImage: base64.StdEncoding.EncodeToString([]byte("segmented-image")),
	}, nil
}

func (s *detectionPythonServiceStub) Detect(string, *service.DetectRequest) (*service.DetectResult, error) {
	return nil, nil
}

func (s *detectionPythonServiceStub) Preprocess(string, string) (*service.PreprocessResult, error) {
	return nil, nil
}

func (s *detectionPythonServiceStub) Segment(string, *service.SegmentRequest) (*service.SegmentResult, error) {
	return nil, nil
}

func (s *detectionPythonServiceStub) EnqueueVideoFrameDetect(*service.VideoFrameDetectEnqueueRequest) (*service.VideoFrameDetectEnqueueResponse, error) {
	return nil, nil
}

func TestDetectionUseCaseUploadAndDetectCalculatesPhysicalDimensions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Bridge{}, &model.Defect{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	bridgeRepo := persistence.NewBridgeRepository(db)
	defectRepo := persistence.NewDefectRepository(db)

	uploadsDir := t.TempDir()
	fileService := persistence.NewLocalFileStorage(uploadsDir)
	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)
	pythonService := &detectionPythonServiceStub{}

	detectionUseCase := usecase.NewDetectionUseCase(defectService, bridgeService, pythonService, fileService)
	t.Cleanup(detectionUseCase.Shutdown)

	currentUser := &model.User{ID: 1, Role: "user"}
	if err := db.Create(currentUser).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	bridge := &model.Bridge{
		BridgeName: "测试桥梁",
		BridgeCode: "BRIDGE_TEST_001",
		UserID:     currentUser.ID,
		Address:    "测试地址",
		Longitude:  118.78,
		Latitude:   32.04,
		BridgeType: "梁桥",
	}
	if err := db.Create(bridge).Error; err != nil {
		t.Fatalf("create bridge: %v", err)
	}

	fileHeader := createDetectionUploadImage(t, 100, 50)
	result, err := detectionUseCase.UploadAndDetect(&dto.DetectionUploadRequest{
		Image:      fileHeader,
		BridgeID:   bridge.ID,
		ModelName:  "baseline",
		PixelRatio: 0.1,
	}, currentUser)
	if err != nil {
		t.Fatalf("upload and detect failed: %v", err)
	}

	if result.TotalDefects != 1 {
		t.Fatalf("unexpected total defects: %d", result.TotalDefects)
	}
	if result.PersistenceStatus != "pending" {
		t.Fatalf("expected pending persistence status, got %s", result.PersistenceStatus)
	}
	if result.PersistenceTaskID == "" {
		t.Fatalf("expected persistence task id")
	}
	if len(result.DefectSummary) != 1 || result.DefectSummary["裂缝"] != 1 {
		t.Fatalf("unexpected defect summary: %#v", result.DefectSummary)
	}
	if pythonService.receivedImagePath == "" {
		t.Fatalf("python service did not receive image path")
	}
	if pythonService.receivedSegmentModelType != "student" {
		t.Fatalf("expected default segment model type student, got %q", pythonService.receivedSegmentModelType)
	}
	if _, err := os.Stat(pythonService.receivedImagePath); err != nil {
		t.Fatalf("expected resolved image path to exist, got err: %v", err)
	}
	if strings.HasPrefix(pythonService.receivedImagePath, "images/") {
		t.Fatalf("expected resolved file path, got stored relative path: %s", pythonService.receivedImagePath)
	}
	if !strings.HasPrefix(result.ImagePath, "images/") {
		t.Fatalf("expected stored image path, got: %s", result.ImagePath)
	}
	if !strings.HasPrefix(result.ResultPath, "results/") {
		t.Fatalf("expected stored result path, got: %s", result.ResultPath)
	}
	expectedResultImageBase64 := base64.StdEncoding.EncodeToString([]byte("segmented-image"))
	if result.ResultImageBase64 != expectedResultImageBase64 {
		t.Fatalf("unexpected result image base64: %q", result.ResultImageBase64)
	}
	if result.Defects[0].ID != 0 {
		t.Fatalf("expected async defect id to be 0 before persistence, got %d", result.Defects[0].ID)
	}

	var bbox service.BBoxData
	if err := json.Unmarshal([]byte(result.Defects[0].BBox), &bbox); err != nil {
		t.Fatalf("unmarshal bbox json: %v", err)
	}

	if result.Defects[0].DefectType != "裂缝" {
		t.Fatalf("unexpected defect type: %s", result.Defects[0].DefectType)
	}
	if bbox.X != 40 || bbox.Y != 15 || bbox.Width != 20 || bbox.Height != 20 {
		t.Fatalf("unexpected bbox pixels: %#v", bbox)
	}
	if bbox.YOLOCoords != [4]float64{0.5, 0.5, 0.2, 0.4} {
		t.Fatalf("unexpected bbox yolo coords: %#v", bbox.YOLOCoords)
	}
	if result.Defects[0].Length != 0 || result.Defects[0].Width != 0 || result.Defects[0].Area != 4 {
		t.Fatalf("unexpected physical metrics: length=%v width=%v area=%v", result.Defects[0].Length, result.Defects[0].Width, result.Defects[0].Area)
	}

	status, err := detectionUseCase.WaitForPersistenceTask(result.PersistenceTaskID, 2*time.Second)
	if err != nil {
		t.Fatalf("wait for persistence task: %v", err)
	}
	if status.Status != "completed" {
		t.Fatalf("expected completed persistence status, got %s (err=%s)", status.Status, status.ErrorMessage)
	}
	if len(status.SavedDefectIDs) != 1 || status.SavedDefectIDs[0] == 0 {
		t.Fatalf("unexpected saved defect ids: %#v", status.SavedDefectIDs)
	}

	if _, err := os.Stat(filepath.Join(uploadsDir, result.ImagePath)); err != nil {
		t.Fatalf("expected persisted image to exist, got err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(uploadsDir, result.ResultPath)); err != nil {
		t.Fatalf("expected persisted result image to exist, got err: %v", err)
	}

	var dbDefects []model.Defect
	if err := db.Find(&dbDefects).Error; err != nil {
		t.Fatalf("list persisted defects: %v", err)
	}
	if len(dbDefects) != 1 {
		t.Fatalf("expected 1 persisted defect, got %d", len(dbDefects))
	}
}

func TestDetectionUseCaseUploadAndDetectPreservesAlgorithmMeasurements(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Bridge{}, &model.Defect{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	bridgeRepo := persistence.NewBridgeRepository(db)
	defectRepo := persistence.NewDefectRepository(db)

	uploadsDir := t.TempDir()
	fileService := persistence.NewLocalFileStorage(uploadsDir)
	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)
	pythonService := &detectionPythonServiceStub{
		result: &service.PythonDetectionResult{
			Success:      true,
			TotalDefects: 1,
			Defects: []service.DefectDetection{
				{
					DefectType: "Crack",
					BBox: service.BBoxData{
						YOLOCoords: [4]float64{0.5, 0.5, 0.2, 0.4},
					},
					Area:       0.4502,
					Confidence: 0.91,
				},
			},
			ResultImage: base64.StdEncoding.EncodeToString([]byte("segmented-image")),
		},
	}

	detectionUseCase := usecase.NewDetectionUseCase(defectService, bridgeService, pythonService, fileService)
	t.Cleanup(detectionUseCase.Shutdown)

	currentUser := &model.User{ID: 1, Role: "user"}
	if err := db.Create(currentUser).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	bridge := &model.Bridge{
		BridgeName: "测试桥梁",
		BridgeCode: "BRIDGE_TEST_002",
		UserID:     currentUser.ID,
		Address:    "测试地址",
		Longitude:  118.78,
		Latitude:   32.04,
		BridgeType: "梁桥",
	}
	if err := db.Create(bridge).Error; err != nil {
		t.Fatalf("create bridge: %v", err)
	}

	fileHeader := createDetectionUploadImage(t, 100, 50)
	result, err := detectionUseCase.UploadAndDetect(&dto.DetectionUploadRequest{
		Image:      fileHeader,
		BridgeID:   bridge.ID,
		ModelName:  "baseline",
		PixelRatio: 0.1,
	}, currentUser)
	if err != nil {
		t.Fatalf("upload and detect failed: %v", err)
	}

	var bbox service.BBoxData
	if err := json.Unmarshal([]byte(result.Defects[0].BBox), &bbox); err != nil {
		t.Fatalf("unmarshal bbox json: %v", err)
	}

	if bbox.X != 40 || bbox.Y != 15 || bbox.Width != 20 || bbox.Height != 20 {
		t.Fatalf("unexpected bbox pixels: %#v", bbox)
	}
	if result.Defects[0].Area != 0.4502 {
		t.Fatalf("expected algorithm area to be preserved, got %v", result.Defects[0].Area)
	}
	if result.Defects[0].Length != 0 || result.Defects[0].Width != 0 {
		t.Fatalf("expected length/width to remain algorithm-provided values, got length=%v width=%v", result.Defects[0].Length, result.Defects[0].Width)
	}
}

func TestNewDetectionUseCaseWithPersistenceWorkersUsesConfiguredValue(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	bridgeRepo := persistence.NewBridgeRepository(db)
	defectRepo := persistence.NewDefectRepository(db)
	fileService := persistence.NewLocalFileStorage(t.TempDir())
	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)
	pythonService := &detectionPythonServiceStub{}

	uc := usecase.NewDetectionUseCaseWithPersistenceWorkers(defectService, bridgeService, pythonService, fileService, 4)
	defer uc.Shutdown()

	if uc.PersistenceWorkerCount() != 4 {
		t.Fatalf("expected configured worker count 4, got %d", uc.PersistenceWorkerCount())
	}
}

func TestNewDetectionUseCaseWithPersistenceWorkersFallsBackToDefault(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	bridgeRepo := persistence.NewBridgeRepository(db)
	defectRepo := persistence.NewDefectRepository(db)
	fileService := persistence.NewLocalFileStorage(t.TempDir())
	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)
	pythonService := &detectionPythonServiceStub{}

	uc := usecase.NewDetectionUseCaseWithPersistenceWorkers(defectService, bridgeService, pythonService, fileService, 0)
	defer uc.Shutdown()

	if uc.PersistenceWorkerCount() != 2 {
		t.Fatalf("expected default worker count 2, got %d", uc.PersistenceWorkerCount())
	}
}

func createDetectionUploadImage(t *testing.T, width, height int) *multipart.FileHeader {
	t.Helper()

	tempDir := t.TempDir()
	imagePath := filepath.Join(tempDir, "sample.png")

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 200, B: 200, A: 255})
		}
	}

	file, err := os.Create(imagePath)
	if err != nil {
		t.Fatalf("create sample image: %v", err)
	}
	if err := png.Encode(file, img); err != nil {
		file.Close()
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
