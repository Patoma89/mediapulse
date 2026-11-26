import { useEffect, useRef, useState } from 'react';
import { AlertTriangle, Download, FolderOpen, Loader2, X } from 'lucide-react';
import {
  AddFiles,
  ApplyOptionsToQueued,
  CancelCurrent,
  CancelFFmpegDownload,
  CheckForUpdates,
  ClearCompleted,
  ClearHistory,
  DownloadFFmpeg,
  GetAppInfo,
  GetConfig,
  GetHistory,
  GetPresets,
  GetQueue,
  GetSystemStatus,
  OpenExternalURL,
  OpenMediaFile,
  PickFFmpegDir,
  PickFiles,
  PickOutputDir,
  RefreshHardware,
  RemoveJob,
  RetryJob,
  RevealInExplorer,
  StartQueue,
  UpdateConfig,
  UpdateJobOptions,
} from '../wailsjs/go/main/App';
import { EventsOn, OnFileDrop, OnFileDropOff } from '../wailsjs/runtime/runtime';
import { Sidebar } from './components/Sidebar';
import { TopBar } from './components/TopBar';
import { ConverterView, formatBytes } from './views/ConverterView';
import { PresetsView } from './views/PresetsView';
import { HistoryView } from './views/HistoryView';
import { SettingsView } from './views/SettingsView';
import { tr } from './i18n';
import type {
  AppConfig,
  AppInfo,
  ConvertOptions,
  FFmpegDownloadProgress,
  HistorySummary,
  NavSection,
  PresetDefinition,
  PresetID,
  QueueState,
  SystemStatus,
  UpdateCheckResult,
} from './types';

function makeDefaultOptions(preset: PresetID, encoderMode: string, outputDir: string): ConvertOptions {
  return {
    presetId: preset,
    targetSizeMb: preset === 'discord_25mb' ? 25 : preset === 'target_size' ? 50 : 10,
    outputFormat: preset === 'audio_extract' ? 'mp3' : 'mp4',
    videoCodec: 'auto',
    qualityCrf: preset === 'compress_hq' ? 22 : 23,
    resolution: 'original',
    fps: 0,
    rotation: 0,
    mute: false,
    trimStartSec: 0,
    trimEndSec: 0,
    audioFormat: 'mp3',
    audioBitrateKbps: 192,
    encoderMode: encoderMode || 'auto',
    outputDir: outputDir || '',
  };
}

