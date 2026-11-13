// Transaction detail (US-026/027, wireframe 08). Type-relevant fields + status
// timeline. Cancel is enabled ONLY for a cancellable pending MF Buy/Sell or DCA
// Buy (BR-13); otherwise the button is disabled with the BR-13 explanation.
// TXN_NOT_FOUND (EX-15) renders a not-found state.

import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { historyService } from '../services/history';
import { catalogService } from '../services/catalog';
import type { Transaction } from '../types';
import { useI18n } from '../i18n/useI18n';
import {
  formatDateBangkok,
  formatDateTimeBangkok,
  formatMoney,
  formatSignedTHB,
  formatTHB,
  signedAmount,
  type WalletTxnType,
} from '../lib/format';
import { ApiClientError } from '../lib/errors';
import PageHeader from '../components/ui/PageHeader';
import { Card, CardBody, CardHeader, CardTitle } from '../components/ui/Card';
import Button from '../components/ui/Button';
import { LoadingState, ErrorState, EmptyState } from '../components/ui/StateViews';
import StatusBadge from '../components/ui/StatusBadge';
import StatusTimeline from '../components/history/StatusTimeline';

// BR-13: cancel allowed only for a pending MF (fund) Buy/Sell or DCA Buy.
function isCancellable(txn: Transaction): boolean {
  const open = txn.status === 'submitted' || txn.status === 'processing';
  const cancellableType =
    (txn.type === 'buy' || txn.type === 'sell' || txn.type === 'dca_buy') &&
    txn.assetType === 'fund';
  return open && cancellableType;
}

function Row({ label, value, tone }: { label: string; value: string; tone?: string }): JSX.Element {
  return (
    <div className="flex justify-between gap-4 py-1.5">
      <dt className="text-text-muted">{label}</dt>
      <dd className={`text-right font-medium ${tone ?? ''}`}>{value}</dd>
    </div>
  );
}

export default function TransactionDetail(): JSX.Element {
  const { t } = useI18n();
  const { txnRef } = useParams();
  const queryClient = useQueryClient();
  const [cancelError, setCancelError] = useState<string | null>(null);
  const [cancelling, setCancelling] = useState(false);

  const detailQuery = useQuery({
    queryKey: ['txn', txnRef],
    queryFn: () => historyService.getTransaction(txnRef!),
    enabled: !!txnRef,
  });

  const txn = detailQuery.data?.transaction;

  const assetQuery = useQuery({
    queryKey: ['asset', txn?.assetType, txn?.assetId],
    queryFn: () => catalogService.getAsset(txn!.assetType!, txn!.assetId!),
    enabled: !!txn?.assetType && !!txn?.assetId,
  });

  if (detailQuery.isLoading) return <LoadingState />;
  if (detailQuery.isError) {
    const notFound =
      detailQuery.error instanceof ApiClientError && detailQuery.error.code === 'TXN_NOT_FOUND';
    return notFound ? (
      <EmptyState message={t('errors.TXN_NOT_FOUND')}>
        <Link to="/history">
          <Button variant="primary">{t('page.history.title')}</Button>
        </Link>
      </EmptyState>
    ) : (
      <ErrorState onRetry={() => void detailQuery.refetch()} />
    );
  }

  if (!txn || !detailQuery.data) return <ErrorState />;

  const asset = assetQuery.data;
  const assetLabel = asset ? `${asset.code} · ${asset.name}` : (txn.assetType ?? '');
  const signed = signedAmount(txn.amount, txn.type as WalletTxnType);
  const cancellable = isCancellable(txn);

  async function handleCancel(): Promise<void> {
    if (!txnRef) return;
    setCancelError(null);
    setCancelling(true);
    try {
      await historyService.cancelTransaction(txnRef);
      await queryClient.invalidateQueries({ queryKey: ['txn', txnRef] });
      await queryClient.invalidateQueries({ queryKey: ['history'] });
    } catch (err) {
      setCancelError(err instanceof ApiClientError ? err.message : t('errors.UNKNOWN'));
    } finally {
      setCancelling(false);
    }
  }

  return (
    <div>
      <PageHeader
        title={txn.txnRef}
        breadcrumb={
          <Link to="/history" className="text-sm text-primary hover:underline">
            {t('page.history.title')} /
          </Link>
        }
      />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_340px]">
        {/* Fields */}
        <Card>
          <CardBody>
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
              <span className="inline-block rounded bg-surface-muted px-2 py-0.5 text-xs font-semibold uppercase text-primary">
                {txn.type.replace('_', ' ')}
              </span>
              <StatusBadge status={txn.status} />
            </div>

            <p className={`text-2xl font-semibold ${signed >= 0 ? 'text-success' : 'text-danger'}`}>
              {formatSignedTHB(txn.amount, txn.type as WalletTxnType)}
            </p>
            {assetLabel && <p className="mt-1 text-sm text-text-muted">{assetLabel}</p>}

            <dl className="mt-4 divide-y divide-border text-sm">
              <Row label="Reference no." value={txn.txnRef} />
              <Row label="Order date/time" value={formatDateTimeBangkok(txn.occurredAt)} />
              {txn.tradeDate && <Row label="Trade date" value={formatDateBangkok(txn.tradeDate)} />}
              {txn.settlementDate && (
                <Row label="Settlement date" value={formatDateBangkok(txn.settlementDate)} />
              )}
              {txn.quantity && <Row label="Quantity" value={formatMoney(txn.quantity)} />}
              {txn.price && <Row label="Price / NAV" value={formatMoney(txn.price)} />}
              {txn.accruedInterest && (
                <Row label="Accrued interest" value={formatTHB(txn.accruedInterest)} />
              )}
              <Row label="Fee" value={formatTHB(txn.fee)} />
              <Row label="Net amount" value={formatTHB(txn.netAmount)} />
              {txn.realizedPl != null && (
                <Row
                  label="Realized P/L"
                  value={`${Number(txn.realizedPl) >= 0 ? '+' : ''}${formatTHB(txn.realizedPl)}`}
                  tone={Number(txn.realizedPl) >= 0 ? 'text-success' : 'text-danger'}
                />
              )}
              {txn.dcaPlanId && (
                <Row
                  label="DCA plan / round"
                  value={`#${txn.dcaPlanId} · round ${txn.dcaRound ?? '—'}`}
                />
              )}
              <Row label="Channel" value="Web" />
            </dl>
          </CardBody>
        </Card>

        {/* Timeline + cancel */}
        <Card className="h-fit">
          <CardHeader>
            <CardTitle>{t('page.history.statusTimeline')}</CardTitle>
          </CardHeader>
          <CardBody>
            <StatusTimeline
              items={detailQuery.data.statusHistory}
              terminalReached={
                txn.status === 'completed' || txn.status === 'failed' || txn.status === 'cancelled'
              }
            />

            <p className="mt-5 text-xs text-text-muted">{t('page.history.cancelHint')}</p>

            {cancelError && (
              <p role="alert" className="mt-2 text-sm text-danger">
                {cancelError}
              </p>
            )}

            <Button
              variant="outline"
              className="mt-3 w-full"
              disabled={!cancellable || cancelling}
              onClick={handleCancel}
              title={cancellable ? undefined : t('page.history.cancelHint')}
            >
              {cancelling ? t('common.loading') : t('page.history.cancelOrder')}
              {!cancellable ? ' (disabled)' : ''}
            </Button>
          </CardBody>
        </Card>
      </div>
    </div>
  );
}
