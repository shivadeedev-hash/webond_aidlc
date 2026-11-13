// Insight dashboard (US-023, wireframe 02). Total value / P&L / cash cards,
// allocation pie, value-trend chart with range toggle (1M/6M/1Y/3Y/5Y/All —
// Add-on), top movers, DCA summary and upcoming events.

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { portfolioService } from '../services/portfolio';
import { walletService } from '../services/wallet';
import type { TrendPeriod } from '../types';
import { useI18n } from '../i18n/useI18n';
import { formatDateBangkok, formatDateTimeBangkok, formatTHB } from '../lib/format';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import SummaryCard from '../components/ui/SummaryCard';
import PLValue from '../components/ui/PLValue';
import { Card, CardBody, CardHeader, CardTitle } from '../components/ui/Card';
import { LoadingState, ErrorState } from '../components/ui/StateViews';
import AllocationPie from '../components/charts/AllocationPie';
import TrendChart from '../components/charts/TrendChart';
import { ArrowUp, ArrowDown } from '../components/ui/icons';

const RANGES: TrendPeriod[] = ['1M', '6M', '1Y', '3Y', '5Y', 'all'];

export default function Dashboard(): JSX.Element {
  const { t } = useI18n();
  const { user } = useAuth();
  const [range, setRange] = useState<TrendPeriod>('6M');

  const dashQuery = useQuery({
    queryKey: ['dashboard', range],
    queryFn: () => portfolioService.getDashboard(range),
  });
  const walletQuery = useQuery({ queryKey: ['wallet'], queryFn: () => walletService.getWallet() });

  if (dashQuery.isLoading) return <LoadingState />;
  if (dashQuery.isError || !dashQuery.data) {
    return <ErrorState onRetry={() => void dashQuery.refetch()} />;
  }

  const d = dashQuery.data;
  const available = walletQuery.data?.available ?? '0';

  return (
    <div>
      {/* Friendly welcome band (Add-on 2026-11-12c) */}
      <section className="relative mb-6 overflow-hidden rounded-xl bg-brand px-6 py-7 text-text-inverse shadow-lg sm:px-8">
        <span
          aria-hidden="true"
          className="absolute -right-12 -top-16 h-52 w-52 rounded-full bg-text-inverse/10"
        />
        <span
          aria-hidden="true"
          className="absolute -bottom-24 right-32 h-56 w-56 rounded-full bg-text-inverse/5"
        />
        <div className="relative flex flex-wrap items-end justify-between gap-6">
          <div>
            <p className="text-sm font-medium text-text-inverse">
              {t('page.dashboard.hello')}
              {user?.fullName ? `, ${user.fullName}` : ''}
            </p>
            <h1 className="mt-1 text-2xl font-bold tracking-tight sm:text-3xl">
              {t('page.dashboard.title')}
            </h1>
            <p className="mt-4 text-sm text-text-inverse">{t('page.dashboard.totalValue')}</p>
            <p className="text-3xl font-extrabold tracking-tight sm:text-4xl">
              {formatTHB(d.totals.totalValue)}
            </p>
          </div>
          <div className="flex flex-wrap gap-3">
            <Link
              to="/market"
              className="inline-flex min-h-[44px] items-center rounded-full bg-surface px-5 text-sm font-semibold text-primary shadow-sm transition-standard duration-standard hover:bg-primary-soft"
            >
              {t('page.dashboard.ctaInvest')}
            </Link>
            <Link
              to="/dca"
              className="inline-flex min-h-[44px] items-center rounded-full border border-text-inverse/60 px-5 text-sm font-semibold text-text-inverse transition-standard duration-standard hover:bg-text-inverse/10"
            >
              {t('page.dashboard.ctaDca')}
            </Link>
          </div>
        </div>
      </section>

      {/* Totals */}
      {/* Total value is shown in the welcome band above, so it is not repeated here. */}
      <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <SummaryCard label={t('page.dashboard.totalPl')}>
          <PLValue amount={d.totals.totalPl} pct={d.totals.totalPlPct} />
        </SummaryCard>
        <SummaryCard label={t('page.dashboard.cashAvailable')}>{formatTHB(available)}</SummaryCard>
        <SummaryCard label={t('page.dashboard.asOf')}>
          <span className="text-sm font-normal text-text-muted">
            {formatDateTimeBangkok(d.asOf)}
          </span>
        </SummaryCard>
      </div>

      {/* Allocation + trend */}
      <div className="mb-6 grid grid-cols-1 gap-6 lg:grid-cols-[320px_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>{t('page.dashboard.allocation')}</CardTitle>
          </CardHeader>
          <CardBody>
            <AllocationPie slices={d.allocation} />
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t('page.dashboard.valueTrend')}</CardTitle>
            <div
              role="group"
              aria-label="Trend range"
              className="inline-flex overflow-hidden rounded-md border border-border"
            >
              {RANGES.map((r) => (
                <button
                  key={r}
                  type="button"
                  aria-pressed={r === range}
                  onClick={() => setRange(r)}
                  className={[
                    'min-h-[36px] px-2.5 text-xs font-medium transition-standard duration-standard',
                    r === range
                      ? 'bg-primary text-text-inverse'
                      : 'bg-surface text-text hover:bg-surface-muted',
                  ].join(' ')}
                >
                  {r === 'all' ? 'All' : r}
                </button>
              ))}
            </div>
          </CardHeader>
          <CardBody>
            <TrendChart points={d.trend} />
          </CardBody>
        </Card>
      </div>

      {/* Bottom row: top movers, DCA summary, upcoming events */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>{t('page.dashboard.topMovers')}</CardTitle>
          </CardHeader>
          <CardBody>
            <ul className="flex flex-col gap-2 text-sm">
              {d.topMovers.map((m) => {
                const n = Number(m.changePct);
                const up = n >= 0;
                return (
                  <li
                    key={`${m.assetType}:${m.assetId}`}
                    className="flex items-center justify-between gap-3"
                  >
                    <span className="font-medium">{m.assetCode}</span>
                    <span
                      className={`inline-flex items-center gap-1 font-medium ${up ? 'text-success' : 'text-danger'}`}
                    >
                      {up ? <ArrowUp size={14} /> : <ArrowDown size={14} />}
                      {up ? '+' : ''}
                      {n.toFixed(1)}%
                    </span>
                  </li>
                );
              })}
            </ul>
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t('page.dashboard.dcaSummary')}</CardTitle>
          </CardHeader>
          <CardBody>
            <dl className="flex flex-col gap-2 text-sm">
              <div className="flex justify-between">
                <dt className="text-text-muted">{t('page.dashboard.investedToDate')}</dt>
                <dd className="font-medium">{formatTHB(d.dcaSummary.totalInvested)}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-text-muted">{t('page.dashboard.activePlans')}</dt>
                <dd className="font-medium">{d.dcaSummary.activePlans}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-text-muted">Monthly commitment</dt>
                <dd className="font-medium">{formatTHB(d.dcaSummary.monthlyCommitment)}</dd>
              </div>
            </dl>
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t('page.dashboard.upcoming')}</CardTitle>
          </CardHeader>
          <CardBody>
            <ul className="flex flex-col gap-2 text-sm">
              {d.upcomingEvents.map((e, i) => (
                <li key={i} className="flex items-center justify-between gap-3">
                  <span>
                    <span className="text-text-muted">{formatDateBangkok(e.date)}</span>{' '}
                    <span className="font-medium">{e.label}</span>
                  </span>
                  {e.amount && (
                    <span className="whitespace-nowrap text-text-muted">{formatTHB(e.amount)}</span>
                  )}
                </li>
              ))}
            </ul>
          </CardBody>
        </Card>
      </div>
    </div>
  );
}
