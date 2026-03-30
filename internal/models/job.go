package models

import "time"

// JobStatus represents the lifecycle state of a download job
type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

// Format represents the requested output format
type Format string

const (
	FormatMP4  Format = "mp4"
	FormatMKV  Format = "mkv"
	FormatWebM Format = "webm"
	FormatMP3  Format = "mp3"  // audio extraction
	FormatM4A  Format = "m4a"  // audio extraction
	FormatWAV  Format = "wav"  // audio extraction
	FormatOGG  Format = "ogg"  // audio extraction
	FormatBest Format = "best" // let yt-dlp decide
)

// Quality represents video quality preference
type Quality string

const (
	QualityBest   Quality = "best"
	Quality1080p  Quality = "1080p"
	Quality720p   Quality = "720p"
	Quality480p   Quality = "480p"
	Quality360p   Quality = "360p"
	QualityAudio  Quality = "audio" // audio only
)

// Job holds all the information about a download job
type Job struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Format      Format    `json:"format"`
	Quality     Quality   `json:"quality"`
	AudioOnly   bool      `json:"audio_only"`
	Status      JobStatus `json:"status"`
	Progress    float64   `json:"progress"`    // 0.0 – 100.0
	Title       string    `json:"title"`       // resolved video title
	FilePath    string    `json:"file_path"`   // absolute path on disk
	FileSize    int64     `json:"file_size"`   // bytes
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// DownloadRequest is the JSON payload for POST /download
type DownloadRequest struct {
	URL       string  `json:"url"`
	Format    Format  `json:"format"`
	Quality   Quality `json:"quality"`
	AudioOnly bool    `json:"audio_only"`
}

// JobResponse is the API response for a single job
type JobResponse struct {
	Job     *Job   `json:"job"`
	Message string `json:"message,omitempty"`
}

// ErrorResponse is a standardised API error envelope
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
}
