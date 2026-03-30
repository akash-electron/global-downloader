package routes

import (
	"net/http"

	"global-downloader/internal/handlers"
	"global-downloader/internal/services"
	"global-downloader/internal/storage"
)

// SetupRoutes wires all HTTP routes to handler functions
func SetupRoutes(svc *services.DownloaderService, store *storage.FileStore) *http.ServeMux {
	h := handlers.New(svc, store)

	mux := http.NewServeMux()

	// Core download API
	mux.HandleFunc("/download", h.Download)      // POST  — enqueue a download job
	mux.HandleFunc("/job/", h.GetJob)            // GET   — poll job status + progress
	mux.HandleFunc("/jobs", h.ListJobs)          // GET   — list all jobs

	// File serving
	mux.HandleFunc("/file/", h.ServeFile)        // GET   — stream a finished file
	mux.HandleFunc("/files", h.ListFiles)        // GET   — list all stored files

	// Observability
	mux.HandleFunc("/health", h.Health)          // GET   — liveness probe

	return mux
}