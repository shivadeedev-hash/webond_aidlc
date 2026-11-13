// Catalog (asset) service interface + env-selected implementation.
// (api-spec.md GET /api/assets, /api/assets/{type}/{id}, .../prices.)

import { DATA_SOURCE } from '../../lib/constants';
import type {
  Asset,
  AssetListItem,
  AssetType,
  PricePeriod,
  PriceSeries,
  ReturnPeriod,
  SortOrder,
} from '../../types';
import type { Paginated } from '../../types/common';
import { mockCatalogService } from './mock';
import { httpCatalogService } from './http';

export interface AssetQueryParams {
  type?: AssetType;
  q?: string;
  riskLevel?: number;
  category?: string;
  period?: ReturnPeriod;
  sort?: 'performance' | 'name';
  order?: SortOrder;
  page?: number;
  pageSize?: number;
}

export interface CatalogService {
  listAssets(params?: AssetQueryParams): Promise<Paginated<AssetListItem>>;
  getAsset(type: AssetType, id: number): Promise<AssetListItem>;
  getPriceSeries(type: AssetType, id: number, period: PricePeriod): Promise<PriceSeries>;
}

export type { Asset };

export const catalogService: CatalogService =
  DATA_SOURCE === 'api' ? httpCatalogService : mockCatalogService;
