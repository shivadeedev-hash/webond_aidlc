// Mock order implementation. Builds synthetic orders so the trade UI can be
// exercised without a backend; gold uses a 30s quote lock (US-017, EX-08).

import pricesData from '../../mocks/prices.json';
import { ApiClientError } from '../../lib/errors';
import { mockResolve } from '../mockUtils';
import type { OrderService } from './index';
import type { GoldQuote, GoldQuoteRequest, Order, OrderMode, PlaceOrderRequest } from '../../types';
import type { AssetQuote } from '../../types';

const quotes = pricesData.quotes as Record<string, AssetQuote>;

const GOLD_QUOTE_TTL_MS = 30_000;
let orderSeq = 90000;

interface PendingQuote {
  req: GoldQuoteRequest;
  quote: GoldQuote;
}
const pendingQuotes = new Map<string, PendingQuote>();

function nextTxnRef(): string {
  orderSeq += 1;
  const now = new Date();
  const yy = String(now.getFullYear()).slice(2);
  const mm = String(now.getMonth() + 1).padStart(2, '0');
  const dd = String(now.getDate()).padStart(2, '0');
  return `WB${yy}${mm}${dd}${String(orderSeq).padStart(7, '0')}`;
}

function buildOrder(
  partial: Partial<Order> & Pick<Order, 'assetType' | 'assetId' | 'side' | 'orderMode'>,
): Order {
  const now = new Date().toISOString();
  return {
    id: orderSeq,
    txnRef: nextTxnRef(),
    userId: 1,
    amount: null,
    units: null,
    weightGrams: null,
    price: null,
    fee: '0.0000',
    netAmount: null,
    status: 'completed',
    dcaPlanId: null,
    tradeDate: now.slice(0, 10),
    settlementDate: null,
    realizedPl: null,
    riskAck: false,
    createdAt: now,
    updatedAt: now,
    ...partial,
  };
}

function amountFor(mode: OrderMode, req: PlaceOrderRequest): string | null {
  if (mode === 'amount') return req.amount ?? null;
  return req.amount ?? null;
}

export const mockOrderService: OrderService = {
  async placeOrder(req: PlaceOrderRequest): Promise<Order> {
    if (req.assetType === 'gold') {
      throw new ApiClientError(
        'VALIDATION_ERROR',
        'Gold orders must use the quote flow (requestGoldQuote → confirmGoldQuote).',
      );
    }
    const order = buildOrder({
      assetType: req.assetType,
      assetId: req.assetId,
      side: req.side,
      orderMode: req.orderMode,
      amount: amountFor(req.orderMode, req),
      units: req.units ?? null,
      netAmount: req.amount ?? null,
      riskAck: req.riskAck ?? false,
      status: req.assetType === 'fund' ? 'pending' : 'completed',
    });
    return mockResolve(order);
  },

  async requestGoldQuote(req: GoldQuoteRequest): Promise<GoldQuote> {
    const q = quotes[`gold:${req.assetId}`];
    if (!q) throw new ApiClientError('PRICE_UNAVAILABLE');
    const price =
      req.side === 'buy' ? (q.askPrice ?? q.latestPrice) : (q.bidPrice ?? q.latestPrice);
    const priceNum = Number(price);
    const weight = req.weightGrams ? Number(req.weightGrams) : Number(req.amount ?? 0) / priceNum;
    const total = req.amount ? Number(req.amount) : weight * priceNum;
    const quoteId = `q_${Date.now()}`;
    const quote: GoldQuote = {
      quoteId,
      price: priceNum.toFixed(4),
      weight: weight.toFixed(4),
      total: total.toFixed(4),
      expiresAt: new Date(Date.now() + GOLD_QUOTE_TTL_MS).toISOString(),
    };
    pendingQuotes.set(quoteId, { req, quote });
    return mockResolve(quote);
  },

  async confirmGoldQuote(quoteId: string): Promise<Order> {
    const pending = pendingQuotes.get(quoteId);
    if (!pending) throw new ApiClientError('QUOTE_EXPIRED');
    if (Date.now() > new Date(pending.quote.expiresAt).getTime()) {
      pendingQuotes.delete(quoteId);
      throw new ApiClientError('QUOTE_EXPIRED');
    }
    pendingQuotes.delete(quoteId);
    const order = buildOrder({
      assetType: 'gold',
      assetId: pending.req.assetId,
      side: pending.req.side,
      orderMode: 'weight',
      weightGrams: pending.quote.weight,
      price: pending.quote.price,
      amount: pending.quote.total,
      netAmount: pending.quote.total,
      status: 'completed',
    });
    return mockResolve(order);
  },

  async getOrder(txnRef: string): Promise<Order> {
    // No standalone order store in mock mode beyond this session; surface a 404.
    throw new ApiClientError('TXN_NOT_FOUND', `Order ${txnRef} not found.`, { status: 404 });
  },
};
