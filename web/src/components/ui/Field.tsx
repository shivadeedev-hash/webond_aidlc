// Labelled form controls with accessible error wiring. Inputs are token-styled
// and keep a 44px tap target. Hint/error live OUTSIDE the <label> so the label's
// accessible name is exactly its text; they are linked via aria-describedby.

import { useId } from 'react';
import type { InputHTMLAttributes, ReactNode, SelectHTMLAttributes } from 'react';

const inputBase =
  'min-h-[48px] w-full rounded-md border border-border bg-surface px-4 text-text shadow-sm ' +
  'transition-standard duration-standard placeholder:text-text-muted hover:border-primary/50 ' +
  'focus-visible:border-primary focus-visible:ring-4 focus-visible:ring-primary/10';

export function TextField({
  label,
  error,
  hint,
  id,
  className = '',
  ...props
}: InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  error?: string;
  hint?: string;
}): JSX.Element {
  const reactId = useId();
  const fieldId = id ?? reactId;
  const errId = `${fieldId}-err`;
  const hintId = `${fieldId}-hint`;
  return (
    <div className="flex flex-col gap-1 text-sm">
      <label htmlFor={fieldId} className="font-medium">
        {label}
      </label>
      <input
        id={fieldId}
        className={`${inputBase} ${error ? 'border-danger' : ''} ${className}`}
        aria-invalid={error ? true : undefined}
        aria-describedby={
          [error ? errId : null, hint ? hintId : null].filter(Boolean).join(' ') || undefined
        }
        {...props}
      />
      {hint && !error && (
        <span id={hintId} className="text-xs text-text-muted">
          {hint}
        </span>
      )}
      {error && (
        <span id={errId} role="alert" className="text-xs text-danger">
          {error}
        </span>
      )}
    </div>
  );
}

export function SelectField({
  label,
  error,
  id,
  children,
  className = '',
  ...props
}: SelectHTMLAttributes<HTMLSelectElement> & {
  label: string;
  error?: string;
  children: ReactNode;
}): JSX.Element {
  const reactId = useId();
  const fieldId = id ?? reactId;
  const errId = `${fieldId}-err`;
  return (
    <div className="flex flex-col gap-1 text-sm">
      <label htmlFor={fieldId} className="font-medium">
        {label}
      </label>
      <select
        id={fieldId}
        className={`${inputBase} ${error ? 'border-danger' : ''} ${className}`}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? errId : undefined}
        {...props}
      >
        {children}
      </select>
      {error && (
        <span id={errId} role="alert" className="text-xs text-danger">
          {error}
        </span>
      )}
    </div>
  );
}
