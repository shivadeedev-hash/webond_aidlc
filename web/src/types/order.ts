// Order DTOs (data-model.md: orders). Covers buy/sell for all asset types plus
// the gold 30s quote flow (api-spec.md: /orders/quote, /orders/{quoteId}/confirm).

import type { AssetType, Decimal, ISODate, ISODateTime, OrderSide } from './common';

export type OrderMode = 'amount' | 'units' | 'weight';

export type OrderStatus = 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled';

export interface Order {
  id: number;
  /** WB + yyMMdd + 7 running digits (BR-16). */
  txnRef: string;
  userId: number;
  assetType: AssetType;
  assetId: number;
  side: OrderSide;
  orderMode: OrderMode;
  amount: Decimal | null;
  units: Decimal | null;
  weightGrams: Decimal | null;
  /** NAV / clean price / gold price at execution. */
  price: Decimal | null;
  /** All fees 0 in phase 1 (BR-17); column retained. */
  fee: Decimal;
  netAmount: Decimal | null;
  status: OrderStatus;
  dcaPlanId: number | null;
  tradeDate: ISODate | null;
  settlementDate: ISODate | null;
  realizedPl: Decimal | null;
  riskAck: boolean;
  createdAt: ISODateTime;
  updatedAt: ISODateTime;
}

/** Body for POST /api/orders. Exactly one of amount|units|weightGrams per mode. */
export interface PlaceOrderRequest {
  assetType: AssetType;
  assetId: number;
  side: OrderSide;
  orderMode: OrderMode;
  amount?: Decimal;
  units?: Decimal;
  weightGrams?: Decimal;
  /** Acknowledge a risk mismatch when required (EX-12). */
  riskAck?: boolean;
  /** Idempotency-Key header value (EX-11, NFR-05). */
  idempotencyKey?: string;
}

/** Body for POST /api/orders/quote (gold only). */
export interface GoldQuoteRequest {
  assetType: 'gold';
  assetId: number;
  side: OrderSide;
  amount?: Decimal;
  weightGrams?: Decimal;
}

/** 30-second gold price lock (US-017, EX-08). */
export interface GoldQuote {
  quoteId: string;
  price: Decimal;
  weight: Decimal;
  total: Decimal;
  expiresAt: ISODateTime;
}
