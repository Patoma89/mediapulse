package presets

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"mediapulse/internal/ffmpeg"
)

// PresetID enumerates built-in conversion profiles.
type PresetID string

const (
	PresetDiscord10MB PresetID = "discord_10mb"
	PresetDiscord25MB PresetID = "discord_25mb"
	PresetWhatsApp    PresetID = "whatsapp"
	PresetTVSmart     PresetID = "tv_smart"
	PresetTargetSize  PresetID = "target_size"
	PresetCompressHQ  PresetID = "compress_hq"
	PresetAudioOnly   PresetID = "audio_extract"
	PresetCustom      PresetID = "custom"
)

// PresetDefinition describes a preset for the UI catalog.
type PresetDefinition struct {
	ID             PresetID `json:"id"`
	NameES         string   `json:"nameEs"`
	NameEN         string   `json:"nameEn"`
	DescES         string   `json:"descEs"`
	DescEN         string   `json:"descEn"`
	Badge          string   `json:"badge"`
	DefaultFormat  string   `json:"defaultFormat"`
	DefaultSizeMB  float64  `json:"defaultSizeMb"`
	Accent         string   `json:"accent"`
}

// ConvertOptions configures how a queued item is processed.
type ConvertOptions struct {
	PresetID         PresetID `json:"presetId"`
	TargetSizeMB     float64  `json:"targetSizeMb"`
	OutputFormat     string   `json:"outputFormat"`     // mp4 | mkv | mov | webm | mp3 | m4a | wav
	VideoCodec       string   `json:"videoCodec"`       // auto | h264 | hevc | copy
	QualityCRF       int      `json:"qualityCrf"`       // 16..34 (default 23)
	Resolution       string   `json:"resolution"`       // original | 1080p | 720p | 480p
	FPS              int      `json:"fps"`              // 0 = original, 60, 30, 24
	Rotation         int      `json:"rotation"`         // 0, 90, 180, 270
	Mute             bool     `json:"mute"`
	TrimStartSec     float64  `json:"trimStartSec"`
	TrimEndSec       float64  `json:"trimEndSec"`
	AudioFormat      string   `json:"audioFormat"`      // mp3 | m4a | wav
	AudioBitrateKbps int      `json:"audioBitrateKbps"` // 0 = auto
	EncoderMode      string   `json:"encoderMode"`      // auto | nvenc | amf | qsv | cpu
	OutputDir        string   `json:"outputDir"`
}

// TVStrategy explains what the Smart TV preset will do with the input file.
type TVStrategy struct {
	Mode                string `json:"mode"` // remux | audio_only | transcode
	Reason              string `json:"reason"`
	EstimatedSpeedLabel string `json:"estimatedSpeedLabel"`
}

// BitrateBudget holds calculated video/audio bitrates for size-constrained outputs.
type BitrateBudget struct {
	VideoBitrateKbps int    `json:"videoBitrateKbps"`
	AudioBitrateKbps int    `json:"audioBitrateKbps"`
	SuggestedHeight  int    `json:"suggestedHeight"`
	Warning          string `json:"warning,omitempty"`
}

// ConversionPlan is the resolved execution plan passed to FFmpeg and previewed in UI.
type ConversionPlan struct {
	OutputPath           string      `json:"outputPath"`
	Extension            string      `json:"extension"`
	UsesTwoPass          bool        `json:"usesTwoPass"`
	EffectiveEncoder     string      `json:"effectiveEncoder"`
	EncoderBackend       string      `json:"encoderBackend"` // copy | nvenc | amf | qsv | cpu | audio
	TVStrategy           *TVStrategy `json:"tvStrategy,omitempty"`
	VideoBitrateKbps     int         `json:"videoBitrateKbps"`
	AudioBitrateKbps     int         `json:"audioBitrateKbps"`
	EffectiveDurationSec float64     `json:"effectiveDurationSec"`
	EstimatedSizeMB      float64     `json:"estimatedSizeMb"`
	Warning              string      `json:"warning,omitempty"`
	Pass1Args            []string    `json:"pass1Args,omitempty"`
	Pass2Args            []string    `json:"pass2Args,omitempty"`
	SingleArgs           []string    `json:"singleArgs,omitempty"`
}

