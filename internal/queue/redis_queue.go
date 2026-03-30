package queue

import (
	"fmt"
	"sync"

	"global-downloader/internal/models"
)

// JobStore is a thread-safe in-memory job registry and work queue.
// For production scale, swap the channel + map for a Redis-backed implementation.
type JobStore struct {
	mu      sync.RWMutex
	jobs    map[string]*models.Job
	workCh  chan *models.Job
}

// NewJobStore creates a JobStore with capacity maxQueue buffered slots
func NewJobStore(maxQueue int) *JobStore {
	return &JobStore{
		jobs:   make(map[string]*models.Job),
		workCh: make(chan *models.Job, maxQueue),
	}
}

// Enqueue adds a job to the registry and places it on the work channel.
// Returns ErrQueueFull if the buffered channel is at capacity.
func (s *JobStore) Enqueue(job *models.Job) error {
	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()

	select {
	case s.workCh <- job:
		return nil
	default:
		return fmt.Errorf("queue is full, try again later")
	}
}

// Get returns the job with the given ID, or nil if not found
func (s *JobStore) Get(id string) (*models.Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[id]
	return job, ok
}

// Update applies a mutator function to a job while holding the write lock
func (s *JobStore) Update(id string, fn func(*models.Job)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	if !ok {
		return fmt.Errorf("job %s not found", id)
	}
	fn(job)
	return nil
}

// List returns a snapshot of all jobs
func (s *JobStore) List() []*models.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		out = append(out, j)
	}
	return out
}

// WorkChannel exposes the read-only work channel for workers to consume
func (s *JobStore) WorkChannel() <-chan *models.Job {
	return s.workCh
}
