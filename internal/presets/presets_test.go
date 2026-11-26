package presets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mediapulse/internal/ffmpeg"
)

func sampleHW(hasNVENC bool) ffmpeg.HardwareInfo {
	return ffmpeg.HardwareInfo{
		GPUName: "NVIDIA GeForce RTX 3090",
		Encoders: []ffmpeg.EncoderCapability{
			{Name: "h264_nvenc", Codec: "h264", Backend: "nvenc", Available: hasNVENC},
			{Name: "hevc_nvenc", Codec: "hevc", Backend: "nvenc", Available: hasNVENC},
			{Name: "libx264", Codec: "h264", Backend: "cpu", Available: true},
			{Name: "libx265", Codec: "hevc", Backend: "cpu", Available: true},
		},
		Recommended: map[bool]string{true: "nvenc", false: "cpu"}[hasNVENC],
	}
}

func TestCalculateTargetBitrate(t *testing.T) {
	// 30s clip targeting 10 MB should comfortably fit under 10 MB
	b := CalculateTargetBitrate(30, 10, false)
	if b.AudioBitrateKbps != 128 {
		t.Fatalf("expected 128k audio for 30s clip, got %d", b.AudioBitrateKbps)
	}
	// Total budget: 10 * 1024 * 1024 * 8 * 0.95 / 30 / 1000 = ~2656 kbps total -> ~2528 kbps video
	if b.VideoBitrateKbps < 2300 || b.VideoBitrateKbps > 2700 {
		t.Fatalf("unexpected video bitrate for 30s/10MB: %d", b.VideoBitrateKbps)
	}
	if b.Warning != "" {
		t.Fatalf("did not expect warning for 30s/10MB, got %q", b.Warning)
	}

	// Very long 600s clip for 10 MB should clamp at minimum 120 kbps and warn
	longB := CalculateTargetBitrate(600, 10, false)
	if longB.VideoBitrateKbps != 120 {
		t.Fatalf("expected min clamp 120 kbps, got %d", longB.VideoBitrateKbps)
	}
	if longB.Warning == "" {
		t.Fatalf("expected warning for 600s/10MB")
	}

	// Muted clip assigns 0 audio bitrate
	muted := CalculateTargetBitrate(60, 10, true)
	if muted.AudioBitrateKbps != 0 {
		t.Fatalf("expected 0 audio bitrate when muted, got %d", muted.AudioBitrateKbps)
	}
}

func TestAnalyzeTVStrategy(t *testing.T) {
	opts := DefaultOptions(PresetTVSmart)

	// Case 1: H.264 8-bit + AAC MKV -> instant remux
	mkvH264 := ffmpeg.MediaInfo{
		HasVideo:    true,
		VideoCodec:  "h264",
		PixelFormat: "yuv420p",
		HasAudio:    true,
		AudioCodec:  "aac",
	}
	st1 := AnalyzeTVStrategy(mkvH264, opts)
	if st1.Mode != "remux" {
		t.Fatalf("expected remux, got %s (%s)", st1.Mode, st1.Reason)
	}

	// Case 2: H.264 8-bit + DTS audio -> audio_only
	mkvDTS := mkvH264
	mkvDTS.AudioCodec = "dts"
	st2 := AnalyzeTVStrategy(mkvDTS, opts)
	if st2.Mode != "audio_only" {
		t.Fatalf("expected audio_only, got %s", st2.Mode)
	}

	// Case 3: HEVC 10-bit -> full transcode
	mkvHEVC := ffmpeg.MediaInfo{
		HasVideo:    true,
		VideoCodec:  "hevc",
		PixelFormat: "yuv420p10le",
		HasAudio:    true,
		AudioCodec:  "aac",
	}
	st3 := AnalyzeTVStrategy(mkvHEVC, opts)
	if st3.Mode != "transcode" {
		t.Fatalf("expected transcode for HEVC 10-bit, got %s", st3.Mode)
	}

	// Case 4: H.264 with resolution change -> transcode
	opts720 := opts
	opts720.Resolution = "720p"
	st4 := AnalyzeTVStrategy(mkvH264, opts720)
	if st4.Mode != "transcode" {
		t.Fatalf("expected transcode when scaling, got %s", st4.Mode)
	}
}