// DefaultOptions returns sensible defaults for a given preset.
func DefaultOptions(preset PresetID) ConvertOptions {
	opts := ConvertOptions{
		PresetID:     preset,
		OutputFormat: "mp4",
		VideoCodec:   "auto",
		QualityCRF:   23,
		Resolution:   "original",
		FPS:          0,
		Rotation:     0,
		AudioFormat:  "mp3",
		EncoderMode:  "auto",
	}
	switch preset {
	case PresetDiscord10MB:
		opts.TargetSizeMB = 10
	case PresetDiscord25MB:
		opts.TargetSizeMB = 25
	case PresetTargetSize:
		opts.TargetSizeMB = 50
	case PresetAudioOnly:
		opts.OutputFormat = "mp3"
		opts.AudioFormat = "mp3"
		opts.AudioBitrateKbps = 192
	case PresetCompressHQ:
		opts.QualityCRF = 22
	}
	return opts
}

// Catalog returns all built-in presets for UI display.
func Catalog() []PresetDefinition {
	return []PresetDefinition{
		{
			ID:            PresetDiscord10MB,
			NameES:        "Discord 10 MB",
			NameEN:        "Discord 10 MB",
			DescES:        "Calcula bitrate exacto con margen de seguridad para entrar debajo de 10 MB sin Nitro.",
			DescEN:        "Calculates exact bitrate with safety headroom to stay strictly under 10 MB without Nitro.",
			Badge:         "< 10 MB",
			DefaultFormat: "mp4",
			DefaultSizeMB: 10,
			Accent:        "violet",
		},
		{
			ID:            PresetDiscord25MB,
			NameES:        "Discord 25 MB",
			NameEN:        "Discord 25 MB",
			DescES:        "Ajuste automático para límite de 25 MB (Nitro Basic / Boost nivel 1).",
			DescEN:        "Auto-targets under 25 MB with balanced audio and crisp H.264/NVENC video.",
			Badge:         "< 25 MB",
			DefaultFormat: "mp4",
			DefaultSizeMB: 25,
			Accent:        "violet",
		},
		{
			ID:            PresetWhatsApp,
			NameES:        "WhatsApp",
			NameEN:        "WhatsApp",
			DescES:        "MP4 H.264 Main + yuv420p + AAC + faststart. Cero pantallas negras o errores al enviar.",
			DescEN:        "MP4 H.264 Main + yuv420p + AAC + faststart. Prevents black frames and upload errors.",
			Badge:         "MP4 Safe",
			DefaultFormat: "mp4",
			Accent:        "emerald",
		},
		{
			ID:            PresetTVSmart,
			NameES:        "TV Smart (MKV → MP4)",
			NameEN:        "Smart TV (MKV → MP4)",
			DescES:        "Si el MKV ya es H.264 hace remux en ~2 segundos sin perder calidad. Si no, recodifica compatible.",
			DescEN:        "Instant lossless remux in ~2s when MKV already contains H.264; transcodes only when needed.",
			Badge:         "Smart Remux",
			DefaultFormat: "mp4",
			Accent:        "cyan",
		},
		{
			ID:            PresetTargetSize,
			NameES:        "Tamaño Exacto",
			NameEN:        "Target Size",
			DescES:        "Elegís cuántos MB querés que pese el video final (ej. 50 MB, 100 MB).",
			DescEN:        "Pick the exact output size in MB (e.g. 50 MB, 100 MB) and MediaPulse handles the bitrate.",
			Badge:         "MB Exacto",
			DefaultFormat: "mp4",
			DefaultSizeMB: 50,
			Accent:        "amber",
		},
		{
			ID:            PresetCompressHQ,
			NameES:        "Comprimir HQ",
			NameEN:        "Compress HQ",
			DescES:        "Reduce archivos pesados manteniendo nitidez visual con NVENC HQ o x264.",
			DescEN:        "Shrinks large recordings while preserving visual sharpness using NVENC HQ or x264.",
			Badge:         "Calidad HQ",
			DefaultFormat: "mp4",
			Accent:        "cyan",
		},
		{
			ID:            PresetAudioOnly,
			NameES:        "Extraer Audio",
			NameEN:        "Extract Audio",
			DescES:        "Extrae la pista de audio en MP3, M4A (AAC) o WAV sin recodificar video.",
			DescEN:        "Extracts the audio track cleanly to MP3, M4A (AAC) or WAV.",
			Badge:         "MP3 / WAV",
			DefaultFormat: "mp3",
			Accent:        "rose",
		},
		{
			ID:            PresetCustom,
			NameES:        "Manual / Custom",
			NameEN:        "Custom",
			DescES:        "Control total de formato (MP4/MKV/MOV/WEBM), códec H.264/HEVC/Copiar y calidad.",
			DescEN:        "Full control over container (MP4/MKV/MOV/WEBM), codec H.264/HEVC/Copy and CRF.",
			Badge:         "Pro",
			DefaultFormat: "mp4",
			Accent:        "slate",
		},
	}
}

