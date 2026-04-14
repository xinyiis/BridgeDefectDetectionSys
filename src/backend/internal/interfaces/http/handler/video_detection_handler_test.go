package handler_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/application/usecase"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/persistence"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/interfaces/http/handler"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/interfaces/http/middleware"
	"golang.org/x/net/websocket"
	"gorm.io/gorm"
)

type stubFrameExtractor struct{}

func (s *stubFrameExtractor) CountFrames(videoPath string, fps float64) (int, error) {
	return 0, nil
}

func (s *stubFrameExtractor) ExtractFramesStream(videoPath, outputDir string, fps float64, onFrame func(string, int) error) error {
	return nil
}

type sequenceFrameExtractor struct {
	frames []string
}

func (s *sequenceFrameExtractor) CountFrames(videoPath string, fps float64) (int, error) {
	return len(s.frames), nil
}

func (s *sequenceFrameExtractor) ExtractFramesStream(videoPath, outputDir string, fps float64, onFrame func(string, int) error) error {
	for idx, frame := range s.frames {
		if err := onFrame(frame, idx+1); err != nil {
			return err
		}
	}
	return nil
}

type stubPythonService struct{}

func (s *stubPythonService) DetectDefect(string, string, float64) (*service.PythonDetectionResult, error) {
	return nil, nil
}

func (s *stubPythonService) Detect(string, *service.DetectRequest) (*service.DetectResult, error) {
	return &service.DetectResult{
		Status:      "success",
		ModelUsed:   "baseline",
		YOLOBBoxes:  []service.BBoxItem{},
		ImageResult: base64.StdEncoding.EncodeToString([]byte("mock")),
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
	return &service.VideoFrameDetectEnqueueResponse{
		Status:    "accepted",
		RequestID: req.RequestID,
		TaskID:    req.TaskID,
	}, nil
}

type streamingPythonService struct{}

func (s *streamingPythonService) DetectDefect(string, string, float64) (*service.PythonDetectionResult, error) {
	return nil, nil
}

func (s *streamingPythonService) Detect(string, *service.DetectRequest) (*service.DetectResult, error) {
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
		ImageResult: base64.StdEncoding.EncodeToString([]byte("streaming-result")),
	}, nil
}

func (s *streamingPythonService) Preprocess(string, string) (*service.PreprocessResult, error) {
	return &service.PreprocessResult{Status: "success"}, nil
}

func (s *streamingPythonService) Segment(string, *service.SegmentRequest) (*service.SegmentResult, error) {
	return &service.SegmentResult{Status: "success"}, nil
}

func (s *streamingPythonService) EnqueueVideoFrameDetect(req *service.VideoFrameDetectEnqueueRequest) (*service.VideoFrameDetectEnqueueResponse, error) {
	if req == nil {
		return nil, nil
	}
	return &service.VideoFrameDetectEnqueueResponse{
		Status:    "accepted",
		RequestID: req.RequestID,
		TaskID:    req.TaskID,
	}, nil
}

func setupVideoTaskRouter(db *gorm.DB) *gin.Engine {
	return setupVideoTaskRouterWithDeps(db, &stubFrameExtractor{}, &stubPythonService{})
}

