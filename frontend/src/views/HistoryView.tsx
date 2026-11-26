import { FolderOpen, Trash2 } from 'lucide-react';
import type { HistorySummary, UILang } from '../types';
import { tr } from '../i18n';
import { formatBytes } from './ConverterView';

interface HistoryViewProps {
  lang: UILang;
  summary: HistorySummary;
  onClearHistory: () => void;
  onRevealFile: (path: string) => void;
}

export function HistoryView({
  lang,
  summary,
  onClearHistory,
  onRevealFile,
}: HistoryViewProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-slate-100">
            {tr(lang, 'historyTitle')}
          </h2>
          <p className="text-xs text-slate-400 mt-0.5">{tr(lang, 'historySub')}</p>
        </div>
        {summary.totalJobs > 0 && (
          <button
            type="button"
            onClick={onClearHistory}
            className="px-3 py-1.5 rounded-lg text-xs font-medium bg-rose-500/15 hover:bg-rose-500/25 text-rose-200 border border-rose-500/30 flex items-center gap-1.5 cursor-pointer"
          >
            <Trash2 className="w-3.5 h-3.5" />
            <span>{tr(lang, 'clearHistory')}</span>
          </button>
        )}
      </div>

      {/* Stat Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <StatCard label={tr(lang, 'totalConverted')} value={String(summary.totalJobs)} />
        <StatCard label={tr(lang, 'totalInput')} value={formatBytes(summary.totalInputBytes)} />
        <StatCard label={tr(lang, 'totalOutput')} value={formatBytes(summary.totalOutputBytes)} />
        <StatCard
          label={tr(lang, 'totalSaved')}
          value={formatBytes(Math.max(0, summary.totalSavedBytes))}
          accent="emerald"
        />
      </div>

      {/* Entries Table */}
      <div className="mp-card rounded-xl overflow-hidden">
        {summary.entries.length === 0 ? (
          <div className="py-12 text-center text-xs text-slate-400">
            {tr(lang, 'noHistory')}
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-white/[0.03] text-slate-400 border-b border-white/[0.07]">
                <tr>
                  <th className="px-4 py-2.5 font-medium">Archivo</th>
                  <th className="px-4 py-2.5 font-medium">Preset</th>
                  <th className="px-4 py-2.5 font-medium">Encoder</th>
                  <th className="px-4 py-2.5 font-medium">Tamaño</th>
                  <th className="px-4 py-2.5 font-medium">Tiempo</th>
                  <th className="px-4 py-2.5 font-medium text-right">Acción</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/[0.05]">
                {summary.entries.map((e) => (
                  <tr key={e.id} className="hover:bg-white/[0.02]">
                    <td className="px-4 py-2.5 text-slate-200 max-w-[260px] truncate">
                      {e.outputName}
                    </td>
                    <td className="px-4 py-2.5 font-mono-telemetry text-violet-300">
                      {e.presetId}
                    </td>
                    <td className="px-4 py-2.5 font-mono-telemetry text-slate-300">
                      {e.encoder}
                    </td>
                    <td className="px-4 py-2.5 font-mono-telemetry text-slate-300">
                      {formatBytes(e.inputBytes)} →{' '}
                      <span className="text-emerald-300">{formatBytes(e.outputBytes)}</span>
                    </td>
                    <td className="px-4 py-2.5 font-mono-telemetry text-slate-400">
                      {e.processSec.toFixed(1)}s
                    </td>
                    <td className="px-4 py-2.5 text-right">
                      <button
                        type="button"
                        onClick={() => onRevealFile(e.outputPath)}
                        className="inline-flex items-center gap-1 px-2 py-1 rounded bg-white/[0.05] hover:bg-white/[0.12] text-slate-200 cursor-pointer"
                      >
                        <FolderOpen className="w-3.5 h-3.5" />
                        <span>{tr(lang, 'openFolder')}</span>
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

function StatCard({
  label,
  value,
  accent,
}: {
  label: string;
  value: string;
  accent?: 'emerald';
}) {
  return (
    <div className="mp-card rounded-xl p-3.5">
      <p className="text-[11px] text-slate-400">{label}</p>
      <p
        className={`text-lg font-semibold font-mono-telemetry mt-1 ${
          accent === 'emerald' ? 'text-emerald-300' : 'text-slate-100'
        }`}
      >
        {value}
      </p>
    </div>
  );
}
