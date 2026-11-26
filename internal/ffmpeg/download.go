package ffmpeg

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// DownloadProgress reports real-time status for an opt-in FFmpeg download.
type DownloadProgress struct {
	Stage           string  `json:"stage"` // idle | connecting | downloading | extracting | verifying | done | error | cancelled
	Percent         float64 `json:"percent"`
	DownloadedBytes int64   `json:"downloadedBytes"`
	TotalBytes      int64   `json:"totalBytes"`
	Source          string  `json:"source,omitempty"`
	DestDir         string  `json:"destDir,omitempty"`
	Error           string  `json:"error,omitempty"`
}

// DefaultDownloadURLs lists official static Windows x64 FFmpeg release archives.
var DefaultDownloadURLs = []string{
	"https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip",
	"https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip",
}

// DownloadOptions configures opt-in FFmpeg/FFprobe download and extraction.
type DownloadOptions struct {
	DestDir        string
	URLs           []string
	UserAgent      string
	MinZipBytes    int64
	SkipVerifyExec bool
	HTTPClient     *http.Client
}

// DefaultDownloadDir returns %LOCALAPPDATA%\MediaPulse\bin (or user cache fallback) and creates it.
func DefaultDownloadDir() (string, error) {
	var dir string
	if localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); localAppData != "" {
		dir = filepath.Join(localAppData, "MediaPulse", "bin")
	} else {
		cacheDir, err := os.UserCacheDir()
		if err != nil || strings.TrimSpace(cacheDir) == "" {
			cacheDir = os.TempDir()
		}
		dir = filepath.Join(cacheDir, "MediaPulse", "bin")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// Download fetches FFmpeg and FFprobe from configured mirrors into DestDir when explicitly requested by the user.
func Download(ctx context.Context, opts DownloadOptions, onProgress func(DownloadProgress)) (Binaries, error) {
	emit := func(p DownloadProgress) {
		if onProgress != nil {
			onProgress(p)
		}
	}

	destDir := strings.TrimSpace(opts.DestDir)
	if destDir == "" {
		var err error
		destDir, err = DefaultDownloadDir()
		if err != nil {
			emit(DownloadProgress{Stage: "error", Error: err.Error()})
			return Binaries{}, fmt.Errorf("no se pudo crear carpeta destino: %w", err)
		}
	} else if err := os.MkdirAll(destDir, 0o755); err != nil {
		emit(DownloadProgress{Stage: "error", DestDir: destDir, Error: err.Error()})
		return Binaries{}, fmt.Errorf("no se pudo crear carpeta destino: %w", err)
	}

	urls := opts.URLs
	minBytes := opts.MinZipBytes
	if len(urls) == 0 {
		urls = DefaultDownloadURLs
		if minBytes <= 0 {
			minBytes = 5 * 1024 * 1024
		}
	}

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: 30 * time.Second,
			},
		}
	}

	userAgent := strings.TrimSpace(opts.UserAgent)
	if userAgent == "" {
		userAgent = "MediaPulse"
	}

	var zipPath string
	var lastErr error

	for _, rawURL := range urls {
		if ctx.Err() != nil {
			emit(DownloadProgress{Stage: "cancelled", DestDir: destDir})
			return Binaries{}, ctx.Err()
		}

		host := rawURL
		if parsed, err := url.Parse(rawURL); err == nil && parsed.Host != "" {
			host = parsed.Host
		}

		emit(DownloadProgress{
			Stage:   "connecting",
			Percent: 0,
			Source:  host,
			DestDir: destDir,
		})

		downloadedPath, err := downloadArchiveToTemp(ctx, client, rawURL, userAgent, host, destDir, minBytes, emit)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
				emit(DownloadProgress{Stage: "cancelled", Source: host, DestDir: destDir})
				return Binaries{}, err
			}
			lastErr = err
			continue
		}
		zipPath = downloadedPath
		break
	}

	if zipPath == "" {
		if lastErr == nil {
			lastErr = fmt.Errorf("no hay URLs de descarga disponibles")
		}
		emit(DownloadProgress{Stage: "error", DestDir: destDir, Error: lastErr.Error()})
		return Binaries{}, lastErr
	}
	defer func() {
		_ = os.Remove(zipPath)
	}()

	emit(DownloadProgress{
		Stage:   "extracting",
		Percent: 95,
		DestDir: destDir,
	})

	ffmpegName, ffprobeName := exeNames()
	ffmpegFinal, ffprobeFinal, err := extractBinariesFromZip(ctx, zipPath, destDir, ffmpegName, ffprobeName)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			emit(DownloadProgress{Stage: "cancelled", DestDir: destDir})
			return Binaries{}, err
		}
		emit(DownloadProgress{Stage: "error", DestDir: destDir, Error: err.Error()})
		return Binaries{}, err
	}

	emit(DownloadProgress{
		Stage:   "verifying",
		Percent: 98,
		DestDir: destDir,
	})

	ver := "downloaded"
	if !opts.SkipVerifyExec {
		readVer, verErr := readVersion(ffmpegFinal)
		if verErr != nil {
			_ = os.Remove(ffmpegFinal)
			_ = os.Remove(ffprobeFinal)
			errMsg := fmt.Sprintf("el binario descargado fallo la verificacion: %v", verErr)
			emit(DownloadProgress{Stage: "error", DestDir: destDir, Error: errMsg})
			return Binaries{}, fmt.Errorf("%s", errMsg)
		}
		ver = readVer
	}

	emit(DownloadProgress{
		Stage:   "done",
		Percent: 100,
		DestDir: destDir,
	})

	return Binaries{
		FFmpegPath:  ffmpegFinal,
		FFprobePath: ffprobeFinal,
		Source:      "appdata",
		Version:     ver,
		Ready:       true,
	}, nil
}

