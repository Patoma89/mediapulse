package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/options"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"mediapulse/internal/config"
	"mediapulse/internal/ffmpeg"
	"mediapulse/internal/history"
	"mediapulse/internal/presets"
	"mediapulse/internal/queue"
)

const appVersion = "1.0.0"

// AppInfo exposes static runtime metadata and local privacy guarantee.
type AppInfo struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
	ConfigPath    string `json:"configPath"`
	ZeroTelemetry bool   `json:"zeroTelemetry"`
}

// SystemStatus reports FFmpeg binary availability and detected GPU encoders.
type SystemStatus struct {
	Binaries ffmpeg.Binaries     `json:"binaries"`
	Hardware ffmpeg.HardwareInfo `json:"hardware"`
}

// AddFilesResult returns newly queued jobs and any per-file errors.
type AddFilesResult struct {
	Added  []queue.Job `json:"added"`
	Errors []string    `json:"errors"`
}

// App is the main Wails backend bound to the React frontend.
type App struct {
	ctx        context.Context
	store      *config.Store
	histStore  *history.Store
	configPath string
	mu         sync.RWMutex
	bins       ffmpeg.Binaries
	hw         ffmpeg.HardwareInfo
	queueMgr   *queue.Manager
}

// NewApp initializes local stores, locates FFmpeg, and sets up the queue manager.
func NewApp() *App {
	store, err := config.NewStore()
	cfgPath := ""
	customDir := ""
	if err == nil {
		cfgPath = store.Path()
		customDir = store.Get().CustomFFmpegDir
	}

	hist, _ := history.NewStore()
	bins := ffmpeg.Locate(customDir)

	a := &App{
		store:      store,
		histStore:  hist,
		configPath: cfgPath,
		bins:       bins,
		hw: ffmpeg.HardwareInfo{
			GPUName:     "Detectando...",
			Recommended: "cpu",
		},
	}

	a.queueMgr = queue.NewManager(bins, a.hw, hist, queue.Events{
		OnState: func(st queue.QueueState) {
			if a.ctx != nil {
				wailsruntime.EventsEmit(a.ctx, "queue:state", st)
			}
		},
		OnProgress: func(jobID string, p ffmpeg.ProgressUpdate) {
			if a.ctx != nil {
				wailsruntime.EventsEmit(a.ctx, "job:progress", map[string]interface{}{
					"jobId":    jobID,
					"progress": p,
				})
			}
		},
		OnJobDone: func(job queue.Job) {
			if a.ctx != nil {
				wailsruntime.EventsEmit(a.ctx, "job:done", job)
			}
			if a.store != nil && a.store.Get().OpenFolderOnDone && job.Status == queue.StatusDone {
				st := a.queueMgr.State()
				if st.Pending == 0 && job.OutputPath != "" {
					_ = a.RevealInExplorer(job.OutputPath)
				}
			}
		},
	})

	return a
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go func() {
		a.RefreshHardware()
	}()
}

func (a *App) shutdown(ctx context.Context) {
	if a.queueMgr != nil {
		a.queueMgr.Cancel()
	}
}

func (a *App) onSecondInstance(_ options.SecondInstanceData) {
	if a.ctx != nil {
		wailsruntime.WindowUnminimise(a.ctx)
		wailsruntime.Show(a.ctx)
		wailsruntime.WindowSetAlwaysOnTop(a.ctx, true)
		wailsruntime.WindowSetAlwaysOnTop(a.ctx, false)
	}
}

// GetAppInfo returns application metadata.
func (a *App) GetAppInfo() AppInfo {
	return AppInfo{
		Name:          "MediaPulse",
		Version:       appVersion,
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		ConfigPath:    a.configPath,
		ZeroTelemetry: true,
	}
}

// GetSystemStatus returns cached FFmpeg binary and GPU encoder status.
func (a *App) GetSystemStatus() SystemStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return SystemStatus{
		Binaries: a.bins,
		Hardware: a.hw,
	}
}

// RefreshHardware re-locates FFmpeg and runs live encoder capability probes.
func (a *App) RefreshHardware() SystemStatus {
	customDir := ""
	if a.store != nil {
		customDir = a.store.Get().CustomFFmpegDir
	}
	bins := ffmpeg.Locate(customDir)
	hw := ffmpeg.DetectHardware(context.Background(), bins.FFmpegPath)

	a.mu.Lock()
	a.bins = bins
	a.hw = hw
	a.mu.Unlock()

	if a.queueMgr != nil {
		a.queueMgr.SetEnvironment(bins, hw)
	}

	st := SystemStatus{Binaries: bins, Hardware: hw}
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "system:status", st)
	}
	return st
}

// GetConfig returns the persisted user configuration.
func (a *App) GetConfig() config.Config {
	if a.store == nil {
		return config.DefaultConfig()
	}
	return a.store.Get()
}

// UpdateConfig persists user configuration and refreshes FFmpeg if directory changed.
func (a *App) UpdateConfig(cfg config.Config) (config.Config, error) {
	prevCustom := ""
	if a.store != nil {
		prevCustom = a.store.Get().CustomFFmpegDir
		if err := a.store.Save(cfg); err != nil {
			return a.store.Get(), err
		}
		cfg = a.store.Get()
	}
	if strings.TrimSpace(prevCustom) != strings.TrimSpace(cfg.CustomFFmpegDir) {
		a.RefreshHardware()
	}
	return cfg, nil
}

