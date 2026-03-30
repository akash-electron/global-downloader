#!/usr/bin/env bash
# install_tools.sh — installs yt-dlp and ffmpeg on the host machine
set -euo pipefail

OS="$(uname -s)"
ARCH="$(uname -m)"

echo "==> Detected OS: $OS / Arch: $ARCH"

# ── yt-dlp ──────────────────────────────────────────────────────────────────
install_ytdlp() {
  echo "==> Installing yt-dlp..."
  if command -v yt-dlp &>/dev/null; then
    echo "    yt-dlp already installed: $(yt-dlp --version)"
    echo "    Updating to latest..."
    yt-dlp -U || true
    return
  fi

  case "$OS" in
    Linux)
      curl -fsSL https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp \
        -o /usr/local/bin/yt-dlp
      chmod +x /usr/local/bin/yt-dlp
      ;;
    Darwin)
      if command -v brew &>/dev/null; then
        brew install yt-dlp
      else
        curl -fsSL https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos \
          -o /usr/local/bin/yt-dlp
        chmod +x /usr/local/bin/yt-dlp
      fi
      ;;
    *)
      echo "ERROR: Unsupported OS: $OS. Install yt-dlp manually from https://github.com/yt-dlp/yt-dlp"
      exit 1
      ;;
  esac

  echo "    yt-dlp installed: $(yt-dlp --version)"
}

# ── ffmpeg ───────────────────────────────────────────────────────────────────
install_ffmpeg() {
  echo "==> Installing ffmpeg..."
  if command -v ffmpeg &>/dev/null; then
    echo "    ffmpeg already installed: $(ffmpeg -version 2>&1 | head -1)"
    return
  fi

  case "$OS" in
    Linux)
      if command -v apt-get &>/dev/null; then
        apt-get update -qq && apt-get install -y ffmpeg
      elif command -v yum &>/dev/null; then
        yum install -y ffmpeg
      elif command -v apk &>/dev/null; then
        apk add --no-cache ffmpeg
      else
        echo "ERROR: Cannot detect package manager. Install ffmpeg manually."
        exit 1
      fi
      ;;
    Darwin)
      if command -v brew &>/dev/null; then
        brew install ffmpeg
      else
        echo "ERROR: Homebrew not found. Install it from https://brew.sh then run: brew install ffmpeg"
        exit 1
      fi
      ;;
    *)
      echo "ERROR: Unsupported OS: $OS. Install ffmpeg manually from https://ffmpeg.org"
      exit 1
      ;;
  esac

  echo "    ffmpeg installed: $(ffmpeg -version 2>&1 | head -1)"
}

install_ytdlp
install_ffmpeg

echo ""
echo "✅  All tools installed successfully."
echo "    yt-dlp : $(yt-dlp --version)"
echo "    ffmpeg : $(ffmpeg -version 2>&1 | head -1)"
