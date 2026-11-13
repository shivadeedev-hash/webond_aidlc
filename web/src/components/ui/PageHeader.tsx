// Page title row with an optional action slot. Keeps page headings consistent
// (design-system §8: section header pattern).

import type { ReactNode } from 'react';

export default function PageHeader({
  title,
  breadcrumb,
  action,
}: {
  title: string;
  breadcrumb?: ReactNode;
  action?: ReactNode;
}): JSX.Element {
  return (
    <div className="mb-6 flex flex-wrap items-center justify-between gap-3">
      <div className="flex items-center gap-2">
        {breadcrumb}
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">{title}</h1>
      </div>
      {action}
    </div>
  );
}
