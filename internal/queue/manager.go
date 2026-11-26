package queue

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"mediapulse/internal/ffmpeg"
	"mediapulse/internal/history"
	"mediapulse/internal/presets"
)

// JobStatus enumerates lifecycle states of a queued media job.
type JobStatus string

const (
	StatusQueued    JobStatus = "queued"
	StatusRunning   JobStatus = "running"
	StatusDone      JobStatus = "done"
	StatusError     JobStatus = "error"
	StatusCancelled JobStatus = "cancelled"
)

// Job represents a single media file in the conversion queue.
type Job struct {
	ID              string                 `json:"id"`
	Info            ffmpeg.MediaInfo       `json:"info"`
	Options         presets.ConvertOptions `json:"options"`
	Plan            presets.ConversionPlan `json:"plan"`
	Status          JobStatus              `json:"status"`
	ProgressPercent float64                `json:"progressPercent"`
	Speed           string                 `json:"speed"`
	ETASec          float64                `json:"etaSec"`
	Pass            int                    `json:"pass"`
	OutputPath      string                 `json:"outputPath"`
	OutputBytes     int64                  `json:"outputBytes"`
	ProcessSec      float64                `json:"processSec"`
	FallbackNotice  string                 `json:"fallbackNotice,omitempty"`
	Error           string                 `json:"error,omitempty"`
}

// QueueState is the full snapshot sent to the UI.
type QueueState struct {
	Jobs      []Job `json:"jobs"`
	Running   bool  `json:"running"`
	DoneCount int   `json:"doneCount"`
	Pending   int   `json:"pending"`
}

// Events defines callbacks emitted by Manager to Wails or tests.
type Events struct {
	OnState    func(QueueState)
	OnProgress func(jobID string, p ffmpeg.ProgressUpdate)
	OnJobDone  func(job Job)
}

// Manager orchestrates probing, planning, and sequential execution of conversion jobs.
type Manager struct {
	mu           sync.Mutex
	jobs         []Job
	nextSeq      int
	running      bool
	cancelCur    context.CancelFunc
	binaries     ffmpeg.Binaries
	hw           ffmpeg.HardwareInfo
	historyStore *history.Store
	events       Events
}

// NewManager creates a conversion queue manager.
func NewManager(bins ffmpeg.Binaries, hw ffmpeg.HardwareInfo, hist *history.Store, events Events) *Manager {
	return &Manager{
		jobs:         make([]Job, 0),
		binaries:     bins,
		hw:           hw,
		historyStore: hist,
		events:       events,
	}
}

// SetEnvironment updates the active FFmpeg binaries and detected hardware.
func (m *Manager) SetEnvironment(bins ffmpeg.Binaries, hw ffmpeg.HardwareInfo) {
	m.mu.Lock()
	m.binaries = bins
	m.hw = hw
	for i := range m.jobs {
		if m.jobs[i].Status == StatusQueued {
			outPath := presets.BuildOutputPath(m.jobs[i].Info.Path, m.jobs[i].Options)
			m.jobs[i].OutputPath = outPath
			m.jobs[i].Plan = presets.BuildPlan(m.jobs[i].Info, m.jobs[i].Options, m.hw, outPath, "")
		}
	}
	m.mu.Unlock()
	m.emitState()
}

// State returns a snapshot of the current queue.
func (m *Manager) State() QueueState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.snapshotLocked()
}

func (m *Manager) snapshotLocked() QueueState {
	jobsCopy := make([]Job, len(m.jobs))
	copy(jobsCopy, m.jobs)
	done := 0
	pending := 0
	for _, j := range jobsCopy {
		if j.Status == StatusDone {
			done++
		} else if j.Status == StatusQueued || j.Status == StatusRunning {
			pending++
		}
	}
	return QueueState{
		Jobs:      jobsCopy,
		Running:   m.running,
		DoneCount: done,
		Pending:   pending,
	}
}

