import { useState, useEffect } from 'react';
import {
  RefreshCw,
  CheckCircle2,
  XCircle,
  Save,
  Download,
  ExternalLink,
  ArrowUpCircle,
  ShieldCheck,
  X,
} from 'lucide-react';
import type {
  AppConfig,
  FFmpegDownloadProgress,
  SystemStatus,
  UILang,
  UpdateCheckResult,
} from '../types';
import { tr } from '../i18n';
import { formatBytes } from './ConverterView';

interface SettingsViewProps {
  lang: UILang;
  appVersion: string;
  status: SystemStatus;
  config: AppConfig;
  ffmpegDownload: FFmpegDownloadProgress;
  onDownloadFFmpeg: () => Promise<void>;
  onCancelFFmpegDownload: () => Promise<void>;
  checkingUpdates: boolean;
  updateResult: UpdateCheckResult | null;
  onCheckUpdates: () => Promise<void>;
  onOpenExternalURL: (url: string) => Promise<void>;
  onRefreshHardware: () => Promise<void>;
  onSaveConfig: (cfg: AppConfig) => Promise<void>;
  onPickOutputDir: () => Promise<string>;
  onPickFFmpegDir: () => Promise<string>;
}

export function SettingsView({
  lang,
  appVersion,
  status,
  config,
  ffmpegDownload,
  onDownloadFFmpeg,
  onCancelFFmpegDownload,
  checkingUpdates,
  updateResult,
  onCheckUpdates,
  onOpenExternalURL,
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
          <div className="rounded-lg bg-[#090D16] border border-white/[0.08] p-3 space-y-2">
            <span className="text-slate-400">FFmpeg Binary</span>
            <p className="font-mono-telemetry text-slate-100 break-all">
              {status.binaries?.ffmpegPath || status.binaries?.error || 'No encontrado'}
            </p>
            {status.binaries?.ready ? (
              <p className="text-[11px] text-emerald-300 font-mono-telemetry">
                v{status.binaries.version} ({status.binaries.source})
              </p>
            ) : (
              (() => {
                const isDownloading =
                  ffmpegDownload.stage === 'connecting' ||
                  ffmpegDownload.stage === 'downloading' ||
                  ffmpegDownload.stage === 'extracting' ||
                  ffmpegDownload.stage === 'verifying';
                if (isDownloading) {
                  return (
                    <div className="space-y-1.5 pt-1">
                      <div className="h-1.5 w-full rounded-full bg-white/[0.08] overflow-hidden">
                        <div
                          className="h-full bg-gradient-to-r from-violet-500 to-cyan-400 transition-all duration-150"
                          style={{ width: `${Math.min(100, Math.max(4, ffmpegDownload.percent))}%` }}
                        />
                      </div>
                      <div className="flex items-center justify-between gap-2 text-[11px] text-cyan-200 font-mono-telemetry">
                        <span>
                          {ffmpegDownload.stage === 'downloading' && ffmpegDownload.totalBytes > 0
                            ? `${formatBytes(ffmpegDownload.downloadedBytes)} / ${formatBytes(ffmpegDownload.totalBytes)} (${Math.round(ffmpegDownload.percent)}%)`
                            : `${tr(lang, 'downloadingFfmpeg')} (${Math.round(ffmpegDownload.percent)}%)`}
                        </span>
                        <button
                          type="button"
                          onClick={onCancelFFmpegDownload}
                          className="px-2 py-0.5 rounded bg-rose-500/20 hover:bg-rose-500/30 text-rose-200 border border-rose-500/35 flex items-center gap-1 cursor-pointer"
                        >
                          <X className="w-3 h-3" />
                          <span>{tr(lang, 'cancelQueue')}</span>
                        </button>
                      </div>
                    </div>
                  );
                }
                return (
                  <div className="space-y-1.5 pt-1">
                    <button
                      type="button"
                      onClick={onDownloadFFmpeg}
                      className="px-3 py-1.5 rounded-lg text-xs font-medium bg-violet-600 hover:bg-violet-500 text-white inline-flex items-center gap-1.5 cursor-pointer transition-colors"
                    >
                      <Download className="w-3.5 h-3.5" />
                      <span>
                        {ffmpegDownload.stage === 'error'
                          ? tr(lang, 'retryDownload')
                          : tr(lang, 'downloadFfmpeg')}
                      </span>
                    </button>
                    {ffmpegDownload.stage === 'error' && ffmpegDownload.error && (
                      <p className="text-[11px] text-rose-300 font-mono-telemetry break-all">
                        {tr(lang, 'ffmpegDownloadError')} {ffmpegDownload.error}
                      </p>
                    )}
                  </div>
                );
              })()
            )}
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

      {/* Manual Opt-In Updates Card (0% Telemetry) */}
      <div className="mp-card rounded-xl p-4 space-y-3">
        <div className="flex items-start justify-between gap-3 flex-wrap">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <ArrowUpCircle className="w-4 h-4 text-violet-400" />
              <h3 className="text-sm font-semibold text-slate-100">
                {tr(lang, 'updatesTitle')}
              </h3>
              <span className="font-mono-telemetry text-[11px] px-2 py-0.5 rounded bg-white/[0.06] text-slate-300 border border-white/[0.08]">
                {tr(lang, 'currentVersionLabel')}: v{appVersion}
              </span>
            </div>
            <p className="text-xs text-slate-400">{tr(lang, 'updatesSub')}</p>
          </div>

          <button
            type="button"
            onClick={onCheckUpdates}
            disabled={checkingUpdates}
            className="px-3.5 py-2 rounded-lg text-xs font-semibold bg-cyan-500/15 hover:bg-cyan-500/25 disabled:opacity-50 text-cyan-200 border border-cyan-500/30 flex items-center gap-1.5 cursor-pointer transition-colors"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${checkingUpdates ? 'animate-spin' : ''}`} />
            <span>
              {checkingUpdates ? tr(lang, 'checkingUpdates') : tr(lang, 'checkUpdates')}
            </span>
          </button>
        </div>

        <div className="flex items-center gap-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20 px-3 py-2 text-[11px] text-emerald-300">
          <ShieldCheck className="w-4 h-4 shrink-0" />
          <span>{tr(lang, 'updatesPrivacyNote')}</span>
        </div>

        {updateResult && (
          <>
            {updateResult.error ? (
              <div className="rounded-lg bg-rose-500/10 border border-rose-500/25 p-3 text-xs text-rose-200 flex items-center justify-between gap-3 flex-wrap">
                <span>
                  {tr(lang, 'updateCheckFailed')} {updateResult.error}
                </span>
                <button
                  type="button"
                  onClick={() => onOpenExternalURL(updateResult.releaseUrl)}
                  className="px-2.5 py-1 rounded-md bg-white/[0.08] hover:bg-white/[0.15] text-slate-100 inline-flex items-center gap-1.5 cursor-pointer"
                >
                  <ExternalLink className="w-3.5 h-3.5" />
                  <span>{tr(lang, 'openGithubRelease')}</span>
                </button>
              </div>
            ) : updateResult.updateAvailable ? (
              <div className="rounded-lg bg-violet-500/15 border border-violet-500/35 p-3.5 space-y-2">
                <div className="flex items-center justify-between gap-3 flex-wrap">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-semibold text-emerald-300">
                      {tr(lang, 'updateAvailable')} v{updateResult.latestVersion}
                    </span>
                    {updateResult.publishedAt && (
                      <span className="font-mono-telemetry text-[11px] text-slate-400">
                        {updateResult.publishedAt.slice(0, 10)}
                      </span>
                    )}
                  </div>
                  <button
                    type="button"
                    onClick={() => onOpenExternalURL(updateResult.releaseUrl)}
                    className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-violet-600 hover:bg-violet-500 text-white inline-flex items-center gap-1.5 cursor-pointer transition-colors"
                  >
                    <ExternalLink className="w-3.5 h-3.5" />
                    <span>{tr(lang, 'openGithubRelease')}</span>
                  </button>
                </div>
                {updateResult.releaseNotes && (
                  <p className="text-[11px] text-slate-300 whitespace-pre-line font-mono-telemetry bg-[#090D16] border border-white/[0.07] rounded-md p-2.5">
                    {updateResult.releaseNotes}
                  </p>
                )}
              </div>
            ) : (
              <div className="rounded-lg bg-emerald-500/10 border border-emerald-500/25 px-3 py-2 text-xs text-emerald-200 flex items-center justify-between gap-2">
                <span className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
                  <span>{tr(lang, 'updateUpToDate')}</span>
                </span>
                {updateResult.latestVersion && (
                  <span className="font-mono-telemetry text-[11px] text-slate-300">
                    {tr(lang, 'latestVersionLabel')}: v{updateResult.latestVersion}
                  </span>
                )}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
