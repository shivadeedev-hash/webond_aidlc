// Thai/English locale switch. A segmented control (not emoji flags) so it stays
// accessible and license-free. Current locale is marked with aria-pressed.

import { useI18n } from '../../i18n/useI18n';
import type { Locale } from '../../lib/i18n';

const LABELS: Record<Locale, string> = { th: 'ไทย', en: 'EN' };

export default function LocaleToggle(): JSX.Element {
  const { locale, setLocale, locales } = useI18n();

  return (
    <div
      role="group"
      aria-label="Language"
      className="inline-flex overflow-hidden rounded-full border border-border bg-surface p-0.5"
    >
      {locales.map((loc) => {
        const active = loc === locale;
        return (
          <button
            key={loc}
            type="button"
            aria-pressed={active}
            onClick={() => setLocale(loc)}
            className={[
              'min-h-[40px] min-w-[44px] rounded-full px-3 text-sm font-medium transition-standard duration-standard',
              active ? 'bg-primary text-text-inverse' : 'text-text-muted hover:text-primary',
            ].join(' ')}
          >
            {LABELS[loc]}
          </button>
        );
      })}
    </div>
  );
}
