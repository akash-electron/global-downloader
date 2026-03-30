package workers

import (
	"os"
	"time"

	"global-downloader/internal/downloader"
	"global-downloader/internal/models"
	"global-downloader/internal/queue"
	"global-downloader/pkg/logger"
)

// DownloadWorkerPool manages a pool of concurrent download workers
type DownloadWorkerPool struct {
	store      *queue.JobStore
	ytDlp      *downloader.YtDlp
	numWorkers int
}

// NewDownloadWorkerPool creates (but does not start) a pool
func NewDownloadWorkerPool(store *queue.JobStore, ytDlp *downloader.YtDlp, numWorkers int) *DownloadWorkerPool {
	return &DownloadWorkerPool{
		store:      store,
		ytDlp:      ytDlp,
		numWorkers: numWorkers,
	}
}

// Start launches numWorkers goroutines that consume jobs from the queue
func (p *DownloadWorkerPool) Start() {
	for i := range p.numWorkers {
		go p.runWorker(i)
	}
	logger.Info("worker pool started", "workers", p.numWorkers)
}

// runWorker blocks on the work channel and processes jobs sequentially
func (p *DownloadWorkerPool) runWorker(id int) {
	logger.Info("worker started", "worker_id", id)

	for job := range p.store.WorkChannel() {
		p.processJob(id, job)
	}
}

// processJob executes a single download job end-to-end
func (p *DownloadWorkerPool) processJob(workerID int, job *models.Job) {
	logger.Info("processing job", "worker_id", workerID, "job_id", job.ID, "url", job.URL)

	// Mark as processing
	_ = p.store.Update(job.ID, func(j *models.Job) {
		j.Status = models.StatusProcessing
		j.UpdatedAt = time.Now()
	})

	// Progress callback — updates job in place
	onProgress := func(pct float64, title string) {
		_ = p.store.Update(job.ID, func(j *models.Job) {
			j.Progress = pct
			if title != "" {
				j.Title = title
			}
			j.UpdatedAt = time.Now()
		})
	}

	filePath, err := p.ytDlp.Download(job, onProgress)
	if err != nil {
		logger.Error("download failed", "worker_id", workerID, "job_id", job.ID, "err", err)
		_ = p.store.Update(job.ID, func(j *models.Job) {
			j.Status = models.StatusFailed
			j.Error = err.Error()
			j.UpdatedAt = time.Now()
		})
		return
	}

	// Harvest file size
	var fileSize int64
	if info, err := os.Stat(filePath); err == nil {
		fileSize = info.Size()
	}

	now := time.Now()
	_ = p.store.Update(job.ID, func(j *models.Job) {
		j.Status = models.StatusCompleted
		j.Progress = 100
		j.FilePath = filePath
		j.FileSize = fileSize
		j.UpdatedAt = now
		j.CompletedAt = &now
	})

	logger.Info("job completed", "worker_id", workerID, "job_id", job.ID, "file", filePath, "bytes", fileSize)
}
