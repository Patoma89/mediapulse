package ffmpeg_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"mediapulse/internal/ffmpeg"
	"mediapulse/internal/presets"
	"mediapulse/internal/queue"
)

func TestParseProgressLine(t *testing.T) {
	st := &ffmpeg.ProgressState{}
	lines := []string{
		"out_time_us=2500000",
		"total_size=123456",
		"speed=12.4x",
		"progress=continue",
	}
	var done bool
	for _, l := range lines {
		done = ffmpeg.ParseProgressLine(l, st)
	}
	if !done {
		t.Fatalf("expected progress=continue to complete block")
	}
	if st.OutTimeUS != 2500000 || st.TotalSize != 123456 || st.Speed != "12.4x" {
		t.Fatalf("unexpected progress state: %+v", st)
	}
}

func TestIntegrationWithRealFFmpeg(t *testing.T) {
	bins := ffmpeg.Locate("")
	if !bins.Ready {
		t.Skipf("FFmpeg binaries not present locally (%s); skipping real integration test", bins.Error)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	hw := ffmpeg.DetectHardware(ctx, bins.FFmpegPath)
	t.Logf("Detected GPU=%q, Recommended=%q", hw.GPUName, hw.Recommended)

	tmpDir := t.TempDir()
	srcMKV := filepath.Join(tmpDir, "sample.mkv")

	// Generate a synthetic 4-second 1280x720 H.264 + AAC MKV test clip
	genCmd := exec.CommandContext(
		ctx,
		bins.FFmpegPath,
		"-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc=duration=4:size=1280x720:rate=30",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=4",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-preset", "ultrafast",
		"-c:a", "aac", "-b:a", "128k",
		"-y", srcMKV,
	)
	ffmpeg.ConfigureCmd(genCmd)
	if out, err := genCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed generating test MKV: %v (%s)", err, string(out))
	}

	info, err := ffmpeg.Probe(ctx, bins.FFprobePath, srcMKV)
	if err != nil {
		t.Fatalf("Probe failed: %v", err)
	}
	if info.VideoCodec != "h264" || info.Width != 1280 || info.Height != 720 || info.AudioCodec != "aac" {
		t.Fatalf("unexpected probe metadata: %+v", info)
	}

	mgr := queue.NewManager(bins, hw, nil, queue.Events{})

	// 1. Smart TV Remux (MKV -> MP4 lossless copy)
	tvOpts := presets.DefaultOptions(presets.PresetTVSmart)
	tvOpts.OutputDir = tmpDir

	// 2. Discord 10 MB
	discOpts := presets.DefaultOptions(presets.PresetDiscord10MB)
	discOpts.OutputDir = tmpDir

	// 3. WhatsApp with trim + rotation + mute
	waOpts := presets.DefaultOptions(presets.PresetWhatsApp)
	waOpts.OutputDir = tmpDir
	waOpts.TrimStartSec = 0.5
	waOpts.TrimEndSec = 2.5
	waOpts.Rotation = 90
	waOpts.Mute = true

	// 4. Audio Extract MP3
	audOpts := presets.DefaultOptions(presets.PresetAudioOnly)
	audOpts.OutputDir = tmpDir
	audOpts.AudioFormat = "mp3"

	added1, errs1 := mgr.AddFiles(ctx, []string{srcMKV}, tvOpts)
	added2, errs2 := mgr.AddFiles(ctx, []string{srcMKV}, discOpts)
	added3, errs3 := mgr.AddFiles(ctx, []string{srcMKV}, waOpts)
	added4, errs4 := mgr.AddFiles(ctx, []string{srcMKV}, audOpts)

	if len(errs1)+len(errs2)+len(errs3)+len(errs4) > 0 || len(added1) != 1 || len(added2) != 1 || len(added3) != 1 || len(added4) != 1 {
		t.Fatalf("AddFiles failures: %v %v %v %v", errs1, errs2, errs3, errs4)
	}

	if added1[0].Plan.TVStrategy == nil || added1[0].Plan.TVStrategy.Mode != "remux" {
		t.Fatalf("expected Smart TV remux strategy, got %+v", added1[0].Plan.TVStrategy)
	}

	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for queue completion
	deadline := time.Now().Add(75 * time.Second)
	for time.Now().Before(deadline) {
		st := mgr.State()
		if !st.Running && st.Pending == 0 {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}

	finalState := mgr.State()
	if finalState.Running || finalState.DoneCount != 4 {
		t.Fatalf("expected 4 completed jobs, got state=%+v", finalState)
	}

	for _, j := range finalState.Jobs {
		if j.Status != queue.StatusDone {
			t.Fatalf("job %s failed: %s", j.Options.PresetID, j.Error)
		}
		st, statErr := os.Stat(j.OutputPath)
		if statErr != nil || st.Size() == 0 {
			t.Fatalf("missing output file %s: %v", j.OutputPath, statErr)
		}
		outInfo, pErr := ffmpeg.Probe(ctx, bins.FFprobePath, j.OutputPath)
		if pErr != nil {
			t.Fatalf("probe output %s failed: %v", j.OutputPath, pErr)
		}

		switch j.Options.PresetID {
		case presets.PresetDiscord10MB:
			if st.Size() >= 10*1024*1024 {
				t.Fatalf("Discord 10MB output exceeded 10MB: %d bytes", st.Size())
			}
		case presets.PresetWhatsApp:
			if outInfo.HasAudio {
				t.Fatalf("expected muted WhatsApp output")
			}
			if outInfo.PixelFormat != "yuv420p" || outInfo.VideoCodec != "h264" {
				t.Fatalf("WhatsApp output incompatible: %+v", outInfo)
			}
			// Rotated 90 deg from 1280x720 -> 720x1280
			if outInfo.Width != 720 || outInfo.Height != 1280 {
				t.Fatalf("expected rotated 720x1280, got %dx%d", outInfo.Width, outInfo.Height)
			}
		case presets.PresetAudioOnly:
			if outInfo.HasVideo || !outInfo.HasAudio || outInfo.AudioCodec != "mp3" {
				t.Fatalf("expected MP3 audio only, got %+v", outInfo)
			}
		}
	}
}
