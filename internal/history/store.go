package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry records a single completed media conversion.
type Entry struct {
	ID            string  `json:"id"`
	Timestamp     string  `json:"timestamp"`
	InputName     string  `json:"inputName"`
	InputPath     string  `json:"inputPath"`
	OutputName    string  `json:"outputName"`
	OutputPath    string  `json:"outputPath"`
	PresetID      string  `json:"presetId"`
	Encoder       string  `json:"encoder"`
	InputBytes    int64   `json:"inputBytes"`
	OutputBytes   int64   `json:"outputBytes"`
	SavedBytes    int64   `json:"savedBytes"`
	ProcessSec    float64 `json:"processSec"`
}

// Summary contains persisted conversion entries and aggregate savings.
type Summary struct {
	TotalJobs        int     `json:"totalJobs"`
	TotalInputBytes  int64   `json:"totalInputBytes"`
	TotalOutputBytes int64   `json:"totalOutputBytes"`
	TotalSavedBytes  int64   `json:"totalSavedBytes"`
	Entries          []Entry `json:"entries"`
}

// Store persists local conversion history without any telemetry.
type Store struct {
	mu      sync.RWMutex
	path    string
	entries []Entry
}

const maxEntries = 200

// NewStore opens or creates %APPDATA%\MediaPulse\history.json.
func NewStore() (*Store, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil || baseDir == "" {
		baseDir = os.TempDir()
	}
	dir := filepath.Join(baseDir, "MediaPulse")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "history.json")
	s := &Store{
		path:    path,
		entries: make([]Entry, 0),
	}
	_ = s.load()
	return s, nil
}

// Add prepends a completed entry and writes history.json.
func (s *Store) Add(e Entry) error {
	if e.Timestamp == "" {
		e.Timestamp = time.Now().Format(time.RFC3339)
	}
	e.SavedBytes = e.InputBytes - e.OutputBytes

	s.mu.Lock()
	s.entries = append([]Entry{e}, s.entries...)
	if len(s.entries) > maxEntries {
		s.entries = s.entries[:maxEntries]
	}
	snapshot := make([]Entry, len(s.entries))
	copy(snapshot, s.entries)
	s.mu.Unlock()

	return s.persist(snapshot)
}

// Summary returns aggregate statistics and recent entries.
func (s *Store) Summary() Summary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := Summary{
		TotalJobs: len(s.entries),
		Entries:   make([]Entry, len(s.entries)),
	}
	copy(out.Entries, s.entries)
	for _, e := range s.entries {
		out.TotalInputBytes += e.InputBytes
		out.TotalOutputBytes += e.OutputBytes
		out.TotalSavedBytes += e.SavedBytes
	}
	return out
}

// Clear removes all stored history entries.
func (s *Store) Clear() error {
	s.mu.Lock()
	s.entries = make([]Entry, 0)
	s.mu.Unlock()
	return s.persist([]Entry{})
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var loaded []Entry
	if err := json.Unmarshal(data, &loaded); err != nil {
		return err
	}
	s.mu.Lock()
	s.entries = loaded
	s.mu.Unlock()
	return nil
}

func (s *Store) persist(entries []Entry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}
