// Market catalog (US-011/012, wireframe 03). PUBLIC page — guests can browse.
// Asset-type tabs (Fund/Bond/Gold), search, risk/category filters, period-return
// toggle (3M/6M/1Y/3Y/5Y) + sort by performance, and cards linking to detail.

import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { catalogService } from '../services/catalog';
import type { AssetType } from '../types';
import { useI18n } from '../i18n/useI18n';
import PageHeader from '../components/ui/PageHeader';
import { LoadingState, ErrorState, EmptyState } from '../components/ui/StateViews';
import AssetCard from '../components/market/AssetCard';
import FilterBar, { type MarketFilters } from '../components/market/FilterBar';

const TABS: { type: AssetType; labelKey: string }[] = [
  { type: 'fund', labelKey: 'page.market.tab.fund' },
  { type: 'bond', labelKey: 'page.market.tab.bond' },
  { type: 'gold', labelKey: 'page.market.tab.gold' },
];

const DEFAULT_FILTERS: MarketFilters = {
  q: '',
  riskLevel: '',
  category: '',
  period: '1Y',
  sortPerformance: false,
  order: 'desc',
};

export default function Market(): JSX.Element {
  const { t } = useI18n();
  const [type, setType] = useState<AssetType>('fund');
  const [filters, setFilters] = useState<MarketFilters>(DEFAULT_FILTERS);

  const params = useMemo(
    () => ({
      type,
      q: filters.q.trim() || undefined,
      riskLevel: filters.riskLevel === '' ? undefined : filters.riskLevel,
      category: type === 'fund' && filters.category ? filters.category : undefined,
      period: filters.period,
      sort: filters.sortPerformance ? ('performance' as const) : undefined,
      order: filters.order,
    }),
    [type, filters],
  );

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['market', params],
    queryFn: () => catalogService.listAssets(params),
  });

  function updateFilters(next: Partial<MarketFilters>): void {
    setFilters((f) => ({ ...f, ...next }));
  }

  function handleTab(next: AssetType): void {
    setType(next);
    // Category only applies to funds; clear it when leaving the Fund tab.
    if (next !== 'fund') setFilters((f) => ({ ...f, category: '' }));
  }

  const items = data?.items ?? [];

  return (
    <div>
      <PageHeader title={t('page.market.title')} />

      {/* Asset-type tabs */}
      <div role="tablist" aria-label={t('page.market.title')} className="mb-5 flex flex-wrap gap-3">
        {TABS.map((tab) => {
          const active = tab.type === type;
          return (
            <button
              key={tab.type}
              role="tab"
              aria-selected={active}
              type="button"
              onClick={() => handleTab(tab.type)}
              className={[
                'min-h-[44px] rounded-full border px-5 text-sm font-medium transition-standard duration-standard',
                active
                  ? 'border-primary bg-primary text-text-inverse'
                  : 'border-primary bg-surface text-primary hover:bg-surface-muted',
              ].join(' ')}
            >
              {t(tab.labelKey)}
            </button>
          );
        })}
      </div>

      <div className="mb-6">
        <FilterBar type={type} filters={filters} onChange={updateFilters} />
      </div>

      {isLoading && <LoadingState />}
      {isError && <ErrorState onRetry={() => void refetch()} />}
      {!isLoading && !isError && items.length === 0 && <EmptyState />}

      {!isLoading && !isError && items.length > 0 && (
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {items.map((item) => (
            <AssetCard key={`${item.assetType}:${item.id}`} item={item} period={filters.period} />
          ))}
        </div>
      )}
    </div>
  );
}
