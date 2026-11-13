// Portfolio (US-022, wireframe 06). Summary cards (market value, total cost,
// unrealized P/L, realized P/L YTD) + holdings table. Bond footer shows accrued
// interest / next coupon / maturity for any bond holding.

import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { portfolioService } from '../services/portfolio';
import { historyService } from '../services/history';
import { catalogService } from '../services/catalog';
import { useI18n } from '../i18n/useI18n';
import { formatDateBangkok, formatTHB } from '../lib/format';
import PageHeader from '../components/ui/PageHeader';
import SummaryCard from '../components/ui/SummaryCard';
import PLValue from '../components/ui/PLValue';
import { Card, CardBody } from '../components/ui/Card';
import { LoadingState, ErrorState, EmptyState } from '../components/ui/StateViews';
import HoldingsTable from '../components/portfolio/HoldingsTable';

function currentYearRealizedPl(
  txns: { realizedPl: string | null; occurredAt: string; type: string }[],
): number {
  const year = new Date().getFullYear();
  return txns
    .filter((x) => x.realizedPl != null && new Date(x.occurredAt).getFullYear() === year)
    .reduce((s, x) => s + Number(x.realizedPl), 0);
}

export default function Portfolio(): JSX.Element {
  const { t } = useI18n();

  const portfolioQuery = useQuery({
    queryKey: ['portfolio'],
    queryFn: () => portfolioService.getPortfolio(),
  });

  // Realized P/L (YTD) is derived from the history feed (no dedicated field).
  const historyQuery = useQuery({
    queryKey: ['history', { pageSize: 100 }],
    queryFn: () => historyService.listTransactions({ pageSize: 100 }),
  });

  const holdings = portfolioQuery.data?.holdings ?? [];
  const bondHolding = holdings.find((h) => h.assetType === 'bond');

  // Fetch the bond master for maturity/coupon display in the footer.
  const bondQuery = useQuery({
    queryKey: ['asset', 'bond', bondHolding?.assetId],
    queryFn: () => catalogService.getAsset('bond', bondHolding!.assetId),
    enabled: bondHolding !== undefined,
  });

  const realizedYtd = useMemo(
    () => currentYearRealizedPl(historyQuery.data?.items ?? []),
    [historyQuery.data],
  );

  if (portfolioQuery.isLoading) return <LoadingState />;
  if (portfolioQuery.isError || !portfolioQuery.data) {
    return <ErrorState onRetry={() => void portfolioQuery.refetch()} />;
  }

  const { summary } = portfolioQuery.data;
  const bondAsset = bondQuery.data;

  return (
    <div>
      <PageHeader title={t('page.portfolio.title')} />

      <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <SummaryCard label={t('page.portfolio.marketValue')}>
          {formatTHB(summary.totalMarketValue)}
        </SummaryCard>
        <SummaryCard label={t('page.portfolio.totalCost')}>
          {formatTHB(summary.totalCost)}
        </SummaryCard>
        <SummaryCard label={t('page.portfolio.unrealizedPl')}>
          <PLValue amount={summary.totalUnrealizedPl} pct={summary.totalUnrealizedPlPct} />
        </SummaryCard>
        <SummaryCard label={t('page.portfolio.realizedPl')}>
          <PLValue amount={realizedYtd} />
        </SummaryCard>
      </div>

      <Card>
        <CardBody className="p-0">
          {holdings.length === 0 ? <EmptyState /> : <HoldingsTable holdings={holdings} />}
        </CardBody>
      </Card>

      {bondHolding && (
        <p className="mt-4 text-sm text-text-muted">
          Bond ({bondHolding.assetCode}): accrued interest settled at coupon
          {bondHolding.nextEventDate &&
            ` · next coupon ${formatDateBangkok(bondHolding.nextEventDate)}`}
          {bondAsset &&
            bondAsset.assetType === 'bond' &&
            ` · maturity ${formatDateBangkok(bondAsset.maturityDate)}`}
          {bondAsset &&
            bondAsset.assetType === 'bond' &&
            ` · coupon ${Number(bondAsset.couponRate).toFixed(2)}%`}
        </p>
      )}
    </div>
  );
}