func downloadArchiveToTemp(
	ctx context.Context,
	client *http.Client,
	rawURL string,
	userAgent string,
	host string,
	destDir string,
	minBytes int64,
	emit func(DownloadProgress),
) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("servidor %s devolvio HTTP %d", host, resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "mediapulse_ffmpeg_*.zip")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()

	cleanup := true
	defer func() {
		_ = tmpFile.Close()
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 64*1024)
	lastEmit := time.Time{}

	for {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
				return "", writeErr
			}
			downloaded += int64(n)
			if time.Since(lastEmit) >= 120*time.Millisecond {
				lastEmit = time.Now()
				pct := 0.0
				if total > 0 {
					pct = (float64(downloaded) / float64(total)) * 90.0
					if pct > 90 {
						pct = 90
					}
				}
				emit(DownloadProgress{
					Stage:           "downloading",
					Percent:         pct,
					DownloadedBytes: downloaded,
					TotalBytes:      total,
					Source:          host,
					DestDir:         destDir,
				})
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return "", readErr
		}
	}

	if minBytes > 0 && downloaded < minBytes {
		return "", fmt.Errorf("archivo descargado demasiado pequeno (%d bytes)", downloaded)
	}

	if err := tmpFile.Close(); err != nil {
		return "", err
	}

	emit(DownloadProgress{
		Stage:           "downloading",
		Percent:         90,
		DownloadedBytes: downloaded,
		TotalBytes:      total,
		Source:          host,
		DestDir:         destDir,
	})

	cleanup = false
	return tmpPath, nil
}

func extractBinariesFromZip(ctx context.Context, zipPath, destDir, ffmpegName, ffprobeName string) (string, string, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", "", fmt.Errorf("archivo zip invalido: %w", err)
	}
	defer zr.Close()

	var ffmpegFinal string
	var ffprobeFinal string
	var tmpFiles []string

	cleanupTmps := func() {
		for _, p := range tmpFiles {
			_ = os.Remove(p)
		}
	}

	targets := map[string]*string{
		strings.ToLower(ffmpegName):  &ffmpegFinal,
		strings.ToLower(ffprobeName): &ffprobeFinal,
	}

	for _, f := range zr.File {
		if ctx.Err() != nil {
			cleanupTmps()
			return "", "", ctx.Err()
		}
		if f.FileInfo().IsDir() {
			continue
		}

		base := path.Base(strings.ReplaceAll(f.Name, `\`, "/"))
		destPtr, wanted := targets[strings.ToLower(base)]
		if !wanted || *destPtr != "" {
			continue
		}

		targetName := ffmpegName
		if strings.EqualFold(base, ffprobeName) {
			targetName = ffprobeName
		}

		finalPath := filepath.Join(destDir, targetName)
		tmpPath := filepath.Join(destDir, targetName+".tmp")
		tmpFiles = append(tmpFiles, tmpPath)

		if err := extractSingleZipEntry(ctx, f, tmpPath); err != nil {
			cleanupTmps()
			return "", "", err
		}

		_ = os.Remove(finalPath)
		if err := os.Rename(tmpPath, finalPath); err != nil {
			cleanupTmps()
			return "", "", err
		}
		*destPtr = finalPath
	}

	if ffmpegFinal == "" || ffprobeFinal == "" {
		if ffmpegFinal != "" {
			_ = os.Remove(ffmpegFinal)
		}
		if ffprobeFinal != "" {
			_ = os.Remove(ffprobeFinal)
		}
		cleanupTmps()
		return "", "", fmt.Errorf("el archivo comprimido no contiene %s y %s", ffmpegName, ffprobeName)
	}

	return ffmpegFinal, ffprobeFinal, nil
}

func extractSingleZipEntry(ctx context.Context, f *zip.File, tmpPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}

	buf := make([]byte, 64*1024)
	for {
		if ctx.Err() != nil {
			_ = out.Close()
			_ = os.Remove(tmpPath)
			return ctx.Err()
		}
		n, readErr := rc.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				_ = out.Close()
				_ = os.Remove(tmpPath)
				return writeErr
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			_ = out.Close()
			_ = os.Remove(tmpPath)
			return readErr
		}
	}

	return out.Close()
}
