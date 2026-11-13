// History filter row (wireframe 07): tab pills (All/Orders/DCA/Cash/Income),
// search box, and asset-type / txn-type / status / date-range / sort selects.
// The parent syncs all of these to the URL query string (US-025).

import type {
  AssetType,
  HistorySort,
  HistoryStatus,
  HistoryTab,
  HistoryType,
  SortOrder,
} from './historyQuery';
import { useI18n } from '../../i18n/useI18n';
import { SearchIcon } from '../ui/icons';

export interface HistoryUiFilters {
  tab: HistoryTab;
  q: string;
  assetType: AssetType | '';
  type: HistoryType | '';
  status: HistoryStatus | '';
  range: '7D' | '1M' | '3M' | '6M' | '1Y';
  sort: HistorySort;
  order: SortOrder;
}

const TABS: { key: HistoryTab; labelKey: string }[] = [
  { key: 'all', labelKey: 'page.history.tab.all' },
  { key: 'orders', labelKey: 'page.history.tab.orders' },
  { key: 'dca', labelKey: 'page.history.tab.dca' },
  { key: 'cash', labelKey: 'page.history.tab.cash' },
  { key: 'income', labelKey: 'page.history.tab.income' },
];

export default function HistoryFilterBar({
  filters,
  onChange,
}: {
  filters: HistoryUiFilters;
  onChange: (next: Partial<HistoryUiFilters>) => void;
}): JSX.Element {
  const { t } = useI18n();

  return (
    <div className="flex flex-col gap-3">
      {/* Tabs + search */}
      <div className="flex flex-col gap-3 lg:flex-row lg:items-center">
        <div role="tablist" aria-label={t('page.history.title')} className="flex flex-wrap gap-2">
          {TABS.map((tab) => {
            const active = filters.tab === tab.key;
            return (
              <button
                key={tab.key}
                role="tab"
                aria-selected={active}
                type="button"
                onClick={() => onChange({ tab: tab.key })}
                className={[
                  'min-h-[40px] rounded-full border px-4 text-sm font-medium transition-standard duration-standard',
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

        <div className="relative lg:ml-auto lg:w-80">
          <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-text-muted">
            <SearchIcon size={18} />
          </span>
          <input
            type="search"
            value={filters.q}
            onChange={(e) => onChange({ q: e.target.value })}
            placeholder={t('page.history.search')}
            aria-label={t('page.history.search')}
            className="min-h-[44px] w-full rounded-md border border-border bg-surface pl-10 pr-3 text-text placeholder:text-text-muted"
          />
        </div>
      </div>

      {/* Selects */}
      <div className="flex flex-wrap items-center gap-3 rounded-md border border-border bg-surface-muted p-3 text-sm">
        <label className="flex items-center gap-2">
          <span className="text-text-muted">Asset type</span>
          <select
            value={filters.assetType}
            onChange={(e) => onChange({ assetType: e.target.value as AssetType | '' })}
            className="min-h-[40px] rounded-md border border-border bg-surface px-2"
            aria-label="Asset type"
          >
            <option value="">All</option>
            <option value="fund">Fund</option>
            <option value="bond">Bond</option>
            <option value="gold">Gold</option>
          </select>
        </label>

        <label className="flex items-center gap-2">
          <span className="text-text-muted">Txn type</span>
          <select
            value={filters.type}
            onChange={(e) => onChange({ type: e.target.value as HistoryType | '' })}
            className="min-h-[40px] rounded-md border border-border bg-surface px-2"
            aria-label="Transaction type"
          >
            <option value="">All</option>
            <option value="buy">Buy</option>
            <option value="sell">Sell</option>
            <option value="dca_buy">DCA Buy</option>
            <option value="coupon">Coupon</option>
            <option value="maturity">Maturity</option>
            <option value="deposit">Deposit</option>
            <option value="withdraw">Withdraw</option>
          </select>
        </label>

        <label className="flex items-center gap-2">
          <span className="text-text-muted">Status</span>
          <select
            value={filters.status}
            onChange={(e) => onChange({ status: e.target.value as HistoryStatus | '' })}
            className="min-h-[40px] rounded-md border border-border bg-surface px-2"
            aria-label="Status"
          >
            <option value="">All</option>
            <option value="submitted">Submitted</option>
            <option value="processing">Processing</option>
            <option value="completed">Completed</option>
            <option value="failed">Failed</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </label>

        <label className="flex items-center gap-2">
          <span className="text-text-muted">Date</span>
          <select
            value={filters.range}
            onChange={(e) => onChange({ range: e.target.value as HistoryUiFilters['range'] })}
            className="min-h-[40px] rounded-md border border-border bg-surface px-2"
            aria-label="Date range"
          >
            <option value="7D">Last 7D</option>
            <option value="1M">Last 1M</option>
            <option value="3M">Last 3M</option>
            <option value="6M">Last 6M</option>
            <option value="1Y">Last 1Y</option>
          </select>
        </label>

        <label className="flex items-center gap-2">
          <span className="text-text-muted">Sort</span>
          <select
            value={`${filters.sort}:${filters.order}`}
            onChange={(e) => {
              const [sort, order] = e.target.value.split(':') as [HistorySort, SortOrder];
              onChange({ sort, order });
            }}
            className="min-h-[40px] rounded-md border border-border bg-surface px-2"
            aria-label="Sort"
          >
            <option value="date:desc">Newest</option>
            <option value="date:asc">Oldest</option>
            <option value="amount:desc">Amount high → low</option>
            <option value="amount:asc">Amount low → high</option>
          </select>
        </label>
      </div>
    </div>
  );
}