export default function App() {
  const [section, setSection] = useState<NavSection>('converter');
  const [appInfo, setAppInfo] = useState<AppInfo>({
    name: 'MediaPulse',
    version: '1.1.0',
    os: 'windows',
    arch: 'amd64',
    configPath: '',
    zeroTelemetry: true,
  });
  const [status, setStatus] = useState<SystemStatus>({
    binaries: { ffmpegPath: '', ffprobePath: '', source: '', version: '', ready: false },
    hardware: { gpuName: 'Detectando...', encoders: [], recommended: 'cpu' },
  });
  const [config, setConfig] = useState<AppConfig>({
    language: 'es',
    theme: 'dark',
    preferredEncoder: 'auto',
    defaultOutputDir: '',
    customFfmpegDir: '',
    openFolderOnDone: false,
    defaultPreset: 'discord_10mb',
  });
  const [presets, setPresets] = useState<PresetDefinition[]>([]);
  const [queue, setQueue] = useState<QueueState>({
    jobs: [],
    running: false,
    doneCount: 0,
    pending: 0,
  });
  const [history, setHistory] = useState<HistorySummary>({
    totalJobs: 0,
    totalInputBytes: 0,
    totalOutputBytes: 0,
    totalSavedBytes: 0,
    entries: [],
  });
  const [activeOptions, setActiveOptions] = useState<ConvertOptions>(() =>
    makeDefaultOptions('discord_10mb', 'auto', '')
  );
  const [dropError, setDropError] = useState<string>('');
  const [ffmpegDownload, setFfmpegDownload] = useState<FFmpegDownloadProgress>({
    stage: 'idle',
    percent: 0,
    downloadedBytes: 0,
    totalBytes: 0,
  });
  const [cliPendingCount, setCliPendingCount] = useState<number>(0);
  const [checkingUpdates, setCheckingUpdates] = useState<boolean>(false);
  const [updateResult, setUpdateResult] = useState<UpdateCheckResult | null>(null);

  const optionsRef = useRef(activeOptions);
  optionsRef.current = activeOptions;

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', config.theme);
  }, [config.theme]);

  useEffect(() => {
    Promise.all([
      GetAppInfo(),
      GetSystemStatus(),
      GetConfig(),
      GetPresets(),
      GetQueue(),
      GetHistory(),
    ]).then(([info, sys, cfg, pres, q, hist]) => {
      setAppInfo(info as AppInfo);
      setStatus(sys as SystemStatus);
      const typedCfg = cfg as AppConfig;
      setConfig(typedCfg);
      setPresets(pres as PresetDefinition[]);
      setQueue(q as QueueState);
      setHistory(hist as HistorySummary);
      setActiveOptions(
        makeDefaultOptions(
          typedCfg.defaultPreset || 'discord_10mb',
          typedCfg.preferredEncoder || 'auto',
          typedCfg.defaultOutputDir || ''
        )
      );
    });

    const offSys = EventsOn('system:status', (sys: SystemStatus) => {
      setStatus(sys);
      if (sys.binaries?.ready) {
        setCliPendingCount(0);
      }
    });
    const offQueue = EventsOn('queue:state', (q: QueueState) => {
      setQueue(q);
    });
    const offProg = EventsOn(
      'job:progress',
      (payload: { jobId: string; progress: { percent: number; speed: string; etaSec: number; pass: number } }) => {
        setQueue((prev) => ({
          ...prev,
          jobs: prev.jobs.map((j) =>
            j.id === payload.jobId
              ? {
                  ...j,
                  progressPercent: payload.progress.percent,
                  speed: payload.progress.speed,
                  etaSec: payload.progress.etaSec,
                  pass: payload.progress.pass,
                }
              : j
          ),
        }));
      }
    );
    const offDone = EventsOn('job:done', () => {
      GetHistory().then((h) => setHistory(h as HistorySummary));
    });
    const offFFmpegDL = EventsOn('ffmpeg:download-progress', (prog: FFmpegDownloadProgress) => {
      setFfmpegDownload(prog);
    });
    const offCLIFiles = EventsOn(
      'cli:files-received',
      (payload: { addedCount: number; pending: number; errors?: string[]; needsFFmpeg?: boolean }) => {
        setSection('converter');
        if (payload.needsFFmpeg || payload.pending > 0) {
          setCliPendingCount(payload.pending);
        } else {
          setCliPendingCount(0);
        }
        if (payload.errors && payload.errors.length > 0) {
          setDropError(payload.errors.join(' | '));
        }
      }
    );

    OnFileDrop((_x, _y, paths) => {
      if (paths && paths.length > 0) {
        AddFiles(paths, optionsRef.current as never).then((res) => {
          if (res.errors && res.errors.length > 0) {
            setDropError(res.errors.join(' | '));
          } else {
            setDropError('');
          }
        });
      }
    }, true);

    return () => {
      offSys();
      offQueue();
      offProg();
      offDone();
      offFFmpegDL();
      offCLIFiles();
      OnFileDropOff();
    };
  }, []);

  const handleSelectPreset = (id: PresetID) => {
    const next = makeDefaultOptions(id, config.preferredEncoder, activeOptions.outputDir);
    setActiveOptions(next);
  };

  const handlePickFiles = async () => {
    const paths = await PickFiles();
    if (paths && paths.length > 0) {
      const res = await AddFiles(paths, activeOptions as never);
      if (res.errors && res.errors.length > 0) {
        setDropError(res.errors.join(' | '));
      } else {
        setDropError('');
      }
    }
  };

  const handleToggleLang = async () => {
    const nextLang = config.language === 'es' ? 'en' : 'es';
    const updated = await UpdateConfig({ ...config, language: nextLang } as never);
    setConfig(updated as AppConfig);
  };

  const handleToggleTheme = async () => {
    const nextTheme = config.theme === 'dark' ? 'light' : 'dark';
    const updated = await UpdateConfig({ ...config, theme: nextTheme } as never);
    setConfig(updated as AppConfig);
  };

  const handleDownloadFFmpeg = async () => {
    setFfmpegDownload({
      stage: 'connecting',
      percent: 0,
      downloadedBytes: 0,
      totalBytes: 0,
    });
    try {
      const sys = (await DownloadFFmpeg()) as SystemStatus;
      setStatus(sys);
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      if (msg.toLowerCase().includes('cancel')) {
        setFfmpegDownload((prev) => ({ ...prev, stage: 'cancelled', error: '' }));
      } else {
        setFfmpegDownload((prev) => ({ ...prev, stage: 'error', error: msg }));
      }
    }
  };

  const handleCancelFFmpegDownload = async () => {
    await CancelFFmpegDownload();
  };

  const handlePickAndSaveFFmpegDir = async () => {
    const dir = await PickFFmpegDir();
    if (dir) {
      const updated = (await UpdateConfig({ ...config, customFfmpegDir: dir } as never)) as AppConfig;
      setConfig(updated);
      const sys = (await RefreshHardware()) as SystemStatus;
      setStatus(sys);
    }
  };

  const handleCheckUpdates = async () => {
    setCheckingUpdates(true);
    try {
      const res = (await CheckForUpdates()) as UpdateCheckResult;
      setUpdateResult(res);
    } finally {
      setCheckingUpdates(false);
    }
  };

  const handleOpenExternalURL = async (url: string) => {
    await OpenExternalURL(url);
  };

  const isFFmpegDownloading =
    ffmpegDownload.stage === 'connecting' ||
    ffmpegDownload.stage === 'downloading' ||
    ffmpegDownload.stage === 'extracting' ||
    ffmpegDownload.stage === 'verifying';

  const stageLabel = () => {
    switch (ffmpegDownload.stage) {
      case 'connecting':
        return tr(config.language, 'ffmpegStageConnecting');
      case 'downloading':
        return tr(config.language, 'ffmpegStageDownloading');
      case 'extracting':
        return tr(config.language, 'ffmpegStageExtracting');
      case 'verifying':
        return tr(config.language, 'ffmpegStageVerifying');
      default:
        return '';
    }
  };

  return (
    <div className="h-screen w-screen flex bg-[#090D16] text-slate-100 overflow-hidden">
      <Sidebar
        activeSection={section}
        onSelectSection={setSection}
        version={appInfo.version}
        lang={config.language}
        pendingCount={queue.pending}
        historyCount={history.totalJobs}
        running={queue.running}
        updateAvailable={Boolean(updateResult?.updateAvailable)}
        latestVersion={updateResult?.latestVersion}
        onOpenUpdates={() => setSection('settings')}
      />

      <div className="flex-1 flex flex-col min-w-0">
        <TopBar
          section={section}
          status={status}
          lang={config.language}
          theme={config.theme}
          ffmpegDownload={ffmpegDownload}
          onClickFFmpegStatus={() => setSection('converter')}
          onToggleLang={handleToggleLang}
          onToggleTheme={handleToggleTheme}
        />

        <main className="flex-1 overflow-y-auto mp-scroll p-6">
          {(!status.binaries?.ready || isFFmpegDownloading) && (
            <div className="mb-4 rounded-xl bg-rose-500/10 border border-rose-500/30 p-4 space-y-3">
              <div className="flex items-start justify-between gap-4 flex-wrap">
                <div className="flex items-start gap-3 min-w-0">
                  <div className="h-9 w-9 rounded-lg bg-rose-500/15 border border-rose-500/30 flex items-center justify-center shrink-0 mt-0.5">
                    {isFFmpegDownloading ? (
                      <Loader2 className="w-5 h-5 text-cyan-300 animate-spin" />
                    ) : (
                      <AlertTriangle className="w-5 h-5 text-rose-300" />
                    )}
                  </div>
                  <div className="space-y-1 min-w-0">
                    <h3 className="text-sm font-semibold text-rose-200">
                      {tr(config.language, 'ffmpegMissingTitle')}
                    </h3>
                    <p className="text-xs text-slate-300 leading-relaxed">
                      {tr(config.language, 'ffmpegMissingDesc')}
                    </p>
                    {cliPendingCount > 0 && (
                      <p className="text-xs text-amber-300 font-medium">
                        {cliPendingCount} {tr(config.language, 'cliPendingFfmpeg')}
                      </p>
                    )}
                  </div>
                </div>

                {!isFFmpegDownloading && (
                  <div className="flex items-center gap-2 shrink-0">
                    <button
                      type="button"
                      onClick={handlePickAndSaveFFmpegDir}
                      className="px-3 py-2 rounded-lg text-xs font-medium bg-white/[0.06] hover:bg-white/[0.12] text-slate-200 border border-white/[0.1] flex items-center gap-1.5 cursor-pointer transition-colors"
                    >
                      <FolderOpen className="w-3.5 h-3.5" />
                      <span>{tr(config.language, 'chooseFfmpegFolder')}</span>
                    </button>
                    <button
                      type="button"
                      onClick={handleDownloadFFmpeg}
                      className="px-3.5 py-2 rounded-lg text-xs font-semibold bg-violet-600 hover:bg-violet-500 text-white flex items-center gap-1.5 cursor-pointer transition-colors"
                    >
                      <Download className="w-3.5 h-3.5" />
                      <span>
                        {ffmpegDownload.stage === 'error'
                          ? tr(config.language, 'retryDownload')
                          : tr(config.language, 'downloadFfmpeg')}
                      </span>
                    </button>
                  </div>
                )}
              </div>

              {isFFmpegDownloading && (
                <div className="space-y-1.5">
                  <div className="h-2 w-full rounded-full bg-white/[0.08] overflow-hidden">
                    <div
                      className="h-full bg-gradient-to-r from-violet-500 to-cyan-400 transition-all duration-150"
                      style={{ width: `${Math.min(100, Math.max(4, ffmpegDownload.percent))}%` }}
                    />
                  </div>
                  <div className="flex items-center justify-between gap-2 text-[11px] font-mono-telemetry text-slate-300">
                    <span>
                      {stageLabel()}
                      {ffmpegDownload.source ? ` (${ffmpegDownload.source})` : ''}
                    </span>
                    <div className="flex items-center gap-3">
                      <span>
                        {ffmpegDownload.totalBytes > 0
                          ? `${formatBytes(ffmpegDownload.downloadedBytes)} / ${formatBytes(ffmpegDownload.totalBytes)} · `
                          : ffmpegDownload.downloadedBytes > 0
                          ? `${formatBytes(ffmpegDownload.downloadedBytes)} · `
                          : ''}
                        {Math.round(ffmpegDownload.percent)}%
                      </span>
                      <button
                        type="button"
                        onClick={handleCancelFFmpegDownload}
                        className="px-2 py-0.5 rounded bg-rose-500/20 hover:bg-rose-500/35 text-rose-200 border border-rose-500/35 inline-flex items-center gap-1 cursor-pointer"
                      >
                        <X className="w-3 h-3" />
                        <span>{tr(config.language, 'cancelQueue')}</span>
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {ffmpegDownload.stage === 'error' && ffmpegDownload.error && (
                <p className="text-xs text-rose-300 font-mono-telemetry break-all">
                  {tr(config.language, 'ffmpegDownloadError')} {ffmpegDownload.error}
                </p>
              )}

              {ffmpegDownload.stage === 'cancelled' && (
                <p className="text-xs text-amber-300">
                  {tr(config.language, 'ffmpegDownloadCancelled')}
                </p>
              )}
            </div>
          )}

          {dropError && (
            <div className="mb-4 rounded-lg bg-rose-500/15 border border-rose-500/30 px-3 py-2 text-xs text-rose-200 flex items-center justify-between">
              <span>{dropError}</span>
              <button
                type="button"
                onClick={() => setDropError('')}
                className="ml-3 text-rose-300 hover:text-white cursor-pointer"
              >
                ✕
              </button>
            </div>
          )}

          {section === 'converter' && (
            <ConverterView
              lang={config.language}
              status={status}
              presets={presets}
              queue={queue}
              activeOptions={activeOptions}
              onChangeOptions={setActiveOptions}
              onSelectPreset={handleSelectPreset}
              onPickFiles={handlePickFiles}
              onPickOutputDir={async () => {
                const dir = await PickOutputDir();
                if (dir) setActiveOptions({ ...activeOptions, outputDir: dir });
              }}
              onApplyToAllQueued={() => ApplyOptionsToQueued(activeOptions as never)}
              onStartQueue={() => StartQueue()}
              onCancelQueue={() => CancelCurrent()}
              onClearCompleted={() => ClearCompleted()}
              onRemoveJob={(id) => RemoveJob(id)}
              onRetryJob={(id) => RetryJob(id)}
              onRevealFile={(p) => RevealInExplorer(p)}
              onOpenFile={(p) => OpenMediaFile(p)}
              onUpdateJobOptions={(id, opts) => UpdateJobOptions(id, opts as never)}
            />
          )}

          {section === 'presets' && (
            <PresetsView
              lang={config.language}
              presets={presets}
              activePreset={activeOptions.presetId}
              onUsePreset={(id) => {
                handleSelectPreset(id);
                setSection('converter');
              }}
            />
          )}

          {section === 'history' && (
            <HistoryView
              lang={config.language}
              summary={history}
              onClearHistory={async () => {
                await ClearHistory();
                const h = await GetHistory();
                setHistory(h as HistorySummary);
              }}
              onRevealFile={(p) => RevealInExplorer(p)}
            />
          )}

          {section === 'settings' && (
            <SettingsView
              lang={config.language}
              appVersion={appInfo.version}
              status={status}
              config={config}
              ffmpegDownload={ffmpegDownload}
              onDownloadFFmpeg={handleDownloadFFmpeg}
              onCancelFFmpegDownload={handleCancelFFmpegDownload}
              checkingUpdates={checkingUpdates}
              updateResult={updateResult}
              onCheckUpdates={handleCheckUpdates}
              onOpenExternalURL={handleOpenExternalURL}
              onRefreshHardware={async () => {
                const s = await RefreshHardware();
                setStatus(s as SystemStatus);
              }}
              onSaveConfig={async (nextCfg) => {
                const saved = (await UpdateConfig(nextCfg as never)) as AppConfig;
                setConfig(saved);
                setActiveOptions((prev) => ({
                  ...prev,
                  encoderMode: saved.preferredEncoder,
                  outputDir: prev.outputDir || saved.defaultOutputDir,
                }));
              }}
              onPickOutputDir={() => PickOutputDir()}
              onPickFFmpegDir={() => PickFFmpegDir()}
            />
          )}
        </main>
      </div>
    </div>
  );
}
