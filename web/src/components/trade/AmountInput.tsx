// Friendly amount field for orders (Add-on 2026-11-12d). A text input (not
// type=number) so "1,000" or "฿1000" work; shows thousands separators while
// typing, a unit suffix, quick-pick chips and a specific inline message.
import { useId } from 'react';
import { groupAmount, sanitizeAmount } from '../../lib/amountInput';

export interface QuickPick {
  label: string;
  value: string; // plain numeric string
}

export default function AmountInput({
  label,
  unit,
  value,
  onChange,
  decimals,
  picks = [],
  picksLabel = 'Quick amounts',
  hint,
  error,
  disabled,
}: {
  label: string;
  /** Suffix shown inside the field, e.g. "บาท", "หน่วย", "กรัม". */
  unit: string;
  /** Plain numeric string, no grouping. */
  value: string;
  onChange: (clean: string) => void;
  decimals: number;
  picks?: QuickPick[];
  /** Accessible name of the quick-pick group. */
  picksLabel?: string;
  hint?: string;
  error?: string | null;
  disabled?: boolean;
}): JSX.Element {
  const id = useId();
  const hintId = `${id}-hint`;
  const errId = `${id}-err`;
  return (
    <div className="flex flex-col gap-2 text-sm">
      <label htmlFor={id} className="font-medium">
        {label}
      </label>
      <div
        className={[
          'flex min-h-[56px] items-center rounded-md border bg-surface px-4 shadow-sm',
          'transition-standard duration-standard focus-within:ring-4 focus-within:ring-primary/10',
          error ? 'border-danger' : 'border-border focus-within:border-primary',
          disabled ? 'opacity-60' : '',
        ].join(' ')}
      >
        <input
          id={id}
          type="text"
          inputMode="decimal"
          autoComplete="off"
          placeholder="0"
          value={groupAmount(value)}
          onChange={(e) => onChange(sanitizeAmount(e.target.value, decimals))}
          disabled={disabled}
          aria-invalid={error ? true : undefined}
          aria-describedby={
            [error ? errId : null, hint ? hintId : null].filter(Boolean).join(' ') || undefined
          }
          className="w-full min-w-0 bg-transparent text-2xl font-bold tracking-tight text-text outline-none placeholder:text-text-muted/50"
        />
        <span className="ml-3 shrink-0 text-base font-medium text-text-muted">{unit}</span>
      </div>
      {picks.length > 0 && (
        <div className="flex flex-wrap gap-2" role="group" aria-label={picksLabel}>
          {picks.map((p) => (
            <button
              key={p.label}
              type="button"
              disabled={disabled}
              onClick={() => onChange(p.value)}
              aria-pressed={value === p.value}
              className={[
                'min-h-[44px] rounded-full border px-4 text-xs font-semibold transition-standard duration-standard',
                value === p.value
                  ? 'border-primary bg-primary-soft text-primary'
                  : 'border-border text-text-muted hover:border-primary hover:text-primary',
              ].join(' ')}
            >
              {p.label}
            </button>
          ))}
        </div>
      )}
      {error ? (
        <p id={errId} role="alert" className="text-sm text-danger">
          {error}
        </p>
      ) : (
        hint && (
          <p id={hintId} className="text-xs text-text-muted">
            {hint}
          </p>
        )
      )}
    </div>
  );
}