func setupVideoTaskRouterWithDeps(db *gorm.DB, extractor interface {
	CountFrames(string, float64) (int, error)
	ExtractFramesStream(string, string, float64, func(string, int) error) error
}, pythonService service.PythonService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	store := cookie.NewStore([]byte("test-secret-key"))
	r.Use(sessions.Sessions("test_session", store))

	userRepo := persistence.NewUserRepository(db)
	bridgeRepo := persistence.NewBridgeRepository(db)
	defectRepo := persistence.NewDefectRepository(db)
	taskRepo := persistence.NewVideoTaskRepository(db)
	frameTaskRepo := persistence.NewVideoFrameTaskRepository(db)
	observationRepo := persistence.NewDefectObservationRepository(db)
	fileService := persistence.NewLocalFileStorage("./test_uploads")

	userService := service.NewUserService(userRepo)
	bridgeService := service.NewBridgeService(db, bridgeRepo, fileService)
	defectService := service.NewDefectService(db, defectRepo, bridgeRepo)

	authUseCase := usecase.NewAuthUseCase(userService)
	videoUseCase := usecase.NewVideoDetectionUseCase(
		defectService,
		bridgeService,
		pythonService,
		fileService,
		extractor,
		taskRepo,
		frameTaskRepo,
		observationRepo,
	)

	authHandler := handler.NewAuthHandler(authUseCase)
	videoHandler := handler.NewVideoDetectionHandler(videoUseCase)
	videoStreamHandler := handler.NewVideoStreamHandler(videoUseCase)
	videoCallbackHandler := handler.NewVideoCallbackHandler(videoUseCase)

	api := r.Group("/api/v1")
	api.POST("/auth/login", authHandler.Login)

	auth := api.Group("")
	auth.Use(middleware.AuthRequired(db))
	{
		detection := auth.Group("/detection")
		detection.POST("/video/upload", videoHandler.UploadVideo)
		detection.POST("/video/:task_id/start", videoHandler.StartTask)
		detection.GET("/video/:task_id", videoHandler.GetTask)
		detection.POST("/video/:task_id/cancel", videoHandler.CancelTask)
		detection.GET("/video/ws", videoStreamHandler.Stream)
	}

	callback := api.Group("/detection/video/callback")
	{
		callback.POST("/frame-queued", videoCallbackHandler.TaskQueued)
		callback.POST("/frame-started", videoCallbackHandler.TaskStarted)
		callback.POST("/task-queued", videoCallbackHandler.TaskQueued)
		callback.POST("/task-started", videoCallbackHandler.TaskStarted)
		callback.POST("/frame-result", videoCallbackHandler.FrameResult)
		callback.POST("/task-completed", videoCallbackHandler.TaskCompleted)
		callback.POST("/task-failed", videoCallbackHandler.TaskFailed)
	}

	return r
}

func createTestVideoFile(t *testing.T) string {
	t.Helper()

	dir := "./test_uploads/videos"
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	path := filepath.Join(dir, "test_video.mp4")
	if err := os.WriteFile(path, []byte("fake-video-content"), 0644); err != nil {
		t.Fatalf("write video failed: %v", err)
	}

	return path
}

func createJPEGFrame(t *testing.T, dir string, name string) string {
	t.Helper()

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir frame dir failed: %v", err)
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
			img.Set(x, y, color.RGBA{R: 220, G: 220, B: 220, A: 255})
		}
	}

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode frame failed: %v", err)
	}

	return framePath
}