func TestBuildPlanVariants(t *testing.T) {
	info := ffmpeg.MediaInfo{
		Path:        `C:\clips\gameplay.mkv`,
		FileName:    "gameplay.mkv",
		SizeBytes:   80 * 1024 * 1024,
		DurationSec: 45,
		HasVideo:    true,
		VideoCodec:  "hevc",
		PixelFormat: "yuv420p",
		Width:       1920,
		Height:      1080,
		HasAudio:    true,
		AudioCodec:  "opus",
	}

	// Discord 10MB on RTX NVENC -> single command with multipass fullres
	dOpts := DefaultOptions(PresetDiscord10MB)
	dPlan := BuildPlan(info, dOpts, sampleHW(true), `C:\clips\gameplay_discord10mb.mp4`, "")
	if dPlan.UsesTwoPass {
		t.Fatalf("NVENC should use single command multipass, not 2-pass files")
	}
	if dPlan.EffectiveEncoder != "h264_nvenc" {
		t.Fatalf("expected h264_nvenc, got %s", dPlan.EffectiveEncoder)
	}
	joined := strings.Join(dPlan.SingleArgs, " ")
	if !strings.Contains(joined, "-multipass fullres") {
		t.Fatalf("expected -multipass fullres in NVENC args: %s", joined)
	}

	// Discord 10MB on CPU -> 2-pass args
	cpuPlan := BuildPlan(info, dOpts, sampleHW(false), `C:\clips\gameplay_discord10mb.mp4`, "passlog_1")
	if !cpuPlan.UsesTwoPass || len(cpuPlan.Pass1Args) == 0 || len(cpuPlan.Pass2Args) == 0 {
		t.Fatalf("CPU target size must use 2-pass")
	}
	if cpuPlan.EffectiveEncoder != "libx264" {
		t.Fatalf("expected libx264, got %s", cpuPlan.EffectiveEncoder)
	}

	// WhatsApp -> enforces yuv420p + faststart + aac
	wOpts := DefaultOptions(PresetWhatsApp)
	wOpts.Rotation = 90
	wOpts.Mute = true
	wPlan := BuildPlan(info, wOpts, sampleHW(true), `C:\clips\gameplay_whatsapp.mp4`, "")
	wJoined := strings.Join(wPlan.SingleArgs, " ")
	if !strings.Contains(wJoined, "transpose=1") || !strings.Contains(wJoined, "-an") || !strings.Contains(wJoined, "+faststart") {
		t.Fatalf("WhatsApp args missing expected flags: %s", wJoined)
	}

	// Audio extract -> -vn and libmp3lame
	aOpts := DefaultOptions(PresetAudioOnly)
	aOpts.TrimStartSec = 5
	aOpts.TrimEndSec = 20
	aPlan := BuildPlan(info, aOpts, sampleHW(true), `C:\clips\gameplay_audio.mp3`, "")
	aJoined := strings.Join(aPlan.SingleArgs, " ")
	if !strings.Contains(aJoined, "-vn") || !strings.Contains(aJoined, "libmp3lame") || !strings.Contains(aJoined, "-ss 5.000") {
		t.Fatalf("Audio extract args missing expected flags: %s", aJoined)
	}
	if aPlan.EffectiveDurationSec != 15 {
		t.Fatalf("expected 15s effective duration, got %.2f", aPlan.EffectiveDurationSec)
	}
}

func TestBuildOutputPathDeduplication(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "video.mkv")
	if err := os.WriteFile(inPath, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := DefaultOptions(PresetTVSmart)
	p1 := BuildOutputPath(inPath, opts)
	if filepath.Base(p1) != "video_tv.mp4" {
		t.Fatalf("expected video_tv.mp4, got %s", filepath.Base(p1))
	}
	if err := os.WriteFile(p1, []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}

	p2 := BuildOutputPath(inPath, opts)
	if filepath.Base(p2) != "video_tv_2.mp4" {
		t.Fatalf("expected video_tv_2.mp4, got %s", filepath.Base(p2))
	}
}
