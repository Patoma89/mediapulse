import { useState, useEffect } from 'react';
import {
  UploadCloud,
  Play,
  Square,
  Trash2,
  FolderOpen,
  RotateCw,
  VolumeX,
  Scissors,
  Zap,
  AlertTriangle,
  CheckCircle2,
  RefreshCw,
  FileVideo,
  Music,
  Sliders,
} from 'lucide-react';
import type {
  ConvertOptions,
  Job,
  PresetDefinition,
  PresetID,
  QueueState,
  SystemStatus,
  UILang,
} from '../types';
import { tr } from '../i18n';

interface ConverterViewProps {
  lang: UILang;
  status: SystemStatus;
  presets: PresetDefinition[];
  queue: QueueState;
  activeOptions: ConvertOptions;
  onChangeOptions: (opts: ConvertOptions) => void;
  onSelectPreset: (id: PresetID) => void;
  onPickFiles: () => void;
  onPickOutputDir: () => void;
  onApplyToAllQueued: () => void;
  onStartQueue: () => void;
  onCancelQueue: () => void;
  onClearCompleted: () => void;
  onRemoveJob: (id: string) => void;
  onRetryJob: (id: string) => void;
  onRevealFile: (path: string) => void;
  onOpenFile: (path: string) => void;
  onUpdateJobOptions: (id: string, opts: ConvertOptions) => void;
}

export function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  let idx = 0;
  let val = bytes;
  while (val >= 1024 && idx < units.length - 1) {
    val /= 1024;
    idx++;
  }
  return `${val.toFixed(idx === 0 ? 0 : val >= 100 ? 0 : 1)} ${units[idx]}`;
}

export function formatDuration(sec: number): string {
  if (!sec || sec <= 0) return '0:00';
  const s = Math.floor(sec % 60);
  const m = Math.floor((sec / 60) % 60);
  const h = Math.floor(sec / 3600);
  if (h > 0) {
    return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
  }
  return `${m}:${String(s).padStart(2, '0')}`;
}

function parseTimeInput(raw: string): number {
  const cleaned = raw.trim();
  if (!cleaned) return 0;
  if (cleaned.includes(':')) {
    const parts = cleaned.split(':').map((p) => Number(p));
    if (parts.some((n) => Number.isNaN(n) || n < 0)) return 0;
    if (parts.length === 2) return parts[0] * 60 + parts[1];
    if (parts.length === 3) return parts[0] * 3600 + parts[1] * 60 + parts[2];
  }
  const num = Number(cleaned);
  return Number.isFinite(num) && num >= 0 ? num : 0;
}

