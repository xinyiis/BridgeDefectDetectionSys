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
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}

	return &ffmpegFrameExtractor{
		ffmpegPath:  ffmpegPath,
		ffprobePath: "ffprobe",
	}
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
