package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Config holds local-only user preferences for MediaPulse.
type Config struct {
	Language         string `json:"language"`         // es | en
	Theme            string `json:"theme"`            // dark | light
	PreferredEncoder string `json:"preferredEncoder"` // auto | nvenc | amf | qsv | cpu
	DefaultOutputDir string `json:"defaultOutputDir"`
	CustomFFmpegDir  string `json:"customFfmpegDir"`
	OpenFolderOnDone bool   `json:"openFolderOnDone"`
	DefaultPreset    string `json:"defaultPreset"`
}

// DefaultConfig returns sensible defaults for a fresh MediaPulse installation.
func DefaultConfig() Config {
	return Config{
		Language:         "es",
		Theme:            "dark",
		PreferredEncoder: "auto",
		DefaultPreset:    "discord_10mb",
	}
}

// Store manages atomic persistence of Config in %APPDATA%\MediaPulse\config.json.
type Store struct {
	mu   sync.RWMutex
	path string
	cfg  Config
}

// NewStore opens or initializes the local configuration store.
func NewStore() (*Store, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil || baseDir == "" {
		baseDir = os.TempDir()
	}
	dir := filepath.Join(baseDir, "MediaPulse")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "config.json")
	s := &Store{
		path: path,
		cfg:  DefaultConfig(),
	}
	_ = s.load()
	return s, nil
}

// Path returns the full config.json path.
func (s *Store) Path() string {
	return s.path
}

// Get returns a copy of the current configuration.
func (s *Store) Get() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

// Save normalizes, stores, and persists the configuration.
func (s *Store) Save(cfg Config) error {
	cfg = normalize(cfg)
	s.mu.Lock()
	s.cfg = cfg
	s.mu.Unlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	s.mu.Lock()
	s.cfg = normalize(cfg)
	s.mu.Unlock()
	return nil
}

func normalize(cfg Config) Config {
	if cfg.Language != "en" {
		cfg.Language = "es"
	}
	if cfg.Theme != "light" {
		cfg.Theme = "dark"
	}
	switch cfg.PreferredEncoder {
	case "auto", "nvenc", "amf", "qsv", "cpu":
	default:
		cfg.PreferredEncoder = "auto"
	}
	if cfg.DefaultPreset == "" {
		cfg.DefaultPreset = "discord_10mb"
	}
	return cfg
}
