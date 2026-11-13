// Return-period segmented control for the Market list (US-011 Add-on):
// 3M / 6M / 1Y / 3Y / 5Y. Selected period is both the display and the sort key.

import type { ReturnPeriod } from '../../types';

const PERIODS: ReturnPeriod[] = ['3M', '6M', '1Y', '3Y', '5Y'];

export default function PeriodToggle({
  value,
  onChange,
  label,
}: {
  value: ReturnPeriod;
  onChange: (p: ReturnPeriod) => void;
  label: string;
}): JSX.Element {
  return (
    <div
      role="group"
      aria-label={label}
      className="inline-flex overflow-hidden rounded-md border border-border"
    >
      {PERIODS.map((p) => {
        const active = p === value;
        return (
          <button
            key={p}
            type="button"
            aria-pressed={active}
            onClick={() => onChange(p)}
            className={[
              'min-h-[40px] px-3 text-sm font-medium transition-standard duration-standard',
              active
                ? 'bg-primary text-text-inverse'
                : 'bg-surface text-text hover:bg-surface-muted',
            ].join(' ')}
          >
            {p}
          </button>
        );
      })}
    </div>
  );
}
