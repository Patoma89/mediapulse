package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Binaries describes the located FFmpeg/FFprobe executables and their origin.
type Binaries struct {
	FFmpegPath  string `json:"ffmpegPath"`
	FFprobePath string `json:"ffprobePath"`
	Source      string `json:"source"`
	Version     string `json:"version"`
	Ready       bool   `json:"ready"`
	Error       string `json:"error,omitempty"`
}

func exeNames() (string, string) {
	if runtime.GOOS == "windows" {
		return "ffmpeg.exe", "ffprobe.exe"
	}
	return "ffmpeg", "ffprobe"
}

// Locate searches deterministic locations for ffmpeg and ffprobe executables.
func Locate(customDir string) Binaries {
	ffmpegName, ffprobeName := exeNames()

	type candidate struct {
		dir    string
		source string
	}
	var candidates []candidate

	if strings.TrimSpace(customDir) != "" {
		candidates = append(candidates, candidate{dir: strings.TrimSpace(customDir), source: "custom"})
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			candidate{dir: filepath.Join(exeDir, "bin"), source: "bundled"},
			candidate{dir: exeDir, source: "bundled"},
		)
	}

	if cwd, err := os.Getwd(); err == nil {
		// Walk up a couple of levels so `go test ./...` inside internal/... finds build/ffmpeg.
		dir := cwd
		for i := 0; i < 4; i++ {
			candidates = append(candidates,
				candidate{dir: filepath.Join(dir, "build", "ffmpeg"), source: "dev-build"},
				candidate{dir: filepath.Join(dir, "bin"), source: "dev-bin"},
			)
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		candidates = append(candidates, candidate{
			dir:    filepath.Join(localAppData, "MediaPulse", "bin"),
			source: "appdata",
		})
	}

	seen := make(map[string]bool)
	for _, c := range candidates {
		cleanDir := filepath.Clean(c.dir)
		if seen[cleanDir] {
			continue
		}
		seen[cleanDir] = true

		ff := filepath.Join(cleanDir, ffmpegName)
		fp := filepath.Join(cleanDir, ffprobeName)
		if fileExists(ff) && fileExists(fp) {
			ver, err := readVersion(ff)
			if err == nil {
				return Binaries{
					FFmpegPath:  ff,
					FFprobePath: fp,
					Source:      c.source,
					Version:     ver,
					Ready:       true,
				}
			}
		}
	}

	if ff, err1 := exec.LookPath(ffmpegName); err1 == nil {
		if fp, err2 := exec.LookPath(ffprobeName); err2 == nil {
			ver, err := readVersion(ff)
			if err == nil {
				return Binaries{
					FFmpegPath:  ff,
					FFprobePath: fp,
					Source:      "path",
					Version:     ver,
					Ready:       true,
				}
			}
		}
	}

	return Binaries{
		Ready: false,
		Error: fmt.Sprintf("No se encontro %s y %s en bin/ ni en PATH", ffmpegName, ffprobeName),
	}
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func readVersion(ffmpegPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, ffmpegPath, "-version")
	ConfigureCmd(cmd)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return "unknown", nil
	}
	first := strings.TrimSpace(lines[0])
	// Example: "ffmpeg version 7.1-essentials_build-www.gyan.dev Copyright ..."
	if idx := strings.Index(first, " Copyright"); idx > 0 {
		first = first[:idx]
	}
	first = strings.TrimPrefix(first, "ffmpeg version ")
	return strings.TrimSpace(first), nil
}
