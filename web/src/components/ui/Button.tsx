// Token-backed button with primary / secondary / outline / ghost variants.
// All variants keep tap target ≥ 44px and rely on the global focus-visible ring.

import type { ButtonHTMLAttributes } from 'react';

type Variant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger';

const VARIANTS: Record<Variant, string> = {
  primary:
    'bg-primary text-text-inverse hover:bg-primary-hover active:bg-primary-active disabled:opacity-60',
  secondary: 'bg-surface-muted text-text hover:bg-border disabled:opacity-60',
  outline: 'border border-primary text-primary hover:bg-surface-muted disabled:opacity-60',
  ghost: 'text-primary hover:bg-surface-muted disabled:opacity-60',
  danger:
    'border border-danger text-danger hover:bg-danger hover:text-text-inverse disabled:opacity-60',
};

export default function Button({
  variant = 'primary',
  className = '',
  type = 'button',
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: Variant }): JSX.Element {
  return (
    <button
      type={type}
      className={[
        'inline-flex min-h-[44px] items-center justify-center gap-2 rounded-full px-5 text-sm font-semibold',
        'transition-standard duration-standard disabled:cursor-not-allowed',
        VARIANTS[variant],
        className,
      ].join(' ')}
      {...props}
    />
  );
}
