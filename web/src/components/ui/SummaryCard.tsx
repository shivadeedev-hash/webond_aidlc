// Compact metric card used on Portfolio and Dashboard summary rows (wireframes
// 02 / 06). Value may carry its own color node (e.g. a PLValue).

import type { ReactNode } from 'react';

export default function SummaryCard({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}): JSX.Element {
  return (
    <div className="relative overflow-hidden rounded-lg border border-border bg-surface p-5 shadow-sm transition-standard duration-standard hover:-translate-y-0.5 hover:shadow-md">
      <span aria-hidden="true" className="absolute inset-x-0 top-0 h-1 bg-brand" />
      <p className="text-sm font-medium text-text-muted">{label}</p>
      <div className="mt-2 text-2xl font-bold tracking-tight">{children}</div>
    </div>
  );
}
