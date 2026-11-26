package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// MediaInfo holds probed metadata for a local audio/video file.
type MediaInfo struct {
	Path            string  `json:"path"`
	FileName        string  `json:"fileName"`
	Extension       string  `json:"extension"`
	SizeBytes       int64   `json:"sizeBytes"`
	DurationSec     float64 `json:"durationSec"`
	BitrateKbps     int     `json:"bitrateKbps"`
	Container       string  `json:"container"`
	HasVideo        bool    `json:"hasVideo"`
	VideoCodec      string  `json:"videoCodec"`
	VideoProfile    string  `json:"videoProfile"`
	PixelFormat     string  `json:"pixelFormat"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	FPS             float64 `json:"fps"`
	BitDepth        int     `json:"bitDepth"`
	HasAudio        bool    `json:"hasAudio"`
	AudioCodec      string  `json:"audioCodec"`
	AudioChannels   int     `json:"audioChannels"`
	AudioSampleRate int     `json:"audioSampleRate"`
	AudioBitrateKbps int    `json:"audioBitrateKbps"`
}

type ffprobeOutput struct {
	Format struct {
		FormatName string `json:"format_name"`
		Duration   string `json:"duration"`
		Size       string `json:"size"`
		BitRate    string `json:"bit_rate"`
	} `json:"format"`
	Streams []struct {
		CodecType      string `json:"codec_type"`
		CodecName      string `json:"codec_name"`
		Profile        string `json:"profile"`
		PixFmt         string `json:"pix_fmt"`
		Width          int    `json:"width"`
		Height         int    `json:"height"`
		AvgFrameRate   string `json:"avg_frame_rate"`
		RFrameRate     string `json:"r_frame_rate"`
		BitsPerRaw     string `json:"bits_per_raw_sample"`
		Channels       int    `json:"channels"`
		SampleRate     string `json:"sample_rate"`
		BitRate        string `json:"bit_rate"`
		Disposition    struct {
			AttachedPic int `json:"attached_pic"`
		} `json:"disposition"`
	} `json:"streams"`
}

// Probe runs ffprobe on mediaPath and extracts normalized stream metadata.
func Probe(ctx context.Context, ffprobePath string, mediaPath string) (MediaInfo, error) {
	if strings.TrimSpace(ffprobePath) == "" {
		return MediaInfo{}, fmt.Errorf("ffprobe no disponible")
	}
	stat, err := os.Stat(mediaPath)
	if err != nil {
		return MediaInfo{}, fmt.Errorf("no se puede acceder al archivo: %w", err)
	}
	if stat.IsDir() {
		return MediaInfo{}, fmt.Errorf("la ruta es una carpeta, no un archivo")
	}

	probeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		probeCtx,
		ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		mediaPath,
	)
	ConfigureCmd(cmd)
	out, err := cmd.Output()
	if err != nil {
		return MediaInfo{}, fmt.Errorf("ffprobe fallo: %w", err)
	}

	var parsed ffprobeOutput
	if err := json.Unmarshal(out, &parsed); err != nil {
		return MediaInfo{}, fmt.Errorf("error parseando metadata de ffprobe: %w", err)
	}

	info := MediaInfo{
		Path:      mediaPath,
		FileName:  filepath.Base(mediaPath),
		Extension: strings.ToLower(strings.TrimPrefix(filepath.Ext(mediaPath), ".")),
		SizeBytes: stat.Size(),
		Container: strings.ToLower(parsed.Format.FormatName),
	}

	if d, err := strconv.ParseFloat(strings.TrimSpace(parsed.Format.Duration), 64); err == nil && d > 0 {
		info.DurationSec = d
	}
	if br, err := strconv.ParseInt(strings.TrimSpace(parsed.Format.BitRate), 10, 64); err == nil && br > 0 {
		info.BitrateKbps = int(br / 1000)
	}

	for _, s := range parsed.Streams {
		switch s.CodecType {
		case "video":
			if s.Disposition.AttachedPic == 1 || info.HasVideo {
				continue
			}
			info.HasVideo = true
			info.VideoCodec = strings.ToLower(strings.TrimSpace(s.CodecName))
			info.VideoProfile = strings.TrimSpace(s.Profile)
			info.PixelFormat = strings.ToLower(strings.TrimSpace(s.PixFmt))
			info.Width = s.Width
			info.Height = s.Height
			info.FPS = parseFrameRate(s.AvgFrameRate)
			if info.FPS <= 0 {
				info.FPS = parseFrameRate(s.RFrameRate)
			}
			info.BitDepth = parseBitDepth(s.BitsPerRaw, info.PixelFormat)
		case "audio":
			if info.HasAudio {
				continue
			}
			info.HasAudio = true
			info.AudioCodec = strings.ToLower(strings.TrimSpace(s.CodecName))
			info.AudioChannels = s.Channels
			if sr, err := strconv.Atoi(strings.TrimSpace(s.SampleRate)); err == nil {
				info.AudioSampleRate = sr
			}
			if br, err := strconv.ParseInt(strings.TrimSpace(s.BitRate), 10, 64); err == nil && br > 0 {
				info.AudioBitrateKbps = int(br / 1000)
			}
		}
	}

	if !info.HasVideo && !info.HasAudio {
		return MediaInfo{}, fmt.Errorf("el archivo no contiene pistas de video ni audio validas")
	}

	return info, nil
}

func parseFrameRate(expr string) float64 {
	expr = strings.TrimSpace(expr)
	if expr == "" || expr == "0/0" {
		return 0
	}
	if strings.Contains(expr, "/") {
		parts := strings.SplitN(expr, "/", 2)
		num, err1 := strconv.ParseFloat(parts[0], 64)
		den, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 == nil && err2 == nil && den > 0 {
			fps := num / den
			if fps > 0 && fps < 1000 {
				return fps
			}
		}
		return 0
	}
	v, err := strconv.ParseFloat(expr, 64)
	if err != nil || v <= 0 || v >= 1000 {
		return 0
	}
	return v
}

func parseBitDepth(rawBits string, pixFmt string) int {
	if b, err := strconv.Atoi(strings.TrimSpace(rawBits)); err == nil && b > 0 {
		return b
	}
	switch {
	case strings.Contains(pixFmt, "12le"), strings.Contains(pixFmt, "12be"):
		return 12
	case strings.Contains(pixFmt, "10le"), strings.Contains(pixFmt, "10be"), strings.Contains(pixFmt, "p010"):
		return 10
	default:
		return 8
	}
}
