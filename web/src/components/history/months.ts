// Month options for the statement picker. A <select> is used instead of
// <input type="month">, which desktop Safari does not support.

import type { Locale } from '../../lib/i18n';

export interface MonthOption {
  /** yyyyMM, the form the statement endpoint expects. */
  value: string;
  /** Human label, e.g. "October 2026" / "ตุลาคม 2026" (Gregorian year, BR-15). */
  label: string;
}

/**
 * The current month (Asia/Bangkok) and the `count - 1` months before it,
 * newest first. 12 options matches the 12-month history range limit (BR-12).
 */
export function recentMonths(count: number, locale: Locale, now: Date = new Date()): MonthOption[] {
  const parts = new Intl.DateTimeFormat('en-US', {
    timeZone: 'Asia/Bangkok',
    year: 'numeric',
    month: 'numeric',
  }).formatToParts(now);
  let year = Number(parts.find((p) => p.type === 'year')?.value);
  let month = Number(parts.find((p) => p.type === 'month')?.value);

  const label = new Intl.DateTimeFormat(locale === 'th' ? 'th-TH-u-ca-gregory' : 'en-US', {
    month: 'long',
    year: 'numeric',
    timeZone: 'UTC',
  });

  const out: MonthOption[] = [];
  for (let i = 0; i < count; i++) {
    out.push({
      value: `${year}${String(month).padStart(2, '0')}`,
      label: label.format(new Date(Date.UTC(year, month - 1, 1))),
    });
    month -= 1;
    if (month === 0) {
      month = 12;
      year -= 1;
    }
  }
  return out;
}
