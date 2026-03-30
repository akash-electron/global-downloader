package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"global-downloader/internal/models"
	"global-downloader/internal/services"
	"global-downloader/internal/storage"
	"global-downloader/pkg/logger"
)

// Handler holds dependencies for all HTTP handlers
type Handler struct {
	svc   *services.DownloaderService
	store *storage.FileStore
}

// New creates a Handler with its dependencies injected
func New(svc *services.DownloaderService, store *storage.FileStore) *Handler {
	return &Handler{svc: svc, store: store}
}

// ─── POST /download ──────────────────────────────────────────────────────────

// Download accepts a download request, enqueues a job, and returns the job ID
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	job, err := h.svc.CreateJob(&req)
	if err != nil {
		logger.Error("create job failed", "err", err)
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Info("job enqueued", "job_id", job.ID, "url", job.URL)
	writeJSON(w, http.StatusAccepted, &models.JobResponse{
		Job:     job,
		Message: "job enqueued",
	})
}

// ─── GET /job/{id} ───────────────────────────────────────────────────────────

// GetJob returns the current state of a download job
func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractPathParam(r.URL.Path, "/job/")
	if id == "" {
		writeError(w, "job id is required", http.StatusBadRequest)
		return
	}

	job, err := h.svc.GetJob(id)
	if err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, &models.JobResponse{Job: job})
}

// ─── GET /jobs ────────────────────────────────────────────────────────────────

// ListJobs returns all tracked jobs
func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs := h.svc.ListJobs()
	writeJSON(w, http.StatusOK, map[string]any{
		"jobs":  jobs,
		"count": len(jobs),
	})
}

// ─── GET /file/{filename} ─────────────────────────────────────────────────────

// ServeFile streams a completed download file to the client
func (h *Handler) ServeFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := extractPathParam(r.URL.Path, "/file/")
	if filename == "" {
		writeError(w, "filename is required", http.StatusBadRequest)
		return
	}

	info, err := h.store.Stat(filename)
	if err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}

	f, err := os.Open(info.Path)
	if err != nil {
		writeError(w, "could not open file", http.StatusInternalServerError)
		return
	}
	defer func() {
		f.Close()
		// Auto-delete the file immediately after streaming finishes / connection drops
		if err := h.store.Delete(filepath.Base(info.Path)); err != nil {
			logger.Error("failed to auto-delete file", "file", info.Path, "err", err)
		} else {
			logger.Info("auto-deleted file after streaming", "file", info.Path)
		}
	}()

	w.Header().Set("Content-Type", info.MimeType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(info.Path)+`"`)
	http.ServeContent(w, r, filepath.Base(info.Path), info.CreatedAt, f)
}

// ─── GET /files ───────────────────────────────────────────────────────────────

// ListFiles lists all files currently in the download store
func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	files, err := h.store.List()
	if err != nil {
		writeError(w, "failed to list files", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"files": files,
		"count": len(files),
	})
}

// ─── GET /health ─────────────────────────────────────────────────────────────

// Health returns a simple liveness check response
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "global-downloader",
	})
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Error("writeJSON encode failed", "err", err)
	}
}

func writeError(w http.ResponseWriter, msg string, code int) {
	writeJSON(w, code, &models.ErrorResponse{Error: msg, Code: code})
}

// extractPathParam gets the segment of path after prefix
func extractPathParam(path, prefix string) string {
	trimmed := strings.TrimPrefix(path, prefix)
	return strings.Trim(trimmed, "/")
}