// EffectiveDuration returns the clip duration taking TrimStartSec and TrimEndSec into account.
func EffectiveDuration(info ffmpeg.MediaInfo, opts ConvertOptions) float64 {
	dur := info.DurationSec
	if dur <= 0 {
		dur = 1
	}
	start := math.Max(0, opts.TrimStartSec)
	end := opts.TrimEndSec
	if end <= 0 || end > dur {
		end = dur
	}
	if end <= start {
		return math.Max(0.5, dur-start)
	}
	return math.Max(0.5, end-start)
}

// AnalyzeTVStrategy evaluates whether a file can be remuxed losslessly or needs audio/full transcode for TVs.
func AnalyzeTVStrategy(info ffmpeg.MediaInfo, opts ConvertOptions) TVStrategy {
	hasVideoTransform := (opts.Resolution != "" && opts.Resolution != "original") ||
		opts.FPS > 0 ||
		opts.Rotation != 0

	if !info.HasVideo {
		return TVStrategy{
			Mode:                "transcode",
			Reason:              "Sin pista de video; se generara salida estandar.",
			EstimatedSpeedLabel: "Rapido",
		}
	}

	isH264 := info.VideoCodec == "h264"
	is8Bit420 := info.PixelFormat == "" || info.PixelFormat == "yuv420p" || info.PixelFormat == "yuvj420p"

	if isH264 && is8Bit420 && !hasVideoTransform {
		if !info.HasAudio || opts.Mute || info.AudioCodec == "aac" || info.AudioCodec == "mp3" {
			return TVStrategy{
				Mode:                "remux",
				Reason:              "Video H.264 8-bit y audio compatible: copia directa a MP4 sin recodificar (0% perdida).",
				EstimatedSpeedLabel: "~2 seg (Remux)",
			}
		}
		return TVStrategy{
			Mode:                "audio_only",
			Reason:              fmt.Sprintf("Video H.264 compatible (%s), solo se convierte audio (%s → AAC) sin tocar el video.", info.VideoCodec, strings.ToUpper(info.AudioCodec)),
			EstimatedSpeedLabel: "Ultra rapido (Audio AAC)",
		}
	}

	reason := "Recodificacion completa a H.264 + AAC MP4 para maxima compatibilidad en TVs."
	if hasVideoTransform {
		reason = "Hay cambios de resolucion/FPS/rotacion; se recodifica a H.264 + AAC compatible con TV."
	} else if info.VideoCodec != "" {
		reason = fmt.Sprintf("Codec origen %s (%s): se recodifica a H.264 8-bit + AAC para compatibilidad universal en TV.", strings.ToUpper(info.VideoCodec), info.PixelFormat)
	}

	return TVStrategy{
		Mode:                "transcode",
		Reason:              reason,
		EstimatedSpeedLabel: "GPU / Recodificado",
	}
}