export function ConverterView({
  lang,
  status,
  presets,
  queue,
  activeOptions,
  onChangeOptions,
  onSelectPreset,
  onPickFiles,
  onPickOutputDir,
  onApplyToAllQueued,
  onStartQueue,
  onCancelQueue,
  onClearCompleted,
  onRemoveJob,
  onRetryJob,
  onRevealFile,
  onOpenFile,
  onUpdateJobOptions,
}: ConverterViewProps) {
  const [trimStartStr, setTrimStartStr] = useState('0');
  const [trimEndStr, setTrimEndStr] = useState('0');
  const [draggingOver, setDraggingOver] = useState(false);

  useEffect(() => {
    setTrimStartStr(activeOptions.trimStartSec > 0 ? String(activeOptions.trimStartSec) : '0');
    setTrimEndStr(activeOptions.trimEndSec > 0 ? String(activeOptions.trimEndSec) : '0');
  }, [activeOptions.trimStartSec, activeOptions.trimEndSec]);

  const showTargetSize =
    activeOptions.presetId === 'target_size' ||
    activeOptions.presetId === 'discord_10mb' ||
    activeOptions.presetId === 'discord_25mb';

  const isAudioPreset = activeOptions.presetId === 'audio_extract';
  const showCustomCodec =
    activeOptions.presetId === 'custom' || activeOptions.presetId === 'compress_hq';

  const accentBadge = (id: PresetID, active: boolean) => {
    if (!active) {
      return 'bg-white/[0.03] hover:bg-white/[0.06] border-white/[0.08] text-slate-300';
    }
    switch (id) {
      case 'discord_10mb':
      case 'discord_25mb':
        return 'bg-violet-500/20 border-violet-400/50 text-violet-100 shadow-[0_0_20px_rgba(139,92,246,0.12)]';
      case 'whatsapp':
        return 'bg-emerald-500/20 border-emerald-400/50 text-emerald-100';
      case 'tv_smart':
        return 'bg-cyan-500/20 border-cyan-400/50 text-cyan-100';
      case 'target_size':
        return 'bg-amber-500/20 border-amber-400/50 text-amber-100';
      case 'audio_extract':
        return 'bg-rose-500/20 border-rose-400/50 text-rose-100';
      default:
        return 'bg-cyan-500/15 border-cyan-400/45 text-cyan-100';
    }
  };

  return (
    <div
      className="space-y-5 mp-drop-zone"
      onDragOver={(e) => {
        e.preventDefault();
        setDraggingOver(true);
      }}
      onDragLeave={() => setDraggingOver(false)}
      onDrop={() => setDraggingOver(false)}
    >
      {/* Drop Zone */}
      <div
        onClick={onPickFiles}
        className={`mp-card mp-drop-zone rounded-xl p-5 border-2 border-dashed transition-all cursor-pointer flex items-center justify-between gap-4 ${
          draggingOver
            ? 'border-violet-400 bg-violet-500/10'
            : 'border-white/[0.12] hover:border-violet-400/50 hover:bg-white/[0.02]'
        }`}
      >
        <div className="flex items-center gap-4 min-w-0">
          <div className="h-12 w-12 rounded-xl bg-violet-500/15 border border-violet-500/30 flex items-center justify-center shrink-0">
            <UploadCloud className="w-6 h-6 text-violet-300" />
          </div>
          <div className="min-w-0">
            <h2 className="text-sm font-semibold text-slate-100">
              {tr(lang, 'dropTitle')}
            </h2>
            <p className="text-xs text-slate-400 mt-0.5 truncate">
              {tr(lang, 'dropSub')}
            </p>
          </div>
        </div>
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            onPickFiles();
          }}
          disabled={!status.binaries?.ready}
          className="shrink-0 px-4 py-2 rounded-lg text-xs font-semibold bg-violet-600 hover:bg-violet-500 disabled:opacity-40 text-white transition-colors cursor-pointer"
        >
          + {tr(lang, 'addVideos')}
        </button>
      </div>

      {/* Preset Selector Grid */}
      <div className="mp-card rounded-xl p-4 space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
            {tr(lang, 'presetTitle')}
          </span>
          <span className="text-[11px] text-slate-400 font-mono-telemetry">
            {status.hardware?.recommended !== 'cpu'
              ? `GPU ${status.hardware?.recommended?.toUpperCase()} Ready`
              : 'CPU x264'}
          </span>
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
          {presets.map((p) => {
            const active = activeOptions.presetId === p.id;
            return (
              <button
                key={p.id}
                type="button"
                onClick={() => onSelectPreset(p.id)}
                className={`text-left rounded-lg p-2.5 border transition-all cursor-pointer ${accentBadge(
                  p.id,
                  active
                )}`}
              >
                <div className="flex items-center justify-between gap-1">
                  <span className="text-xs font-semibold truncate">
                    {lang === 'en' ? p.nameEn : p.nameEs}
                  </span>
                  <span className="font-mono-telemetry text-[10px] px-1.5 py-0.5 rounded bg-black/25 border border-white/10 shrink-0">
                    {p.badge}
                  </span>
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Quick Adjustments & Trim Panel */}
      <div className="mp-card rounded-xl p-4 space-y-3">
        <div className="flex items-center justify-between gap-2 flex-wrap">
          <div className="flex items-center gap-2">
            <Sliders className="w-4 h-4 text-cyan-400" />
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-300">
              {tr(lang, 'quickAdjust')}
            </span>
          </div>
          {queue.pending > 0 && !queue.running && (
            <button
              type="button"
              onClick={onApplyToAllQueued}
              className="px-2.5 py-1 rounded-md text-xs font-medium bg-cyan-500/15 hover:bg-cyan-500/25 text-cyan-200 border border-cyan-500/30 transition-colors cursor-pointer"
            >
              {tr(lang, 'applyToAll')} ({queue.pending})
            </button>
          )}
        </div>

        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-3">
          {/* Target MB when applicable */}
          {showTargetSize && (
            <div>
              <label className="block text-[11px] text-slate-400 mb-1">
                {tr(lang, 'targetSizeMb')}
              </label>
              <input
                type="number"
                min={1}
                max={4096}
                value={activeOptions.targetSizeMb || 10}
                onChange={(e) =>
                  onChangeOptions({
                    ...activeOptions,
                    targetSizeMb: Math.max(1, Number(e.target.value) || 10),
                  })
                }
                className="w-full px-2.5 py-1.5 rounded-lg bg-[#090D16] border border-white/[0.1] text-xs font-mono-telemetry text-slate-100 focus:outline-none focus:border-violet-400"
              />
            </div>
          )}

          {/* Audio format */}
          {isAudioPreset && (
            <div>
              <label className="block text-[11px] text-slate-400 mb-1">
                {tr(lang, 'audioFormat')}
              </label>
              <select
                value={activeOptions.audioFormat || 'mp3'}
                onChange={(e) =>
                  onChangeOptions({
                    ...activeOptions,
                    audioFormat: e.target.value,
                    outputFormat: e.target.value,
                  })
                }
                className="w-full px-2.5 py-1.5 rounded-lg bg-[#090D16] border border-white/[0.1] text-xs text-slate-100"
              >
                <option value="mp3">MP3 (192k)</option>
                <option value="m4a">M4A (AAC)</option>
                <option value="wav">WAV (Lossless)</option>
              </select>
            </div>
          )}

          {/* Custom Format / Codec / CRF */}
          {activeOptions.presetId === 'custom' && (
            <div>
              <label className="block text-[11px] text-slate-400 mb-1">
                {tr(lang, 'format')}
              </label>
              <select
                value={activeOptions.outputFormat || 'mp4'}
                onChange={(e) =>
                  onChangeOptions({ ...activeOptions, outputFormat: e.target.value })
                }
                className="w-full px-2.5 py-1.5 rounded-lg bg-[#090D16] border border-white/[0.1] text-xs text-slate-100"
              >
                <option value="mp4">MP4</option>
                <option value="mkv">MKV</option>
                <option value="mov">MOV</option>
                <option value="webm">WEBM (VP9)</option>
              </select>
            </div>
          )}

          {showCustomCodec && (
            <>
              <div>
                <label className="block text-[11px] text-slate-400 mb-1">
                  {tr(lang, 'codec')}
                </label>
                <select
                  value={activeOptions.videoCodec || 'auto'}
                  onChange={(e) =>
                    onChangeOptions({ ...activeOptions, videoCodec: e.target.value })
                  }
                  className="w-full px-2.5 py-1.5 rounded-lg bg-[#090D16] border border-white/[0.1] text-xs text-slate-100"
                >
                  <option value="auto">Auto (H.264)</option>
                  <option value="h264">H.264 (AVC)</option>
                  <option value="hevc">H.265 (HEVC)</option>
                  {activeOptions.presetId === 'custom' && (
                    <option value="copy">Copiar stream</option>
                  )}
                </select>
              </div>
              <div>
                <label className="block text-[11px] text-slate-400 mb-1">
                  {tr(lang, 'quality')}: <span className="font-mono-telemetry">{activeOptions.qualityCrf || 23}</span>
                </label>
                <input
                  type="range"
                  min={16}
                  max={32}
                  value={activeOptions.qualityCrf || 23}
                  onChange={(e) =>
                    onChangeOptions({ ...activeOptions, qualityCrf: Number(e.target.value) })
                  }
                  className="w-full accent-violet-500 mt-1.5"
                />
              </div>
            </>
          )}

          {/* Resolution */}
          {!isAudioPreset && (
            <div>
              <label className="block text-[11px] text-slate-400 mb-1">
                {tr(lang, 'resolution')}
              </label>
              <select
                value={activeOptions.resolution || 'original'}
                onChange={(e) =>
                  onChangeOptions({ ...activeOptions, resolution: e.target.value })
                }
                className="w-full px-2.5 py-1.5 rounded-lg bg-[#090D16] border border-white/[0.1] text-xs text-slate-100"
              >
                <option value="original">Original</option>
                <option value="1080p">1080p</option>
                <option value="720p">720p</option>
                <option value="480p">480p</option>
              </select>
            </div>
          )}

          {/* FPS */}
          {!isAudioPreset && (
            <div>
              <label className="block text-[11px] text-slate-400 mb-1">
                {tr(lang, 'fps')}
              </label>
              <select
                value={activeOptions.fps || 0}
                onChange={(e) =>
                  onChangeOptions({ ...activeOptions, fps: Number(e.target.value) })
                }
                className="w-full px-2.5 py-1.5 rounded-lg bg-[#090D16] border border-white/[0.1] text-xs text-slate-100"
              >
                <option value={0}>Original</option>
                <option value={60}>60 fps</option>
                <option value={30}>30 fps</option>
                <option value={24}>24 fps</option>
              </select>
            </div>
          )}

          {/* Rotation */}
          {!isAudioPreset && (
            <div>
              <label className="block text-[11px] text-slate-400 mb-1 flex items-center gap-1">
                <RotateCw className="w-3 h-3" /> {tr(lang, 'rotation')}
              </label>
              <select
                value={activeOptions.rotation || 0}
                onChange={(e) =>
                  onChangeOptions({ ...activeOptions, rotation: Number(e.target.value) })
                }
                className="w-full px-2.5 py-1.5 rounded-lg bg-[#090D16] border border-white/[0.1] text-xs text-slate-100"
              >
                <option value={0}>0°</option>
                <option value={90}>90°</option>
                <option value={180}>180°</option>
                <option value={270}>270°</option>
              </select>
            </div>
          )}

          {/* Trim Start */}
          <div>
            <label className="block text-[11px] text-slate-400 mb-1 flex items-center gap-1">
              <Scissors className="w-3 h-3" /> {tr(lang, 'trimStart')}
            </label>
            <input
              type="text"
              value={trimStartStr}
              onChange={(e) => {
                setTrimStartStr(e.target.value);
                onChangeOptions({
                  ...activeOptions,
                  trimStartSec: parseTimeInput(e.target.value),
                });
              }}
              placeholder="0 or 0:15"
              className="w-full px-2.5 py-1.5 rounded-lg bg-[#090D16] border border-white/[0.1] text-xs font-mono-telemetry text-slate-100"
            />
          </div>

          {/* Trim End */}
          <div>
            <label className="block text-[11px] text-slate-400 mb-1 flex items-center gap-1">
              <Scissors className="w-3 h-3" /> {tr(lang, 'trimEnd')}
            </label>
            <input
              type="text"
              value={trimEndStr}
              onChange={(e) => {
                setTrimEndStr(e.target.value);
                onChangeOptions({
                  ...activeOptions,
                  trimEndSec: parseTimeInput(e.target.value),
                });
              }}
              placeholder="0 or 1:30"
              className="w-full px-2.5 py-1.5 rounded-lg bg-[#090D16] border border-white/[0.1] text-xs font-mono-telemetry text-slate-100"
            />
          </div>

          {/* Mute Toggle */}
          {!isAudioPreset && (
            <div className="flex items-end">
              <button
                type="button"
                onClick={() =>
                  onChangeOptions({ ...activeOptions, mute: !activeOptions.mute })
                }
                className={`w-full px-2.5 py-1.5 rounded-lg border text-xs font-medium flex items-center justify-center gap-1.5 cursor-pointer transition-colors ${
                  activeOptions.mute
                    ? 'bg-rose-500/20 border-rose-500/40 text-rose-200'
                    : 'bg-[#090D16] border-white/[0.1] text-slate-300 hover:border-white/25'
                }`}
              >
                <VolumeX className="w-3.5 h-3.5" />
                <span>{tr(lang, 'muteAudio')}</span>
              </button>
            </div>
          )}

          {/* Output folder */}
          <div className="col-span-2">
            <label className="block text-[11px] text-slate-400 mb-1">
              {tr(lang, 'outputDir')}
            </label>
            <div className="flex items-center gap-1.5">
              <div className="flex-1 px-2.5 py-1.5 rounded-lg bg-[#090D16] border border-white/[0.1] text-xs text-slate-300 truncate font-mono-telemetry">
                {activeOptions.outputDir || tr(lang, 'sameFolder')}
              </div>
              <button
                type="button"
                onClick={onPickOutputDir}
                className="px-2.5 py-1.5 rounded-lg bg-white/[0.06] hover:bg-white/[0.12] text-xs text-slate-200 border border-white/[0.08] cursor-pointer"
              >
                {tr(lang, 'chooseFolder')}
              </button>
              {activeOptions.outputDir && (
                <button
                  type="button"
                  onClick={() => onChangeOptions({ ...activeOptions, outputDir: '' })}
                  className="px-2 py-1.5 rounded-lg bg-white/[0.04] hover:bg-white/[0.1] text-xs text-slate-400 border border-white/[0.08] cursor-pointer"
                >
                  {tr(lang, 'resetFolder')}
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Queue Section */}
      <div className="mp-card rounded-xl p-4 space-y-3">
        <div className="flex items-center justify-between gap-3 flex-wrap">
          <div className="flex items-center gap-2">
            <h3 className="text-sm font-semibold text-slate-100">
              {tr(lang, 'queueTitle')}
            </h3>
            <span className="font-mono-telemetry text-xs px-2 py-0.5 rounded-full bg-white/[0.06] text-slate-300">
              {queue.jobs.length}
            </span>
          </div>

          <div className="flex items-center gap-2">
            {queue.doneCount > 0 && !queue.running && (
              <button
                type="button"
                onClick={onClearCompleted}
                className="px-3 py-1.5 rounded-lg text-xs font-medium bg-white/[0.05] hover:bg-white/[0.1] text-slate-300 border border-white/[0.08] cursor-pointer transition-colors"
              >
                {tr(lang, 'clearDone')}
              </button>
            )}

            {queue.running ? (
              <button
                type="button"
                onClick={onCancelQueue}
                className="px-4 py-1.5 rounded-lg text-xs font-semibold bg-rose-600 hover:bg-rose-500 text-white flex items-center gap-1.5 cursor-pointer transition-colors"
              >
                <Square className="w-3.5 h-3.5 fill-current" />
                <span>{tr(lang, 'cancelQueue')}</span>
              </button>
            ) : (
              <button
                type="button"
                onClick={onStartQueue}
                disabled={queue.pending === 0 || !status.binaries?.ready}
                className="px-4 py-1.5 rounded-lg text-xs font-semibold bg-emerald-600 hover:bg-emerald-500 disabled:opacity-40 text-white flex items-center gap-1.5 cursor-pointer transition-colors"
              >
                <Play className="w-3.5 h-3.5 fill-current" />
                <span>
                  {tr(lang, 'startQueue')} {queue.pending > 0 ? `(${queue.pending})` : ''}
                </span>
              </button>
            )}
          </div>
        </div>

        {queue.jobs.length === 0 ? (
          <div className="py-10 text-center text-xs text-slate-400">
            {tr(lang, 'queueEmpty')}
          </div>
        ) : (
          <div className="space-y-2.5">
            {queue.jobs.map((job) => (
              <JobCard
                key={job.id}
                job={job}
                lang={lang}
                activeOptions={activeOptions}
                onApplyActive={() => onUpdateJobOptions(job.id, activeOptions)}
                onRemove={() => onRemoveJob(job.id)}
                onRetry={() => onRetryJob(job.id)}
                onReveal={() => onRevealFile(job.outputPath)}
                onOpen={() => onOpenFile(job.outputPath)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function JobCard({
  job,
  lang,
  activeOptions,
  onApplyActive,
  onRemove,
  onRetry,
  onReveal,
  onOpen,
}: {
  job: Job;
  lang: UILang;
  activeOptions: ConvertOptions;
  onApplyActive: () => void;
  onRemove: () => void;
  onRetry: () => void;
  onReveal: () => void;
  onOpen: () => void;
}) {
  const { info, plan, status } = job;
  const tv = plan.tvStrategy;
  const savedPct =
    job.status === 'done' && info.sizeBytes > 0 && job.outputBytes > 0
      ? Math.round(((info.sizeBytes - job.outputBytes) / info.sizeBytes) * 100)
      : 0;

  const isAudioOnly = !info.hasVideo || job.options.presetId === 'audio_extract';
  const differsFromActive =
    job.status === 'queued' && job.options.presetId !== activeOptions.presetId;

  return (
    <div className="rounded-xl bg-[#090D16] border border-white/[0.08] p-3.5 space-y-2.5">
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-start gap-3 min-w-0">
          <div className="h-9 w-9 rounded-lg bg-white/[0.04] border border-white/[0.08] flex items-center justify-center shrink-0 mt-0.5">
            {isAudioOnly ? (
              <Music className="w-4 h-4 text-rose-300" />
            ) : (
              <FileVideo className="w-4 h-4 text-cyan-300" />
            )}
          </div>
          <div className="min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="text-sm font-medium text-slate-100 truncate max-w-[360px]">
                {info.fileName}
              </span>
              <span className="font-mono-telemetry text-[10px] px-1.5 py-0.5 rounded bg-violet-500/15 text-violet-200 border border-violet-500/25">
                {job.options.presetId}
              </span>
              <span className="font-mono-telemetry text-[10px] px-1.5 py-0.5 rounded bg-white/[0.05] text-slate-300 border border-white/[0.08]">
                {plan.effectiveEncoder}
              </span>

              {tv && (
                <span
                  title={tv.reason}
                  className={`font-mono-telemetry text-[10px] px-1.5 py-0.5 rounded border flex items-center gap-1 ${
                    tv.mode === 'remux'
                      ? 'bg-emerald-500/15 text-emerald-300 border-emerald-500/35'
                      : tv.mode === 'audio_only'
                      ? 'bg-cyan-500/15 text-cyan-300 border-cyan-500/35'
                      : 'bg-amber-500/15 text-amber-300 border-amber-500/35'
                  }`}
                >
                  <Zap className="w-3 h-3" />
                  {tv.mode === 'remux'
                    ? tr(lang, 'tvRemux')
                    : tv.mode === 'audio_only'
                    ? tr(lang, 'tvAudioOnly')
                    : tr(lang, 'tvTranscode')}
                </span>
              )}
            </div>

            <div className="flex items-center gap-3 text-[11px] text-slate-400 font-mono-telemetry mt-1 flex-wrap">
              <span>{formatBytes(info.sizeBytes)}</span>
              <span>{formatDuration(info.durationSec)}</span>
              {info.hasVideo && (
                <span>
                  {info.width}×{info.height} · {info.videoCodec.toUpperCase()}
                </span>
              )}
              {info.hasAudio && <span>{info.audioCodec.toUpperCase()}</span>}
              {plan.videoBitrateKbps > 0 && (
                <span className="text-violet-300">→ {plan.videoBitrateKbps} kbps</span>
              )}
            </div>
          </div>
        </div>

        {/* Action buttons */}
        <div className="flex items-center gap-1.5 shrink-0">
          {differsFromActive && (
            <button
              type="button"
              onClick={onApplyActive}
              className="px-2 py-1 rounded text-[11px] bg-cyan-500/15 hover:bg-cyan-500/25 text-cyan-200 border border-cyan-500/30 cursor-pointer"
            >
              → {activeOptions.presetId}
            </button>
          )}
          {status === 'done' && (
            <>
              <button
                type="button"
                onClick={onOpen}
                className="px-2.5 py-1 rounded-md text-xs bg-white/[0.06] hover:bg-white/[0.12] text-slate-200 border border-white/[0.08] cursor-pointer"
              >
                {tr(lang, 'playFile')}
              </button>
              <button
                type="button"
                onClick={onReveal}
                className="px-2.5 py-1 rounded-md text-xs bg-emerald-500/15 hover:bg-emerald-500/25 text-emerald-200 border border-emerald-500/30 flex items-center gap-1 cursor-pointer"
              >
                <FolderOpen className="w-3.5 h-3.5" />
                <span>{tr(lang, 'openFolder')}</span>
              </button>
            </>
          )}
          {(status === 'error' || status === 'cancelled') && (
            <button
              type="button"
              onClick={onRetry}
              className="px-2.5 py-1 rounded-md text-xs bg-amber-500/15 hover:bg-amber-500/25 text-amber-200 border border-amber-500/30 flex items-center gap-1 cursor-pointer"
            >
              <RefreshCw className="w-3.5 h-3.5" />
              <span>{tr(lang, 'retry')}</span>
            </button>
          )}
          {status !== 'running' && (
            <button
              type="button"
              onClick={onRemove}
              title={tr(lang, 'remove')}
              className="p-1.5 rounded-md bg-white/[0.04] hover:bg-rose-500/20 text-slate-400 hover:text-rose-300 border border-white/[0.06] cursor-pointer transition-colors"
            >
              <Trash2 className="w-3.5 h-3.5" />
            </button>
          )}
        </div>
      </div>

      {/* Warning or Fallback notice */}
      {(plan.warning || job.fallbackNotice) && (
        <div className="flex items-center gap-1.5 text-[11px] text-amber-300 bg-amber-500/10 border border-amber-500/25 rounded-md px-2.5 py-1">
          <AlertTriangle className="w-3.5 h-3.5 shrink-0" />
          <span>{job.fallbackNotice || plan.warning}</span>
        </div>
      )}

      {/* Progress / Result bar */}
      {status === 'running' && (
        <div className="space-y-1">
          <div className="h-2 w-full rounded-full bg-white/[0.06] overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-violet-500 to-cyan-400 transition-all duration-150"
              style={{ width: `${Math.min(100, Math.max(2, job.progressPercent))}%` }}
            />
          </div>
          <div className="flex items-center justify-between text-[11px] font-mono-telemetry text-slate-400">
            <span>
              {job.progressPercent.toFixed(1)}%
              {plan.usesTwoPass ? ` · Pass ${job.pass || 1}/2` : ''}
            </span>
            <span>
              {job.speed ? `${job.speed} · ` : ''}
              {job.etaSec > 0 ? `ETA ${Math.ceil(job.etaSec)}s` : ''}
            </span>
          </div>
        </div>
      )}

      {status === 'done' && (
        <div className="flex items-center justify-between text-[11px] font-mono-telemetry text-emerald-300 bg-emerald-500/10 border border-emerald-500/25 rounded-md px-2.5 py-1.5">
          <span className="flex items-center gap-1.5 truncate">
            <CheckCircle2 className="w-3.5 h-3.5 shrink-0" />
            <span className="truncate">{job.outputPath}</span>
          </span>
          <span className="shrink-0 ml-2">
            {formatBytes(job.outputBytes)}
            {savedPct > 0 ? ` (-${savedPct}%)` : ''} · {job.processSec.toFixed(1)}s
          </span>
        </div>
      )}

      {status === 'error' && (
        <div className="text-[11px] text-rose-300 bg-rose-500/10 border border-rose-500/25 rounded-md px-2.5 py-1.5 font-mono-telemetry break-all">
          {job.error}
        </div>
      )}
    </div>
  );
}
