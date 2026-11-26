import { Sparkles, ArrowRight } from 'lucide-react';
import type { PresetDefinition, PresetID, UILang } from '../types';
import { tr } from '../i18n';

interface PresetsViewProps {
  lang: UILang;
  presets: PresetDefinition[];
  activePreset: PresetID;
  onUsePreset: (id: PresetID) => void;
}

export function PresetsView({
  lang,
  presets,
  activePreset,
  onUsePreset,
}: PresetsViewProps) {
  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-base font-semibold text-slate-100">
          {tr(lang, 'presetsPageTitle')}
        </h2>
        <p className="text-xs text-slate-400 mt-0.5">{tr(lang, 'presetsPageSub')}</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3.5">
        {presets.map((p) => {
          const isCurrent = p.id === activePreset;
          return (
            <div
              key={p.id}
              className={`mp-card rounded-xl p-4 flex flex-col justify-between gap-3 border ${
                isCurrent ? 'border-violet-400/50' : ''
              }`}
            >
              <div className="space-y-1.5">
                <div className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <Sparkles className="w-4 h-4 text-violet-400" />
                    <h3 className="text-sm font-semibold text-slate-100">
                      {lang === 'en' ? p.nameEn : p.nameEs}
                    </h3>
                  </div>
                  <span className="font-mono-telemetry text-[11px] px-2 py-0.5 rounded bg-white/[0.06] text-cyan-200 border border-white/[0.08]">
                    {p.badge}
                  </span>
                </div>
                <p className="text-xs text-slate-400 leading-relaxed">
                  {lang === 'en' ? p.descEn : p.descEs}
                </p>
              </div>

              <div className="flex items-center justify-between pt-2 border-t border-white/[0.06]">
                <span className="font-mono-telemetry text-[11px] text-slate-400 uppercase">
                  .{p.defaultFormat}
                </span>
                <button
                  type="button"
                  onClick={() => onUsePreset(p.id)}
                  className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-violet-600/80 hover:bg-violet-500 text-white flex items-center gap-1.5 cursor-pointer transition-colors"
                >
                  <span>{tr(lang, 'usePreset')}</span>
                  <ArrowRight className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
