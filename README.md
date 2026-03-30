# Global Downloader

A high-performance universal video and audio downloader backend built in Go.

This service utilizes `yt-dlp` and `ffmpeg` underneath to support downloading videos, extracting audio, and formatting from thousands of websites (YouTube, Instagram, TikTok, Soundcloud, etc.).

It features an asynchronous job-queue architecture, allowing to handle high loads without blocking API requests, making it suitable for production environments.

## Features

* **Async Worker Pool**: Non-blocking downloads using a concurrent worker system.
* **Job Queue Tracking**: Monitor the real-time progress, status, and metadata of your downloads with unique identifiers (`UUID`).
* **Universal Platform Support**: Uses `yt-dlp` allowing downloads from any supported site.
* **Media Formatting & Audio Extraction**: Select quality (`720p`, `1080p`, etc) and extract high-quality audio (`mp3`, `m4a`, `wav`, etc) using `ffmpeg`.
* **Local Storage Management**: Includes a built-in clean-up routine to automatically delete stale downloads to prevent disks from filling up.
* **Docker Ready**: A minimal, multi-stage Alpine/Python Docker build to keep images slim while bundling the required binaries.

## API Documentation

### 1. Request a Download

**`POST /download`**

Initiate a new video or audio download job.

**Request Body**
```json
{
  "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
  "format": "mp4",          // mp4, mkv, webm, mp3, m4a...
  "quality": "1080p",       // best, 1080p, 720p...
  "audio_only": false       // set to true to force audio extraction
}
```

**Response**
```json
{
  "job": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "url": "https://www.youtube.com/watch?v=...",
    "format": "mp4",
    "quality": "1080p",
    "status": "pending",
    "progress": 0,
    // ...
  },
  "message": "job enqueued"
}
```

### 2. Check Job Status

**`GET /job/{id}`**

Fetch the real-time status and progress of a specific job.

**Response**
```json
{
  "job": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "status": "processing",   // pending -> processing -> completed / failed
    "progress": 45.5,         // Percentage: 0 - 100
    "title": "Rick Astley - Never Gonna Give You Up (Official Music Video)",
    // ...
  }
}
```

### 3. List All Jobs

**`GET /jobs`**

View a snapshot of all tracked jobs currently in the system memory.

### 4. Fetch the File

**`GET /file/{filename}`**

Once a job hits the `completed` status, it will expose `file_path`. You can use the file's basename to stream and save the processed file from the server.

### 5. List Downloaded Files

**`GET /files`**

Lists all generated files currently stored on the server before they are cleaned up.

## Quickstart

### Native (macOS/Linux)

1. **Install Prerequisites (yt-dlp & ffmpeg)**
   ```bash
   ./scripts/install_tools.sh
   ```

2. **Run Service**
   ```bash
   go run ./cmd/api
   ```

### Docker Compose

The easiest way to get everything running out of the box with the necessary runtime binaries:

```bash
docker-compose -f deployments/docker/docker-compose.yml up --build
```

The API will be available at `http://localhost:8080`.

## Architecture Details

* **`internal/models`**: Domain data structures (Job, Request payloads, Responses).
* **`internal/downloader`**: The core interaction boundary with `yt-dlp`, interpreting stdout for live progress tracking.
* **`internal/workers`**: Concurrency pool that pops items off the `internal/queue` to perform long-running file I/O operations asynchronously.
* **`internal/storage`**: Handles file cleanup and lookup.
* **`internal/handlers`**: Translates HTTP requests to internal Service parameters.
