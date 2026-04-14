package external

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	domainservice "github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
)

func TestHTTPPythonServiceSegmentUsesBBoxesJSONField(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "sample.jpg")
	if err := os.WriteFile(imagePath, []byte("fake-image"), 0644); err != nil {
		t.Fatalf("write temp image: %v", err)
	}

	const expectedBBoxes = `[{"box_id":0,"class_idx":0,"class_name":"Crack","yolo_coords":[0.5,0.5,0.3,0.1]}]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.URL.Path != "/algo/segment" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parse multipart form: %v", err)
		}

		if got := r.FormValue("bboxes_json"); got != expectedBBoxes {
			t.Fatalf("unexpected bboxes_json: %q", got)
		}

		if got := r.FormValue("yolo_bboxes"); got != "" {
			t.Fatalf("unexpected legacy yolo_bboxes field: %q", got)
		}

		if got := r.FormValue("alpha"); got != "0.50" {
			t.Fatalf("unexpected alpha: %q", got)
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("missing file field: %v", err)
		}
		defer file.Close()

		if _, err := io.ReadAll(file); err != nil {
			t.Fatalf("read uploaded file: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","fusion_image":"ok","individual_masks":[]}`))
	}))
	defer server.Close()

	pythonService := NewHTTPPythonService(server.URL)
	result, err := pythonService.Segment(imagePath, &domainservice.SegmentRequest{
		BBoxesJSON: expectedBBoxes,
		Alpha:      0.5,
	})
	if err != nil {
		t.Fatalf("segment request failed: %v", err)
	}

	if result.Status != "success" {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

func TestHTTPPythonServiceDetectDefectComposesDetectAndSegment(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "sample.jpg")
	if err := os.WriteFile(imagePath, []byte("fake-image"), 0644); err != nil {
		t.Fatalf("write temp image: %v", err)
	}

	const detectResponse = `{"status":"success","model_used":"baseline","yolo_bboxes":[{"box_id":0,"class_idx":0,"class_name":"Crack","yolo_coords":[0.5,0.5,0.3,0.1],"confidence":0.93}],"image_results":"detect-image"}`
	const expectedBBoxes = `[{"box_id":0,"class_idx":0,"class_name":"Crack","yolo_coords":[0.5,0.5,0.3,0.1],"confidence":0.93}]`
	const segmentResponse = `{"status":"success","fusion_image":"segment-image","individual_masks":[{"box_index":0,"label":"Crack","mask_base64":"mask"}]}`

	var detectCalled bool
	var segmentCalled bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		switch r.URL.Path {
		case "/algo/detect":
			detectCalled = true
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Fatalf("parse detect multipart form: %v", err)
			}
			if got := r.FormValue("model_name"); got != "baseline" {
				t.Fatalf("unexpected model name: %q", got)
			}
			if got := r.FormValue("conf"); got != "0.25" {
				t.Fatalf("unexpected conf: %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(detectResponse))
		case "/algo/segment":
			segmentCalled = true
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Fatalf("parse segment multipart form: %v", err)
			}
			if got := r.FormValue("bboxes_json"); got != expectedBBoxes {
				t.Fatalf("unexpected bboxes_json: %q", got)
			}
			if got := r.FormValue("alpha"); got != "0.50" {
				t.Fatalf("unexpected alpha: %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(segmentResponse))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	pythonService := NewHTTPPythonService(server.URL)
	result, err := pythonService.DetectDefect(imagePath, "baseline", 0.1)
	if err != nil {
		t.Fatalf("detect defect failed: %v", err)
	}

	if !detectCalled {
		t.Fatalf("detect endpoint was not called")
	}
	if !segmentCalled {
		t.Fatalf("segment endpoint was not called")
	}
	if !result.Success {
		t.Fatalf("expected success result")
	}
	if result.TotalDefects != 1 {
		t.Fatalf("unexpected total defects: %d", result.TotalDefects)
	}
	if result.ResultImage != "segment-image" {
		t.Fatalf("unexpected result image: %q", result.ResultImage)
	}
	if len(result.Defects) != 1 {
		t.Fatalf("unexpected defect count: %d", len(result.Defects))
	}
	if result.Defects[0].DefectType != "Crack" {
		t.Fatalf("unexpected defect type: %s", result.Defects[0].DefectType)
	}
	if result.Defects[0].BBox.YOLOCoords != [4]float64{0.5, 0.5, 0.3, 0.1} {
		t.Fatalf("unexpected yolo coords: %#v", result.Defects[0].BBox.YOLOCoords)
	}
	if result.Defects[0].Confidence != 0.93 {
		t.Fatalf("unexpected confidence: %v", result.Defects[0].Confidence)
	}
}

func TestHTTPPythonServiceEnqueueVideoFrameDetectUsesJSONContract(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.URL.Path != "/algo/video/frames/detect" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("unexpected content-type: %s", got)
		}

		var req domainservice.VideoFrameDetectEnqueueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode json body failed: %v", err)
		}
		if req.RequestID != "frame_req_1" {
			t.Fatalf("unexpected request_id: %s", req.RequestID)
		}
		if req.TaskID != "video_task_1" {
			t.Fatalf("unexpected task_id: %s", req.TaskID)
		}
		if req.FrameNo != 12 {
			t.Fatalf("unexpected frame_no: %d", req.FrameNo)
		}
		if req.CallbackBaseURL != "http://backend/api/v1/detection/video/callback" {
			t.Fatalf("unexpected callback_base_url: %s", req.CallbackBaseURL)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"accepted","request_id":"frame_req_1","task_id":"video_task_1","queued_at":"2026-04-14T00:00:00Z"}`))
	}))
	defer server.Close()

	pythonService := NewHTTPPythonService(server.URL)
	result, err := pythonService.EnqueueVideoFrameDetect(&domainservice.VideoFrameDetectEnqueueRequest{
		RequestID:       "frame_req_1",
		TaskID:          "video_task_1",
		BridgeID:        101,
		FrameNo:         12,
		TimestampMS:     12000,
		FrameRef:        "file:///tmp/frame_000012.jpg",
		ModelName:       "baseline",
		Conf:            0.25,
		CallbackBaseURL: "http://backend/api/v1/detection/video/callback",
	})
	if err != nil {
		t.Fatalf("enqueue frame detect failed: %v", err)
	}
	if result.Status != "accepted" {
		t.Fatalf("unexpected status: %s", result.Status)
	}
	if result.RequestID != "frame_req_1" {
		t.Fatalf("unexpected response request_id: %s", result.RequestID)
	}
}