// AddFiles probes each valid path and appends it to the queue.
func (m *Manager) AddFiles(ctx context.Context, paths []string, defaultOpts presets.ConvertOptions) ([]Job, []string) {
	m.mu.Lock()
	bins := m.binaries
	hw := m.hw
	m.mu.Unlock()

	var added []Job
	var errs []string

	if !bins.Ready {
		return nil, []string{"FFmpeg/FFprobe no estan listos"}
	}

	for _, rawPath := range paths {
		p := strings.TrimSpace(rawPath)
		if p == "" {
			continue
		}
		info, err := ffmpeg.Probe(ctx, bins.FFprobePath, p)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", filepath.Base(p), err))
			continue
		}

		m.mu.Lock()
		m.nextSeq++
		id := fmt.Sprintf("job-%d-%d", time.Now().UnixMilli(), m.nextSeq)
		opts := defaultOpts
		outPath := presets.BuildOutputPath(info.Path, opts)
		plan := presets.BuildPlan(info, opts, hw, outPath, "")
		job := Job{
			ID:         id,
			Info:       info,
			Options:    opts,
			Plan:       plan,
			Status:     StatusQueued,
			OutputPath: outPath,
		}
		m.jobs = append(m.jobs, job)
		added = append(added, job)
		m.mu.Unlock()
	}

	m.emitState()
	return added, errs
}

// UpdateJobOptions updates conversion options and recomputes the plan for a queued/error/cancelled job.
func (m *Manager) UpdateJobOptions(jobID string, opts presets.ConvertOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.jobs {
		if m.jobs[i].ID == jobID {
			if m.jobs[i].Status == StatusRunning {
				return fmt.Errorf("no se puede editar un trabajo en ejecucion")
			}
			m.jobs[i].Options = opts
			outPath := presets.BuildOutputPath(m.jobs[i].Info.Path, opts)
			m.jobs[i].OutputPath = outPath
			m.jobs[i].Plan = presets.BuildPlan(m.jobs[i].Info, opts, m.hw, outPath, "")
			if m.jobs[i].Status != StatusQueued {
				m.jobs[i].Status = StatusQueued
				m.jobs[i].ProgressPercent = 0
				m.jobs[i].Error = ""
			}
			go m.emitState()
			return nil
		}
	}
	return fmt.Errorf("trabajo no encontrado: %s", jobID)
}

// ApplyOptionsToQueued applies options to every currently queued item.
func (m *Manager) ApplyOptionsToQueued(opts presets.ConvertOptions) {
	m.mu.Lock()
	for i := range m.jobs {
		if m.jobs[i].Status == StatusQueued {
			m.jobs[i].Options = opts
			outPath := presets.BuildOutputPath(m.jobs[i].Info.Path, opts)
			m.jobs[i].OutputPath = outPath
			m.jobs[i].Plan = presets.BuildPlan(m.jobs[i].Info, opts, m.hw, outPath, "")
		}
	}
	m.mu.Unlock()
	m.emitState()
}

// RemoveJob removes a non-running job from the queue.
func (m *Manager) RemoveJob(jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, j := range m.jobs {
		if j.ID == jobID {
			if j.Status == StatusRunning {
				return fmt.Errorf("cancela el trabajo antes de eliminarlo")
			}
			m.jobs = append(m.jobs[:i], m.jobs[i+1:]...)
			go m.emitState()
			return nil
		}
	}
	return nil
}

// ClearCompleted removes all finished, errored, or cancelled jobs.
func (m *Manager) ClearCompleted() {
	m.mu.Lock()
	kept := make([]Job, 0, len(m.jobs))
	for _, j := range m.jobs {
		if j.Status == StatusQueued || j.Status == StatusRunning {
			kept = append(kept, j)
		}
	}
	m.jobs = kept
	m.mu.Unlock()
	m.emitState()
}

