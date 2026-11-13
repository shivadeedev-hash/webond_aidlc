// Two-column auth shell (wireframe 01): a friendly red brand panel on the left,
// the auth card on the right. Collapses to a single column on mobile.
import type { ReactNode } from 'react';
import { useI18n } from '../../i18n/useI18n';

const PERKS = ['auth.perk1', 'auth.perk2', 'auth.perk3'] as const;

export default function AuthShell({ children }: { children: ReactNode }): JSX.Element {
  const { t } = useI18n();
  return (
    <div className="grid grid-cols-1 overflow-hidden rounded-xl border border-border bg-surface shadow-lg md:grid-cols-2">
      <div className="relative overflow-hidden bg-brand px-6 py-8 text-text-inverse sm:px-10 md:py-12">
        {/* soft decorative circles (white, low opacity) */}
        <span
          aria-hidden="true"
          className="absolute -right-16 -top-16 h-56 w-56 rounded-full bg-text-inverse/10"
        />
        <span
          aria-hidden="true"
          className="absolute -bottom-20 -left-10 h-64 w-64 rounded-full bg-text-inverse/5"
        />
        <div className="relative">
          <p className="text-4xl font-extrabold tracking-tight">{t('app.name')}</p>
          <p className="mt-3 max-w-sm text-lg">{t('app.tagline')}</p>
          <ul className="mt-8 hidden space-y-3 text-sm md:block">
            {PERKS.map((k) => (
              <li key={k} className="flex items-center gap-3">
                <span className="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-text-inverse/15">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                    <path
                      d="M5 12.5l4.5 4.5L19 7.5"
                      stroke="currentColor"
                      strokeWidth="2.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                </span>
                {t(k)}
              </li>
            ))}
          </ul>
        </div>
      </div>
      <div className="w-full p-6 sm:p-10">{children}</div>
    </div>
  );
}
