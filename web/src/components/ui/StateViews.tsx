// Shared loading / error / empty state blocks. Every page that fetches data
// uses these so states are consistent and accessible (design-system policy §
// "Empty / loading / error states for every page that fetches data").

import type { ReactNode } from 'react';
import { useI18n } from '../../i18n/useI18n';
import { AlertTriangle, InfoIcon } from './icons';

export function LoadingState({ label }: { label?: string }): JSX.Element {
  const { t } = useI18n();
  return (
    <div
      role="status"
      aria-live="polite"
      className="flex items-center justify-center gap-3 py-12 text-text-muted"
    >
      <span
        aria-hidden="true"
        className="h-5 w-5 animate-spin rounded-full border-2 border-border border-t-primary"
      />
      <span>{label ?? t('common.loading')}</span>
    </div>
  );
}

export function ErrorState({
  message,
  onRetry,
}: {
  message?: string;
  onRetry?: () => void;
}): JSX.Element {
  const { t } = useI18n();
  return (
    <div role="alert" className="flex flex-col items-center gap-3 py-12 text-center">
      <span className="text-danger">
        <AlertTriangle size={28} />
      </span>
      <p className="text-text">{message ?? t('common.error')}</p>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="min-h-[44px] rounded-md border border-border px-4 text-sm font-medium transition-standard duration-standard hover:bg-surface-muted"
        >
          {t('common.retry')}
        </button>
      )}
    </div>
  );
}

export function EmptyState({
  message,
  children,
}: {
  message?: string;
  children?: ReactNode;
}): JSX.Element {
  const { t } = useI18n();
  return (
    <div className="flex flex-col items-center gap-3 py-12 text-center text-text-muted">
      <span className="text-text-muted">
        <InfoIcon size={28} />
      </span>
      <p>{message ?? t('common.empty')}</p>
      {children}
    </div>
  );
}

/** Small skeleton block for inline loading placeholders. */
export function Skeleton({ className = '' }: { className?: string }): JSX.Element {
  return (
    <div aria-hidden="true" className={`animate-pulse rounded-md bg-surface-muted ${className}`} />
  );
}
