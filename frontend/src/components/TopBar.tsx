import { Cpu, CheckCircle2, AlertTriangle, Sun, Moon, RefreshCw } from 'lucide-react';
import type { FFmpegDownloadProgress, NavSection, SystemStatus, UILang } from '../types';
import { tr } from '../i18n';

interface TopBarProps {
  section: NavSection;
  status: SystemStatus;
  lang: UILang;
  theme: 'dark' | 'light';
  ffmpegDownload: FFmpegDownloadProgress;
  onClickFFmpegStatus: () => void;
  onToggleLang: () => void;
  onToggleTheme: () => void;
}

export function TopBar({
  section,
  status,
  lang,
  theme,
  ffmpegDownload,
  onClickFFmpegStatus,
  onToggleLang,
  onToggleTheme,
}: TopBarProps) {
  const titleMap: Record<NavSection, string> = {
    converter: tr(lang, 'navConverter'),
    presets: tr(lang, 'presetsPageTitle'),
    history: tr(lang, 'historyTitle'),
    settings: tr(lang, 'settingsTitle'),
  };

  const rec = status.hardware?.recommended?.toUpperCase() || 'CPU';
  const gpuName = status.hardware?.gpuName || 'CPU';
  const shortGpu = gpuName.replace('NVIDIA GeForce ', '').replace('NVIDIA ', '');

  return (
    <header className="h-16 shrink-0 px-6 flex items-center justify-between border-b border-white/[0.07] bg-[#0B101B]/80 backdrop-blur">
      <div className="flex items-center gap-3 min-w-0">
        <h1 className="text-base font-semibold tracking-tight text-slate-100 truncate">
          {titleMap[section]}
        </h1>
      </div>

      <div className="flex items-center gap-2.5">
        {/* GPU Pill */}
        <div
          title={gpuName}
          className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-mono-telemetry border ${
            rec !== 'CPU'
              ? 'bg-violet-500/15 text-violet-200 border-violet-500/30'
              : 'bg-white/[0.05] text-slate-300 border-white/[0.08]'
          }`}
        >
          <Cpu className="w-3.5 h-3.5" />
          <span className="truncate max-w-[160px]">{shortGpu}</span>
          <span className="text-[10px] opacity-80">· {rec}</span>
        </div>

        {/* FFmpeg Pill */}
        {status.binaries?.ready ? (
          <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border bg-emerald-500/10 text-emerald-300 border-emerald-500/25">
            <CheckCircle2 className="w-3.5 h-3.5" />
            <span>{tr(lang, 'ffmpegReady')}</span>
          </div>
        ) : (
          (() => {
            const isDownloading =
              ffmpegDownload.stage === 'connecting' ||
              ffmpegDownload.stage === 'downloading' ||
              ffmpegDownload.stage === 'extracting' ||
              ffmpegDownload.stage === 'verifying';
            return (
              <button
                type="button"
                onClick={onClickFFmpegStatus}
                className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border cursor-pointer transition-colors ${
                  isDownloading
                    ? 'bg-cyan-500/15 text-cyan-200 border-cyan-500/35 hover:bg-cyan-500/25'
                    : 'bg-rose-500/15 text-rose-300 border-rose-500/30 hover:bg-rose-500/25'
                }`}
              >
                {isDownloading ? (
                  <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                ) : (
                  <AlertTriangle className="w-3.5 h-3.5" />
                )}
                <span>
                  {isDownloading
                    ? `${tr(lang, 'downloadingFfmpeg')} ${Math.round(ffmpegDownload.percent)}%`
                    : tr(lang, 'ffmpegMissing')}
                </span>
              </button>
            );
          })()
        )}

        {/* Language Toggle */}
        <button
          type="button"
          onClick={onToggleLang}
          className="px-2.5 py-1 rounded-lg text-xs font-mono-telemetry bg-white/[0.05] hover:bg-white/[0.1] text-slate-200 border border-white/[0.08] cursor-pointer transition-colors"
        >
          {lang.toUpperCase()}
        </button>

        {/* Theme Toggle */}
        <button
          type="button"
          onClick={onToggleTheme}
          aria-label="Toggle theme"
          className="p-1.5 rounded-lg bg-white/[0.05] hover:bg-white/[0.1] text-slate-300 border border-white/[0.08] cursor-pointer transition-colors"
        >
          {theme === 'dark' ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
        </button>
      </div>
    </header>
  );
}
