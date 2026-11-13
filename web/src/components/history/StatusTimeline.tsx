// Status timeline for the transaction detail (wireframe 08). Renders the ordered
// status-history entries (Submitted → Processing → Completed/Failed/Cancelled)
// with filled/open markers and timestamps.

import type { TransactionStatusHistoryItem } from '../../types';
import { formatDateTimeBangkok } from '../../lib/format';

export default function StatusTimeline({
  items,
  terminalReached,
}: {
  items: TransactionStatusHistoryItem[];
  /** Whether the final step is reached (filled) vs expected (open). */
  terminalReached: boolean;
}): JSX.Element {
  if (items.length === 0) {
    return <p className="text-sm text-text-muted">—</p>;
  }

  return (
    <ol className="relative flex flex-col gap-6 pl-6">
      <span aria-hidden="true" className="absolute bottom-2 left-[7px] top-2 w-0.5 bg-border" />
      {items.map((item, i) => {
        const isLast = i === items.length - 1;
        const filled = !isLast || terminalReached;
        return (
          <li key={item.id} className="relative">
            <span
              aria-hidden="true"
              className={[
                'absolute -left-6 top-0.5 inline-block h-4 w-4 rounded-full border-2',
                filled ? 'border-primary bg-primary' : 'border-primary bg-surface',
              ].join(' ')}
            />
            <p className="text-sm font-medium capitalize">{item.status}</p>
            <p className="text-xs text-text-muted">
              {formatDateTimeBangkok(item.changedAt)}
              {item.note ? ` · ${item.note}` : ''}
            </p>
          </li>
        );
      })}
    </ol>
  );
}
