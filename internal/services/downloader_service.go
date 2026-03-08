package services

import (
	"fmt"
	"os/exec"
)

func DownloadVideo(url string) (string, error) {

	cmd := exec.Command(
		"yt-dlp",
		"-f", "bv*+ba/b",
		"--merge-output-format", "mp4",
		"--no-playlist",
		"-o", "downloads/%(title)s.%(ext)s",
		url,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp error: %s", string(output))
	}

	return string(output), nil
}