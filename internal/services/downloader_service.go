package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"global-downloader/internal/models"
	"global-downloader/internal/queue"
)

// DownloaderService orchestrates job creation and enqueueing
type DownloaderService struct {
	store *queue.JobStore
}

// NewDownloaderService creates a DownloaderService backed by the given store
func NewDownloaderService(store *queue.JobStore) *DownloaderService {
	return &DownloaderService{store: store}
}

// CreateJob validates a download request, creates a Job, and enqueues it
func (s *DownloaderService) CreateJob(req *models.DownloadRequest) (*models.Job, error) {
	if req.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	// Apply defaults
	if req.Format == "" {
		req.Format = models.FormatMP4
	}
	if req.Quality == "" {
		req.Quality = models.QualityBest
	}
	if req.AudioOnly {
		// Override format to MP3 if audio-only and no audio format set
		if req.Format == models.FormatMP4 || req.Format == models.FormatMKV || req.Format == models.FormatWebM {
			req.Format = models.FormatMP3
		}
	}

	now := time.Now()
	job := &models.Job{
		ID:        uuid.New().String(),
		URL:       req.URL,
		Format:    req.Format,
		Quality:   req.Quality,
		AudioOnly: req.AudioOnly,
		Status:    models.StatusPending,
		Progress:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.store.Enqueue(job); err != nil {
		return nil, fmt.Errorf("could not enqueue job: %w", err)
	}

	return job, nil
}

// GetJob returns a job by ID
func (s *DownloaderService) GetJob(id string) (*models.Job, error) {
	job, ok := s.store.Get(id)
	if !ok {
		return nil, fmt.Errorf("job not found: %s", id)
	}
	return job, nil
}

// ListJobs returns all jobs currently tracked
func (s *DownloaderService) ListJobs() []*models.Job {
	return s.store.List()
}