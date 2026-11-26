import { useState, useEffect } from 'react';
import { RefreshCw, CheckCircle2, XCircle, Save } from 'lucide-react';
import type { AppConfig, SystemStatus, UILang } from '../types';
import { tr } from '../i18n';

interface SettingsViewProps {
  lang: UILang;
  status: SystemStatus;
  config: AppConfig;
  onRefreshHardware: () => Promise<void>;
  onSaveConfig: (cfg: AppConfig) => Promise<void>;
  onPickOutputDir: () => Promise<string>;
  onPickFFmpegDir: () => Promise<string>;
}

export function SettingsView({
  lang,
  status,
  config,
  onRefreshHardware,
  onSaveConfig,
  onPickOutputDir,
  onPickFFmpegDir,
}: SettingsViewProps) {
  const [draft, setDraft] = useState<AppConfig>(config);
  const [refreshing, setRefreshing] = useState(false);
  const [savedToast, setSavedToast] = useState(false);

  useEffect(() => {
    setDraft(config);
  }, [config]);

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await onRefreshHardware();
    } finally {
      setRefreshing(false);
    }
  };

  const handleSave = async () => {
    await onSaveConfig(draft);
    setSavedToast(true);
    setTimeout(() => setSavedToast(false), 2500);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-slate-100">
            {tr(lang, 'settingsTitle')}
          </h2>
          <p className="text-xs text-slate-400 mt-0.5">{tr(lang, 'settingsSub')}</p>
        </div>
        <button
          type="button"
          onClick={handleRefresh}
          disabled={refreshing}
          className="px-3 py-1.5 rounded-lg text-xs font-medium bg-cyan-500/15 hover:bg-cyan-500/25 text-cyan-200 border border-cyan-500/30 flex items-center gap-1.5 cursor-pointer"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${refreshing ? 'animate-spin' : ''}`} />
          <span>{tr(lang, 'redetectGpu')}</span>
        </button>
      </div>

      {/* FFmpeg + GPU Card */}
      <div className="mp-card rounded-xl p-4 space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
          <div className="rounded-lg bg-[#090D16] border border-white/[0.08] p-3 space-y-1">
            <span className="text-slate-400">FFmpeg Binary</span>
            <p className="font-mono-telemetry text-slate-100 break-all">
              {status.binaries?.ffmpegPath || status.binaries?.error || 'No encontrado'}
            </p>
            <p className="text-[11px] text-emerald-300 font-mono-telemetry">
              {status.binaries?.ready
                ? `v${status.binaries.version} (${status.binaries.source})`
                : ''}
            </p>
          </div>
          <div className="rounded-lg bg-[#090D16] border border-white/[0.08] p-3 space-y-1">
            <span className="text-slate-400">GPU Detectada</span>
            <p className="font-mono-telemetry text-violet-200 font-semibold">
              {status.hardware?.gpuName || 'CPU'}
            </p>
            <p className="text-[11px] text-slate-400 font-mono-telemetry">
              Recomendado: {status.hardware?.recommended?.toUpperCase()}
            </p>
          </div>
        </div>

        {/* Encoders Matrix */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
          {(status.hardware?.encoders || []).map((enc) => (
            <div
              key={enc.name}
              className={`rounded-lg p-2.5 border text-xs flex items-center justify-between ${
                enc.available
                  ? 'bg-emerald-500/10 border-emerald-500/25 text-emerald-200'
                  : 'bg-white/[0.02] border-white/[0.06] text-slate-500'
              }`}
            >
              <div className="min-w-0">
                <p className="font-mono-telemetry font-medium truncate">{enc.name}</p>
                <p className="text-[10px] opacity-80 truncate">{enc.label}</p>
              </div>
              {enc.available ? (
                <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
              ) : (
                <XCircle className="w-4 h-4 text-slate-600 shrink-0" />
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Preferences Form */}
      <div className="mp-card rounded-xl p-4 space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
          <div>
            <label className="block text-slate-400 mb-1">
              {tr(lang, 'preferredEncoder')}
            </label>
            <select
              value={draft.preferredEncoder}
              onChange={(e) => setDraft({ ...draft, preferredEncoder: e.target.value })}
              className="w-full px-3 py-2 rounded-lg bg-[#090D16] border border-white/[0.1] text-slate-100"
            >
              <option value="auto">Auto (NVENC → AMF → QSV → CPU)</option>
              <option value="nvenc">NVIDIA NVENC</option>
              <option value="amf">AMD AMF</option>
              <option value="qsv">Intel QuickSync (QSV)</option>
              <option value="cpu">CPU (x264 / x265)</option>
            </select>
          </div>

          <div className="flex items-end">
            <label className="flex items-center gap-2.5 cursor-pointer py-2">
              <input
                type="checkbox"
                checked={draft.openFolderOnDone}
                onChange={(e) =>
                  setDraft({ ...draft, openFolderOnDone: e.target.checked })
                }
                className="w-4 h-4 accent-violet-500 rounded"
              />
              <span className="text-slate-200">{tr(lang, 'openFolderOnDone')}</span>
            </label>
          </div>

          <div>
            <label className="block text-slate-400 mb-1">
              {tr(lang, 'defaultOutputDir')}
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={draft.defaultOutputDir}
                onChange={(e) =>
                  setDraft({ ...draft, defaultOutputDir: e.target.value })
                }
                placeholder={tr(lang, 'sameFolder')}
                className="flex-1 px-3 py-2 rounded-lg bg-[#090D16] border border-white/[0.1] text-slate-100 font-mono-telemetry"
              />
              <button
                type="button"
                onClick={async () => {
                  const dir = await onPickOutputDir();
                  if (dir) setDraft({ ...draft, defaultOutputDir: dir });
                }}
                className="px-3 py-2 rounded-lg bg-white/[0.06] hover:bg-white/[0.12] text-slate-200 border border-white/[0.08] cursor-pointer"
              >
                {tr(lang, 'chooseFolder')}
              </button>
            </div>
          </div>

          <div>
            <label className="block text-slate-400 mb-1">
              {tr(lang, 'customFfmpegDir')}
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={draft.customFfmpegDir}
                onChange={(e) =>
                  setDraft({ ...draft, customFfmpegDir: e.target.value })
                }
                placeholder="bin\\ffmpeg.exe (auto)"
                className="flex-1 px-3 py-2 rounded-lg bg-[#090D16] border border-white/[0.1] text-slate-100 font-mono-telemetry"
              />
              <button
                type="button"
                onClick={async () => {
                  const dir = await onPickFFmpegDir();
                  if (dir) setDraft({ ...draft, customFfmpegDir: dir });
                }}
                className="px-3 py-2 rounded-lg bg-white/[0.06] hover:bg-white/[0.12] text-slate-200 border border-white/[0.08] cursor-pointer"
              >
                {tr(lang, 'chooseFolder')}
              </button>
            </div>
          </div>
        </div>

        <div className="flex items-center justify-end gap-3 pt-2 border-t border-white/[0.06]">
          {savedToast && (
            <span className="text-xs text-emerald-300">{tr(lang, 'savedNotice')}</span>
          )}
          <button
            type="button"
            onClick={handleSave}
            className="px-4 py-2 rounded-lg text-xs font-semibold bg-violet-600 hover:bg-violet-500 text-white flex items-center gap-1.5 cursor-pointer"
          >
            <Save className="w-3.5 h-3.5" />
            <span>{tr(lang, 'saveSettings')}</span>
          </button>
        </div>
      </div>
    </div>
  );
}
