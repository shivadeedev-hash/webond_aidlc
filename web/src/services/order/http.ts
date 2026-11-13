// HTTP order implementation — same interface as the mock, API shapes mapped
// through apiMappers.
import { apiClient } from '../../lib/apiClient';
import type { OrderService } from './index';
import type { GoldQuote, GoldQuoteRequest, Order, PlaceOrderRequest } from '../../types';
import { toOrder, type ApiOrder } from '../apiMappers';

function idempotencyHeader(key?: string): { headers?: Record<string, string> } {
  const value = key ?? crypto.randomUUID();
  return { headers: { 'Idempotency-Key': value } };
}

export const httpOrderService: OrderService = {
  async placeOrder(req: PlaceOrderRequest): Promise<Order> {
    const { idempotencyKey, ...body } = req;
    const { data } = await apiClient.post<ApiOrder>(
      '/orders',
      body,
      idempotencyHeader(idempotencyKey),
    );
    return toOrder(data);
  },
  async requestGoldQuote(req: GoldQuoteRequest): Promise<GoldQuote> {
    // Same field names on both sides: quoteId, price, weight, total, expiresAt.
    const { data } = await apiClient.post<GoldQuote>('/orders/quote', req);
    return data;
  },
  async confirmGoldQuote(quoteId: string, idempotencyKey?: string): Promise<Order> {
    const { data } = await apiClient.post<ApiOrder>(
      `/orders/${quoteId}/confirm`,
      undefined,
      idempotencyHeader(idempotencyKey),
    );
    return toOrder(data);
  },
  async getOrder(txnRef: string): Promise<Order> {
    const { data } = await apiClient.get<ApiOrder>(`/orders/${txnRef}`);
    return toOrder(data);
  },
};
