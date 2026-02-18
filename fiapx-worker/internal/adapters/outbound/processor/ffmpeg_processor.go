package processor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"video-processor-worker/internal/core/ports"
)

type ffmpegProcessor struct {
	tempDir string
}

func NewFFmpegProcessor() ports.VideoProcessor {
	return &ffmpegProcessor{
		tempDir: "temp",
	}
}

func (p *ffmpegProcessor) ExtractFrames(videoPath string, timestamp string) ([]string, error) {
	tempOutputDir := filepath.Join(p.tempDir, timestamp)
	os.MkdirAll(tempOutputDir, 0755)
	// We don't remove it here because the service might need the frames for zipping
	// Actually, the storage should probably handle temp files?
	// For now, let's keep it simple. The service should probably be responsible for cleanup if it's not in the adapter.

	framePattern := filepath.Join(tempOutputDir, "frame_%04d.png")

	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vf", "fps=1",
		"-y",
		framePattern,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg error: %w, output: %s", err, string(output))
	}

	frames, err := filepath.Glob(filepath.Join(tempOutputDir, "*.png"))
	if err != nil {
		return nil, err
	}

	if len(frames) == 0 {
		return nil, fmt.Errorf("no frames extracted")
	}

	return frames, nil
}
