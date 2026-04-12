package video

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// FrameExtractor 定义视频帧提取能力。
type FrameExtractor interface {
	CountFrames(videoPath string, fps float64) (int, error)
	ExtractFramesStream(videoPath, outputDir string, fps float64, onFrame func(framePath string, frameNo int) error) error
}

type ffmpegFrameExtractor struct {
	ffmpegPath  string
	ffprobePath string
}

// NewFrameExtractor 创建 FFmpeg 帧提取器。
func NewFrameExtractor(ffmpegPath string) FrameExtractor {
	resolvedFFmpegPath, resolvedFFprobePath := resolveFFmpegPaths(ffmpegPath)

	return &ffmpegFrameExtractor{
		ffmpegPath:  resolvedFFmpegPath,
		ffprobePath: resolvedFFprobePath,
	}
}

func resolveFFmpegPaths(ffmpegPath string) (string, string) {
	requestedPath := strings.TrimSpace(ffmpegPath)
	if requestedPath != "" && requestedPath != "ffmpeg" {
		return requestedPath, resolveFFprobePath(filepath.Dir(requestedPath))
	}

	if envFFmpegPath := strings.TrimSpace(os.Getenv("FFMPEG_PATH")); envFFmpegPath != "" {
		return envFFmpegPath, resolveFFprobePath(filepath.Dir(envFFmpegPath))
	}

	if localFFmpegPath, localFFprobePath, ok := findLocalFFmpegBinaries(); ok {
		return localFFmpegPath, localFFprobePath
	}

	return resolveBinaryPath("ffmpeg"), resolveBinaryPath(resolveEnvOrDefault("FFPROBE_PATH", "ffprobe"))
}

func resolveFFprobePath(ffmpegDir string) string {
	if envFFprobePath := strings.TrimSpace(os.Getenv("FFPROBE_PATH")); envFFprobePath != "" {
		return envFFprobePath
	}

	if ffmpegDir != "" {
		ffprobePath := filepath.Join(ffmpegDir, "ffprobe")
		if isUsableBinary(ffprobePath) {
			return ffprobePath
		}
	}

	return resolveBinaryPath("ffprobe")
}

func findLocalFFmpegBinaries() (string, string, bool) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", "", false
	}

	currentDir := workingDir
	for {
		ffmpegPath := filepath.Join(currentDir, ".local-tools", "ffmpeg", "bin", "ffmpeg")
		ffprobePath := filepath.Join(currentDir, ".local-tools", "ffmpeg", "bin", "ffprobe")
		if isUsableBinary(ffmpegPath) && isUsableBinary(ffprobePath) {
			return ffmpegPath, ffprobePath, true
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}

		currentDir = parentDir
	}

	return "", "", false
}

func resolveEnvOrDefault(envName, defaultValue string) string {
	if envValue := strings.TrimSpace(os.Getenv(envName)); envValue != "" {
		return envValue
	}

	return defaultValue
}

func resolveBinaryPath(binary string) string {
	if absolutePath, err := exec.LookPath(binary); err == nil {
		return absolutePath
	}

	return binary
}

func isUsableBinary(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil || fileInfo.IsDir() {
		return false
	}

	return fileInfo.Mode().Perm()&0111 != 0
}

func (e *ffmpegFrameExtractor) CountFrames(videoPath string, fps float64) (int, error) {
	if fps <= 0 {
		fps = 1
	}

	output, err := exec.Command(
		e.ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	).Output()
	if err != nil {
		return 0, fmt.Errorf("获取视频时长失败: %w", err)
	}

	durationText := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationText, 64)
	if err != nil {
		return 0, fmt.Errorf("解析视频时长失败: %w", err)
	}

	return int(math.Ceil(duration * fps)), nil
}

func (e *ffmpegFrameExtractor) ExtractFramesStream(
	videoPath, outputDir string,
	fps float64,
	onFrame func(framePath string, frameNo int) error,
) error {
	if fps <= 0 {
		fps = 1
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建帧目录失败: %w", err)
	}

	outputPattern := filepath.Join(outputDir, "frame_%06d.jpg")
	cmd := exec.Command(
		e.ffmpegPath,
		"-i", videoPath,
		"-vf", fmt.Sprintf("fps=%.2f", fps),
		"-q:v", "2",
		"-start_number", "1",
		outputPattern,
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 FFmpeg 失败: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	frameNo := 1
	extractorDone := false

	for {
		framePath := filepath.Join(outputDir, fmt.Sprintf("frame_%06d.jpg", frameNo))
		if _, err := os.Stat(framePath); err == nil {
			if err := onFrame(framePath, frameNo); err != nil {
				if cmd.Process != nil {
					_ = cmd.Process.Kill()
				}
				<-done
				return err
			}
			frameNo++
			continue
		}

		if extractorDone {
			break
		}

		select {
		case err := <-done:
			if err != nil {
				return fmt.Errorf("FFmpeg 提帧失败: %w", err)
			}
			extractorDone = true
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}
