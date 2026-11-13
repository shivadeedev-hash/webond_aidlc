// A single transaction row in the History list (wireframe 07). Shows time, a type
// badge, asset code/name, signed amount (+/− per BR-15 wallet view), qty @ price,
// and a status badge. Links to the transaction detail.

import { Link } from 'react-router-dom';
import type { Transaction } from '../../types';
import {
  formatDateTimeBangkok,
  formatMoney,
  formatSignedTHB,
  signedAmount,
  type WalletTxnType,
} from '../../lib/format';
import StatusBadge from '../ui/StatusBadge';
import { ChevronRight } from '../ui/icons';

const TYPE_BADGE: Record<Transaction['type'], string> = {
  buy: 'BUY',
  sell: 'SELL',
  dca_buy: 'DCA BUY',
  coupon: 'COUPON',
  maturity: 'MATURITY',
  deposit: 'DEPOSIT',
  withdraw: 'WITHDRAW',
};

function timeOf(iso: string): string {
  const full = formatDateTimeBangkok(iso);
  return full.slice(-5); // "HH:mm"
}

export default function TxnRow({
  txn,
  assetLabel,
}: {
  txn: Transaction;
  /** Resolved "CODE Name" for the asset, or null for cash transactions. */
  assetLabel: string | null;
}): JSX.Element {
  const signed = signedAmount(txn.amount, txn.type as WalletTxnType);
  const amountClass = signed >= 0 ? 'text-success' : 'text-danger';

  const qtyAtPrice =
    txn.quantity && txn.price
      ? `${formatMoney(txn.quantity)} @ ${formatMoney(txn.price)}`
      : txn.status === 'processing' || txn.status === 'submitted'
        ? 'awaiting NAV'
        : '—';

  // Mobile: two lines (time · type · amount / asset · status) so nothing is
  // pushed off-screen. From `sm` up the two line wrappers use `display: contents`
  // and the order-* classes restore the single desktop row.
  return (
    <Link
      to={`/history/${txn.txnRef}`}
      className="flex flex-col gap-1 rounded-md border border-border bg-surface p-3 transition-standard duration-standard hover:shadow-sm sm:flex-row sm:items-center sm:gap-3"
    >
      <span className="flex items-center gap-3 sm:contents">
        <span className="w-12 shrink-0 text-sm text-text-muted sm:order-1">
          {timeOf(txn.occurredAt)}
        </span>

        <span className="shrink-0 sm:order-2 sm:w-24">
          <span className="inline-block rounded bg-surface-muted px-2 py-0.5 text-xs font-semibold text-primary">
            {TYPE_BADGE[txn.type]}
          </span>
        </span>

        <span
          className={`ml-auto shrink-0 text-right text-sm font-semibold sm:order-4 sm:ml-0 sm:w-32 ${amountClass}`}
        >
          {formatSignedTHB(txn.amount, txn.type as WalletTxnType)}
        </span>
      </span>

      <span className="flex items-center gap-3 sm:contents">
        <span className="min-w-0 flex-1 sm:order-3">
          <span className="block truncate text-sm font-medium">{assetLabel ?? '—'}</span>
        </span>

        <span className="hidden w-40 shrink-0 text-right text-sm text-text-muted sm:order-5 sm:block">
          {qtyAtPrice}
        </span>

        <span className="shrink-0 text-right sm:order-6 sm:w-32">
          <StatusBadge status={txn.status} />
        </span>

        <span className="shrink-0 text-text-muted sm:order-7">
          <ChevronRight size={18} />
        </span>
      </span>
    </Link>
  );
}
