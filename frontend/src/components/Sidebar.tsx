import {
  Film,
  Sparkles,
  History,
  Settings,
  ShieldCheck,
} from 'lucide-react';
import type { NavSection, UILang } from '../types';
import { tr } from '../i18n';

interface SidebarProps {
  activeSection: NavSection;
  onSelectSection: (section: NavSection) => void;
  version: string;
  lang: UILang;
  pendingCount: number;
  historyCount: number;
  running: boolean;
}

export function Sidebar({
  activeSection,
  onSelectSection,
  version,
  lang,
  pendingCount,
  historyCount,
  running,
}: SidebarProps) {
  const navItems: Array<{ id: NavSection; label: string; icon: typeof Film }> = [
    { id: 'converter', label: tr(lang, 'navConverter'), icon: Film },
    { id: 'presets', label: tr(lang, 'navPresets'), icon: Sparkles },
    { id: 'history', label: tr(lang, 'navHistory'), icon: History },
    { id: 'settings', label: tr(lang, 'navSettings'), icon: Settings },
  ];

  return (
    <aside className="w-60 shrink-0 bg-[#0B101B] border-r border-white/[0.07] flex flex-col justify-between select-none">
      <div>
        {/* Brand Header */}
        <div className="h-16 px-5 flex items-center justify-between border-b border-white/[0.07]">
          <div className="flex items-center gap-2.5">
            <div className="relative flex h-8 w-8 items-center justify-center rounded-lg bg-violet-500/15 border border-violet-500/35">
              <span className="h-2.5 w-2.5 rounded-full bg-violet-400 shadow-[0_0_10px_#8B5CF6]" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-sm font-semibold tracking-tight text-slate-100">
                  MediaPulse
                </span>
                <span className="font-mono-telemetry text-[10px] px-1.5 py-0.5 rounded bg-white/[0.06] text-slate-300 border border-white/[0.08]">
                  v{version}
                </span>
              </div>
              <p className="text-[11px] text-slate-400 leading-none mt-0.5">
                {tr(lang, 'subtitle')}
              </p>
            </div>
          </div>
        </div>

        {/* Navigation Links */}
        <nav className="p-3 space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon;
            const active = activeSection === item.id;

            let badge: React.ReactNode = null;
            if (item.id === 'converter' && pendingCount > 0) {
              badge = (
                <span
                  className={`font-mono-telemetry text-[10px] px-1.5 py-0.5 rounded border ${
                    running
                      ? 'bg-emerald-500/15 text-emerald-300 border-emerald-500/35 animate-pulse'
                      : 'bg-violet-500/15 text-violet-300 border-violet-500/30'
                  }`}
                >
                  {pendingCount}
                </span>
              );
            } else if (item.id === 'history' && historyCount > 0) {
              badge = (
                <span className="font-mono-telemetry text-[10px] px-1.5 py-0.5 rounded bg-white/[0.05] text-slate-400">
                  {historyCount}
                </span>
              );
            }

            return (
              <button
                key={item.id}
                type="button"
                onClick={() => onSelectSection(item.id)}
                className={`w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-sm font-medium transition-colors cursor-pointer ${
                  active
                    ? 'bg-violet-500/15 text-violet-200 border border-violet-500/30'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.04] border border-transparent'
                }`}
              >
                <span className="flex items-center gap-2.5">
                  <Icon className="w-4 h-4 shrink-0" />
                  <span>{item.label}</span>
                </span>
                {badge}
              </button>
            );
          })}
        </nav>
      </div>

      {/* Zero Telemetry Footer */}
      <div className="p-3 border-t border-white/[0.07]">
        <div className="rounded-lg bg-emerald-500/10 border border-emerald-500/20 px-3 py-2 flex items-center gap-2">
          <ShieldCheck className="w-4 h-4 text-emerald-400 shrink-0" />
          <div className="min-w-0">
            <p className="text-[11px] font-medium text-emerald-300 truncate">
              {tr(lang, 'zeroTelemetry')}
            </p>
            <p className="text-[10px] text-slate-400 truncate">Patoma89 Suite</p>
          </div>
        </div>
      </div>
    </aside>
  );
}