// RetryJob marks a failed or cancelled job back to queued.
func (m *Manager) RetryJob(jobID string) {
	m.mu.Lock()
	for i := range m.jobs {
		if m.jobs[i].ID == jobID && (m.jobs[i].Status == StatusError || m.jobs[i].Status == StatusCancelled) {
			m.jobs[i].Status = StatusQueued
			m.jobs[i].ProgressPercent = 0
			m.jobs[i].Error = ""
			m.jobs[i].FallbackNotice = ""
			outPath := presets.BuildOutputPath(m.jobs[i].Info.Path, m.jobs[i].Options)
			m.jobs[i].OutputPath = outPath
			m.jobs[i].Plan = presets.BuildPlan(m.jobs[i].Info, m.jobs[i].Options, m.hw, outPath, "")
		}
	}
	m.mu.Unlock()
	m.emitState()
}

// Start begins processing all queued jobs sequentially in a background goroutine.
func (m *Manager) Start(parentCtx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	if !m.binaries.Ready {
		m.mu.Unlock()
		return fmt.Errorf("FFmpeg no esta disponible")
	}
	hasQueued := false
	for _, j := range m.jobs {
		if j.Status == StatusQueued {
			hasQueued = true
			break
		}
	}
	if !hasQueued {
		m.mu.Unlock()
		return nil
	}

	m.running = true
	runCtx, cancel := context.WithCancel(parentCtx)
	m.cancelCur = cancel
	m.mu.Unlock()

	m.emitState()

	go m.processLoop(runCtx)
	return nil
}

// Cancel stops the currently running conversion and marks pending run cancelled.
func (m *Manager) Cancel() {
	m.mu.Lock()
	if m.cancelCur != nil {
		m.cancelCur()
	}
	m.mu.Unlock()
}

func (m *Manager) processLoop(ctx context.Context) {
	defer func() {
		m.mu.Lock()
		m.running = false
		m.cancelCur = nil
		m.mu.Unlock()
		m.emitState()
	}()

	for {
		if ctx.Err() != nil {
			return
		}

		var nextIdx = -1
		m.mu.Lock()
		for i := range m.jobs {
			if m.jobs[i].Status == StatusQueued {
				nextIdx = i
				m.jobs[i].Status = StatusRunning
				m.jobs[i].ProgressPercent = 0
				m.jobs[i].Error = ""
				m.jobs[i].FallbackNotice = ""
				break
			}
		}
		if nextIdx == -1 {
			m.mu.Unlock()
			return
		}
		job := m.jobs[nextIdx]
		bins := m.binaries
		hw := m.hw
		m.mu.Unlock()

		m.emitState()

		finished := m.executeJob(ctx, bins, hw, job)

		m.mu.Lock()
		for i := range m.jobs {
			if m.jobs[i].ID == finished.ID {
				m.jobs[i] = finished
				break
			}
		}
		m.mu.Unlock()

		m.emitState()
		if m.events.OnJobDone != nil {
			m.events.OnJobDone(finished)
		}
	}
}

