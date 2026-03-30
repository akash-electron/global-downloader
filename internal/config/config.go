package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	Port           string
	DownloadDir    string
	MaxWorkers     int
	MaxQueueSize   int
	YtDlpPath      string
	FfmpegPath     string
	MaxFileAgeMins int // how long to keep files before cleanup (minutes)
}

// Load reads config from environment variables, applying defaults where needed
func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DownloadDir:    getEnv("DOWNLOAD_DIR", "downloads"),
		MaxWorkers:     getEnvInt("MAX_WORKERS", 4),
		MaxQueueSize:   getEnvInt("MAX_QUEUE_SIZE", 100),
		YtDlpPath:      getEnv("YTDLP_PATH", "yt-dlp"),
		FfmpegPath:     getEnv("FFMPEG_PATH", "ffmpeg"),
		MaxFileAgeMins: getEnvInt("MAX_FILE_AGE_MINS", 60),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
