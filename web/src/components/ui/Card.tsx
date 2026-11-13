// Token-styled surface card used across pages (design-system §2.3 card spec).

import type { ReactNode } from 'react';

export function Card({
  children,
  className = '',
  as: Tag = 'section',
}: {
  children: ReactNode;
  className?: string;
  as?: 'section' | 'div' | 'article';
}): JSX.Element {
  return (
    <Tag
      className={`rounded-lg border border-border bg-surface shadow-sm transition-standard duration-standard hover:shadow-md ${className}`}
    >
      {children}
    </Tag>
  );
}

export function CardHeader({
  children,
  className = '',
}: {
  children: ReactNode;
  className?: string;
}): JSX.Element {
  return (
    <div
      className={`flex items-center justify-between gap-3 border-b border-border px-6 py-4 ${className}`}
    >
      {children}
    </div>
  );
}

export function CardTitle({ children }: { children: ReactNode }): JSX.Element {
  return (
    <h2 className="flex items-center gap-2 text-base font-bold">
      <span aria-hidden="true" className="h-4 w-1 rounded-full bg-primary" />
      {children}
    </h2>
  );
}

export function CardBody({
  children,
  className = '',
}: {
  children: ReactNode;
  className?: string;
}): JSX.Element {
  return <div className={`p-6 ${className}`}>{children}</div>;
}
