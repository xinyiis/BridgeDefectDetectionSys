package config

import "testing"

func TestValidateConfigDefaultsPersistenceWorkers(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			DSN: "sqlite.db",
		},
		Session: SessionConfig{
			Secret: "test-secret",
		},
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validate config: %v", err)
	}

	if cfg.Detection.PersistenceWorkers != 2 {
		t.Fatalf("expected default persistence_workers=2, got %d", cfg.Detection.PersistenceWorkers)
	}
	if cfg.VideoDetection.SampleFPS != 3 {
		t.Fatalf("expected default sample_fps=3, got %v", cfg.VideoDetection.SampleFPS)
	}
	if cfg.VideoDetection.TrackWindowSeconds != 3 {
		t.Fatalf("expected default track_window_seconds=3, got %d", cfg.VideoDetection.TrackWindowSeconds)
	}
	if cfg.VideoDetection.TrackCloseSeconds != 4 {
		t.Fatalf("expected default track_close_seconds=4, got %d", cfg.VideoDetection.TrackCloseSeconds)
	}
	if cfg.VideoDetection.TrackIOUThreshold != 0.3 {
		t.Fatalf("expected default track_iou_threshold=0.3, got %v", cfg.VideoDetection.TrackIOUThreshold)
	}
	if cfg.VideoDetection.TrackCenterDistanceThreshold != 0.08 {
		t.Fatalf("expected default track_center_distance_threshold=0.08, got %v", cfg.VideoDetection.TrackCenterDistanceThreshold)
	}
	if cfg.VideoDetection.ConfirmHits != 3 {
		t.Fatalf("expected default confirm_hits=3, got %d", cfg.VideoDetection.ConfirmHits)
	}
	if cfg.VideoDetection.CandidateConfidenceThreshold != 0.1 {
		t.Fatalf("expected default candidate_confidence_threshold=0.1, got %v", cfg.VideoDetection.CandidateConfidenceThreshold)
	}
	if cfg.VideoDetection.PersistConfidenceThreshold != 0.55 {
		t.Fatalf("expected default persist_confidence_threshold=0.55, got %v", cfg.VideoDetection.PersistConfidenceThreshold)
	}
	if cfg.VideoDetection.MaxQueueInflight != 8 {
		t.Fatalf("expected default max_queue_inflight=8, got %d", cfg.VideoDetection.MaxQueueInflight)
	}
	if cfg.VideoDetection.QueuedTimeoutSeconds != 5 {
		t.Fatalf("expected default queued_timeout_seconds=5, got %d", cfg.VideoDetection.QueuedTimeoutSeconds)
	}
	if cfg.VideoDetection.ProgressIdleTimeoutSeconds != 15 {
		t.Fatalf("expected default progress_idle_timeout_seconds=15, got %d", cfg.VideoDetection.ProgressIdleTimeoutSeconds)
	}
	if cfg.VideoDetection.CallbackBaseURL != "http://localhost:8080/api/v1/detection/video/callback" {
		t.Fatalf("expected default callback_base_url, got %s", cfg.VideoDetection.CallbackBaseURL)
	}
	if cfg.Upload.BaseDir != "/autodl-tmp/NLP/source" {
		t.Fatalf("expected default upload.base_dir=/autodl-tmp/NLP/source, got %s", cfg.Upload.BaseDir)
	}
	if cfg.Upload.ImageDir != "images" {
		t.Fatalf("expected default upload.image_dir=images, got %s", cfg.Upload.ImageDir)
	}
	if cfg.Upload.ResultDir != "results" {
		t.Fatalf("expected default upload.result_dir=results, got %s", cfg.Upload.ResultDir)
	}
	if cfg.Upload.MaxSize != 10 {
		t.Fatalf("expected default upload.max_size=10, got %d", cfg.Upload.MaxSize)
	}
}