// CalculateTargetBitrate computes safe video & audio bitrates to fit comfortably under targetMB.
func CalculateTargetBitrate(durationSec float64, targetMB float64, mute bool) BitrateBudget {
	if durationSec <= 0.5 {
		durationSec = 1
	}
	if targetMB <= 0.5 {
		targetMB = 10
	}

	// 0.95 safety factor accounts for MP4 moov/muxer overhead and bitrate overshoot.
	totalBits := targetMB * 1024 * 1024 * 8 * 0.95
	totalKbps := totalBits / durationSec / 1000.0

	audioKbps := 0
	if !mute {
		switch {
		case durationSec <= 90:
			audioKbps = 128
		case durationSec <= 180:
			audioKbps = 96
		case durationSec <= 360:
			audioKbps = 64
		default:
			audioKbps = 48
		}
		// Ensure audio doesn't consume more than 35% of budget on very long clips.
		if float64(audioKbps) > totalKbps*0.35 {
			audioKbps = int(math.Max(32, math.Floor(totalKbps*0.30)))
		}
	}

	videoKbps := int(math.Floor(totalKbps)) - audioKbps
	var warning string
	if videoKbps < 120 {
		videoKbps = 120
		warning = fmt.Sprintf("Duracion larga (%.0fs) para %.0f MB: calidad muy comprimida o recorta el clip.", durationSec, targetMB)
	} else if videoKbps < 350 {
		warning = fmt.Sprintf("Bitrate ajustado (%d kbps) para entrar en %.0f MB; considera bajar a 480p o recortar.", videoKbps, targetMB)
	}

	suggestedHeight := 0
	if videoKbps < 600 {
		suggestedHeight = 480
	} else if videoKbps < 1500 {
		suggestedHeight = 720
	}

	return BitrateBudget{
		VideoBitrateKbps: videoKbps,
		AudioBitrateKbps: audioKbps,
		SuggestedHeight:  suggestedHeight,
		Warning:          warning,
	}
}

// BuildOutputPath generates a safe, non-colliding output path next to the source or in opts.OutputDir.
func BuildOutputPath(inputPath string, opts ConvertOptions) string {
	dir := strings.TrimSpace(opts.OutputDir)
	if dir == "" {
		dir = filepath.Dir(inputPath)
	}
	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	ext := resolveExtension(opts)
	suffix := presetSuffix(opts)

	candidate := filepath.Join(dir, fmt.Sprintf("%s%s.%s", base, suffix, ext))
	if !pathExistsOrSame(candidate, inputPath) {
		return candidate
	}
	for i := 2; i < 1000; i++ {
		alt := filepath.Join(dir, fmt.Sprintf("%s%s_%d.%s", base, suffix, i, ext))
		if !pathExistsOrSame(alt, inputPath) {
			return alt
		}
	}
	return filepath.Join(dir, fmt.Sprintf("%s%s_out.%s", base, suffix, ext))
}

func pathExistsOrSame(candidate string, inputPath string) bool {
	if strings.EqualFold(filepath.Clean(candidate), filepath.Clean(inputPath)) {
		return true
	}
	_, err := os.Stat(candidate)
	return err == nil
}

func resolveExtension(opts ConvertOptions) string {
	switch opts.PresetID {
	case PresetDiscord10MB, PresetDiscord25MB, PresetWhatsApp, PresetTVSmart, PresetTargetSize:
		return "mp4"
	case PresetAudioOnly:
		af := strings.ToLower(strings.TrimSpace(opts.AudioFormat))
		switch af {
		case "m4a", "wav", "mp3":
			return af
		default:
			return "mp3"
		}
	default:
		f := strings.ToLower(strings.TrimSpace(opts.OutputFormat))
		switch f {
		case "mp4", "mkv", "mov", "webm", "mp3", "m4a", "wav":
			return f
		default:
			return "mp4"
		}
	}
}

func presetSuffix(opts ConvertOptions) string {
	switch opts.PresetID {
	case PresetDiscord10MB:
		return "_discord10mb"
	case PresetDiscord25MB:
		return "_discord25mb"
	case PresetWhatsApp:
		return "_whatsapp"
	case PresetTVSmart:
		return "_tv"
	case PresetTargetSize:
		if opts.TargetSizeMB > 0 {
			return fmt.Sprintf("_%dmb", int(math.Round(opts.TargetSizeMB)))
		}
		return "_target"
	case PresetCompressHQ:
		return "_hq"
	case PresetAudioOnly:
		return "_audio"
	default:
		return "_converted"
	}
}

