package video

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFrameExtractorFindsWorkspaceLocalTools(t *testing.T) {
	t.Helper()

	rootDir := t.TempDir()
	backendDir := filepath.Join(rootDir, "BridgeDefectDetectionSys", "src", "backend")
	toolDir := filepath.Join(rootDir, ".local-tools", "ffmpeg", "bin")

	if err := os.MkdirAll(backendDir, 0755); err != nil {
		t.Fatalf("mkdir backend dir: %v", err)
	}
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		t.Fatalf("mkdir tool dir: %v", err)
	}

	ffmpegPath := filepath.Join(toolDir, "ffmpeg")
	ffprobePath := filepath.Join(toolDir, "ffprobe")
	if err := os.WriteFile(ffmpegPath, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatalf("write ffmpeg stub: %v", err)
	}
	if err := os.WriteFile(ffprobePath, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatalf("write ffprobe stub: %v", err)
	}

	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousWD)
	}()

	if err := os.Chdir(backendDir); err != nil {
		t.Fatalf("chdir backend dir: %v", err)
	}

	extractor, ok := NewFrameExtractor("").(*ffmpegFrameExtractor)
	if !ok {
		t.Fatalf("unexpected extractor type")
	}

	if extractor.ffmpegPath != ffmpegPath {
		t.Fatalf("unexpected ffmpeg path: %s", extractor.ffmpegPath)
	}

	if extractor.ffprobePath != ffprobePath {
		t.Fatalf("unexpected ffprobe path: %s", extractor.ffprobePath)
	}
}
