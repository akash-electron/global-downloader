package utils

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// FFmpeg wraps the ffmpeg binary for post-processing tasks
type FFmpeg struct {
	BinaryPath string
}

// NewFFmpeg creates a new FFmpeg instance
func NewFFmpeg(binaryPath string) *FFmpeg {
	return &FFmpeg{BinaryPath: binaryPath}
}

// ExtractAudio extracts the audio from a video file and writes to outputPath.
// format can be "mp3", "aac", "wav", "ogg", "flac".
func (f *FFmpeg) ExtractAudio(inputPath, outputPath, format string) error {
	args := []string{
		"-y",                       // overwrite without prompt
		"-i", inputPath,            // input file
		"-vn",                      // no video stream
		"-acodec", audioCodec(format), // output codec
		outputPath,
	}

	cmd := exec.Command(f.BinaryPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg extract audio failed: %s | %w", string(out), err)
	}
	return nil
}

// ConvertVideo converts a video file to a different container format.
func (f *FFmpeg) ConvertVideo(inputPath, outputPath string) error {
	args := []string{
		"-y",
		"-i", inputPath,
		"-c", "copy", // stream copy (fast, lossless)
		outputPath,
	}

	cmd := exec.Command(f.BinaryPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg convert failed: %s | %w", string(out), err)
	}
	return nil
}

// GetDuration returns the duration of a media file in seconds using ffprobe.
func (f *FFmpeg) GetDuration(filePath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}
	var dur float64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%f", &dur); err != nil {
		return 0, fmt.Errorf("could not parse duration: %w", err)
	}
	return dur, nil
}

// AudioOutputPath returns a sibling output path with the given audio extension
func AudioOutputPath(inputPath, format string) string {
	ext := format
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	base := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
	return base + ext
}

// audioCodec maps a human-readable format name to an ffmpeg codec name
func audioCodec(format string) string {
	switch strings.ToLower(format) {
	case "mp3":
		return "libmp3lame"
	case "aac", "m4a":
		return "aac"
	case "ogg":
		return "libvorbis"
	case "flac":
		return "flac"
	case "wav":
		return "pcm_s16le"
	default:
		return "copy"
	}
}