// BuildPlan compiles FFmpeg command-line arguments and preview metadata for a media item.
func BuildPlan(info ffmpeg.MediaInfo, opts ConvertOptions, hw ffmpeg.HardwareInfo, targetOutputPath string, passLogPrefix string) ConversionPlan {
	ext := resolveExtension(opts)
	effDur := EffectiveDuration(info, opts)
	plan := ConversionPlan{
		OutputPath:           targetOutputPath,
		Extension:            ext,
		EffectiveDurationSec: effDur,
	}

	commonIn := buildInputArgs(info, opts)

	// 1. Audio extraction preset or audio-only container
	if opts.PresetID == PresetAudioOnly || ext == "mp3" || ext == "m4a" || ext == "wav" {
		audioFmt := ext
		if opts.PresetID == PresetAudioOnly {
			audioFmt = resolveExtension(opts)
		}
		aBr := opts.AudioBitrateKbps
		if aBr <= 0 {
			aBr = 192
		}
		args := append([]string{}, commonIn...)
		args = append(args, "-vn")
		switch audioFmt {
		case "m4a":
			args = append(args, "-c:a", "aac", "-b:a", fmt.Sprintf("%dk", aBr))
			plan.EffectiveEncoder = "aac"
		case "wav":
			args = append(args, "-c:a", "pcm_s16le")
			plan.EffectiveEncoder = "pcm_s16le"
			aBr = 1411
		default:
			args = append(args, "-c:a", "libmp3lame", "-b:a", fmt.Sprintf("%dk", aBr))
			plan.EffectiveEncoder = "libmp3lame"
		}
		args = append(args, "-progress", "pipe:1", "-nostats", "-loglevel", "warning", "-y", targetOutputPath)
		plan.SingleArgs = args
		plan.EncoderBackend = "audio"
		plan.AudioBitrateKbps = aBr
		plan.EstimatedSizeMB = float64(aBr) * effDur / 8.0 / 1024.0
		return plan
	}

	// 2. Smart TV preset (remux / audio_only / transcode)
	if opts.PresetID == PresetTVSmart {
		tv := AnalyzeTVStrategy(info, opts)
		plan.TVStrategy = &tv
		if tv.Mode == "remux" {
			args := append([]string{}, commonIn...)
			args = append(args, "-map", "0:v:0", "-c:v", "copy")
			if info.HasAudio && !opts.Mute {
				args = append(args, "-map", "0:a:0?", "-c:a", "copy")
			} else {
				args = append(args, "-an")
			}
			args = append(args, "-movflags", "+faststart", "-progress", "pipe:1", "-nostats", "-loglevel", "warning", "-y", targetOutputPath)
			plan.SingleArgs = args
			plan.EffectiveEncoder = "copy (remux)"
			plan.EncoderBackend = "copy"
			plan.EstimatedSizeMB = float64(info.SizeBytes) / (1024 * 1024)
			return plan
		}
		if tv.Mode == "audio_only" {
			args := append([]string{}, commonIn...)
			args = append(args, "-map", "0:v:0", "-c:v", "copy")
			args = append(args, "-map", "0:a:0?", "-c:a", "aac", "-b:a", "192k", "-ac", "2")
			args = append(args, "-movflags", "+faststart", "-progress", "pipe:1", "-nostats", "-loglevel", "warning", "-y", targetOutputPath)
			plan.SingleArgs = args
			plan.EffectiveEncoder = "copy + aac"
			plan.EncoderBackend = "copy"
			plan.AudioBitrateKbps = 192
			plan.EstimatedSizeMB = float64(info.SizeBytes) / (1024 * 1024)
			return plan
		}
		// Fall through to full H.264 TV-safe transcode
	}

	// 3. Custom video stream copy (when user selects codec=copy with no transforms)
	hasTransform := (opts.Resolution != "" && opts.Resolution != "original") || opts.FPS > 0 || opts.Rotation != 0
	if opts.PresetID == PresetCustom && opts.VideoCodec == "copy" && !hasTransform {
		args := append([]string{}, commonIn...)
		args = append(args, "-c:v", "copy")
		if info.HasAudio && !opts.Mute {
			if ext == "mp4" && info.AudioCodec != "aac" && info.AudioCodec != "mp3" {
				args = append(args, "-c:a", "aac", "-b:a", "192k")
			} else {
				args = append(args, "-c:a", "copy")
			}
		} else {
			args = append(args, "-an")
		}
		if ext == "mp4" || ext == "mov" {
			args = append(args, "-movflags", "+faststart")
		}
		args = append(args, "-progress", "pipe:1", "-nostats", "-loglevel", "warning", "-y", targetOutputPath)
		plan.SingleArgs = args
		plan.EffectiveEncoder = "copy"
		plan.EncoderBackend = "copy"
		plan.EstimatedSizeMB = float64(info.SizeBytes) / (1024 * 1024)
		return plan
	}

	// 4. Size-targeted presets (Discord 10/25 MB, Custom Target MB)
	isSizeTarget := opts.PresetID == PresetDiscord10MB || opts.PresetID == PresetDiscord25MB || opts.PresetID == PresetTargetSize
	if isSizeTarget {
		targetMB := opts.TargetSizeMB
		switch opts.PresetID {
		case PresetDiscord10MB:
			targetMB = 10
		case PresetDiscord25MB:
			targetMB = 25
		}
		if targetMB <= 0 {
			targetMB = 10
		}

		budget := CalculateTargetBitrate(effDur, targetMB, opts.Mute || !info.HasAudio)
		plan.VideoBitrateKbps = budget.VideoBitrateKbps
		plan.AudioBitrateKbps = budget.AudioBitrateKbps
		plan.Warning = budget.Warning
		plan.EstimatedSizeMB = targetMB * 0.95

		// Pick height cap if resolution is original and bitrate is tight
		resOpt := opts.Resolution
		if (resOpt == "" || resOpt == "original") && budget.SuggestedHeight > 0 {
			if info.Height == 0 || info.Height > budget.SuggestedHeight {
				if budget.SuggestedHeight == 480 {
					resOpt = "480p"
				} else {
					resOpt = "720p"
				}
			}
		}

		codecFamily := "h264"
		if opts.VideoCodec == "hevc" {
			codecFamily = "hevc"
		}
		encName, backend := selectEncoder(opts.EncoderMode, codecFamily, hw)
		plan.EffectiveEncoder = encName
		plan.EncoderBackend = backend

		vf := buildVideoFilters(resOpt, opts.FPS, opts.Rotation)
		audioArgs := buildAudioArgs(info, opts.Mute, budget.AudioBitrateKbps)

		if backend == "nvenc" {
			args := append([]string{}, commonIn...)
			if vf != "" {
				args = append(args, "-vf", vf)
			}
			maxRate := int(math.Round(float64(budget.VideoBitrateKbps) * 1.25))
			bufSize := budget.VideoBitrateKbps * 2
			args = append(args,
				"-c:v", encName,
				"-preset", "p5",
				"-tune", "hq",
				"-multipass", "fullres",
				"-rc", "vbr",
				"-b:v", fmt.Sprintf("%dk", budget.VideoBitrateKbps),
				"-maxrate", fmt.Sprintf("%dk", maxRate),
				"-bufsize", fmt.Sprintf("%dk", bufSize),
				"-pix_fmt", "yuv420p",
			)
			args = append(args, audioArgs...)
			args = append(args, "-movflags", "+faststart", "-progress", "pipe:1", "-nostats", "-loglevel", "warning", "-y", targetOutputPath)
			plan.SingleArgs = args
			return plan
		}

		if backend == "cpu" {
			// True 2-pass encoding for exact size targeting on CPU
			plan.UsesTwoPass = true
			if passLogPrefix == "" {
				passLogPrefix = "mediapulse_pass"
			}
			nullOut := "/dev/null"
			if runtime.GOOS == "windows" {
				nullOut = "NUL"
			}

			p1 := append([]string{}, commonIn...)
			if vf != "" {
				p1 = append(p1, "-vf", vf)
			}
			p1 = append(p1,
				"-c:v", encName,
				"-preset", "medium",
				"-b:v", fmt.Sprintf("%dk", budget.VideoBitrateKbps),
				"-pass", "1",
				"-passlogfile", passLogPrefix,
				"-pix_fmt", "yuv420p",
				"-an",
				"-f", "mp4",
				"-progress", "pipe:1", "-nostats", "-loglevel", "warning",
				"-y", nullOut,
			)

			p2 := append([]string{}, commonIn...)
			if vf != "" {
				p2 = append(p2, "-vf", vf)
			}
			p2 = append(p2,
				"-c:v", encName,
				"-preset", "medium",
				"-b:v", fmt.Sprintf("%dk", budget.VideoBitrateKbps),
				"-maxrate", fmt.Sprintf("%dk", int(math.Round(float64(budget.VideoBitrateKbps)*1.3))),
				"-bufsize", fmt.Sprintf("%dk", budget.VideoBitrateKbps*2),
				"-pass", "2",
				"-passlogfile", passLogPrefix,
				"-pix_fmt", "yuv420p",
			)
			p2 = append(p2, audioArgs...)
			p2 = append(p2, "-movflags", "+faststart", "-progress", "pipe:1", "-nostats", "-loglevel", "warning", "-y", targetOutputPath)

			plan.Pass1Args = p1
			plan.Pass2Args = p2
			return plan
		}

		// AMF / QSV single-pass CBR/VBR
		args := append([]string{}, commonIn...)
		if vf != "" {
			args = append(args, "-vf", vf)
		}
		args = append(args,
			"-c:v", encName,
			"-b:v", fmt.Sprintf("%dk", budget.VideoBitrateKbps),
			"-maxrate", fmt.Sprintf("%dk", int(math.Round(float64(budget.VideoBitrateKbps)*1.25))),
			"-bufsize", fmt.Sprintf("%dk", budget.VideoBitrateKbps*2),
			"-pix_fmt", "yuv420p",
		)
		args = append(args, audioArgs...)
		args = append(args, "-movflags", "+faststart", "-progress", "pipe:1", "-nostats", "-loglevel", "warning", "-y", targetOutputPath)
		plan.SingleArgs = args
		return plan
	}

	// 5. Quality / Compatibility modes (WhatsApp, TV transcode, Compress HQ, Custom)
	codecFamily := "h264"
	if opts.PresetID == PresetCustom && opts.VideoCodec == "hevc" {
		codecFamily = "hevc"
	}
	if opts.PresetID == PresetCompressHQ && opts.VideoCodec == "hevc" {
		codecFamily = "hevc"
	}
	// WhatsApp & TV transcode always enforce H.264 for max device compatibility
	if opts.PresetID == PresetWhatsApp || opts.PresetID == PresetTVSmart || ext == "webm" {
		codecFamily = "h264"
	}

	resOpt := opts.Resolution
	if opts.PresetID == PresetWhatsApp && (resOpt == "" || resOpt == "original") {
		if info.Height > 1080 {
			resOpt = "1080p"
		}
	}

	crf := opts.QualityCRF
	if crf <= 0 {
		crf = 23
	}
	if opts.PresetID == PresetWhatsApp {
		crf = 24
	} else if opts.PresetID == PresetTVSmart {
		crf = 21
	}

	audioKbps := 192
	if opts.PresetID == PresetWhatsApp {
		audioKbps = 128
	} else if opts.AudioBitrateKbps > 0 {
		audioKbps = opts.AudioBitrateKbps
	}
	if opts.Mute || !info.HasAudio {
		audioKbps = 0
	}
	plan.AudioBitrateKbps = audioKbps

	vf := buildVideoFilters(resOpt, opts.FPS, opts.Rotation)
	audioArgs := buildAudioArgs(info, opts.Mute, audioKbps)

	if ext == "webm" {
		args := append([]string{}, commonIn...)
		if vf != "" {
			args = append(args, "-vf", vf)
		}
		args = append(args, "-c:v", "libvpx-vp9", "-crf", "30", "-b:v", "0", "-pix_fmt", "yuv420p")
		if info.HasAudio && !opts.Mute {
			args = append(args, "-c:a", "libopus", "-b:a", "128k")
		} else {
			args = append(args, "-an")
		}
		args = append(args, "-progress", "pipe:1", "-nostats", "-loglevel", "warning", "-y", targetOutputPath)
		plan.SingleArgs = args
		plan.EffectiveEncoder = "libvpx-vp9"
		plan.EncoderBackend = "cpu"
		return plan
	}

	encName, backend := selectEncoder(opts.EncoderMode, codecFamily, hw)
	plan.EffectiveEncoder = encName
	plan.EncoderBackend = backend

	args := append([]string{}, commonIn...)
	if vf != "" {
		args = append(args, "-vf", vf)
	}

	switch backend {
	case "nvenc":
		args = append(args,
			"-c:v", encName,
			"-preset", "p5",
			"-tune", "hq",
			"-rc", "vbr",
			"-cq", fmt.Sprintf("%d", crf),
			"-b:v", "0",
			"-spatial-aq", "1",
			"-temporal-aq", "1",
			"-pix_fmt", "yuv420p",
		)
	case "amf":
		args = append(args,
			"-c:v", encName,
			"-quality", "quality",
			"-rc", "cqp",
			"-qp_i", fmt.Sprintf("%d", crf),
			"-qp_p", fmt.Sprintf("%d", crf+2),
			"-pix_fmt", "yuv420p",
		)
	case "qsv":
		args = append(args,
			"-c:v", encName,
			"-preset", "medium",
			"-global_quality", fmt.Sprintf("%d", crf),
			"-pix_fmt", "yuv420p",
		)
	default:
		args = append(args,
			"-c:v", encName,
			"-preset", "medium",
			"-crf", fmt.Sprintf("%d", crf),
			"-pix_fmt", "yuv420p",
		)
	}

	if (opts.PresetID == PresetWhatsApp || opts.PresetID == PresetTVSmart) && codecFamily == "h264" {
		if backend == "cpu" || backend == "nvenc" {
			args = append(args, "-profile:v", "main", "-level:v", "4.1")
		}
	}

	args = append(args, audioArgs...)
	if ext == "mp4" || ext == "mov" {
		args = append(args, "-movflags", "+faststart")
	}
	args = append(args, "-progress", "pipe:1", "-nostats", "-loglevel", "warning", "-y", targetOutputPath)
	plan.SingleArgs = args

	if info.SizeBytes > 0 {
		// Rough heuristic estimate for UI display only
		ratio := 0.55
		if opts.PresetID == PresetWhatsApp {
			ratio = 0.45
		}
		plan.EstimatedSizeMB = (float64(info.SizeBytes) / (1024 * 1024)) * ratio * (effDur / math.Max(1, info.DurationSec))
	}

	return plan
}

