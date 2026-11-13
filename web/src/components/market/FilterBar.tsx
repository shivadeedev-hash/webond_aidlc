// Market filter row: search box, risk filter, fund-category filter, performance
// sort toggle, and the period-return toggle (wireframe 03 + US-011 Add-on).

import type { AssetType, ReturnPeriod, SortOrder } from '../../types';
import { useI18n } from '../../i18n/useI18n';
import { SearchIcon } from '../ui/icons';
import PeriodToggle from './PeriodToggle';

export interface MarketFilters {
  q: string;
  riskLevel: number | '';
  category: string;
  period: ReturnPeriod;
  sortPerformance: boolean;
  order: SortOrder;
}

const FUND_CATEGORIES = ['Global Equity', 'Thai Equity', 'Money Market', 'Fixed Income'];

export default function FilterBar({
  type,
  filters,
  onChange,
}: {
  type: AssetType;
  filters: MarketFilters;
  onChange: (next: Partial<MarketFilters>) => void;
}): JSX.Element {
  const { t } = useI18n();

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col gap-3 md:flex-row md:items-center">
        <div className="relative flex-1">
          <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-text-muted">
            <SearchIcon size={18} />
          </span>
          <input
            type="search"
            value={filters.q}
            onChange={(e) => onChange({ q: e.target.value })}
            placeholder={t('page.market.searchPlaceholder')}
            aria-label={t('page.market.searchPlaceholder')}
            className="min-h-[44px] w-full rounded-md border border-border bg-surface pl-10 pr-3 text-text placeholder:text-text-muted"
          />
        </div>

        <label className="flex items-center gap-2 text-sm">
          <span className="whitespace-nowrap text-text-muted">{t('page.market.filterRisk')}</span>
          <select
            value={filters.riskLevel}
            onChange={(e) =>
              onChange({ riskLevel: e.target.value === '' ? '' : Number(e.target.value) })
            }
            aria-label={t('page.market.filterRisk')}
            className="min-h-[44px] rounded-md border border-border bg-surface px-3 text-text"
          >
            <option value="">{t('page.market.allRisk')}</option>
            {[1, 2, 3, 4, 5].map((r) => (
              <option key={r} value={r}>
                {r}
              </option>
            ))}
          </select>
        </label>

        {type === 'fund' && (
          <label className="flex items-center gap-2 text-sm">
            <span className="whitespace-nowrap text-text-muted">
              {t('page.market.filterCategory')}
            </span>
            <select
              value={filters.category}
              onChange={(e) => onChange({ category: e.target.value })}
              aria-label={t('page.market.filterCategory')}
              className="min-h-[44px] rounded-md border border-border bg-surface px-3 text-text"
            >
              <option value="">{t('page.market.allCategory')}</option>
              {FUND_CATEGORIES.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>
          </label>
        )}
      </div>

      <div className="flex flex-wrap items-center gap-4">
        <div className="flex items-center gap-2">
          <span className="text-sm text-text-muted">{t('page.market.return')}</span>
          <PeriodToggle
            value={filters.period}
            onChange={(p) => onChange({ period: p })}
            label={t('page.market.return')}
          />
        </div>

        <label className="inline-flex min-h-[44px] items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={filters.sortPerformance}
            onChange={(e) => onChange({ sortPerformance: e.target.checked })}
            className="h-5 w-5 rounded border-border accent-[color:var(--color-primary)]"
          />
          <span>{t('page.market.sortPerformance')}</span>
        </label>

        {filters.sortPerformance && (
          <button
            type="button"
            onClick={() => onChange({ order: filters.order === 'desc' ? 'asc' : 'desc' })}
            className="min-h-[44px] rounded-md border border-border px-3 text-sm font-medium transition-standard duration-standard hover:bg-surface-muted"
            aria-label="Toggle sort order"
          >
            {filters.order === 'desc' ? 'High → Low' : 'Low → High'}
          </button>
        )}
      </div>
    </div>
  );
}
