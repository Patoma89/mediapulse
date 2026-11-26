package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ProgressUpdate reports real-time conversion telemetry parsed from FFmpeg.
type ProgressUpdate struct {
	Percent     float64 `json:"percent"`
	Speed       string  `json:"speed"`
	OutTimeSec  float64 `json:"outTimeSec"`
	CurrentSize int64   `json:"currentSize"`
	ETASec      float64 `json:"etaSec"`
	Pass        int     `json:"pass"`
}

// ProgressMap maps raw progress fields emitted by FFmpeg `-progress pipe:1`.
type ProgressState struct {
	OutTimeUS  int64
	TotalSize  int64
	Speed      string
	Progress   string
}

// ParseProgressLine parses a single `key=value` line from FFmpeg `-progress pipe:1`.
func ParseProgressLine(line string, state *ProgressState) (completedBlock bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return false
	}
	k := strings.TrimSpace(parts[0])
	v := strings.TrimSpace(parts[1])

	switch k {
	case "out_time_us":
		if us, err := strconv.ParseInt(v, 10, 64); err == nil && us >= 0 {
			state.OutTimeUS = us
		}
	case "out_time_ms":
		// Some FFmpeg builds emit microseconds under out_time_ms; only fallback if out_time_us wasn't seen.
		if state.OutTimeUS == 0 {
			if us, err := strconv.ParseInt(v, 10, 64); err == nil && us >= 0 {
				state.OutTimeUS = us
			}
		}
	case "total_size":
		if sz, err := strconv.ParseInt(v, 10, 64); err == nil && sz >= 0 {
			state.TotalSize = sz
		}
	case "speed":
		if v != "N/A" {
			state.Speed = v
		}
	case "progress":
		state.Progress = v
		return true
	}
	return false
}

// BuildTempOutputPath returns a temporary sibling file path preserving extension for FFmpeg muxer detection.
func BuildTempOutputPath(finalPath string) string {
	dir := filepath.Dir(finalPath)
	ext := filepath.Ext(finalPath)
	base := strings.TrimSuffix(filepath.Base(finalPath), ext)
	return filepath.Join(dir, fmt.Sprintf(".%s.mediapulse.tmp%s", base, ext))
}

// RunArgs executes a single FFmpeg invocation and streams progress callbacks.
func RunArgs(
	ctx context.Context,
	ffmpegPath string,
	args []string,
	durationSec float64,
	pass int,
	passOffset float64,
	passSpan float64,
	onProgress func(ProgressUpdate),
) error {
	if durationSec <= 0 {
		durationSec = 1
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	ConfigureCmd(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("iniciando ffmpeg: %w", err)
	}

	startTime := time.Now()
	state := &ProgressState{}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if ParseProgressLine(line, state) && onProgress != nil {
			outSec := float64(state.OutTimeUS) / 1_000_000.0
			rawRatio := math.Min(1.0, math.Max(0.0, outSec/durationSec))
			if state.Progress == "end" {
				rawRatio = 1.0
			}
			pct := passOffset + rawRatio*passSpan
			if pct > 99.5 && state.Progress != "end" {
				pct = 99.5
			}
			elapsed := time.Since(startTime).Seconds()
			eta := 0.0
			if rawRatio > 0.02 && rawRatio < 1.0 && elapsed > 0.3 {
				eta = (elapsed / rawRatio) * (1.0 - rawRatio)
			}
			onProgress(ProgressUpdate{
				Percent:     pct,
				Speed:       state.Speed,
				OutTimeSec:  outSec,
				CurrentSize: state.TotalSize,
				ETASec:      eta,
				Pass:        pass,
			})
		}
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		errMsg := strings.TrimSpace(stderrBuf.String())
		if len(errMsg) > 400 {
			errMsg = errMsg[len(errMsg)-400:]
		}
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Errorf("ffmpeg error: %s", errMsg)
	}

	return nil
}

// AtomicFinalize moves tempPath to finalPath safely after a successful FFmpeg run.
func AtomicFinalize(tempPath string, finalPath string) error {
	if err := os.Rename(tempPath, finalPath); err == nil {
		return nil
	}
	// Fallback copy+remove if cross-volume or locked rename quirk
	in, err := os.Open(tempPath)
	if err != nil {
		return err
	}
	out, err := os.Create(finalPath)
	if err != nil {
		_ = in.Close()
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = in.Close()
		_ = os.Remove(finalPath)
		return err
	}
	_ = out.Close()
	_ = in.Close()
	_ = os.Remove(tempPath)
	return nil
}

// CleanupPassLogs removes FFmpeg 2-pass log artifacts matching passLogPrefix.
func CleanupPassLogs(passLogPrefix string) {
	if passLogPrefix == "" {
		return
	}
	matches, _ := filepath.Glob(passLogPrefix + "*")
	for _, m := range matches {
		_ = os.Remove(m)
	}
}