func (m *Manager) executeJob(ctx context.Context, bins ffmpeg.Binaries, hw ffmpeg.HardwareInfo, job Job) Job {
	started := time.Now()
	finalOut := presets.BuildOutputPath(job.Info.Path, job.Options)
	tempOut := ffmpeg.BuildTempOutputPath(finalOut)
	passLogPrefix := filepath.Join(os.TempDir(), fmt.Sprintf("mp_pass_%s", job.ID))
	defer func() {
		_ = os.Remove(tempOut)
		ffmpeg.CleanupPassLogs(passLogPrefix)
	}()

	plan := presets.BuildPlan(job.Info, job.Options, hw, tempOut, passLogPrefix)
	job.OutputPath = finalOut
	job.Plan = plan
	job.Plan.OutputPath = finalOut

	var lastEmit time.Time
	emitProg := func(p ffmpeg.ProgressUpdate) {
		m.mu.Lock()
		for i := range m.jobs {
			if m.jobs[i].ID == job.ID {
				m.jobs[i].ProgressPercent = p.Percent
				m.jobs[i].Speed = p.Speed
				m.jobs[i].ETASec = p.ETASec
				m.jobs[i].Pass = p.Pass
				break
			}
		}
		m.mu.Unlock()

		if m.events.OnProgress != nil && (time.Since(lastEmit) >= 150*time.Millisecond || p.Percent >= 100) {
			lastEmit = time.Now()
			m.events.OnProgress(job.ID, p)
		}
	}

	err := runPlan(ctx, bins.FFmpegPath, plan, emitProg)

	// Automatic CPU fallback if a hardware encoder fails
	if err != nil && ctx.Err() == nil && (plan.EncoderBackend == "nvenc" || plan.EncoderBackend == "amf" || plan.EncoderBackend == "qsv") {
		_ = os.Remove(tempOut)
		ffmpeg.CleanupPassLogs(passLogPrefix)

		cpuOpts := job.Options
		cpuOpts.EncoderMode = "cpu"
		cpuPlan := presets.BuildPlan(job.Info, cpuOpts, hw, tempOut, passLogPrefix)
		job.FallbackNotice = fmt.Sprintf("Fallback a CPU (%s) tras fallo de %s", cpuPlan.EffectiveEncoder, plan.EffectiveEncoder)
		plan = cpuPlan
		job.Plan = cpuPlan
		job.Plan.OutputPath = finalOut
		err = runPlan(ctx, bins.FFmpegPath, cpuPlan, emitProg)
	}

	if ctx.Err() != nil {
		job.Status = StatusCancelled
		job.Error = "Cancelado por el usuario"
		return job
	}

	if err != nil {
		job.Status = StatusError
		job.Error = err.Error()
		return job
	}

	if err := ffmpeg.AtomicFinalize(tempOut, finalOut); err != nil {
		job.Status = StatusError
		job.Error = fmt.Sprintf("No se pudo guardar archivo final: %v", err)
		return job
	}

	var outSize int64
	if st, stErr := os.Stat(finalOut); stErr == nil {
		outSize = st.Size()
	}

	job.Status = StatusDone
	job.ProgressPercent = 100
	job.OutputPath = finalOut
	job.OutputBytes = outSize
	job.ProcessSec = time.Since(started).Seconds()

	if m.historyStore != nil {
		_ = m.historyStore.Add(history.Entry{
			ID:          job.ID,
			Timestamp:   time.Now().Format(time.RFC3339),
			InputName:   job.Info.FileName,
			InputPath:   job.Info.Path,
			OutputName:  filepath.Base(finalOut),
			OutputPath:  finalOut,
			PresetID:    string(job.Options.PresetID),
			Encoder:     plan.EffectiveEncoder,
			InputBytes:  job.Info.SizeBytes,
			OutputBytes: outSize,
			ProcessSec:  job.ProcessSec,
		})
	}

	return job
}

func runPlan(ctx context.Context, ffmpegPath string, plan presets.ConversionPlan, onProgress func(ffmpeg.ProgressUpdate)) error {
	if plan.UsesTwoPass {
		if err := ffmpeg.RunArgs(ctx, ffmpegPath, plan.Pass1Args, plan.EffectiveDurationSec, 1, 0, 45, onProgress); err != nil {
			return err
		}
		return ffmpeg.RunArgs(ctx, ffmpegPath, plan.Pass2Args, plan.EffectiveDurationSec, 2, 45, 55, onProgress)
	}
	return ffmpeg.RunArgs(ctx, ffmpegPath, plan.SingleArgs, plan.EffectiveDurationSec, 1, 0, 100, onProgress)
}

func (m *Manager) emitState() {
	if m.events.OnState != nil {
		m.events.OnState(m.State())
	}
}
