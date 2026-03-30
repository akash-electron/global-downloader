package storage

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"global-downloader/pkg/logger"
)

// FileStore manages downloaded files on local disk
type FileStore struct {
	DownloadDir    string
	MaxFileAgeMins int
}

// FileInfo holds metadata about a stored file
type FileInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
	MimeType  string    `json:"mime_type"`
}

// NewFileStore creates a FileStore rooted at downloadDir
func NewFileStore(downloadDir string, maxAgeMinutes int) (*FileStore, error) {
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create download dir %q: %w", downloadDir, err)
	}
	return &FileStore{
		DownloadDir:    downloadDir,
		MaxFileAgeMins: maxAgeMinutes,
	}, nil
}

// Stat returns metadata for a file identified by its base name
func (s *FileStore) Stat(filename string) (*FileInfo, error) {
	path := filepath.Join(s.DownloadDir, filepath.Base(filename))
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", filename)
		}
		return nil, err
	}

	mime, _ := detectMime(path)
	return &FileInfo{
		Name:      info.Name(),
		Path:      path,
		SizeBytes: info.Size(),
		CreatedAt: info.ModTime(),
		MimeType:  mime,
	}, nil
}

// List returns metadata for all files currently in the store
func (s *FileStore) List() ([]*FileInfo, error) {
	entries, err := os.ReadDir(s.DownloadDir)
	if err != nil {
		return nil, err
	}

	var files []*FileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		path := filepath.Join(s.DownloadDir, e.Name())
		mime, _ := detectMime(path)
		files = append(files, &FileInfo{
			Name:      e.Name(),
			Path:      path,
			SizeBytes: info.Size(),
			CreatedAt: info.ModTime(),
			MimeType:  mime,
		})
	}

	// newest first
	sort.Slice(files, func(i, j int) bool {
		return files[i].CreatedAt.After(files[j].CreatedAt)
	})
	return files, nil
}

// Delete removes a file from the store
func (s *FileStore) Delete(filename string) error {
	path := filepath.Join(s.DownloadDir, filepath.Base(filename))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}

// Cleanup removes files older than MaxFileAgeMins
func (s *FileStore) Cleanup() {
	cutoff := time.Now().Add(-time.Duration(s.MaxFileAgeMins) * time.Minute)
	entries, err := os.ReadDir(s.DownloadDir)
	if err != nil {
		logger.Error("storage cleanup: ReadDir failed", "err", err)
		return
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			path := filepath.Join(s.DownloadDir, e.Name())
			if err := os.Remove(path); err != nil {
				logger.Warn("storage cleanup: remove failed", "path", path, "err", err)
			} else {
				logger.Info("storage cleanup: removed old file", "path", path)
			}
		}
	}
}

// StartCleanupLoop runs Cleanup in the background every intervalMinutes
func (s *FileStore) StartCleanupLoop(intervalMinutes int) {
	go func() {
		t := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
		defer t.Stop()
		for range t.C {
			s.Cleanup()
		}
	}()
}

// detectMime reads the first 512 bytes of a file and returns its MIME type
func detectMime(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "application/octet-stream", err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		return "application/octet-stream", err
	}
	return http.DetectContentType(buf[:n]), nil
}
