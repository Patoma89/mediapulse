import { useEffect, useRef, useState } from 'react';
import {
  AddFiles,
  ApplyOptionsToQueued,
  CancelCurrent,
  ClearCompleted,
  ClearHistory,
  GetAppInfo,
  GetConfig,
  GetHistory,
  GetPresets,
  GetQueue,
  GetSystemStatus,
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
import { ConverterView } from './views/ConverterView';
import { PresetsView } from './views/PresetsView';
import { HistoryView } from './views/HistoryView';
import { SettingsView } from './views/SettingsView';
import type {
  AppConfig,
  AppInfo,
  ConvertOptions,
  HistorySummary,
  NavSection,
  PresetDefinition,
  PresetID,
  QueueState,
  SystemStatus,
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
    version: '1.0.0',
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
      />

      <div className="flex-1 flex flex-col min-w-0">
        <TopBar
          section={section}
          status={status}
          lang={config.language}
          theme={config.theme}
          onToggleLang={handleToggleLang}
          onToggleTheme={handleToggleTheme}
        />

        <main className="flex-1 overflow-y-auto mp-scroll p-6">
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
              status={status}
              config={config}
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
