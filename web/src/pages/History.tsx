// Transaction history list (US-024..028, wireframe 07). Dated groups, tabs +
// filters + search + sort synced to the URL query (US-025). Load-more paging,
// signed amounts + status badges (via TxnRow). Export (Excel) + monthly Statement
// (PDF) are wired to the API through HistoryDownloads (disabled in mock mode).

import { useEffect, useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { historyService } from '../services/history';
import { catalogService } from '../services/catalog';
import type { Transaction } from '../types';
import { useI18n } from '../i18n/useI18n';
import { formatDateBangkok } from '../lib/format';
import PageHeader from '../components/ui/PageHeader';
import Button from '../components/ui/Button';
import { LoadingState, ErrorState, EmptyState } from '../components/ui/StateViews';
import TxnRow from '../components/history/TxnRow';
import HistoryFilterBar, { type HistoryUiFilters } from '../components/history/HistoryFilterBar';
import HistoryDownloads from '../components/history/HistoryDownloads';
import {
  DEFAULT_FILTERS,
  filtersFromParams,
  isDefault,
  paramsFromFilters,
  toServiceQuery,
} from '../components/history/historyQuery';

const PAGE_SIZE = 20;

function groupByDate(txns: Transaction[]): [string, Transaction[]][] {
  const groups = new Map<string, Transaction[]>();
  for (const t of txns) {
    const key = t.occurredAt.slice(0, 10);
    const arr = groups.get(key) ?? [];
    arr.push(t);
    groups.set(key, arr);
  }
  return [...groups.entries()];
}

export default function History(): JSX.Element {
  const { t } = useI18n();
  const [searchParams, setSearchParams] = useSearchParams();

  const filters = useMemo(() => filtersFromParams(searchParams), [searchParams]);
  const [page, setPage] = useState(1);

  // Reset paging whenever the filters change.
  useEffect(() => {
    setPage(1);
  }, [searchParams]);

  const query = useMemo(() => toServiceQuery(filters, page, PAGE_SIZE), [filters, page]);

  const { data, isLoading, isError, refetch, isFetching } = useQuery({
    queryKey: ['history', query],
    queryFn: () => historyService.listTransactions(query),
  });

  // Accumulate pages for infinite-scroll-style "Load more".
  const [rows, setRows] = useState<Transaction[]>([]);
  useEffect(() => {
    if (!data) return;
    setRows((prev) => (page === 1 ? data.items : [...prev, ...data.items]));
  }, [data, page]);

  // Asset-label resolver (code + name) from the catalog.
  const assetsQuery = useQuery({
    queryKey: ['allAssets'],
    queryFn: () => catalogService.listAssets({ pageSize: 100 }),
  });
  const labelMap = useMemo(() => {
    const m = new Map<string, string>();
    for (const a of assetsQuery.data?.items ?? [])
      m.set(`${a.assetType}:${a.id}`, `${a.code} ${a.name}`);
    return m;
  }, [assetsQuery.data]);

  function updateFilters(next: Partial<HistoryUiFilters>): void {
    const merged = { ...filters, ...next };
    setSearchParams(paramsFromFilters(merged), { replace: true });
  }

  function clearFilters(): void {
    setSearchParams(paramsFromFilters(DEFAULT_FILTERS), { replace: true });
  }

  const hasMore = data?.meta.hasMore ?? false;
  const groups = useMemo(() => groupByDate(rows), [rows]);

  return (
    <div>
      <PageHeader
        title={t('page.history.title')}
        action={<HistoryDownloads query={toServiceQuery(filters, 1, PAGE_SIZE)} />}
      />
      <div className="mb-6">
        <HistoryFilterBar filters={filters} onChange={updateFilters} />
      </div>

      {isLoading && rows.length === 0 && <LoadingState />}
      {isError && <ErrorState onRetry={() => void refetch()} />}

      {!isError && !isLoading && rows.length === 0 && (
        <EmptyState>
          <div className="mt-2 flex flex-wrap justify-center gap-3">
            {!isDefault(filters) && (
              <Button variant="outline" onClick={clearFilters}>
                {t('page.history.clearFilters')}
              </Button>
            )}
            <Link to="/market">
              <Button variant="primary">{t('page.history.goMarket')}</Button>
            </Link>
          </div>
        </EmptyState>
      )}

      {rows.length > 0 && (
        <div className="flex flex-col gap-5">
          {groups.map(([date, items]) => (
            <section key={date}>
              <h2 className="mb-2 text-sm font-semibold text-text-muted">
                {formatDateBangkok(date)}
              </h2>
              <div className="flex flex-col gap-2">
                {items.map((txn) => (
                  <TxnRow
                    key={txn.txnRef}
                    txn={txn}
                    assetLabel={
                      txn.assetType && txn.assetId
                        ? (labelMap.get(`${txn.assetType}:${txn.assetId}`) ?? txn.assetType)
                        : null
                    }
                  />
                ))}
              </div>
            </section>
          ))}

          {hasMore && (
            <div className="flex justify-center">
              <Button variant="outline" onClick={() => setPage((p) => p + 1)} disabled={isFetching}>
                {isFetching ? t('common.loading') : t('page.history.loadMore')}
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