func cookiesToHeader(cookies []*http.Cookie) string {
	parts := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		parts = append(parts, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(parts, "; ")
}

func TestVideoTaskLifecycleHandlers(t *testing.T) {
	db := setupDetectionTestDB(t)
	if err := db.AutoMigrate(&model.VideoAnalysisTask{}, &model.DefectObservation{}, &model.VideoFrameTaskRequest{}); err != nil {
		t.Fatalf("migrate video models failed: %v", err)
	}

	router := setupVideoTaskRouter(db)
	defer cleanupTestFiles()

	user := createDetectionTestUser(db, "video_handler_user", "user")
	bridge := createDetectionTestBridge(db, user.ID, "视频桥梁")
	cookies := loginDetectionTest(t, router, "video_handler_user", "123456")

	videoPath := createTestVideoFile(t)
	defer os.Remove(videoPath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(videoPath)
	if err != nil {
		t.Fatalf("open video failed: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("video", "test_video.mp4")
	if err != nil {
		t.Fatalf("create form file failed: %v", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		t.Fatalf("copy video failed: %v", err)
	}

	if err := writer.WriteField("bridge_id", fmt.Sprintf("%d", bridge.ID)); err != nil {
		t.Fatalf("write bridge_id failed: %v", err)
	}
	if err := writer.WriteField("model_name", "baseline"); err != nil {
		t.Fatalf("write model_name failed: %v", err)
	}
	if err := writer.WriteField("pixel_ratio", "0.01"); err != nil {
		t.Fatalf("write pixel_ratio failed: %v", err)
	}
	if err := writer.WriteField("fps", "1"); err != nil {
		t.Fatalf("write fps failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer failed: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/detection/video/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected upload 200, got %d: %s", w.Code, w.Body.String())
	}

	var uploadResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &uploadResp); err != nil {
		t.Fatalf("decode upload response failed: %v", err)
	}
	data := uploadResp["data"].(map[string]any)
	taskID := data["task_id"].(string)

	startReq := httptest.NewRequest("POST", "/api/v1/detection/video/"+taskID+"/start", nil)
	for _, cookie := range cookies {
		startReq.AddCookie(cookie)
	}
	startW := httptest.NewRecorder()
	router.ServeHTTP(startW, startReq)
	if startW.Code != http.StatusOK {
		t.Fatalf("expected start 200, got %d: %s", startW.Code, startW.Body.String())
	}

	getReq := httptest.NewRequest("GET", "/api/v1/detection/video/"+taskID, nil)
	for _, cookie := range cookies {
		getReq.AddCookie(cookie)
	}
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("expected get 200, got %d: %s", getW.Code, getW.Body.String())
	}

	var getResp map[string]any
	if err := json.Unmarshal(getW.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("decode get response failed: %v", err)
	}
	getData := getResp["data"].(map[string]any)
	if getData["status"].(string) != string(model.VideoTaskReadyToProcess) {
		t.Fatalf("expected ready_to_process, got %s", getData["status"].(string))
	}

	cancelReq := httptest.NewRequest("POST", "/api/v1/detection/video/"+taskID+"/cancel", nil)
	for _, cookie := range cookies {
		cancelReq.AddCookie(cookie)
	}
	cancelW := httptest.NewRecorder()
	router.ServeHTTP(cancelW, cancelReq)
	if cancelW.Code != http.StatusOK {
		t.Fatalf("expected cancel 200, got %d: %s", cancelW.Code, cancelW.Body.String())
	}

	var task model.VideoAnalysisTask
	if err := db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		t.Fatalf("load task failed: %v", err)
	}
	if task.Status != model.VideoTaskCanceled {
		t.Fatalf("expected canceled task, got %s", task.Status)
	}
}

func TestUploadVideoRejectsInvalidExtension(t *testing.T) {
	db := setupDetectionTestDB(t)
	if err := db.AutoMigrate(&model.VideoAnalysisTask{}, &model.DefectObservation{}, &model.VideoFrameTaskRequest{}); err != nil {
		t.Fatalf("migrate video models failed: %v", err)
	}

	router := setupVideoTaskRouter(db)
	defer cleanupTestFiles()

	user := createDetectionTestUser(db, "video_invalid_user", "user")
	bridge := createDetectionTestBridge(db, user.ID, "视频桥梁")
	cookies := loginDetectionTest(t, router, "video_invalid_user", "123456")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("video", "invalid.txt")
	if err != nil {
		t.Fatalf("create form file failed: %v", err)
	}
	if _, err := part.Write([]byte("not a video")); err != nil {
		t.Fatalf("write invalid file failed: %v", err)
	}
	_ = writer.WriteField("bridge_id", fmt.Sprintf("%d", bridge.ID))
	_ = writer.WriteField("pixel_ratio", "0.01")
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/detection/video/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUploadVideoRejectsEnableSegment(t *testing.T) {
	db := setupDetectionTestDB(t)
	if err := db.AutoMigrate(&model.VideoAnalysisTask{}, &model.DefectObservation{}, &model.VideoFrameTaskRequest{}); err != nil {
		t.Fatalf("migrate video models failed: %v", err)
	}

	router := setupVideoTaskRouter(db)
	defer cleanupTestFiles()

	user := createDetectionTestUser(db, "video_enable_segment_user", "user")
	bridge := createDetectionTestBridge(db, user.ID, "视频桥梁")
	cookies := loginDetectionTest(t, router, "video_enable_segment_user", "123456")

	videoPath := createTestVideoFile(t)
	defer os.Remove(videoPath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	file, err := os.Open(videoPath)
	if err != nil {
		t.Fatalf("open video failed: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("video", "test_video.mp4")
	if err != nil {
		t.Fatalf("create form file failed: %v", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		t.Fatalf("copy video failed: %v", err)
	}
	_ = writer.WriteField("bridge_id", fmt.Sprintf("%d", bridge.ID))
	_ = writer.WriteField("pixel_ratio", "0.01")
	_ = writer.WriteField("enable_segment", "true")
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer failed: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/detection/video/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVideoUserFlowE2E_WithSyntheticVideoAndWebSocket(t *testing.T) {
	db := setupDetectionTestDB(t)
	if err := db.AutoMigrate(&model.VideoAnalysisTask{}, &model.DefectObservation{}, &model.VideoFrameTaskRequest{}); err != nil {
		t.Fatalf("migrate video models failed: %v", err)
	}

	frameDir := t.TempDir()
	extractor := &sequenceFrameExtractor{
		frames: []string{
			createJPEGFrame(t, frameDir, "frame_001.jpg"),
			createJPEGFrame(t, frameDir, "frame_002.jpg"),
			createJPEGFrame(t, frameDir, "frame_003.jpg"),
		},
	}

	router := setupVideoTaskRouterWithDeps(db, extractor, &streamingPythonService{})
	defer cleanupTestFiles()

	user := createDetectionTestUser(db, "video_e2e_user", "user")
	bridge := createDetectionTestBridge(db, user.ID, "视频黑盒桥梁")
	cookies := loginDetectionTest(t, router, "video_e2e_user", "123456")

	videoPath := createTestVideoFile(t)
	defer os.Remove(videoPath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	file, err := os.Open(videoPath)
	if err != nil {
		t.Fatalf("open video failed: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("video", "synthetic.mp4")
	if err != nil {
		t.Fatalf("create form file failed: %v", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		t.Fatalf("copy video failed: %v", err)
	}
	_ = writer.WriteField("bridge_id", fmt.Sprintf("%d", bridge.ID))
	_ = writer.WriteField("model_name", "baseline")
	_ = writer.WriteField("pixel_ratio", "0.01")
	_ = writer.WriteField("fps", "1")
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer failed: %v", err)
	}

	uploadReq := httptest.NewRequest("POST", "/api/v1/detection/video/upload", body)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cookies {
		uploadReq.AddCookie(cookie)
	}
	uploadW := httptest.NewRecorder()
	router.ServeHTTP(uploadW, uploadReq)
	if uploadW.Code != http.StatusOK {
		t.Fatalf("upload failed: %d %s", uploadW.Code, uploadW.Body.String())
	}

	var uploadResp map[string]any
	if err := json.Unmarshal(uploadW.Body.Bytes(), &uploadResp); err != nil {
		t.Fatalf("decode upload response failed: %v", err)
	}
	uploadData := uploadResp["data"].(map[string]any)
	taskID := uploadData["task_id"].(string)

	startReq := httptest.NewRequest("POST", "/api/v1/detection/video/"+taskID+"/start", nil)
	for _, cookie := range cookies {
		startReq.AddCookie(cookie)
	}
	startW := httptest.NewRecorder()
	router.ServeHTTP(startW, startReq)
	if startW.Code != http.StatusOK {
		t.Fatalf("start failed: %d %s", startW.Code, startW.Body.String())
	}

	var startResp map[string]any
	if err := json.Unmarshal(startW.Body.Bytes(), &startResp); err != nil {
		t.Fatalf("decode start response failed: %v", err)
	}
	startData := startResp["data"].(map[string]any)
	sessionID := startData["session_id"].(string)

	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/detection/video/ws?task_id=" + url.QueryEscape(taskID) + "&session_id=" + url.QueryEscape(sessionID)
	config, err := websocket.NewConfig(wsURL, server.URL)
	if err != nil {
		t.Fatalf("new websocket config failed: %v", err)
	}
	config.Header = http.Header{}
	config.Header.Set("Cookie", cookiesToHeader(cookies))

	conn, err := websocket.DialConfig(config)
	if err != nil {
		t.Fatalf("dial websocket failed: %v", err)
	}
	defer conn.Close()

	seenConnected := false
	seenFrame := false
	seenProgress := false
	seenCompleted := false

	for !seenCompleted {
		var message map[string]any
		if err := websocket.JSON.Receive(conn, &message); err != nil {
			t.Fatalf("receive websocket message failed: %v", err)
		}

		switch message["type"] {
		case "connected":
			seenConnected = true
		case "frame":
			seenFrame = true
			if int(message["defect_count"].(float64)) != 1 {
				t.Fatalf("expected 1 defect in frame, got %v", message["defect_count"])
			}
		case "progress":
			seenProgress = true
		case "completed":
			seenCompleted = true
			if int(message["total_defects"].(float64)) != 1 {
				t.Fatalf("expected 1 total defect, got %v", message["total_defects"])
			}
		case "error":
			t.Fatalf("unexpected websocket error: %v", message["message"])
		}
	}

	if !seenConnected || !seenFrame || !seenProgress || !seenCompleted {
		t.Fatalf("missing websocket lifecycle messages: connected=%v frame=%v progress=%v completed=%v", seenConnected, seenFrame, seenProgress, seenCompleted)
	}

	var task model.VideoAnalysisTask
	if err := db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		t.Fatalf("load task failed: %v", err)
	}
	if task.Status != model.VideoTaskCompleted {
		t.Fatalf("expected completed task, got %s", task.Status)
	}
	if task.ConfirmedDefects != 1 {
		t.Fatalf("expected 1 confirmed defect, got %d", task.ConfirmedDefects)
	}

	var defects []model.Defect
	if err := db.Find(&defects).Error; err != nil {
		t.Fatalf("list defects failed: %v", err)
	}
	if len(defects) != 1 {
		t.Fatalf("expected 1 defect record, got %d", len(defects))
	}
	if defects[0].SourceType != "video" {
		t.Fatalf("expected video source type, got %s", defects[0].SourceType)
	}

	var observationCount int64
	if err := db.Model(&model.DefectObservation{}).Count(&observationCount).Error; err != nil {
		t.Fatalf("count observations failed: %v", err)
	}
	if observationCount != 3 {
		t.Fatalf("expected 3 observations, got %d", observationCount)
	}
}

func TestVideoCallbackFrameResultPersistsObservation(t *testing.T) {
	db := setupDetectionTestDB(t)
	if err := db.AutoMigrate(&model.VideoAnalysisTask{}, &model.DefectObservation{}, &model.VideoFrameTaskRequest{}); err != nil {
		t.Fatalf("migrate video callback models failed: %v", err)
	}

	router := setupVideoTaskRouter(db)
	defer cleanupTestFiles()

	user := createDetectionTestUser(db, "video_callback_user", "user")
	bridge := createDetectionTestBridge(db, user.ID, "回调桥梁")
	task := &model.VideoAnalysisTask{
		TaskID:      "video_task_callback_001",
		UserID:      user.ID,
		BridgeID:    bridge.ID,
		VideoPath:   "videos/callback.mp4",
		FPS:         1,
		ModelName:   "baseline",
		PixelRatio:  0.01,
		Status:      model.VideoTaskDispatching,
		TotalFrames: 3,
	}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("create task failed: %v", err)
	}

	frameDir := t.TempDir()
	framePath := createJPEGFrame(t, frameDir, "frame_045.jpg")

	payload := map[string]any{
		"request_id":       "frame_req_45",
		"task_id":          task.TaskID,
		"frame_no":         45,
		"timestamp_ms":     45000,
		"status":           "success",
		"queue_latency_ms": 180,
		"detect_total_ms":  1500,
		"decode_ms":        120,
		"infer_ms":         1080,
		"postprocess_ms":   300,
		"frame_ref":        "file://" + framePath,
		"yolo_bboxes": []map[string]any{
			{
				"box_id":      0,
				"class_idx":   0,
				"class_name":  "Crack",
				"yolo_coords": []float64{0.5, 0.5, 0.2, 0.2},
				"confidence":  0.91,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal callback payload failed: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/detection/video/callback/frame-result", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected callback 200, got %d: %s", w.Code, w.Body.String())
	}

	var refreshedTask model.VideoAnalysisTask
	if err := db.Where("task_id = ?", task.TaskID).First(&refreshedTask).Error; err != nil {
		t.Fatalf("reload task failed: %v", err)
	}
	if refreshedTask.Status != model.VideoTaskProcessing {
		t.Fatalf("expected processing status, got %s", refreshedTask.Status)
	}
	if refreshedTask.ProcessedFrames != 1 {
		t.Fatalf("expected processed_frames=1, got %d", refreshedTask.ProcessedFrames)
	}

	var frameTask model.VideoFrameTaskRequest
	if err := db.Where("request_id = ?", "frame_req_45").First(&frameTask).Error; err != nil {
		t.Fatalf("load frame task failed: %v", err)
	}
	if frameTask.Status != model.VideoFrameTaskCompleted {
		t.Fatalf("expected completed frame task, got %s", frameTask.Status)
	}

	var observations []model.DefectObservation
	if err := db.Where("task_id = ?", task.TaskID).Find(&observations).Error; err != nil {
		t.Fatalf("load observations failed: %v", err)
	}
	if len(observations) != 1 {
		t.Fatalf("expected 1 observation, got %d", len(observations))
	}
	if observations[0].TrackID == "" {
		t.Fatalf("expected observation track_id to be populated")
	}
}

func TestVideoCallbackTaskCompletedFinalizesOpenTracks(t *testing.T) {
	db := setupDetectionTestDB(t)
	if err := db.AutoMigrate(&model.VideoAnalysisTask{}, &model.DefectObservation{}, &model.VideoFrameTaskRequest{}); err != nil {
		t.Fatalf("migrate video callback models failed: %v", err)
	}

	router := setupVideoTaskRouter(db)
	defer cleanupTestFiles()

	user := createDetectionTestUser(db, "video_callback_complete_user", "user")
	bridge := createDetectionTestBridge(db, user.ID, "完成回调桥梁")
	task := &model.VideoAnalysisTask{
		TaskID:      "video_task_callback_complete",
		UserID:      user.ID,
		BridgeID:    bridge.ID,
		VideoPath:   "videos/callback_complete.mp4",
		FPS:         1,
		ModelName:   "baseline",
		PixelRatio:  0.01,
		Status:      model.VideoTaskDispatching,
		TotalFrames: 3,
	}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("create task failed: %v", err)
	}

	frameDir := t.TempDir()
	framePath1 := createJPEGFrame(t, frameDir, "frame_001.jpg")
	framePath2 := createJPEGFrame(t, frameDir, "frame_002.jpg")
	framePath3 := createJPEGFrame(t, frameDir, "frame_003.jpg")

	callbacks := []map[string]any{
		{
			"request_id":       "frame_req_complete_1",
			"task_id":          task.TaskID,
			"frame_no":         1,
			"timestamp_ms":     0,
			"status":           "success",
			"queue_latency_ms": 10,
			"detect_total_ms":  100,
			"frame_ref":        "file://" + framePath1,
			"yolo_bboxes": []map[string]any{{
				"box_id":      0,
				"class_idx":   0,
				"class_name":  "Crack",
				"yolo_coords": []float64{0.5, 0.5, 0.2, 0.2},
				"confidence":  0.9,
			}},
		},
		{
			"request_id":       "frame_req_complete_2",
			"task_id":          task.TaskID,
			"frame_no":         2,
			"timestamp_ms":     1000,
			"status":           "success",
			"queue_latency_ms": 10,
			"detect_total_ms":  100,
			"frame_ref":        "file://" + framePath2,
			"yolo_bboxes": []map[string]any{{
				"box_id":      0,
				"class_idx":   0,
				"class_name":  "Crack",
				"yolo_coords": []float64{0.5, 0.5, 0.2, 0.2},
				"confidence":  0.9,
			}},
		},
		{
			"request_id":       "frame_req_complete_3",
			"task_id":          task.TaskID,
			"frame_no":         3,
			"timestamp_ms":     2000,
			"status":           "success",
			"queue_latency_ms": 10,
			"detect_total_ms":  100,
			"frame_ref":        "file://" + framePath3,
			"yolo_bboxes": []map[string]any{{
				"box_id":      0,
				"class_idx":   0,
				"class_name":  "Crack",
				"yolo_coords": []float64{0.5, 0.5, 0.2, 0.2},
				"confidence":  0.9,
			}},
		},
	}

	for _, payload := range callbacks {
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal frame callback payload failed: %v", err)
		}
		req := httptest.NewRequest("POST", "/api/v1/detection/video/callback/frame-result", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected frame callback 200, got %d: %s", w.Code, w.Body.String())
		}
	}

	completedBody, err := json.Marshal(map[string]any{
		"task_id":          task.TaskID,
		"status":           "completed",
		"total_frames":     3,
		"processed_frames": 3,
		"failed_frames":    0,
	})
	if err != nil {
		t.Fatalf("marshal completed payload failed: %v", err)
	}

	completedReq := httptest.NewRequest("POST", "/api/v1/detection/video/callback/task-completed", bytes.NewReader(completedBody))
	completedReq.Header.Set("Content-Type", "application/json")
	completedW := httptest.NewRecorder()
	router.ServeHTTP(completedW, completedReq)
	if completedW.Code != http.StatusOK {
		t.Fatalf("expected task completed callback 200, got %d: %s", completedW.Code, completedW.Body.String())
	}

	var refreshedTask model.VideoAnalysisTask
	if err := db.Where("task_id = ?", task.TaskID).First(&refreshedTask).Error; err != nil {
		t.Fatalf("reload task failed: %v", err)
	}
	if refreshedTask.Status != model.VideoTaskCompleted {
		t.Fatalf("expected completed status, got %s", refreshedTask.Status)
	}
	if refreshedTask.ConfirmedDefects != 1 {
		t.Fatalf("expected 1 confirmed defect, got %d", refreshedTask.ConfirmedDefects)
	}

	var defectCount int64
	if err := db.Model(&model.Defect{}).Count(&defectCount).Error; err != nil {
		t.Fatalf("count defects failed: %v", err)
	}
	if defectCount != 1 {
		t.Fatalf("expected 1 persisted defect, got %d", defectCount)
	}
}
