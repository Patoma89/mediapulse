package ffmpeg

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// EncoderCapability describes a single video encoder and whether it passed a live probe.
type EncoderCapability struct {
	Name      string `json:"name"`
	Codec     string `json:"codec"` // h264 | hevc
	Backend   string `json:"backend"` // nvenc | amf | qsv | cpu
	Label     string `json:"label"`
	Available bool   `json:"available"`
}

// HardwareInfo summarizes GPU detection and available hardware/CPU encoders.
type HardwareInfo struct {
	GPUName     string              `json:"gpuName"`
	Encoders    []EncoderCapability `json:"encoders"`
	Recommended string              `json:"recommended"` // nvenc | amf | qsv | cpu
}

// DetectHardware probes the system GPU and tests which FFmpeg hardware encoders actually work.
func DetectHardware(ctx context.Context, ffmpegPath string) HardwareInfo {
	gpuName := detectGPUName(ctx)

	specs := []EncoderCapability{
		{Name: "h264_nvenc", Codec: "h264", Backend: "nvenc", Label: "NVIDIA NVENC H.264"},
		{Name: "hevc_nvenc", Codec: "hevc", Backend: "nvenc", Label: "NVIDIA NVENC HEVC"},
		{Name: "h264_amf", Codec: "h264", Backend: "amf", Label: "AMD AMF H.264"},
		{Name: "hevc_amf", Codec: "hevc", Backend: "amf", Label: "AMD AMF HEVC"},
		{Name: "h264_qsv", Codec: "h264", Backend: "qsv", Label: "Intel QSV H.264"},
		{Name: "hevc_qsv", Codec: "hevc", Backend: "qsv", Label: "Intel QSV HEVC"},
		{Name: "libx264", Codec: "h264", Backend: "cpu", Label: "CPU x264", Available: true},
		{Name: "libx265", Codec: "hevc", Backend: "cpu", Label: "CPU x265", Available: true},
	}

	if strings.TrimSpace(ffmpegPath) != "" {
		var wg sync.WaitGroup
		for i := range specs {
			if specs[i].Backend == "cpu" {
				continue
			}
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				specs[idx].Available = testEncoder(ctx, ffmpegPath, specs[idx].Name)
			}(i)
		}
		wg.Wait()
	}

	recommended := "cpu"
	has := func(name string) bool {
		for _, e := range specs {
			if e.Name == name && e.Available {
				return true
			}
		}
		return false
	}

	switch {
	case has("h264_nvenc"):
		recommended = "nvenc"
	case has("h264_amf"):
		recommended = "amf"
	case has("h264_qsv"):
		recommended = "qsv"
	}

	return HardwareInfo{
		GPUName:     gpuName,
		Encoders:    specs,
		Recommended: recommended,
	}
}

// HasEncoder reports whether a specific FFmpeg encoder name is available.
func (h HardwareInfo) HasEncoder(name string) bool {
	for _, e := range h.Encoders {
		if e.Name == name && e.Available {
			return true
		}
	}
	return false
}

func testEncoder(ctx context.Context, ffmpegPath string, encoder string) bool {
	probeCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		probeCtx,
		ffmpegPath,
		"-hide_banner",
		"-loglevel", "error",
		"-f", "lavfi",
		"-i", "color=c=black:s=256x256:d=0.1",
		"-frames:v", "1",
		"-pix_fmt", "yuv420p",
		"-c:v", encoder,
		"-f", "null",
		"-",
	)
	ConfigureCmd(cmd)
	return cmd.Run() == nil
}

func detectGPUName(ctx context.Context) string {
	nvCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(nvCtx, "nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
	ConfigureCmd(cmd)
	if out, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
			return strings.TrimSpace(lines[0])
		}
	}

	if runtime.GOOS == "windows" {
		wmiCtx, wmiCancel := context.WithTimeout(ctx, 3*time.Second)
		defer wmiCancel()
		ps := exec.CommandContext(
			wmiCtx,
			"powershell",
			"-NoProfile",
			"-NonInteractive",
			"-Command",
			"(Get-CimInstance Win32_VideoController | Where-Object { $_.Name -notmatch 'Microsoft Basic' } | Select-Object -First 1 -ExpandProperty Name)",
		)
		ConfigureCmd(ps)
		if out, err := ps.Output(); err == nil {
			name := strings.TrimSpace(string(out))
			if name != "" {
				return name
			}
		}
	}

	return "CPU"
}