func TestValidateConfigKeepsExplicitPersistenceWorkers(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			DSN: "sqlite.db",
		},
		Session: SessionConfig{
			Secret: "test-secret",
		},
		Detection: DetectionConfig{
			PersistenceWorkers: 4,
		},
		VideoDetection: VideoDetectionConfig{
			SampleFPS:                    2,
			TrackWindowSeconds:           5,
			TrackCloseSeconds:            6,
			TrackIOUThreshold:            0.5,
			TrackCenterDistanceThreshold: 0.1,
			ConfirmHits:                  4,
			CandidateConfidenceThreshold: 0.6,
			PersistConfidenceThreshold:   0.7,
			MaxQueueInflight:             12,
			QueuedTimeoutSeconds:         7,
			ProgressIdleTimeoutSeconds:   21,
			CallbackBaseURL:              "http://backend/api/v1/detection/video/callback",
		},
		Upload: UploadConfig{
			BaseDir:   "/data/source",
			ImageDir:  "images_custom",
			ResultDir: "results_custom",
			MaxSize:   20,
		},
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validate config: %v", err)
	}

	if cfg.Detection.PersistenceWorkers != 4 {
		t.Fatalf("expected persistence_workers to stay 4, got %d", cfg.Detection.PersistenceWorkers)
	}
	if cfg.VideoDetection.SampleFPS != 2 {
		t.Fatalf("expected sample_fps to stay 2, got %v", cfg.VideoDetection.SampleFPS)
	}
	if cfg.VideoDetection.TrackWindowSeconds != 5 {
		t.Fatalf("expected track_window_seconds to stay 5, got %d", cfg.VideoDetection.TrackWindowSeconds)
	}
	if cfg.VideoDetection.TrackCloseSeconds != 6 {
		t.Fatalf("expected track_close_seconds to stay 6, got %d", cfg.VideoDetection.TrackCloseSeconds)
	}
	if cfg.VideoDetection.TrackIOUThreshold != 0.5 {
		t.Fatalf("expected track_iou_threshold to stay 0.5, got %v", cfg.VideoDetection.TrackIOUThreshold)
	}
	if cfg.VideoDetection.TrackCenterDistanceThreshold != 0.1 {
		t.Fatalf("expected track_center_distance_threshold to stay 0.1, got %v", cfg.VideoDetection.TrackCenterDistanceThreshold)
	}
	if cfg.VideoDetection.ConfirmHits != 4 {
		t.Fatalf("expected confirm_hits to stay 4, got %d", cfg.VideoDetection.ConfirmHits)
	}
	if cfg.VideoDetection.CandidateConfidenceThreshold != 0.6 {
		t.Fatalf("expected candidate_confidence_threshold to stay 0.6, got %v", cfg.VideoDetection.CandidateConfidenceThreshold)
	}
	if cfg.VideoDetection.PersistConfidenceThreshold != 0.7 {
		t.Fatalf("expected persist_confidence_threshold to stay 0.7, got %v", cfg.VideoDetection.PersistConfidenceThreshold)
	}
	if cfg.VideoDetection.MaxQueueInflight != 12 {
		t.Fatalf("expected max_queue_inflight to stay 12, got %d", cfg.VideoDetection.MaxQueueInflight)
	}
	if cfg.VideoDetection.QueuedTimeoutSeconds != 7 {
		t.Fatalf("expected queued_timeout_seconds to stay 7, got %d", cfg.VideoDetection.QueuedTimeoutSeconds)
	}
	if cfg.VideoDetection.ProgressIdleTimeoutSeconds != 21 {
		t.Fatalf("expected progress_idle_timeout_seconds to stay 21, got %d", cfg.VideoDetection.ProgressIdleTimeoutSeconds)
	}
	if cfg.VideoDetection.CallbackBaseURL != "http://backend/api/v1/detection/video/callback" {
		t.Fatalf("expected callback_base_url to stay explicit, got %s", cfg.VideoDetection.CallbackBaseURL)
	}
	if cfg.Upload.BaseDir != "/data/source" {
		t.Fatalf("expected upload.base_dir to stay /data/source, got %s", cfg.Upload.BaseDir)
	}
	if cfg.Upload.ImageDir != "images_custom" {
		t.Fatalf("expected upload.image_dir to stay images_custom, got %s", cfg.Upload.ImageDir)
	}
	if cfg.Upload.ResultDir != "results_custom" {
		t.Fatalf("expected upload.result_dir to stay results_custom, got %s", cfg.Upload.ResultDir)
	}
	if cfg.Upload.MaxSize != 20 {
		t.Fatalf("expected upload.max_size to stay 20, got %d", cfg.Upload.MaxSize)
	}
}
