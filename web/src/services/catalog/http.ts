// HTTP catalog implementation — same interface as the mock, API shapes mapped
// through apiMappers.
import { apiClient } from '../../lib/apiClient';
import type { AssetQueryParams, CatalogService } from './index';
import type { AssetListItem, AssetType, PricePeriod, PriceSeries } from '../../types';
import type { Paginated } from '../../types/common';
import {
  toAssetDetail,
  toAssetListItem,
  toPriceSeries,
  type ApiAssetDetail,
  type ApiAssetListItem,
  type ApiPricePoint,
} from '../apiMappers';

export const httpCatalogService: CatalogService = {
  async listAssets(params: AssetQueryParams = {}): Promise<Paginated<AssetListItem>> {
    const { data, meta } = await apiClient.get<ApiAssetListItem[]>('/assets', { ...params });
    const items = (data ?? []).map(toAssetListItem);
    return {
      items,
      meta: {
        page: meta?.page ?? params.page ?? 1,
        pageSize: meta?.pageSize ?? params.pageSize ?? items.length,
        total: meta?.total ?? items.length,
        hasMore: meta?.hasMore ?? false,
      },
    };
  },
  async getAsset(type: AssetType, id: number): Promise<AssetListItem> {
    const { data } = await apiClient.get<ApiAssetDetail>(`/assets/${type}/${id}`);
    return toAssetDetail(data);
  },
  async getPriceSeries(type: AssetType, id: number, period: PricePeriod): Promise<PriceSeries> {
    const { data } = await apiClient.get<ApiPricePoint[]>(`/assets/${type}/${id}/prices`, {
      period,
    });
    return toPriceSeries(type, id, period, data ?? []);
  },
};