func buildInputArgs(info ffmpeg.MediaInfo, opts ConvertOptions) []string {
	var args []string
	if opts.TrimStartSec > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", opts.TrimStartSec))
	}
	args = append(args, "-i", info.Path)
	if opts.TrimEndSec > opts.TrimStartSec && opts.TrimEndSec > 0 {
		args = append(args, "-t", fmt.Sprintf("%.3f", opts.TrimEndSec-math.Max(0, opts.TrimStartSec)))
	}
	return args
}

func buildVideoFilters(resolution string, fps int, rotation int) string {
	var filters []string
	switch rotation {
	case 90:
		filters = append(filters, "transpose=1")
	case 180:
		filters = append(filters, "transpose=1,transpose=1")
	case 270:
		filters = append(filters, "transpose=2")
	}

	switch resolution {
	case "1080p":
		filters = append(filters, "scale=-2:min(ih\\,1080)")
	case "720p":
		filters = append(filters, "scale=-2:min(ih\\,720)")
	case "480p":
		filters = append(filters, "scale=-2:min(ih\\,480)")
	default:
		// Guarantee even dimensions required by yuv420p/H.264 encoders.
		filters = append(filters, "scale=trunc(iw/2)*2:trunc(ih/2)*2")
	}

	if fps > 0 {
		filters = append(filters, fmt.Sprintf("fps=%d", fps))
	}

	return strings.Join(filters, ",")
}

