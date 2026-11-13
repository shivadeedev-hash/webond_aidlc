// Wallet ledger table (wireframe 09). Shows date, type, signed amount and status.
// Wallet rows store their own signed amount (deposit +, withdraw/hold −), so we
// color + sign + label by that stored sign — meaning never relies on color alone.

import type { WalletTransaction, WalletTransactionType } from '../../types';
import { formatDateBangkok, formatTHB } from '../../lib/format';
import { useI18n } from '../../i18n/useI18n';
import { ArrowUp, ArrowDown } from '../ui/icons';
import { EmptyState, LoadingState } from '../ui/StateViews';

const TYPE_LABELS: Record<WalletTransactionType, string> = {
  deposit: 'Deposit',
  withdraw: 'Withdraw',
  hold: 'Reserved (hold)',
  release: 'Release',
  settlement: 'Settlement',
  coupon: 'Coupon',
  maturity: 'Maturity',
};

export default function WalletHistory({
  transactions,
  loading,
}: {
  transactions: WalletTransaction[];
  loading: boolean;
}): JSX.Element {
  const { t } = useI18n();

  if (loading) return <LoadingState />;
  if (transactions.length === 0) return <EmptyState />;

  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[420px] text-sm">
        <thead>
          <tr className="border-b border-border text-left text-text-muted">
            <th scope="col" className="py-2 pr-3 font-medium">
              Date
            </th>
            <th scope="col" className="py-2 pr-3 font-medium">
              Type
            </th>
            <th scope="col" className="py-2 pr-3 text-right font-medium">
              {t('page.dca.col.amount')}
            </th>
            <th scope="col" className="py-2 font-medium">
              {t('page.dca.col.status')}
            </th>
          </tr>
        </thead>
        <tbody>
          {transactions.map((tx) => {
            const n = Number(tx.amount);
            const positive = n >= 0;
            return (
              <tr key={tx.id} className="border-b border-border last:border-0">
                <td className="py-2 pr-3 whitespace-nowrap">{formatDateBangkok(tx.createdAt)}</td>
                <td className="py-2 pr-3">{TYPE_LABELS[tx.type]}</td>
                <td className="py-2 pr-3 text-right">
                  <span
                    className={`inline-flex items-center justify-end gap-1 font-medium ${positive ? 'text-success' : 'text-danger'}`}
                  >
                    {positive ? <ArrowUp size={14} /> : <ArrowDown size={14} />}
                    {positive ? '+' : '-'}
                    {formatTHB(Math.abs(n))}
                  </span>
                </td>
                <td className="py-2 capitalize text-text-muted">{tx.status}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