// GetPresets returns the built-in preset catalog.
func (a *App) GetPresets() []presets.PresetDefinition {
	return presets.Catalog()
}

// PreviewPlan computes a live conversion plan (including Smart TV strategy and bitrate) for a queued job.
func (a *App) PreviewPlan(jobID string, opts presets.ConvertOptions) (presets.ConversionPlan, error) {
	st := a.queueMgr.State()
	a.mu.RLock()
	hw := a.hw
	a.mu.RUnlock()

	for _, j := range st.Jobs {
		if j.ID == jobID {
			outPath := presets.BuildOutputPath(j.Info.Path, opts)
			return presets.BuildPlan(j.Info, opts, hw, outPath, ""), nil
		}
	}
	return presets.ConversionPlan{}, fmt.Errorf("trabajo no encontrado")
}

// PickFiles opens a native Windows multi-file selection dialog for media files.
func (a *App) PickFiles() ([]string, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("contexto no listo")
	}
	return wailsruntime.OpenMultipleFilesDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Seleccionar videos o audios",
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "Archivos multimedia (*.mp4, *.mkv, *.mov, *.avi, *.webm, *.flv, *.wmv, *.ts, *.m4v, *.mp3, *.wav)",
				Pattern:     "*.mp4;*.mkv;*.mov;*.avi;*.webm;*.flv;*.wmv;*.ts;*.m4v;*.mts;*.3gp;*.mp3;*.wav;*.m4a;*.flac",
			},
			{
				DisplayName: "Todos los archivos (*.*)",
				Pattern:     "*.*",
			},
		},
	})
}

// PickOutputDir opens a native folder picker for output directory.
func (a *App) PickOutputDir() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("contexto no listo")
	}
	return wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Seleccionar carpeta de salida",
	})
}

// PickFFmpegDir opens a native folder picker to locate custom ffmpeg.exe/ffprobe.exe.
func (a *App) PickFFmpegDir() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("contexto no listo")
	}
	return wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Seleccionar carpeta que contiene ffmpeg.exe y ffprobe.exe",
	})
}

// AddFiles probes media files and adds them to the conversion queue.
func (a *App) AddFiles(paths []string, defaultOpts presets.ConvertOptions) AddFilesResult {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if defaultOpts.PresetID == "" {
		defaultOpts = presets.DefaultOptions(presets.PresetDiscord10MB)
	}
	added, errs := a.queueMgr.AddFiles(ctx, paths, defaultOpts)
	return AddFilesResult{
		Added:  added,
		Errors: errs,
	}
}

// GetQueue returns the current conversion queue snapshot.
func (a *App) GetQueue() queue.QueueState {
	return a.queueMgr.State()
}

// UpdateJobOptions updates conversion options for one queued job.
func (a *App) UpdateJobOptions(jobID string, opts presets.ConvertOptions) error {
	return a.queueMgr.UpdateJobOptions(jobID, opts)
}

// ApplyOptionsToQueued applies options to all pending jobs in the queue.
func (a *App) ApplyOptionsToQueued(opts presets.ConvertOptions) {
	a.queueMgr.ApplyOptionsToQueued(opts)
}

// RemoveJob removes a job from the queue.
func (a *App) RemoveJob(jobID string) error {
	return a.queueMgr.RemoveJob(jobID)
}

// ClearCompleted clears finished, errored, or cancelled jobs.
func (a *App) ClearCompleted() {
	a.queueMgr.ClearCompleted()
}

// StartQueue starts sequential processing of queued items.
func (a *App) StartQueue() error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.queueMgr.Start(ctx)
}

// CancelCurrent aborts the currently active FFmpeg conversion.
func (a *App) CancelCurrent() {
	a.queueMgr.Cancel()
}

// RetryJob moves a failed/cancelled job back into queued state.
func (a *App) RetryJob(jobID string) {
	a.queueMgr.RetryJob(jobID)
}

// RevealInExplorer opens Windows Explorer selecting the generated file.
func (a *App) RevealInExplorer(targetPath string) error {
	clean := filepath.Clean(strings.TrimSpace(targetPath))
	if clean == "" || clean == "." {
		return fmt.Errorf("ruta vacia")
	}
	if _, err := os.Stat(clean); err != nil {
		dir := filepath.Dir(clean)
		cmd := exec.Command("explorer.exe", dir)
		return cmd.Start()
	}
	cmd := exec.Command("explorer.exe", "/select,", clean)
	return cmd.Start()
}

// OpenMediaFile opens a media file using the default Windows player.
func (a *App) OpenMediaFile(targetPath string) error {
	clean := filepath.Clean(strings.TrimSpace(targetPath))
	if clean == "" || clean == "." {
		return fmt.Errorf("ruta vacia")
	}
	cmd := exec.Command("cmd", "/c", "start", "", clean)
	ffmpeg.ConfigureCmd(cmd)
	return cmd.Start()
}

// GetHistory returns conversion history and aggregate savings.
func (a *App) GetHistory() history.Summary {
	if a.histStore == nil {
		return history.Summary{}
	}
	return a.histStore.Summary()
}

// ClearHistory wipes local conversion history.
func (a *App) ClearHistory() error {
	if a.histStore == nil {
		return nil
	}
	return a.histStore.Clear()
}
