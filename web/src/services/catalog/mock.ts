// Mock catalog implementation: merges asset master JSON with the latest quote
// and exposes price/NAV history for charts.

import fundsData from '../../mocks/funds.json';
import bondsData from '../../mocks/bonds.json';
import goldData from '../../mocks/gold.json';
import pricesData from '../../mocks/prices.json';
import { ApiClientError } from '../../lib/errors';
import { mockResolve, paginate } from '../mockUtils';
import type { AssetQueryParams, CatalogService } from './index';
import type {
  Asset,
  AssetListItem,
  AssetQuote,
  AssetType,
  PricePeriod,
  PricePoint,
  PriceSeries,
} from '../../types';

const assets: Asset[] = [
  ...(fundsData as Asset[]),
  ...(bondsData as Asset[]),
  ...(goldData as Asset[]),
];

type QuoteMap = Record<string, AssetQuote>;
type SeriesMap = Record<string, PricePoint[]>;

const quotes = pricesData.quotes as QuoteMap;
const series = pricesData.series as SeriesMap;

function key(type: AssetType, id: number): string {
  return `${type}:${id}`;
}

function toListItem(asset: Asset): AssetListItem {
  const quote = quotes[key(asset.assetType, asset.id)];
  return { ...asset, quote } as AssetListItem;
}

export const mockCatalogService: CatalogService = {
  async listAssets(params: AssetQueryParams = {}) {
    let rows = assets.filter((a) => a.isActive);

    if (params.type) rows = rows.filter((a) => a.assetType === params.type);
    if (params.riskLevel) rows = rows.filter((a) => a.riskLevel === params.riskLevel);
    if (params.category) {
      rows = rows.filter((a) => a.assetType === 'fund' && a.category === params.category);
    }
    if (params.q) {
      const q = params.q.toLowerCase();
      rows = rows.filter(
        (a) => a.code.toLowerCase().includes(q) || a.name.toLowerCase().includes(q),
      );
    }

    const items = rows.map(toListItem);

    if (params.sort === 'performance') {
      items.sort(
        (a, b) => Number(b.quote?.periodReturnPct ?? 0) - Number(a.quote?.periodReturnPct ?? 0),
      );
    } else if (params.sort === 'name') {
      items.sort((a, b) => a.name.localeCompare(b.name));
    }
    if (params.order === 'asc' && params.sort === 'performance') items.reverse();

    const result = paginate(items, params.page, params.pageSize);
    return mockResolve(result);
  },

  async getAsset(type: AssetType, id: number): Promise<AssetListItem> {
    const asset = assets.find((a) => a.assetType === type && a.id === id);
    if (!asset) throw new ApiClientError('VALIDATION_ERROR', 'Asset not found.', { status: 404 });
    return mockResolve(toListItem(asset));
  },

  async getPriceSeries(type: AssetType, id: number, period: PricePeriod): Promise<PriceSeries> {
    const points = series[key(type, id)] ?? [];
    return mockResolve<PriceSeries>({ assetType: type, assetId: id, period, points });
  },
};
