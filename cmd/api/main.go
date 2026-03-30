package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"global-downloader/internal/config"
	"global-downloader/internal/downloader"
	"global-downloader/internal/queue"
	"global-downloader/internal/routes"
	"global-downloader/internal/services"
	"global-downloader/internal/storage"
	"global-downloader/internal/workers"
	"global-downloader/pkg/logger"
)

func main() {
	// ── Config ──────────────────────────────────────────────────────────────
	cfg := config.Load()
	logger.Info("config loaded",
		"port", cfg.Port,
		"workers", cfg.MaxWorkers,
		"download_dir", cfg.DownloadDir,
	)

	// ── Storage ──────────────────────────────────────────────────────────────
	fileStore, err := storage.NewFileStore(cfg.DownloadDir, cfg.MaxFileAgeMins)
	if err != nil {
		logger.Error("failed to init file store", "err", err)
		os.Exit(1)
	}
	// Prune stale files every 15 minutes
	fileStore.StartCleanupLoop(15)

	// ── Job Queue ─────────────────────────────────────────────────────────────
	jobStore := queue.NewJobStore(cfg.MaxQueueSize)

	// ── Downloader ────────────────────────────────────────────────────────────
	ytDlp := downloader.NewYtDlp(cfg.YtDlpPath, cfg.DownloadDir)

	// ── Worker Pool ───────────────────────────────────────────────────────────
	pool := workers.NewDownloadWorkerPool(jobStore, ytDlp, cfg.MaxWorkers)
	pool.Start()

	// ── Service ───────────────────────────────────────────────────────────────
	svc := services.NewDownloaderService(jobStore)

	// ── HTTP Server ───────────────────────────────────────────────────────────
	router := routes.SetupRoutes(svc, fileStore)

	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Graceful shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-quit
	logger.Info("server shutting down")
}