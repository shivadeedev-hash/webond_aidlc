// Maps the History UI filter state to/from URL search params and to the
// historyService query (US-025: filters kept in the URL query string).

import type {
  AssetType,
  HistorySort,
  HistoryTab,
  SortOrder,
  TransactionStatus,
  TransactionType,
} from '../../types';
import type { HistoryUiFilters } from './HistoryFilterBar';

export type { HistoryTab, HistorySort, SortOrder, AssetType };
export type HistoryType = TransactionType;
export type HistoryStatus = TransactionStatus;

export const DEFAULT_FILTERS: HistoryUiFilters = {
  tab: 'all',
  q: '',
  assetType: '',
  type: '',
  status: '',
  range: '3M',
  sort: 'date',
  order: 'desc',
};

const RANGE_DAYS: Record<HistoryUiFilters['range'], number> = {
  '7D': 7,
  '1M': 30,
  '3M': 90,
  '6M': 180,
  '1Y': 365,
};

const VALID_TABS: HistoryTab[] = ['all', 'orders', 'dca', 'cash', 'income'];

export function filtersFromParams(params: URLSearchParams): HistoryUiFilters {
  const get = (k: string): string => params.get(k) ?? '';
  const tab = get('tab');
  const range = get('range');
  const sort = get('sort');
  const order = get('order');
  return {
    tab: VALID_TABS.includes(tab as HistoryTab) ? (tab as HistoryTab) : 'all',
    q: get('q'),
    assetType: (['fund', 'bond', 'gold'].includes(get('assetType')) ? get('assetType') : '') as
      | AssetType
      | '',
    type: (get('type') || '') as HistoryType | '',
    status: (get('status') || '') as HistoryStatus | '',
    range: (['7D', '1M', '3M', '6M', '1Y'].includes(range)
      ? range
      : '3M') as HistoryUiFilters['range'],
    sort: (sort === 'amount' ? 'amount' : 'date') as HistorySort,
    order: (order === 'asc' ? 'asc' : 'desc') as SortOrder,
  };
}

export function paramsFromFilters(f: HistoryUiFilters): URLSearchParams {
  const params = new URLSearchParams();
  if (f.tab !== 'all') params.set('tab', f.tab);
  if (f.q) params.set('q', f.q);
  if (f.assetType) params.set('assetType', f.assetType);
  if (f.type) params.set('type', f.type);
  if (f.status) params.set('status', f.status);
  if (f.range !== '3M') params.set('range', f.range);
  if (f.sort !== 'date') params.set('sort', f.sort);
  if (f.order !== 'desc') params.set('order', f.order);
  return params;
}

/** Compute the ISO `from` date for a rolling range relative to today. */
export function rangeFrom(range: HistoryUiFilters['range']): string {
  const d = new Date();
  d.setDate(d.getDate() - RANGE_DAYS[range]);
  return d.toISOString().slice(0, 10);
}

/** Translate UI filters into the historyService query object. */
export function toServiceQuery(
  f: HistoryUiFilters,
  page: number,
  pageSize: number,
): import('../../types').HistoryQuery {
  return {
    tab: f.tab,
    assetType: f.assetType || undefined,
    type: f.type || undefined,
    status: f.status || undefined,
    from: rangeFrom(f.range),
    q: f.q.trim() || undefined,
    sort: f.sort,
    order: f.order,
    page,
    pageSize,
  };
}

export function isDefault(f: HistoryUiFilters): boolean {
  return (
    f.tab === 'all' &&
    f.q === '' &&
    f.assetType === '' &&
    f.type === '' &&
    f.status === '' &&
    f.range === '3M' &&
    f.sort === 'date' &&
    f.order === 'desc'
  );
}