func buildAudioArgs(info ffmpeg.MediaInfo, mute bool, bitrateKbps int) []string {
	if mute || !info.HasAudio || bitrateKbps <= 0 {
		return []string{"-an"}
	}
	return []string{"-c:a", "aac", "-b:a", fmt.Sprintf("%dk", bitrateKbps), "-ac", "2"}
}

func selectEncoder(mode string, codecFamily string, hw ffmpeg.HardwareInfo) (string, string) {
	isHEVC := codecFamily == "hevc"
	pickCPU := func() (string, string) {
		if isHEVC {
			return "libx265", "cpu"
		}
		return "libx264", "cpu"
	}

	tryBackend := func(backend string) (string, string, bool) {
		var name string
		switch backend {
		case "nvenc":
			if isHEVC {
				name = "hevc_nvenc"
			} else {
				name = "h264_nvenc"
			}
		case "amf":
			if isHEVC {
				name = "hevc_amf"
			} else {
				name = "h264_amf"
			}
		case "qsv":
			if isHEVC {
				name = "hevc_qsv"
			} else {
				name = "h264_qsv"
			}
		default:
			return "", "", false
		}
		if hw.HasEncoder(name) {
			return name, backend, true
		}
		return "", "", false
	}

	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "cpu" {
		return pickCPU()
	}
	if mode == "nvenc" || mode == "amf" || mode == "qsv" {
		if n, b, ok := tryBackend(mode); ok {
			return n, b
		}
		return pickCPU()
	}

	// Auto priority: nvenc -> amf -> qsv -> cpu
	for _, b := range []string{"nvenc", "amf", "qsv"} {
		if n, back, ok := tryBackend(b); ok {
			return n, back
		}
	}
	return pickCPU()
}
