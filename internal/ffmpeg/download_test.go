package ffmpeg_test

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testing"

	"mediapulse/internal/ffmpeg"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func buildTestFFmpegZip(t *testing.T, includeFFprobe bool) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	ffName := "ffmpeg"
	fpName := "ffprobe"
	if runtime.GOOS == "windows" {
		ffName = "ffmpeg.exe"
		fpName = "ffprobe.exe"
	}

	addEntry := func(name string, content []byte) {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create %s failed: %v", name, err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatalf("zip write %s failed: %v", name, err)
		}
	}

	addEntry("ffmpeg-7.1/bin/"+ffName, []byte("dummy-ffmpeg-binary"))
	if includeFFprobe {
		addEntry("ffmpeg-7.1/bin/"+fpName, []byte("dummy-ffprobe-binary"))
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("zip close failed: %v", err)
	}
	return buf.Bytes()
}

func TestDownloadHappyPathAndStages(t *testing.T) {
	payload := buildTestFFmpegZip(t, true)
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode:    http.StatusOK,
				ContentLength: int64(len(payload)),
				Body:          io.NopCloser(bytes.NewReader(payload)),
				Header:        make(http.Header),
			}, nil
		}),
	}

	destDir := t.TempDir()
	var stages []string
	bins, err := ffmpeg.Download(context.Background(), ffmpeg.DownloadOptions{
		DestDir:        destDir,
		URLs:           []string{"https://mirror.example/ffmpeg.zip"},
		SkipVerifyExec: true,
		HTTPClient:     client,
	}, func(p ffmpeg.DownloadProgress) {
		if len(stages) == 0 || stages[len(stages)-1] != p.Stage {
			stages = append(stages, p.Stage)
		}
	})
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	if !bins.Ready {
		t.Fatalf("expected bins.Ready=true")
	}
	if _, err := os.Stat(bins.FFmpegPath); err != nil {
		t.Fatalf("missing ffmpeg binary: %v", err)
	}
	if _, err := os.Stat(bins.FFprobePath); err != nil {
		t.Fatalf("missing ffprobe binary: %v", err)
	}

	wantStages := []string{"connecting", "downloading", "extracting", "verifying", "done"}
	if len(stages) != len(wantStages) {
		t.Fatalf("unexpected stages %v, want %v", stages, wantStages)
	}
	for i := range wantStages {
		if stages[i] != wantStages[i] {
			t.Fatalf("stage[%d]=%q, want %q (all: %v)", i, stages[i], wantStages[i], stages)
		}
	}
}

func TestDownloadMirrorFallback(t *testing.T) {
	payload := buildTestFFmpegZip(t, true)
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if strings.Contains(r.URL.Host, "bad.example") {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("mirror down")),
					Header:     make(http.Header),
				}, nil
			}
			return &http.Response{
				StatusCode:    http.StatusOK,
				ContentLength: int64(len(payload)),
				Body:          io.NopCloser(bytes.NewReader(payload)),
				Header:        make(http.Header),
			}, nil
		}),
	}

	destDir := t.TempDir()
	bins, err := ffmpeg.Download(context.Background(), ffmpeg.DownloadOptions{
		DestDir:        destDir,
		URLs:           []string{"https://bad.example/ffmpeg.zip", "https://good.example/ffmpeg.zip"},
		SkipVerifyExec: true,
		HTTPClient:     client,
	}, nil)
	if err != nil {
		t.Fatalf("expected mirror fallback to succeed, got: %v", err)
	}
	if !bins.Ready {
		t.Fatalf("expected bins.Ready=true")
	}
}

func TestDownloadCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	destDir := t.TempDir()
	var lastStage string
	_, err := ffmpeg.Download(ctx, ffmpeg.DownloadOptions{
		DestDir:        destDir,
		URLs:           []string{"https://mirror.example/unused.zip"},
		SkipVerifyExec: true,
	}, func(p ffmpeg.DownloadProgress) {
		lastStage = p.Stage
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if lastStage != "cancelled" {
		t.Fatalf("expected cancelled stage, got %q", lastStage)
	}

	entries, _ := os.ReadDir(destDir)
	if len(entries) != 0 {
		t.Fatalf("expected empty destDir on cancellation, found %d entries", len(entries))
	}
}

func TestDownloadMissingBinaryInZip(t *testing.T) {
	payload := buildTestFFmpegZip(t, false)
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode:    http.StatusOK,
				ContentLength: int64(len(payload)),
				Body:          io.NopCloser(bytes.NewReader(payload)),
				Header:        make(http.Header),
			}, nil
		}),
	}

	destDir := t.TempDir()
	_, err := ffmpeg.Download(context.Background(), ffmpeg.DownloadOptions{
		DestDir:        destDir,
		URLs:           []string{"https://mirror.example/partial.zip"},
		SkipVerifyExec: true,
		HTTPClient:     client,
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "no contiene") {
		t.Fatalf("expected missing binary error, got %v", err)
	}

	entries, _ := os.ReadDir(destDir)
	if len(entries) != 0 {
		t.Fatalf("expected clean destDir on missing binary error, found %d entries", len(entries))
	}
}
