// Order service interface + env-selected implementation (api-spec.md /api/orders).

import { DATA_SOURCE } from '../../lib/constants';
import type { GoldQuote, GoldQuoteRequest, Order, PlaceOrderRequest } from '../../types';
import { mockOrderService } from './mock';
import { httpOrderService } from './http';

export interface OrderService {
  placeOrder(req: PlaceOrderRequest): Promise<Order>;
  /** Gold only — returns a 30s price lock (US-017). */
  requestGoldQuote(req: GoldQuoteRequest): Promise<GoldQuote>;
  /** Confirm a gold quote within 30s; throws QUOTE_EXPIRED if late. */
  confirmGoldQuote(quoteId: string, idempotencyKey?: string): Promise<Order>;
  getOrder(txnRef: string): Promise<Order>;
}

export const orderService: OrderService =
  DATA_SOURCE === 'api' ? httpOrderService : mockOrderService;
