export type UILang = 'es' | 'en';
export type NavSection = 'converter' | 'presets' | 'history' | 'settings';

export type PresetID =
  | 'discord_10mb'
  | 'discord_25mb'
  | 'whatsapp'
  | 'tv_smart'
  | 'target_size'
  | 'compress_hq'
  | 'audio_extract'
  | 'custom';

export interface AppInfo {
  name: string;
  version: string;
  os: string;
  arch: string;
  configPath: string;
  zeroTelemetry: boolean;
}

export interface FFmpegBinaries {
  ffmpegPath: string;
  ffprobePath: string;
  source: string;
  version: string;
  ready: boolean;
  error?: string;
}

export interface EncoderCapability {
  name: string;
  codec: string;
  backend: string;
  label: string;
  available: boolean;
}

export interface HardwareInfo {
  gpuName: string;
  encoders: EncoderCapability[];
  recommended: string;
}

export interface SystemStatus {
  binaries: FFmpegBinaries;
  hardware: HardwareInfo;
}

export interface AppConfig {
  language: UILang;
  theme: 'dark' | 'light';
  preferredEncoder: string;
  defaultOutputDir: string;
  customFfmpegDir: string;
  openFolderOnDone: boolean;
  defaultPreset: PresetID;
}

export interface MediaInfo {
  path: string;
  fileName: string;
  extension: string;
  sizeBytes: number;
  durationSec: number;
  bitrateKbps: number;
  container: string;
  hasVideo: boolean;
  videoCodec: string;
  videoProfile: string;
  pixelFormat: string;
  width: number;
  height: number;
  fps: number;
  bitDepth: number;
  hasAudio: boolean;
  audioCodec: string;
  audioChannels: number;
  audioSampleRate: number;
  audioBitrateKbps: number;
}

export interface ConvertOptions {
  presetId: PresetID;
  targetSizeMb: number;
  outputFormat: string;
  videoCodec: string;
  qualityCrf: number;
  resolution: string;
  fps: number;
  rotation: number;
  mute: boolean;
  trimStartSec: number;
  trimEndSec: number;
  audioFormat: string;
  audioBitrateKbps: number;
  encoderMode: string;
  outputDir: string;
}

export interface TVStrategy {
  mode: 'remux' | 'audio_only' | 'transcode';
  reason: string;
  estimatedSpeedLabel: string;
}

export interface ConversionPlan {
  outputPath: string;
  extension: string;
  usesTwoPass: boolean;
  effectiveEncoder: string;
  encoderBackend: string;
  tvStrategy?: TVStrategy;
  videoBitrateKbps: number;
  audioBitrateKbps: number;
  effectiveDurationSec: number;
  estimatedSizeMb: number;
  warning?: string;
}

export interface Job {
  id: string;
  info: MediaInfo;
  options: ConvertOptions;
  plan: ConversionPlan;
  status: 'queued' | 'running' | 'done' | 'error' | 'cancelled';
  progressPercent: number;
  speed: string;
  etaSec: number;
  pass: number;
  outputPath: string;
  outputBytes: number;
  processSec: number;
  fallbackNotice?: string;
  error?: string;
}

export interface QueueState {
  jobs: Job[];
  running: boolean;
  doneCount: number;
  pending: number;
}

export interface PresetDefinition {
  id: PresetID;
  nameEs: string;
  nameEn: string;
  descEs: string;
  descEn: string;
  badge: string;
  defaultFormat: string;
  defaultSizeMb: number;
  accent: string;
}

export interface HistoryEntry {
  id: string;
  timestamp: string;
  inputName: string;
  inputPath: string;
  outputName: string;
  outputPath: string;
  presetId: string;
  encoder: string;
  inputBytes: number;
  outputBytes: number;
  savedBytes: number;
  processSec: number;
}

export interface HistorySummary {
  totalJobs: number;
  totalInputBytes: number;
  totalOutputBytes: number;
  totalSavedBytes: number;
  entries: HistoryEntry[];
}

export interface FFmpegDownloadProgress {
  stage:
    | 'idle'
    | 'connecting'
    | 'downloading'
    | 'extracting'
    | 'verifying'
    | 'done'
    | 'error'
    | 'cancelled';
  percent: number;
  downloadedBytes: number;
  totalBytes: number;
  source?: string;
  destDir?: string;
  error?: string;
}

export interface UpdateCheckResult {
  currentVersion: string;
  latestVersion: string;
  updateAvailable: boolean;
  releaseUrl: string;
  releaseName?: string;
  publishedAt?: string;
  releaseNotes?: string;
  checkedAt: string;
  error?: string;
}
