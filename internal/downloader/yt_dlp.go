package downloader

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"global-downloader/internal/models"
	"global-downloader/pkg/logger"
)

// ProgressCallback is called with download progress (0.0–100.0)
type ProgressCallback func(progress float64, title string)

// YtDlp wraps the yt-dlp binary for downloading videos
type YtDlp struct {
	BinaryPath  string
	DownloadDir string
}

// NewYtDlp creates a new YtDlp instance
func NewYtDlp(binaryPath, downloadDir string) *YtDlp {
	return &YtDlp{
		BinaryPath:  binaryPath,
		DownloadDir: downloadDir,
	}
}

// Download downloads a video/audio for a given job, reporting progress via callback.
// Returns the absolute file path of the saved file.
func (y *YtDlp) Download(job *models.Job, onProgress ProgressCallback) (string, error) {
	if err := os.MkdirAll(y.DownloadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create download dir: %w", err)
	}

	args := y.buildArgs(job)
	logger.Info("starting yt-dlp", "job_id", job.ID, "url", job.URL, "args", args)

	cmd := exec.Command(y.BinaryPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe error: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout for scanning

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start yt-dlp: %w", err)
	}

	var lastLine string
	var title string
	progressRe := regexp.MustCompile(`(\d+\.?\d*)%`)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		lastLine = line
		logger.Debug("yt-dlp", "job_id", job.ID, "line", line)

		// Extract title from "[info]" line
		if strings.Contains(line, "[info]") && strings.Contains(line, "Downloading") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				title = strings.TrimSpace(parts[1])
			}
		}

		// Parse progress percentage
		if matches := progressRe.FindStringSubmatch(line); len(matches) > 1 {
			if pct, err := strconv.ParseFloat(matches[1], 64); err == nil {
				if onProgress != nil {
					onProgress(pct, title)
				}
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("yt-dlp failed: %s | %w", lastLine, err)
	}

	// Resolve the output filename yt-dlp actually wrote
	filePath, err := y.resolveOutputFile(job)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// buildArgs constructs the yt-dlp command-line arguments based on the job spec
func (y *YtDlp) buildArgs(job *models.Job) []string {
	outputTemplate := filepath.Join(y.DownloadDir, "%(title)s.%(ext)s")

	args := []string{
		"--no-playlist",
		"--progress",
		"--newline",
		"--print", "after_move:filepath",
		"--output", outputTemplate,
		"--no-continue", // always re-download (avoids stale partial files)
	}

	if job.AudioOnly {
		// Audio extraction
		audioFmt := string(job.Format)
		if audioFmt == "" || audioFmt == "best" || audioFmt == "mp4" || audioFmt == "mkv" {
			audioFmt = "mp3"
		}
		args = append(args,
			"--extract-audio",
			"--audio-format", audioFmt,
			"--audio-quality", "0", // best
		)
	} else {
		// Video download
		videoFmt := buildVideoFormatSelector(job.Quality)
		mergeFormat := string(job.Format)
		if mergeFormat == "" || mergeFormat == "best" {
			mergeFormat = "mp4"
		}
		args = append(args,
			"-f", videoFmt,
			"--merge-output-format", mergeFormat,
		)
	}

	args = append(args, job.URL)
	return args
}

// buildVideoFormatSelector returns the yt-dlp -f selector for a quality preference
func buildVideoFormatSelector(quality models.Quality) string {
	switch quality {
	case models.Quality1080p:
		return "bestvideo[height<=1080][ext=mp4]+bestaudio[ext=m4a]/bestvideo[height<=1080]+bestaudio/best[height<=1080]"
	case models.Quality720p:
		return "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/bestvideo[height<=720]+bestaudio/best[height<=720]"
	case models.Quality480p:
		return "bestvideo[height<=480][ext=mp4]+bestaudio[ext=m4a]/bestvideo[height<=480]+bestaudio/best[height<=480]"
	case models.Quality360p:
		return "bestvideo[height<=360][ext=mp4]+bestaudio[ext=m4a]/bestvideo[height<=360]+bestaudio/best[height<=360]"
	default:
		return "bestvideo[ext=mp4]+bestaudio[ext=m4a]/bestvideo+bestaudio/best"
	}
}

// resolveOutputFile returns the actual path of the file yt-dlp wrote.
// It uses --print after_move:filepath to echo the final path to stdout;
// we scan that output during Download. Here we do a best-effort glob fallback.
func (y *YtDlp) resolveOutputFile(job *models.Job) (string, error) {
	// yt-dlp prints the final filepath via --print after_move:filepath.
	// If we captured it during scanning, great. Otherwise glob the downloads dir.
	pattern := filepath.Join(y.DownloadDir, "*")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("could not find output file for job %s", job.ID)
	}

	// Return the most recently modified file
	var newest string
	var newestTime int64
	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil {
			continue
		}
		if info.ModTime().Unix() > newestTime {
			newestTime = info.ModTime().Unix()
			newest = m
		}
	}

	if newest == "" {
		return "", fmt.Errorf("no output file found for job %s", job.ID)
	}
	return newest, nil
}

// GetInfo fetches video metadata (title, duration, formats) without downloading
func (y *YtDlp) GetInfo(url string) (map[string]string, error) {
	cmd := exec.Command(y.BinaryPath,
		"--dump-json",
		"--no-playlist",
		url,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp --dump-json failed: %w", err)
	}
	// For now return the raw JSON as a single-entry map.
	// Callers can unmarshal as needed.
	return map[string]string{"json": string(out)}, nil
}
